package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string         `yaml:"env" env:"ENV"`
	Database DatabaseConfig `yaml:"database"`
	RabbitMQ RabbitMQConfig 
}

type DatabaseConfig struct {
	Host     string `yaml:"host"     env:"DB_HOST"`
	Port     int    `yaml:"port"     env:"DB_PORT"`
	Name     string `yaml:"name"     env:"DB_NAME"`
	User     string `yaml:"user"     env:"DB_USER"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
	SSLMode  string `yaml:"sslmode" env:"DB_SSLMODE"`
}

type RabbitMQConfig struct {
	User     string `yaml:"user" env:"RABBITMQ_USER"`       // Логин для подключения к RabbitMQ
	Password string `yaml:"password" env:"RABBITMQ_PASSWORD"` // Пароль для подключения
	Host     string `yaml:"host" env:"RABBITMQ_HOST"`       // Адрес сервера RabbitMQ (например, "localhost")
	Port     int    `yaml:"port" env:"RABBITMQ_PORT"`       // Порт RabbitMQ (обычно 5672)
	VHost    string `yaml:"vhost" env:"RABBITMQ_VHOST"`     // Виртуальный хост в RabbitMQ, для логической сегментации очередей
	Exchange string `yaml:"exchange" env:"RABBITMQ_EXCHANGE"` // Название exchange для публикации сообщений
	Queue    string `yaml:"queue" env:"RABBITMQ_QUEUE"`     // Название очереди, в которую будут публиковаться сообщения
}

type ServerConfig struct {
	Address string `yaml:"address"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	var cfg Config
	var err error

	if configPath != "" {
		err = cleanenv.ReadConfig(configPath, &cfg)
	} else {
		err = cleanenv.ReadEnv(&cfg)
	}

	if err != nil {
		log.Fatal(err)
	}
	return &cfg

}
