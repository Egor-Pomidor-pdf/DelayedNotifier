package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/config"
	rabbitconsumer "github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/rabbitConsumer"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/repository/receivers"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/repository/senders"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/service"
	"github.com/wb-go/wbf/retry"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg, err := config.NewConfig("../config/.env", "")
	fmt.Println(cfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cfg.Env)
	setupLogger(cfg.Env)

	slog.Info("starting app", slog.String("env", cfg.Env))
	slog.Debug("debug messages are enabled")

	rabbitmqRetryStrategy := retry.Strategy{
		Attempts: cfg.RabbitMQRetry.Attempts,
		Delay:    time.Duration(cfg.RabbitMQRetry.DelayMilliseconds) * time.Millisecond,
		Backoff:  cfg.RabbitMQRetry.Backoff,
	}

	ctx := context.Background()

	consumer, _, err := rabbitconsumer.NewRabbitConsumer(ctx, cfg.RabbitMQ, rabbitmqRetryStrategy)
	if err != nil {
		log.Fatal(err)
	}
	receiver := receivers.NewRabbitMQReceiver(consumer, rabbitmqRetryStrategy)
	sender := senders.NewConsoleSender()
	fmt.Println(cfg.CheckPeriod)
	duration, err := time.ParseDuration(cfg.CheckPeriod)
	if err != nil {
		log.Fatalf("invalid check period: %v", err)
	}
	notificationService := service.NewNotificationService(receiver, sender, duration)
	err = notificationService.Run(ctx, cfg.RabbitMQ)
	if err != nil {
		log.Fatal(err)
	}

}

func setupLogger(env string) {
	var handler slog.Handler
	switch env {
	case envLocal:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envDev:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case envProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	slog.SetDefault(slog.New(handler)) // ← Устанавливаем global logger
}
