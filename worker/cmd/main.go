package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/config"
	rabbitconsumer "github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/rabbitConsumer"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/repository/receivers"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/repository/senders"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/service"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	// make context
	ctx := context.Background()
	ctx, ctxStop := signal.NotifyContext(ctx, os.Interrupt)

	cfg, err := config.NewConfig("", "")
	fmt.Println(cfg)
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

	slog.Info("starting app", slog.String("env", cfg.Env))

	// init strategies
	consumerRetryStrategy := config.MakeStrategy(cfg.ConsumerRetry)
	receiverRetryStrategy := config.MakeStrategy(cfg.ReceiverRetry)

	// init consumer
	consumer, _, err := rabbitconsumer.NewRabbitConsumer(ctx, cfg.RabbitMQ, consumerRetryStrategy)
	if err != nil {
		zlog.Logger.Fatal().
			Err(err).
			Msg("failed to create rabbit consumer")
	}

	// init reciver and sender
	receiver := receivers.NewRabbitMQReceiver(consumer, receiverRetryStrategy)
	sender := senders.NewConsoleSender()

	// init duration
	duration, err := time.ParseDuration(cfg.CheckPeriod)
	if err != nil {
		zlog.Logger.Fatal().
			Err(err).
			Str("check_period", cfg.CheckPeriod).
			Msg("invalid check period")
	}

	// init notificationService and run
	notificationService := service.NewNotificationService(receiver, sender, duration)
	err = notificationService.Run(ctx, cfg.RabbitMQ)
	if err != nil {
		zlog.Logger.Fatal().
			Err(err).
			Msg("notification service exited with error")
	}

	ctxStop()

}
