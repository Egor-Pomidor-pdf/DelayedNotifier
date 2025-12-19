package config

import (
	"fmt"
	"time"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/retry"
)

type Config struct {
	Env           string        
	RabbitMQ      RabbitMQConfig 
	ConsumerRetry RetryConfig    
	ReceiverRetry RetryConfig   
	CheckPeriod   string         
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
	myConfig.CheckPeriod = cfg.GetString("CHECK_PERIOD")

	// Retry
	// Consumer retry
	myConfig.ConsumerRetry.Attempts = cfg.GetInt("DELAYED_NOTIFIER_RETRY_CONSUMER_ATTEMPTS")
	myConfig.ConsumerRetry.DelayMilliseconds = cfg.GetInt("DELAYED_NOTIFIER_RETRY_CONSUMER_DELAY_MS")
	myConfig.ConsumerRetry.Backoff = cfg.GetFloat64("DELAYED_NOTIFIER_RETRY_CONSUMER_BACKOFF")

	// Receiver retry
	myConfig.ReceiverRetry.Attempts = cfg.GetInt("DELAYED_NOTIFIER_RETRY_RECEIVER_ATTEMPTS")
	myConfig.ReceiverRetry.DelayMilliseconds = cfg.GetInt("DELAYED_NOTIFIER_RETRY_RECEIVER_DELAY_MS")
	myConfig.ReceiverRetry.Backoff = cfg.GetFloat64("DELAYED_NOTIFIER_RETRY_RECEIVER_BACKOFF")

	return myConfig, nil
}

func MakeStrategy(c RetryConfig) retry.Strategy {
	return retry.Strategy{
		Attempts: c.Attempts,
		Delay:    time.Duration(c.DelayMilliseconds) * time.Millisecond,
		Backoff:  c.Backoff,
	}
}

