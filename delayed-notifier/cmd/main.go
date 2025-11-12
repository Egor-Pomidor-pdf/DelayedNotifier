package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/config"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/db"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	rabbitpublisher "github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/rabbitProducer"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/repository"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/service"
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

	db, err := db.NewPostgresDB(cfg.Database)

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	slog.Info("Successfully connected to PostgreSQL")
	slog.Info("working app...")

	ctx := context.Background()
	StoreRepository := repository.NewRepository(db)

	// 1
	err = StoreRepository.CreateNotify(ctx, model.Notification{
		Recipient:   "111user1@example.com",
		Channel:     "email",
		Message:     "Message 1",
		ScheduledAt: time.Now(),
	})

	if err != nil {
		log.Fatalf("pizda")
	}

	// 2
	_ = StoreRepository.CreateNotify(ctx, model.Notification{
		Recipient:   "222user2@example.com",
		Channel:     "telegram",
		Message:     "Message 2",
		ScheduledAt: time.Now(),
	})

	// 3
	_ = StoreRepository.CreateNotify(ctx, model.Notification{
		Recipient:   "333user3@example.com",
		Channel:     "email",
		Message:     "Message 3",
		ScheduledAt: time.Now(),
	})

	// 4
	_ = StoreRepository.CreateNotify(ctx, model.Notification{
		Recipient:   "444user4@example.com",
		Channel:     "telegram",
		Message:     "Message 4",
		ScheduledAt: time.Now(),
	})

	// 5
	_ = StoreRepository.CreateNotify(ctx, model.Notification{
		Recipient:   "555user5@example.com",
		Channel:     "email",
		Message:     "Message 5",
		ScheduledAt: time.Now(),
	})
	m, err := StoreRepository.FetchFromDb(ctx, time.Now().Add(3*time.Hour))
	if err != nil {
		log.Fatalf("failed to fetch notifications: %v", err)
	}
	publisher, err := rabbitpublisher.NewRabbitProducer(context.Background(), cfg.RabbitMQ, rabbitmqRetryStrategy)
	if err != nil {
		slog.Error("failed to GetRabbitProducer", slog.String("error", err.Error()))
	}

	defer publisher.Close()

	rabbitRepository := repository.NewRabbitRepository(publisher)
	senderInterface := service.NewSendService(StoreRepository, rabbitRepository)
	err = senderInterface.SendBatch(context.Background(), m)
	if err != nil {
		slog.Error("failed to sendBatch", slog.String("error", err.Error()))
	}
	for _, v := range m {
		fmt.Println(*v)
	}

	if err := db.Close(); err != nil {
		slog.Error("failed to close database", slog.String("error", err.Error()))
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
