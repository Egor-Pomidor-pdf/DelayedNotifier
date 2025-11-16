package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/config"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/db"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	rabbitpublisher "github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/rabbitProducer"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/repository"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/service"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	ctx := context.Background()

	ctx, ctxStop := signal.NotifyContext(ctx, os.Interrupt)
	defer ctxStop()

	cfg, err := config.NewConfig("../config/.env", "")
	if err != nil {
		log.Fatal(err)
	}

	zlog.InitConsole()
	err = zlog.SetLevel(cfg.Env)
	if err != nil {
		log.Fatal(fmt.Errorf("error setting log level to '%s': %w", cfg.Env, err))
	}

	zlog.Logger.Info().
		Str("env", cfg.Env).
		Msg("Start app...")

	rabbitmqRetryStrategy := retry.Strategy{
		Attempts: cfg.RabbitMQRetry.Attempts,
		Delay:    time.Duration(cfg.RabbitMQRetry.DelayMilliseconds) * time.Millisecond,
		Backoff:  cfg.RabbitMQRetry.Backoff,
	}

	db, err := db.NewPostgresDB(cfg.Database)
	if err != nil {
		zlog.Logger.Fatal().
			Err(err).
			Msg("failed to connect to database")
	}

	zlog.Logger.Info().Msg("Successfully connected to PostgreSQL")
	zlog.Logger.Info().Msg("working app...")

	StoreRepository := repository.NewRepository(db)
	publisher, err := rabbitpublisher.NewRabbitProducer(ctx, cfg.RabbitMQ, rabbitmqRetryStrategy)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Msg("failed to GetRabbitProducer")
	}
	defer publisher.Close()

	rabbitRepository := repository.NewRabbitRepository(publisher)
	senderInterface := service.NewSendService(StoreRepository, rabbitRepository)



	// КУСОК ДЛЯ ТЕСТОВ НАЧИНАЕТСЯ

	// 1
	err = StoreRepository.CreateNotify(ctx, model.Notification{
		Recipient:   "111user1@example.com",
		Channel:     "email",
		Message:     "Message 1",
		ScheduledAt: time.Now(),
	})
	if err != nil {
		zlog.Logger.Fatal().Msg("pizda")
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
		zlog.Logger.Fatal().
			Err(err).
			Msg("failed to fetch notifications")
	}

	err = senderInterface.SendBatch(ctx, m)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Msg("failed to sendBatch")
	}

	for _, v := range m {
		fmt.Println(*v)
	}

	// КУСОК ДЛЯ ТЕСТОВ ЗАКАННЧИВАЕТСЯ

	

	if err := db.Close(); err != nil {
		zlog.Logger.Error().
			Err(err).
			Msg("failed to close database")
	}
}
