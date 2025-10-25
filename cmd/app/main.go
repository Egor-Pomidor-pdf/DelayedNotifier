package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/config"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/db"
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