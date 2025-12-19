package config

type PostgresConfig struct {
	MasterDSN                    string   `env:"MASTER_DSN"`
	SlaveDSNs                    []string `env:"SLAVE_DSNS" envSeparator:","`
	MaxOpenConnections           int      `env:"MAX_OPEN_CONNECTIONS" envDefault:"3"`
	MaxIdleConnections           int      `env:"MAX_IDLE_CONNECTIONS" envDefault:"5"`
	ConnectionMaxLifetimeSeconds int      `env:"CONNECTION_MAX_LIFETIME_SECONDS" envDefault:"0"`
}

type RabbitMQConfig struct {
	User     string `yaml:"user" env:"USER"`         // Логин для подключения к RabbitMQ
	Password string `yaml:"password" env:"PASSWORD"` // Пароль для подключения
	Host     string `yaml:"host" env:"HOST"`         // Адрес сервера RabbitMQ (например, "localhost")
	Port     int    `yaml:"port" env:"PORT"`         // Порт RabbitMQ (обычно 5672)
	VHost    string `yaml:"vhost" env:"VHOST"`       // Виртуальный хост в RabbitMQ, для логической сегментации очередей
	Exchange string `yaml:"exchange" env:"EXCHANGE"` // Название exchange для публикации сообщений
	Queue    string `yaml:"queue" env:"QUEUE"`       // Название очереди, в которую будут публиковаться сообщения

}

type RedisConfig struct {
	Host       string `yaml:"host" env:"HOST"`             // Адрес Redis (например, "localhost")
	Port       int    `yaml:"port" env:"PORT"`             // Порт Redis (обычно 6379)
	Password   string `yaml:"password" env:"PASSWORD"`     // Пароль, если настроена аутентификация
	DB         int    `yaml:"db" env:"DB"`                 // Номер базы Redis (по умолчанию 0)
	Expiration int    `yaml:"expiration" env:"EXPIRATION"` // Время жизни ключей (TTL)
}

type ServerConfig struct {
	Host string `yaml:"host"` // например, "localhost"
	Port int    `yaml:"port"` // например, 8080
}


type RetryConfig struct {
	Attempts          int     `yaml:"attempts" env:"ATTEMPTS"`
	DelayMilliseconds int     `yaml:"delay_milliseconds" env:"DELAY_MS"`
	Backoff           float64 `yaml:"backoff" env:"BACKOFF"`
}


type LogConfig struct {
	Address string `yaml:"address"`
}
