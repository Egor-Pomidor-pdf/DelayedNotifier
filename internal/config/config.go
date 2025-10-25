package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string         `yaml:"env" env:"ENV"`
	Database DatabaseConfig `yaml:"database"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"     env:"DB_HOST"`
	Port     int    `yaml:"port"     env:"DB_PORT"`
	Name     string `yaml:"name"     env:"DB_NAME"`
	User     string `yaml:"user"     env:"DB_USER"`
	Password string `yaml:"password" env:"DB_PASSWORD"`
	SSLMode  string `yaml:"sslmode" env:"DB_SSLMODE"`
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
