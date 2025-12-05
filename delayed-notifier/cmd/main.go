package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/config"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/internaltypes"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	rabbitpublisher "github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/rabbitProducer"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/repository"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/service"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/postgres"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/types"
	"github.com/wb-go/wbf/dbpg"
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
	zlog.Logger.Debug().Any("straegt", rabbitmqRetryStrategy).Msg("str")

	var postgresDB *dbpg.DB
	err = retry.DoContext(ctx, rabbitmqRetryStrategy, func() error {
		var postgresConnErr error

		postgresDB, postgresConnErr = dbpg.New(cfg.Database.MasterDSN, cfg.Database.SlaveDSNs,
			&dbpg.Options{
				MaxOpenConns:    cfg.Database.MaxOpenConnections,
				MaxIdleConns:    cfg.Database.MaxIdleConnections,
				ConnMaxLifetime: time.Duration(cfg.Database.ConnectionMaxLifetimeSeconds) * time.Second,
			})
		return postgresConnErr
	})

	if err != nil {
		zlog.Logger.Fatal().
			Err(err).
			Msg("failed to connect to database")
	}

	zlog.Logger.Info().Msg("Successfully connected to PostgreSQL")

	migrationsPath := "file://./db/migration"
	// migrationsPath := "file:///app/db/migrations" //для докера

	err = postgres.MigrateUp(cfg.Database.MasterDSN, migrationsPath)
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("couldn't migrate postgres on master DSN")
	}

	for i, dsn := range cfg.Database.SlaveDSNs {
		if len(dsn) == 0 {
			continue
		}
		err = postgres.MigrateUp(dsn, migrationsPath)
		if err != nil {
			zlog.Logger.Fatal().Err(err).Int("dsn_index", i).Msg("couldn't migrate postgres on slave DSN")
		}
	}


	StoreRepository := repository.NewRepository(postgresDB, rabbitmqRetryStrategy)
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
	zlog.Logger.Info().Msg("working app...")

	// 1
	chanNotify, err := internaltypes.NotificationChannelFromString("email")
	if err != nil {
		zlog.Logger.Fatal().Msg("pizda")
	}
	Id := types.GenerateUUID()

	err = StoreRepository.CreateNotify(ctx, model.Notification{
		ID: &Id,
		Recipient:   internaltypes.RecipientFromString("as@gmail.com"),
		Channel:     chanNotify,
		Message:     "Message 1",
		ScheduledAt: time.Now(),
	})
	if err != nil {
		zlog.Logger.Fatal().AnErr("err:", err)
	}

	// // 2
	// _ = StoreRepository.CreateNotify(ctx, model.Notification{
	// 	Recipient:   "222user2@example.com",
	// 	Channel:     "telegram",
	// 	Message:     "Message 2",
	// 	ScheduledAt: time.Now(),
	// })

	// // 3
	// _ = StoreRepository.CreateNotify(ctx, model.Notification{
	// 	Recipient:   "333user3@example.com",
	// 	Channel:     "email",
	// 	Message:     "Message 3",
	// 	ScheduledAt: time.Now(),
	// })

	// // 4
	// _ = StoreRepository.CreateNotify(ctx, model.Notification{
	// 	Recipient:   "444user4@example.com",
	// 	Channel:     "telegram",
	// 	Message:     "Message 4",
	// 	ScheduledAt: time.Now(),
	// })

	// // 5
	// _ = StoreRepository.CreateNotify(ctx, model.Notification{
	// 	Recipient:   "555user5@example.com",
	// 	Channel:     "email",
	// 	Message:     "Message 5",
	// 	ScheduledAt: time.Now(),
	// })

	m, err := StoreRepository.FetchFromDb(ctx, time.Now())
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
	uuid2, err := types.NewUUID("7c2b95c4-61c8-4b46-b030-933721931362")
		if err != nil {
			zlog.Logger.Fatal().Err(err).Msg("failed to create UUID")
		}

	ids := []types.UUID{uuid2,
	}
	err = StoreRepository.MarkAsSent(ctx, ids)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Msg("failed to MarkAsSent")
	}

	for _, v := range m {
		fmt.Println(*v)
	}

	// КУСОК ДЛЯ ТЕСТОВ ЗАКАННЧИВАЕТСЯ

	// if err := postgresDB.; err != nil {
	// 	zlog.Logger.Error().
	// 		Err(err).
	// 		Msg("failed to close database")
	// }
}
