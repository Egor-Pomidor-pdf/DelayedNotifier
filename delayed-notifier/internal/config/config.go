package config

import (
	"fmt"

	"github.com/wb-go/wbf/config"
)

type Config struct {
	Env           string              `yaml:"env" env:"ENV"`
	Database      PostgresConfig      `env-prefix:"POSTGRES_"`
	RabbitMQ      RabbitMQConfig      `env-prefix:"RABBITMQ_"`
	RabbitMQRetry RabbitMQRetryConfig `env-prefix:"RETRY_RABBITMQ_"`
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
	myConfig.RabbitMQ.User = cfg.GetString("DELAYED_NOTIFIER_RABBITMQ_USER")
	myConfig.RabbitMQ.Password = cfg.GetString("DELAYED_NOTIFIER_RABBITMQ_PASSWORD")
	myConfig.RabbitMQ.Host = cfg.GetString("DELAYED_NOTIFIER_RABBITMQ_HOST")
	myConfig.RabbitMQ.Port = cfg.GetInt("DELAYED_NOTIFIER_RABBITMQ_PORT")
	myConfig.RabbitMQ.VHost = cfg.GetString("DELAYED_NOTIFIER_RABBITMQ_VHOST")
	myConfig.RabbitMQ.Exchange = cfg.GetString("DELAYED_NOTIFIER_RABBITMQ_EXCHANGE")
	myConfig.RabbitMQ.Queue = cfg.GetString("DELAYED_NOTIFIER_RABBITMQ_QUEUE")

	// Postgres
	myConfig.Database.MasterDSN = cfg.GetString("DELAYED_NOTIFIER_POSTGRES_MASTER_DSN")
	myConfig.Database.SlaveDSNs = cfg.GetStringSlice("DELAYED_NOTIFIER_POSTGRES_SLAVE_DSNS")
	myConfig.Database.MaxOpenConnections = cfg.GetInt("DELAYED_NOTIFIER_POSTGRES_MAX_OPEN_CONNECTIONS")
	myConfig.Database.MaxIdleConnections = cfg.GetInt("DELAYED_NOTIFIER_POSTGRES_MAX_IDLE_CONNECTIONS")
	myConfig.Database.ConnectionMaxLifetimeSeconds = cfg.GetInt("DELAYED_NOTIFIER_POSTGRES_CONNECTION_MAX_LIFETIME_SECONDS")

	// Retry
	myConfig.RabbitMQRetry.Attempts = cfg.GetInt("DELAYED_NOTIFIER_RETRY_RABBITMQ_RETRY_ATTEMPTS")
	myConfig.RabbitMQRetry.DelayMilliseconds = cfg.GetInt("DELAYED_NOTIFIER_RETRY_RABBITMQ_DELAY_MS")
	myConfig.RabbitMQRetry.Backoff = cfg.GetFloat64("DELAYED_NOTIFIER_RETRY_RABBITMQ_BACKOFF")
	return myConfig, nil
}
