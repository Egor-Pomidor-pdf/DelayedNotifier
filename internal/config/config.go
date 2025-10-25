package config

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
}
