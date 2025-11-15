package rabbitconsumer

import (
	"context"
	"fmt"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/config"
	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

type Consumer struct {
	Conn      *amqp091.Connection
	Channel   *amqp091.Channel
	RabbitCfg config.RabbitMQConfig
}

func NewRabbitConsumer(ctx context.Context, rabbitCfg config.RabbitMQConfig, rabbitmqRetryStrategy retry.Strategy) (*Consumer, error) {
	var conn *amqp091.Connection
	var err error

	// подключаемся с ретраями
	err = retry.DoContext(ctx, rabbitmqRetryStrategy, func() error {
		conn, err = amqp091.Dial(fmt.Sprintf(
			"amqp://%s:%s@%s:%d/%s",
			rabbitCfg.User,
			rabbitCfg.Password,
			rabbitCfg.Host,
			rabbitCfg.Port,
			rabbitCfg.VHost,
		))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("error creating channel: %w", err)
	}

	// объявляем exchange
	if err := ch.ExchangeDeclare(
		rabbitCfg.Exchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("error declaring exchange: %w", err)
	}

	err = retry.DoContext(ctx, rabbitmqRetryStrategy, func() error {
		_, errQ := ch.QueueDeclare(rabbitCfg.Queue, // имя очереди
			true,  // durable
			false, // autoDelete
			false, // exclusive
			false, // noWait
			nil,   // args)
		)
		if errQ == nil {
			errQ = ch.QueueBind(rabbitCfg.Queue, "email", rabbitCfg.Exchange, false, nil)

		}
		return errQ
	},
	)

	if err != nil {
		return nil, fmt.Errorf("error declaring queue '%s': %w", rabbitCfg.Queue, err)
	}

	return &Consumer{
		Conn:    conn,
		Channel: ch,
		RabbitCfg: rabbitCfg,
	}, nil
}

