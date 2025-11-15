package config

import (
	"fmt"

	"github.com/wb-go/wbf/config"
)

type Config struct {
	Env           string              `yaml:"env" env:"ENV"`
	// Database      DatabaseConfig      `yaml:"database"`
	RabbitMQ      RabbitMQConfig      `yaml:"rabbitmq"`
	RabbitMQRetry RabbitMQRetryConfig `yaml:"rabbitmq_retry"`
}

func NewConfig(envFilePath string, configFilePath string) (*Config, error) {
	myConfig := &Config{}

	cfg := config.New()

	if envFilePath != "" {
		if err := cfg.LoadEnvFiles(envFilePath); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}
	cfg.EnableEnv("")

	if configFilePath != "" {
		if err := cfg.LoadConfigFiles(configFilePath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	myConfig.Env = cfg.GetString("ENV")

	// RabbitMQ
	myConfig.RabbitMQ.User = cfg.GetString("RABBITMQ_USER")
	myConfig.RabbitMQ.Password = cfg.GetString("RABBITMQ_PASSWORD")
	myConfig.RabbitMQ.Host = cfg.GetString("RABBITMQ_HOST")
	myConfig.RabbitMQ.Port = cfg.GetInt("RABBITMQ_PORT")
	myConfig.RabbitMQ.VHost = cfg.GetString("RABBITMQ_VHOST")
	myConfig.RabbitMQ.Exchange = cfg.GetString("RABBITMQ_EXCHANGE")
	myConfig.RabbitMQ.Queue = cfg.GetString("RABBITMQ_QUEUE")

	// Postgres
	// myConfig.Database.Host = cfg.GetString("POSTGRES_HOST")
	// myConfig.Database.Port = cfg.GetInt("POSTGRES_PORT")
	// myConfig.Database.Name = cfg.GetString("POSTGRES_DB")
	// myConfig.Database.User = cfg.GetString("POSTGRES_USER")
	// myConfig.Database.Password = cfg.GetString("POSTGRES_PASSWORD")
	// myConfig.Database.SSLMode = cfg.GetString("POSTGRES_SSLMODE")

	// Retry
	myConfig.RabbitMQRetry.Attempts = cfg.GetInt("RABBITMQ_RETRY_ATTEMPTS")
	myConfig.RabbitMQRetry.DelayMilliseconds = cfg.GetInt("RABBITMQ_RETRY_DELAY_MS")
	myConfig.RabbitMQRetry.Backoff = cfg.GetFloat64("RABBITMQ_RETRY_BACKOFF")
	return myConfig, nil
}
