## Project overview

This repository implements a delayed notification system composed of two Go services:

- `delayed-notifier/`: HTTP API service that accepts notification requests, persists them to PostgreSQL, caches them in Redis, schedules them, and publishes due notifications to RabbitMQ.
- `worker/`: background worker that consumes notifications from RabbitMQ, keeps them in an in-memory min-heap ordered by scheduled time, and dispatches them to channel-specific senders near their scheduled time.

Observability and tooling (Postgres, RabbitMQ, Redis, Prometheus, Loki, Grafana, Nginx) are wired together via `docker/docker-compose.yml` and Kubernetes manifests under `k8s/`.

## Services and modules

### delayed-notifier service (HTTP API + scheduler)

- **Entry point**: `delayed-notifier/cmd/main.go`.
  - Initializes config via `internal/config` (environment-driven, using `github.com/wb-go/wbf/config`).
  - Sets up logging with `github.com/wb-go/wbf/zlog` based on `ENV`.
  - Connects to PostgreSQL via `github.com/wb-go/wbf/dbpg` with retry strategies from config, then runs migrations from `delayed-notifier/db/migration` using `pkg/postgres.MigrateUp`.
  - Initializes repositories:
    - `internal/repository.StoreRepository` (Postgres; `CRUDStoreRepositoryInterface` + fetch/mark-as-sent APIs).
    - `internal/rabbitProducer.Publisher` wrapped by `internal/repository.RabbitRepository` for publishing messages to RabbitMQ.
    - `internal/repository.RedisRepository` for notification caching in Redis (JSON-encoded models, key = notification UUID).
  - Initializes services:
    - `internal/service.SendService` runs a background loop that periodically fetches due notifications from Postgres and sends them to RabbitMQ (with DLQ + second-chance resend logic), marking them as `sent` in the DB.
    - `internal/service.CRUDService` implements the core CRUD API:
      - On create: generates UUID, writes to Postgres, best-effort write-through to Redis, and optionally invokes `funcOnCreate` hook asynchronously.
      - On read: attempts Redis first, falls back to Postgres, then repopulates cache.
      - On delete: deletes from Postgres and Redis concurrently via `errgroup`.
  - Starts the background `SendService.Run` in a goroutine and then starts the HTTP server.

- **HTTP layer**:
  - `internal/handler/router.go` builds a `github.com/wb-go/wbf/ginext` engine with:
    - `MetricsMiddleware` (project-defined), logging, recovery.
    - Serves `/` from `/app/internal/static/index.html`.
    - REST endpoints:
      - `POST /notify` → `NotifyHandler.CreateNotification`.
      - `GET /notify` → `NotifyHandler.GetAllNotifications`.
      - `GET /notify/:id` → `NotifyHandler.GetNotification`.
      - `DELETE /notify/:id` → `NotifyHandler.DeleteNotification`.
      - `GET /metrics` → Prometheus metrics via `promhttp.Handler()`.
  - `internal/handler/notfications_handler.go` performs:
    - Request parsing and validation through `internal/dto` into `internal/model.Notification`.
    - Delegation to `CRUDService` via `internal/ports` interfaces.
    - Uniform error → HTTP status + JSON mapping.

- **HTTP server & graceful shutdown**:
  - `pkg/server/server.go` defines `HTTPServer.GracefulRun(ctx, host, port)`:
    - Wraps the router in `net/http.Server`.
    - Starts `ListenAndServe` and a signal listener (using `signal.NotifyContext`) that calls `Shutdown` with a timeout on interrupt.
    - Ensures the listener goroutine exits cleanly before returning.

- **Persistence & scheduling**:
  - `internal/repository/postgress_repository.go` (note the spelling) is the main Postgres repository and is central to understanding storage:
    - Uses the schema `notifier_db.public.notifications` with fields like `id`, `recipient`, `channel`, `message`, `scheduled_at`, `status`, `tries`, `last_error`, timestamps.
    - Key methods:
      - `CreateNotify`: inserts a new notification row.
      - `GetNotify` / `GetAllNotifies`: read single/all notifications, mapping strings into domain types via `internal/internaltypes` and `pkg/types`.
      - `FetchFromDb(needToSendTime)`: selects `pending` notifications with `scheduled_at <= needToSendTime` and `tries <= 3`, ordered by `scheduled_at`. This is the main source of work for `SendService`.
      - `UpdateNotification`: updates all fields by ID with `updated_at = now()`.
      - `MarkAsSent`: bulk-updates `status = 'sent'` for a dynamic set of IDs using `IN ($1, $2, ...)`.
      - `DeleteNotification`: deletes by ID and reports "not found" if no rows were affected.
  - `internal/service/send_service.go` orchestrates fetching and publishing:
    - Maintains `fetchPeriod` and `fetchMaxDiapason` (upper bound horizon for fetching).
    - `Run` sets up a ticker; each tick calls `lifeCycle`.
    - `lifeCycle` computes `dateTimeForSent = now + fetchPeriod`, fetches a batch from Postgres via `FetchFromDb`, and forwards to `SendBatch`.
    - `SendBatch`:
      - Sends many notifications to RabbitMQ via `PublisherRepository.SendMany`, which returns a `pkg/dlq.DLQ[*model.Notification]`.
      - Fire-and-forgets a goroutine that marks the whole batch as sent in Postgres.
      - For items that land in the DLQ, attempts `QuickSend` (one-by-one resend); logs failures and successes.

- **Redis caching**:
  - `internal/repository/reddis_repository.go` (note spelling) is a light wrapper around `github.com/wb-go/wbf/redis`:
    - Keys are notification UUIDs.
    - Values are JSON-encoded `internal/model.Notification`.
    - Expiration duration is configured and passed in at construction.

### worker service (RabbitMQ consumer + in-memory scheduler)

- **Entry point**: `worker/cmd/main.go`.
  - Loads config from `worker/config` (same `wbf/config` pattern) and initializes `zlog`.
  - Builds retry strategies for:
    - RabbitMQ consumer (`ConsumerRetry`).
    - Receiver wrapper (`ReceiverRetry`).
  - Creates a RabbitMQ consumer via `internal/rabbitConsumer.NewRabbitConsumer`, which returns a consumer and channel of `*internal/model.Notification`.
  - Wraps the consumer in a `NotificationReceiver` (`internal/repository/receivers.NewRabbitMQReceiver`).
  - Creates a `NotificationSender` (`internal/repository/senders.NewConsoleSender()` for now; designed to be pluggable per channel).
  - Parses `CheckPeriod` (e.g. `"1s"`, `"100ms"`) into a `time.Duration`.
  - Constructs and runs `internal/service.NotificationService` with the receiver, sender, and check period.

- **In-memory scheduling**:
  - `internal/notificationHeap/notification_heap.go` defines `NotificationHeap` as a heap of `*model.Notification`:
    - Implements `heap.Interface` over the slice.
    - `Less` compares `ScheduledAt` parsed as RFC3339 timestamps to order items (currently using `After`, so the top of the heap is the latest time; combined with how `Peek`/`Pop` are used, this effectively processes earlier scheduled items first).
    - `Push`, `Pop`, and `Peek` are standard heap helpers.
  - `internal/service/notification_service.go` manages the lifecycle:
    - Holds:
      - `NotificationReceiver` (pulls notifications from RabbitMQ).
      - `NotificationSender` (currently a single sender; TODO suggests mapping by channel).
      - `checkPeriod` (tick interval for scanning the heap).
      - A `NotificationHeap` protected by an `RWMutex`.
    - `Run(ctx, rabbitCfg)`:
      - Calls `receiver.StartReceiving(ctx)` to obtain a channel of incoming notifications.
      - Logs queue name and starts `serveHeap` in a separate goroutine.
      - Main loop:
        - On `ctx.Done()`, exits loop and calls `receiver.StopReceiving()`.
        - On incoming notification, locks the heap and pushes it.
    - `serveHeap(ctx)`:
      - Ticks every `checkPeriod`.
      - On each tick:
        - Locks the heap and repeatedly `Peek`s the earliest scheduled notification.
        - If `ScheduledAt + checkPeriod` is still after `now`, breaks (nothing is ready).
        - Otherwise, pops the notification, unlocks the heap, and calls `sendNotification`.
        - After sending, re-locks the heap and continues.
      - `sendNotification` delegates to `NotificationSender.Send`; logs detailed error context (ID, channel, scheduled time) on failure.

- **Shared types & module relationship**:
  - `worker/go.mod` depends on `github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier` for shared DTO/model/types.
  - Changes to shared entities in `delayed-notifier` that are used by `worker` may require updating the version in `worker/go.mod` (or adding a local `replace` when working locally).

## Configuration

Configuration for both services is environment-driven using `github.com/wb-go/wbf/config` with explicit keys (no YAML config is required by default).

### delayed-notifier config (subset of important keys)

All keys are read in `internal/config/config.go`:

- General:
  - `ENV` — log level / environment name (e.g. `dev`, `prod`).
- RabbitMQ (producer side):
  - `DELAYED_NOTIFIER_RABBITMQ_USER`
  - `DELAYED_NOTIFIER_RABBITMQ_PASSWORD`
  - `DELAYED_NOTIFIER_RABBITMQ_HOST`
  - `DELAYED_NOTIFIER_RABBITMQ_PORT`
  - `DELAYED_NOTIFIER_RABBITMQ_VHOST`
  - `DELAYED_NOTIFIER_RABBITMQ_EXCHANGE`
  - `DELAYED_NOTIFIER_RABBITMQ_QUEUE`
- PostgreSQL:
  - `DELAYED_NOTIFIER_POSTGRES_MASTER_DSN`
  - `DELAYED_NOTIFIER_POSTGRES_SLAVE_DSNS` (string slice)
  - `DELAYED_NOTIFIER_POSTGRES_MAX_OPEN_CONNECTIONS`
  - `DELAYED_NOTIFIER_POSTGRES_MAX_IDLE_CONNECTIONS`
  - `DELAYED_NOTIFIER_POSTGRES_CONNECTION_MAX_LIFETIME_SECONDS`
- Redis:
  - `DELAYED_NOTIFIER_REDIS_HOST`
  - `DELAYED_NOTIFIER_REDIS_PORT`
  - `DELAYED_NOTIFIER_REDIS_PASSWORD`
  - `DELAYED_NOTIFIER_REDIS_DB`
  - `DELAYED_NOTIFIER_REDIS_EXPIRATION` (seconds; converted to `time.Duration`).
- HTTP server:
  - `DELAYED_NOTIFIER_SERVER_HOST`
  - `DELAYED_NOTIFIER_SERVER_PORT`
- Retry strategies:
  - `DELAYED_NOTIFIER_RETRY_RABBITMQ_*` (publisher connection/retries).
  - `DELAYED_NOTIFIER_RETRY_POSTGRES_*`.
  - `DELAYED_NOTIFIER_RETRY_STORE_REPO_*`.
  - `DELAYED_NOTIFIER_RETRY_RABBIT_REPO_*`.
  - `DELAYED_NOTIFIER_RETRY_REDIS_REPO_*`.

### worker config (subset of important keys)

Read in `worker/config/config.go`:

- General:
  - `ENV` — same semantics as in `delayed-notifier`.
- RabbitMQ (consumer side):
  - Uses the same RabbitMQ keys as the producer (`DELAYED_NOTIFIER_RABBITMQ_*`).
- Scheduling:
  - `CHECK_PERIOD` — string duration for heap scan interval (e.g. `"1s"`, `"500ms"`).
- Retry strategies:
  - `DELAYED_NOTIFIER_RETRY_CONSUMER_*`.
  - `DELAYED_NOTIFIER_RETRY_RECEIVER_*`.

For local Docker-based development, these env vars are expected to be provided via `config/.env` referenced in `docker/docker-compose.yml`.

## Running and building locally

### Using Go directly (no Docker)

From repository root:

- **Run delayed-notifier HTTP API**:
  - `cd delayed-notifier`
  - `go run ./cmd`

- **Run worker**:
  - `cd worker`
  - `go run ./cmd`

Both binaries rely on environment variables described in the Configuration section.

- **Build binaries**:
  - `cd delayed-notifier && go build ./...`
  - `cd worker && go build ./...`

These commands build all packages within each module (including libraries and the `cmd` entrypoint).

### Using Docker Compose (full stack, recommended for local integration)

From repository root:

- `cd docker`
- `docker compose up`  (or `docker-compose up` depending on your Docker version)

This will start:

- `delayed_notifier` service built from `../delayed-notifier`.
- `worker` service built from `../worker`.
- `postgres_master` with health checks.
- `rabbitmq` with management UI.
- `redis`.
- `promtail`, `loki`, `prometheus`, `grafana` for logging/metrics.
- `nginx` as a front-end proxy.

`delayed_notifier` is exposed on host port `8089` by default (see `docker/docker-compose.yml`). Nginx is exposed on port `80`.

### Kubernetes deployment

Kubernetes manifests live under `k8s/`.

- `k8s/03-delayed-notifier.yaml` — Deployment + Service for the HTTP API.
- `k8s/03-worker.yaml` — Deployment for the worker.

The GitLab CI `deploy_dev` job applies `k8s/` and updates the image of the `delayed-notifier` deployment. To perform a similar manual deploy:

- `kubectl apply -f k8s/`
- Then, if needed, manually `kubectl set image deployment/delayed-notifier delayed-notifier=<image>` in the chosen namespace.

## Testing and CI

### Running tests locally

The GitLab CI `test` stage runs:

- From the repository root: `go test ./...`

This command will run all Go tests across modules (as supported by the Go toolchain for the current layout). Use the same command locally to mirror CI.

To run tests for a single package or a single test:

- `cd delayed-notifier` (or `cd worker`)
  - Run all tests in the module: `go test ./...`
  - Run a specific test in a package: `go test ./path/to/package -run ^TestName$`

### GitLab CI pipeline

Defined in `.gitlab-ci.yml`:

- **test**: runs `go test ./...` in a `golang:1.24` container.
- **build_and_push**: builds a Docker image with `docker build -t "$IMAGE_TAG" .` and pushes it to GitLab Container Registry (note that the current Dockerfiles live in `delayed-notifier/` and `worker/`; adjust CI if necessary).
- **deploy_dev**: uses `kubectl` and GitLab Agent (`.gitlab/agents/minikube/config.yaml`) to:
  - `kubectl apply -f k8s/` in the target namespace.
  - Update the `delayed-notifier` deployment image and wait for rollout.

## Notes on the README

`README.md` currently describes a milestone-based implementation plan and an earlier, more idealized directory structure (e.g. `internal/api`, `internal/worker`, `web/`, etc.). The actual implementation diverges from that plan:

- The real layout uses `internal/handler`, `internal/service`, `internal/repository`, `internal/rabbitProducer`, `internal/dto`, `internal/model`, `internal/internaltypes`, and `pkg/*`.
- When modifying or extending the system, prefer the actual code structure over the illustrative tree in the README.
