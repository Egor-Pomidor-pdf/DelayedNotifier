package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/config"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/handler"
	rabbitpublisher "github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/rabbitProducer"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/repository"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/service"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/postgres"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/server"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

func main() {

	// make context
	ctx := context.Background()
	ctx, ctxStop := signal.NotifyContext(ctx, os.Interrupt)

	// init config
	cfg, err := config.NewConfig("../config/.env", "")
	if err != nil {
		log.Fatal(err)
	}

	// init logger
	zlog.InitConsole()
	err = zlog.SetLevel(cfg.Env)
	if err != nil {
		log.Fatal(fmt.Errorf("error setting log level to '%s': %w", cfg.Env, err))
	}
	zlog.Logger.Info().
		Str("env", cfg.Env).
		Msg("Start app...")

	// stratefies
	rabbitmqRetryStrategy := config.MakeStrategy(cfg.RabbitMQRetry)
	postgresRetryStrategy := config.MakeStrategy(cfg.PostgresRetry)
	storeRepoRetryStrategy := config.MakeStrategy(cfg.StoreRepoRetry)
	rabbitRepoRetryStrategy := config.MakeStrategy(cfg.RabbitRepoRetry)
	redisRepoRetryStrategy := config.MakeStrategy(cfg.RedisRepoRetry)

	// connect to db
	var postgresDB *dbpg.DB
	err = retry.DoContext(ctx, postgresRetryStrategy, func() error {
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

	// create migrations
	migrationsPath := "file://./db/migration"
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

	// init storeRepository and publisher
	StoreRepository := repository.NewRepository(postgresDB, storeRepoRetryStrategy)
	publisher, err := rabbitpublisher.NewRabbitProducer(ctx, cfg.RabbitMQ, rabbitmqRetryStrategy)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Msg("failed to GetRabbitProducer")
	}
	defer publisher.Close()

	// init rabbitRepository and senderService
	rabbitRepository := repository.NewRabbitRepository(publisher, rabbitRepoRetryStrategy)
	senderService := service.NewSendService(StoreRepository, rabbitRepository, 5*time.Second, time.Hour)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		senderService.Run(ctx)
	}()

	// init redis
	redisClient := redis.New("localhost:6379", "", 0)
	redisRepository := repository.NewRedisRepository(redisClient, redisRepoRetryStrategy, time.Hour)

	// inint crud service
	crudService := service.NewCrudService(StoreRepository, redisRepository, nil)
	handl := handler.NewNotifyHandler(crudService)
	router := handler.NewRouter(handl)

	// running server
	httpServer := server.NewHTTPServer(router)
	err = httpServer.GracefulRun(ctx, "localhost", 8089)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Msg("failed GracefulRun server")
	}

	if err != nil {
		zlog.Logger.Error().Msg(fmt.Errorf("http server error: %w", err).Error())
	}

	zlog.Logger.Info().Msg("server gracefully stopped")
	ctxStop()
	wg.Wait()
	zlog.Logger.Info().Msg("background operations gracefully stopped")
}
