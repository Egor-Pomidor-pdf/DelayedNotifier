FROM golang:1.24.6-alpine AS build
WORKDIR /app

# Кэш зависимостей: сначала go.mod/go.sum
COPY go.mod go.sum ./
RUN go mod download

# Потом исходники
COPY . .

# Собираем статический бинарь
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/delayed-notifier ./cmd

# ---- runtime stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /bin/delayed-notifier /bin/delayed-notifier

COPY --from=build /app/db/migration /app/db/migration

COPY internal/static/ /app/internal/static/


# если твой сервер слушает 8089
EXPOSE 8089

ENTRYPOINT ["/bin/delayed-notifier"]


