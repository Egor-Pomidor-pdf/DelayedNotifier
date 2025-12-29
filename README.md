# Delayed Notifier

Сервис отложенных уведомлений, состоящий из двух Go‑сервисов:

- **`delayed-notifier`** — HTTP API:
  - принимает запросы на создание уведомлений;
  - сохраняет их в PostgreSQL;
  - кэширует в Redis;
  - планирует отправку и публикует сообщения в RabbitMQ.
- **`worker`** — воркер:
  - читает сообщения из RabbitMQ;
  - хранит их во внутренней очереди (min‑heap по времени отправки);
  - отправляет уведомления в нужный момент через отправители (senders).

Инфраструктура для локального запуска и мониторинга описана в `docker/docker-compose.yml` и манифестах Kubernetes в каталоге `k8s/`.

---

## Основные сценарии использования

- Клиент отправляет `POST /notify` с данными уведомления и временем, когда его нужно отправить.
- Сервис `delayed-notifier`:
  - валидирует запрос;
  - записывает уведомление в таблицу `notifier_db.public.notifications`;
  - кладет объект в Redis‑кэш для быстрого чтения.
- Фоновый планировщик (`SendService`) периодически:
  - выбирает из PostgreSQL уведомления со статусом `pending`, у которых `scheduled_at` уже наступило;
  - публикует их в RabbitMQ (по ключу маршрутизации, зависящему от канала — email/telegram и т.п.);
  - помечает успешно отправленные уведомления как `sent`.
- Воркер `worker`:
  - читает из очереди RabbitMQ объекты уведомлений;
  - складывает их в кучу (min‑heap) по времени `scheduled_at`;
  - с заданной периодичностью проверяет верхушку кучи и отправляет уведомления, чье время уже наступило.

---

## Архитектура и компоненты

### 1. Сервис `delayed-notifier` (HTTP API + планировщик)

**Точка входа:** `delayed-notifier/cmd/main.go`

#### Основные шаги запуска

1. **Загрузка конфигурации** — пакет `internal/config`:
   - использует `github.com/wb-go/wbf/config`;
   - читает переменные окружения (префиксы `DELAYED_NOTIFIER_*`, `ENV` и др.);
   - формирует структуры:
     - `PostgresConfig`
     - `RedisConfig`
     - `RabbitMQConfig`
     - `ServerConfig`
     - `RetryConfig` для разных подсистем (Postgres, RabbitMQ, репозитории).
2. **Логирование** — `github.com/wb-go/wbf/zlog`:
   - `zlog.InitConsole()` — вывод в консоль;
   - уровень логирования определяется переменной `ENV`.
3. **PostgreSQL + миграции**:
   - подключение через `github.com/wb-go/wbf/dbpg` с использованием retry‑стратегий;
   - миграции из `delayed-notifier/db/migration` через `pkg/postgres.MigrateUp`.
4. **Репозитории:**
   - `internal/repository.StoreRepository` (`postgress_repository.go`):
     - сохраняет, читает, обновляет и удаляет уведомления в таблице `notifier_db.public.notifications`;
     - реализует выборку батча уведомлений к отправке (`FetchFromDb`), `MarkAsSent`, `GetAllNotifies` и др.
   - `internal/rabbitProducer.Publisher` + `internal/repository.RabbitRepository`:
     - обертка над RabbitMQ;
     - методы `SendOne` и `SendMany` отправляют сообщения в обменник с routing key, зависящим от канала (`notification.Channel.String()`).
   - `internal/repository.RedisRepository` (`reddis_repository.go`):
     - кэширует объекты `model.Notification` в Redis (ключ — UUID, значение — JSON);
     - поддерживает запись/чтение/удаление.
5. **Сервисы:**
   - `internal/service.CRUDService` (`crud_service.go`):
     - интерфейс CRUD для уведомлений:
       - **CreateNotification**:
         - генерирует UUID, присваивает его модели;
         - сохраняет запись в Postgres;
         - асинхронно кладет ее в Redis;
         - опциональный callback `funcOnCreate` может, например, триггерить дополнительные действия.
       - **GetNotification**:
         - сначала пробует Redis, при промахе — Postgres;
         - при успешном чтении из Postgres — записывает в Redis.
       - **GetAllNotifications**:
         - читает список только из Postgres.
       - **DeleteNotification**:
         - удаляет запись в Postgres и Redis параллельно (`errgroup`).
   - `internal/service.SendService` (`send_service.go`):
     - фоновый планировщик:
       - по таймеру (`fetchPeriod`) выбирает уведомления из Postgres с `scheduled_at <= now + fetchPeriod`;
       - отправляет их пачкой через RabbitMQ (`SendBatch`);
       - использует очередь DLQ (`pkg/dlq`) для повторных попыток отправки отдельных сообщений.
6. **HTTP‑слой:**
   - `internal/handler/router.go`:
     - использует `github.com/wb-go/wbf/ginext` (обертка над Gin);
     - мидлвары: логирование, panic‑recovery, метрики;
     - роуты:
       - `POST /notify`
       - `GET /notify`
       - `GET /notify/:id`
       - `DELETE /notify/:id`
       - `GET /metrics` — отдаёт метрики Prometheus;
       - `/` — отдает статический файл `internal/static/index.html`.
   - `internal/handler/notfications_handler.go`:
     - `CreateNotification`:
       - принимает JSON тела запроса в `dto.NotificationCreate`;
       - валидирует канал, получателя и дату (`time.RFC3339`);
       - преобразует в `model.Notification` и передает в `CRUDService`.
     - `GetNotification`, `GetAllNotifications`, `DeleteNotification`:
       - работают через интерфейсы из `internal/ports` и DTO (`notification_get.go`, `notification_full.go`).
7. **HTTP‑сервер и graceful shutdown:**
   - `pkg/server/server.go`:
     - создает `http.Server` с переданным роутером;
     - слушает системный сигнал (Ctrl+C и т.п.), делает `Shutdown` с таймаутом;
     - корректно завершает сервер и фоновые горутины.

### 2. Сервис `worker` (потребитель RabbitMQ + in‑memory планировщик)

**Точка входа:** `worker/cmd/main.go`

#### Основные шаги

1. **Загрузка конфигурации** — `worker/config/config.go`:
   - читает `ENV`, параметры RabbitMQ и retry‑настроек;
   - параметр `CHECK_PERIOD` задает период проверки кучи уведомлений (`time.ParseDuration`).
2. **Логирование** — также через `zlog`.
3. **RabbitMQ consumer:**
   - создается через `internal/rabbitConsumer.NewRabbitConsumer`;
   - возвращает объект consumer и канал `chan *model.Notification`.
4. **Приемник и отправитель:**
   - `internal/repository/receivers.NewRabbitMQReceiver` — адаптер над consumer’ом;
   - `internal/repository/senders.NewConsoleSender()` — пример отправителя (пишет в консоль; архитектура подразумевает отправителей по каналам уведомления).
5. **Сервис уведомлений** — `internal/service.NotificationService`:
   - содержит:
     - `NotificationReceiver` — источник уведомлений (из RabbitMQ);
     - `NotificationSender` — отправитель (пока один на все каналы);
     - `checkPeriod` — период цикла;
     - `notificationHeap` (`internal/notificationHeap.NotificationHeap`) — кучу уведомлений;
     - `heapMutex` — `sync.RWMutex` для защиты кучи.
   - метод `Run(ctx, rabbitCfg)`:
     - запускает `receiver.StartReceiving(ctx)` и получает канал уведомлений;
     - параллельно запускает `serveHeap(ctx)`;
     - в основном цикле:
       - по `ctx.Done()` — аккуратно останавливается и вызывает `receiver.StopReceiving()`;
       - по новому уведомлению — кладет его в кучу.
   - метод `serveHeap(ctx)`:
     - по таймеру (`checkPeriod`) просматривает верхушку кучи;
     - если уведомление пора отправлять (`scheduled_at` уже наступило), извлекает его и вызывает `sendNotification`;
     - если «еще рано» — завершает цикл обработки до следующего тика.
   - метод `sendNotification`:
     - делегирует отправку в `NotificationSender.Send`;
     - логирует ошибки с ID, каналом и временем.
6. **Куча уведомлений** — `internal/notificationHeap/notification_heap.go`:
   - реализует `heap.Interface` для `[]*model.Notification`;
   - `Less` сравнивает `ScheduledAt` как время; реализует приоритет по времени отправки;
   - `Peek` возвращает следующий элемент без удаления.

---

## Хранение данных

Основная таблица в PostgreSQL: `notifier_db.public.notifications`.

Согласно запросам в репозитории, в таблице используются как минимум поля:

- `id` — UUID уведомления;
- `recipient` — получатель (email, telegram id и т.п.);
- `channel` — канал доставки (например, `"email"`, `"telegram"`);
- `message` — текст уведомления;
- `scheduled_at` — дата/время запланированной отправки (`timestamp`);
- `status` — строковый статус (`pending`, `sent` и т.п.);
- `tries` — число попыток отправки;
- `last_error` — текст последней ошибки (nullable).

Полная схема задается SQL‑миграцией `delayed-notifier/db/migration/001_init.up.sql`.

---

## Конфигурация

### Общие переменные

- `ENV` — окружение и уровень логирования (`dev`, `prod` и др.).

### delayed-notifier

Читаются в `internal/config/config.go` (префикс `DELAYED_NOTIFIER_`):

**RabbitMQ:**

- `DELAYED_NOTIFIER_RABBITMQ_USER`
- `DELAYED_NOTIFIER_RABBITMQ_PASSWORD`
- `DELAYED_NOTIFIER_RABBITMQ_HOST`
- `DELAYED_NOTIFIER_RABBITMQ_PORT`
- `DELAYED_NOTIFIER_RABBITMQ_VHOST`
- `DELAYED_NOTIFIER_RABBITMQ_EXCHANGE`
- `DELAYED_NOTIFIER_RABBITMQ_QUEUE`

**Postgres:**

- `DELAYED_NOTIFIER_POSTGRES_MASTER_DSN`
- `DELAYED_NOTIFIER_POSTGRES_SLAVE_DSNS` (список строк)
- `DELAYED_NOTIFIER_POSTGRES_MAX_OPEN_CONNECTIONS`
- `DELAYED_NOTIFIER_POSTGRES_MAX_IDLE_CONNECTIONS`
- `DELAYED_NOTIFIER_POSTGRES_CONNECTION_MAX_LIFETIME_SECONDS`

**Redis:**

- `DELAYED_NOTIFIER_REDIS_HOST`
- `DELAYED_NOTIFIER_REDIS_PORT`
- `DELAYED_NOTIFIER_REDIS_PASSWORD`
- `DELAYED_NOTIFIER_REDIS_DB`
- `DELAYED_NOTIFIER_REDIS_EXPIRATION` — время жизни записи в кэше (в миллисекундах/секундах, далее приводится к `time.Duration`).

**HTTP‑сервер:**

- `DELAYED_NOTIFIER_SERVER_HOST`
- `DELAYED_NOTIFIER_SERVER_PORT`

**Retry‑настройки:**

- `DELAYED_NOTIFIER_RETRY_RABBITMQ_*`
- `DELAYED_NOTIFIER_RETRY_POSTGRES_*`
- `DELAYED_NOTIFIER_RETRY_STORE_REPO_*`
- `DELAYED_NOTIFIER_RETRY_RABBIT_REPO_*`
- `DELAYED_NOTIFIER_RETRY_REDIS_REPO_*`

### worker

Читаются в `worker/config/config.go`:

- `ENV`
- те же RabbitMQ‑переменные `DELAYED_NOTIFIER_RABBITMQ_*`
- `CHECK_PERIOD` — период проверки кучи (например, `"1s"`, `"200ms"`)
- retry‑настройки:
  - `DELAYED_NOTIFIER_RETRY_CONSUMER_*`
  - `DELAYED_NOTIFIER_RETRY_RECEIVER_*`

Для локального запуска через Docker все эти переменные задаются в `config/.env`, который подключается в `docker/docker-compose.yml`.

---

## Запуск и сборка

### 1. Прямой запуск Go‑бинарников

#### delayed-notifier

```bash
cd delayed-notifier
go run ./cmd
```

#### worker

```bash
cd worker
go run ./cmd
```

Перед запуском убедитесь, что выставлены все необходимые переменные окружения (Postgres, Redis, RabbitMQ, `ENV` и т.д.).

Для сборки бинарников:

```bash
cd delayed-notifier && go build ./...
cd ../worker && go build ./...
```

### 2. Запуск через Docker Compose (рекомендуется для локального окружения)

```bash
cd docker
docker compose up
```

Поднимаются:

- `delayed_notifier` (HTTP API) — образ собирается из `../delayed-notifier/Dockerfile`;
- `worker` — из `../worker/Dockerfile`;
- `postgres_master`;
- `rabbitmq` с management‑панелью;
- `redis`;
- стек логов и метрик: `promtail`, `loki`, `prometheus`, `grafana`;
- `nginx` — фронтовой прокси.

Порты по умолчанию (см. `docker/docker-compose.yml`):

- HTTP API `delayed_notifier` — 8089;
- Nginx — 80;
- RabbitMQ management — 15672, AMQP — 5672;
- Postgres — 5435;
- Redis — 6379;
- Prometheus — 9090;
- Grafana — 3000;
- Loki, Promtail — 3100/9080.

### 3. Kubernetes

Основные манифесты:

- `k8s/03-delayed-notifier.yaml` — Deployment + Service для `delayed-notifier`:
  - Deployment (`delayed-notifier-deployment`) в namespace `delayed-notifier`;
  - Service с типом `ClusterIP`, порт 80 → `targetPort` 8089.
- `k8s/03-worker.yaml` — Deployment для `worker`.

Для ручного деплоя:

```bash
kubectl apply -f k8s/
```

Обновление образа `delayed-notifier`:

```bash
kubectl -n delayed-notifier set image deployment/delayed-notifier-deployment delayed-notifier=<IMAGE>
kubectl -n delayed-notifier rollout status deployment/delayed-notifier-deployment --timeout=120s
```

GitLab CI использует похожую последовательность в job’е `deploy_dev` (`.gitlab-ci.yml`).

---

## HTTP API

### Базовый URL

- при локальном запуске без Nginx: `http://localhost:8089`
- при использовании Nginx из docker‑compose: `http://localhost/`

### 1. Создание уведомления

`POST /notify`

**Тело запроса (`NotificationCreate`):**

```json
{
  "recipient": "user@example.com",
  "channel": "email",
  "message": "Текст уведомления",
  "scheduled_at": "2025-01-01T12:00:00Z"
}
```

Поля:

- `recipient` — куда отправляем (email, telegram id и т.п.);
- `channel` — строка канала (`email`, `telegram`, и др., интерпретация в `internal/internaltypes`);
- `message` — текст;
- `scheduled_at` — время отправки в формате `RFC3339` (ISO 8601).

**Ответ (успех, 201):**

Тело — объект `NotificationFull`, включая:

- `id`
- `recipient`
- `channel`
- `message`
- `scheduled_at`
- `status`

### 2. Получение одного уведомления

`GET /notify/:id`

- `id` — UUID уведомления.

**Ответ (200):** объект `NotificationFull`.

Если формат UUID неверен или объект не найден — возвращаются коды 400/500 с JSON‑описанием ошибки.

### 3. Получение всех уведомлений

`GET /notify`

**Ответ (200):** список объектов `NotificationFull`.

### 4. Удаление уведомления

`DELETE /notify/:id`

- при успехе — статус `204 No Content`;
- при ошибках — JSON с описанием.

### 5. Метрики

`GET /metrics`

- отдает метрики Prometheus (через `promhttp.Handler()`).

---

## Тесты и CI

В `.gitlab-ci.yml` определены стадии:

- **test**:
  - запускается в образе `golang:1.24`;
  - выполняет:

    ```bash
    go test ./...
    ```

- **build_and_push**:
  - собирает Docker‑образ и пушит его в GitLab Container Registry (`IMAGE_TAG` основан на `CI_COMMIT_SHA`).
- **deploy_dev**:
  - использует `kubectl` и GitLab Agent (`.gitlab/agents/minikube/config.yaml`);
  - применяет манифесты `k8s/`;
  - обновляет образ деплоймента `delayed-notifier` и ждет успешного rollout’а.

Локально для запуска всех тестов достаточно:

```bash
go test ./...
```

из корня репозитория или из отдельного модуля (`delayed-notifier`, `worker`).

---

## Замечания по исходному README

Текущий `README.md` в репозитории — это в основном план реализации и пример желаемой структуры директорий. Фактическая структура проекта немного иная:

- используются пакеты:
  - `internal/handler`
  - `internal/service`
  - `internal/repository`
  - `internal/dto`
  - `internal/model`
  - `internal/internaltypes`
  - `internal/rabbitProducer`
  - `pkg/server`, `pkg/postgres`, `pkg/types`, `pkg/dlq`
- при доработках и рефакторинге стоит ориентироваться именно на эту фактическую структуру.
=======