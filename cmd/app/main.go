package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/external/db"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/config"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/repository"
	"github.com/joho/godotenv"
)

const (
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	cfg := config.MustLoad()
	fmt.Println()
	setupLogger(cfg.Env) 

	slog.Info("starting app", slog.String("env", cfg.Env))
	slog.Debug("debug messages are enabled")

	db, err := db.NewPostgresDB(cfg.Database)
	
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	slog.Info("Successfully connected to PostgreSQL")
	slog.Info("working app...")

	ctx := context.Background()
	rep := repository.NewRepository(db)

	// 1
err = rep.CreateNotify(ctx, model.Notification{
    Recipient:   "111user1@example.com",
    Channel:     "email",
    Message:     "Message 1",
    ScheduledAt: time.Now().Add(1 * time.Hour),
})

if err != nil {
	log.Fatalf("pizda")
}

// 2
_ = rep.CreateNotify(ctx, model.Notification{
    Recipient:   "222user2@example.com",
    Channel:     "telegram",
    Message:     "Message 2",
    ScheduledAt: time.Now().Add(2 * time.Hour),
})

// 3
_ = rep.CreateNotify(ctx, model.Notification{
    Recipient:   "333user3@example.com",
    Channel:     "email",
    Message:     "Message 3",
    ScheduledAt: time.Now().Add(3 * time.Hour),
})

// 4
_ = rep.CreateNotify(ctx, model.Notification{
    Recipient:   "444user4@example.com",
    Channel:     "telegram",
    Message:     "Message 4",
    ScheduledAt: time.Now().Add(4 * time.Hour),
})

// 5
_ = rep.CreateNotify(ctx, model.Notification{
    Recipient:   "555user5@example.com",
    Channel:     "email",
    Message:     "Message 5",
    ScheduledAt: time.Now().Add(5 * time.Hour),
})
	m, err := rep.FetchFromDb(ctx, time.Now().Add(3 * time.Hour))
	if err != nil {
		log.Fatalf("failed to fetch notifications: %v", err)
	}
	for _, v := range m {
		fmt.Println(*v)

	}


	if err := db.Close();err != nil {
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
    slog.SetDefault(slog.New(handler))  // ← Устанавливаем global logger
}