package config

import (
	"fmt"
	"time"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/retry"
)

type Config struct {
	Env             string         `yaml:"env" env:"ENV"`
	Database        PostgresConfig `env-prefix:"POSTGRES_"`
	Redis           RedisConfig    `env-prefix:"REDIS_"`
	RabbitMQ        RabbitMQConfig `env-prefix:"RABBITMQ_"`
	RabbitMQRetry   RetryConfig    `env-prefix:"RETRY_RABBITMQ_"`
	PostgresRetry   RetryConfig    `env-prefix:"RETRY_POSTGRES_"`
	StoreRepoRetry  RetryConfig    `env-prefix:"RETRY_STORE_REPO_"`
	RabbitRepoRetry RetryConfig    `env-prefix:"RETRY_RABBIT_REPO_"`
	RedisRepoRetry  RetryConfig    `env-prefix:"RETRY_REDIS_REPO_"`
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

	// Redis
	myConfig.Redis.Host = cfg.GetString("DELAYED_NOTIFIER_REDIS_HOST")
	myConfig.Redis.Port = cfg.GetInt("DELAYED_NOTIFIER_REDIS_PORT")
	myConfig.Redis.Password = cfg.GetString("DELAYED_NOTIFIER_REDIS_PASSWORD")
	myConfig.Redis.DB = cfg.GetInt("DELAYED_NOTIFIER_REDIS_DB")
	myConfig.Redis.Expiration = cfg.GetInt("DELAYED_NOTIFIER_REDIS_EXPIRATION")

	// Retry
	// RabbitMQ retry
	myConfig.RabbitMQRetry.Attempts = cfg.GetInt("DELAYED_NOTIFIER_RETRY_RABBITMQ_RETRY_ATTEMPTS")
	myConfig.RabbitMQRetry.DelayMilliseconds = cfg.GetInt("DELAYED_NOTIFIER_RETRY_RABBITMQ_DELAY_MS")
	myConfig.RabbitMQRetry.Backoff = cfg.GetFloat64("DELAYED_NOTIFIER_RETRY_RABBITMQ_BACKOFF")

	// Postgres retry
	myConfig.PostgresRetry.Attempts = cfg.GetInt("DELAYED_NOTIFIER_RETRY_POSTGRES_ATTEMPTS")
	myConfig.PostgresRetry.DelayMilliseconds = cfg.GetInt("DELAYED_NOTIFIER_RETRY_POSTGRES_DELAY_MS")
	myConfig.PostgresRetry.Backoff = cfg.GetFloat64("DELAYED_NOTIFIER_RETRY_POSTGRES_BACKOFF")

	// StoreRepository retry
	myConfig.StoreRepoRetry.Attempts = cfg.GetInt("DELAYED_NOTIFIER_RETRY_STORE_REPO_ATTEMPTS")
	myConfig.StoreRepoRetry.DelayMilliseconds = cfg.GetInt("DELAYED_NOTIFIER_RETRY_STORE_REPO_DELAY_MS")
	myConfig.StoreRepoRetry.Backoff = cfg.GetFloat64("DELAYED_NOTIFIER_RETRY_STORE_REPO_BACKOFF")

	// RabbitRepository retry
	myConfig.RabbitRepoRetry.Attempts = cfg.GetInt("DELAYED_NOTIFIER_RETRY_RABBIT_REPO_ATTEMPTS")
	myConfig.RabbitRepoRetry.DelayMilliseconds = cfg.GetInt("DELAYED_NOTIFIER_RETRY_RABBIT_REPO_DELAY_MS")
	myConfig.RabbitRepoRetry.Backoff = cfg.GetFloat64("DELAYED_NOTIFIER_RETRY_RABBIT_REPO_BACKOFF")

	// RedisRepository retry
	myConfig.RedisRepoRetry.Attempts = cfg.GetInt("DELAYED_NOTIFIER_RETRY_REDIS_REPO_ATTEMPTS")
	myConfig.RedisRepoRetry.DelayMilliseconds = cfg.GetInt("DELAYED_NOTIFIER_RETRY_REDIS_REPO_DELAY_MS")
	myConfig.RedisRepoRetry.Backoff = cfg.GetFloat64("DELAYED_NOTIFIER_RETRY_REDIS_REPO_BACKOFF")

	return myConfig, nil
}

func MakeStrategy(c RetryConfig) retry.Strategy {
	return retry.Strategy{
		Attempts: c.Attempts,
		Delay:    time.Duration(c.DelayMilliseconds) * time.Millisecond,
		Backoff:  c.Backoff,
		// если в retry.Strategy есть поле MaxDelay, можно использовать:
		// MaxDelay: time.Duration(c.MaxDelayMs) * time.Millisecond,
	}
}
