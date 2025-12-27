FROM golang:1.24.6-alpine AS build
WORKDIR /app

# Кэш зависимостей: сначала go.mod/go.sum
COPY go.mod go.sum ./
RUN go mod download

# Потом исходники
COPY . .

# Собираем статический бинарь
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/worker ./cmd

# ---- runtime stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /bin/worker /bin/worker


ENTRYPOINT ["/bin/worker"]


