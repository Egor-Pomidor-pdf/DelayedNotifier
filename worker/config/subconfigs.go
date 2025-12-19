package config

type RabbitMQConfig struct {
	User     string `yaml:"user" env:"RABBITMQ_USER"`         // Логин для подключения к RabbitMQ
	Password string `yaml:"password" env:"RABBITMQ_PASSWORD"` // Пароль для подключения
	Host     string `yaml:"host" env:"RABBITMQ_HOST"`         // Адрес сервера RabbitMQ (например, "localhost")
	Port     int    `yaml:"port" env:"RABBITMQ_PORT"`         // Порт RabbitMQ (обычно 5672)
	VHost    string `yaml:"vhost" env:"RABBITMQ_VHOST"`       // Виртуальный хост в RabbitMQ, для логической сегментации очередей
	Exchange string `yaml:"exchange" env:"RABBITMQ_EXCHANGE"` // Название exchange для публикации сообщений
	Queue    string `yaml:"queue" env:"RABBITMQ_QUEUE"`       // Название очереди, в которую будут публиковаться сообщения
	AutoAck bool 

}

type RetryConfig struct {
	Attempts          int     `yaml:"attempts" env:"ATTEMPTS"`
	DelayMilliseconds int     `yaml:"delay_milliseconds" env:"DELAY_MS"`
	Backoff           float64 `yaml:"backoff" env:"BACKOFF"`
}



