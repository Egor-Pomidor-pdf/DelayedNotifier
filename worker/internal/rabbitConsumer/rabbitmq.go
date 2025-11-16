package rabbitconsumer

import (
	"context"
	"fmt"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/config"
	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

type Consumer struct {
	Chan *amqp091.Channel
	Cfg  config.RabbitMQConfig
}

func NewRabbitConsumer(ctx context.Context, rabbitCfg config.RabbitMQConfig, rabbitmqRetryStrategy retry.Strategy) (*Consumer, *amqp091.Channel, error) {
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
		return nil, nil, fmt.Errorf("error connecting to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("error creating channel: %w", err)
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
		return nil, nil, fmt.Errorf("error declaring exchange: %w", err)
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
		return nil, nil, fmt.Errorf("error declaring queue '%s': %w", rabbitCfg.Queue, err)
	}

	return &Consumer{
		Chan: ch,
		Cfg:  rabbitCfg,
	}, ch, nil
}

func (c *Consumer) ConsumeWithRetry(ctx context.Context, out chan amqp091.Delivery, retryStrategy retry.Strategy) error {
	for {
		var deliveries <-chan amqp091.Delivery

		err := retry.DoContext(ctx, retryStrategy, func() error {
			var err error
			deliveries, err = c.Chan.Consume(
				c.Cfg.Queue, // имя очереди
				"",          // consumer — пустая строка, RabbitMQ сгенерирует уникальный тег
				false,       // autoAck
				false,       // exclusive
				false,       // noLocal (не поддерживается RabbitMQ, оставляем false)
				false,       // noWait
				nil,         // args
			)
			return err
		})
		if err != nil {
			// контекст завершён — выходим
			return err
		}

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case m, ok := <-deliveries:
				if !ok {
					// можно в будующем добавить reconnect
					return fmt.Errorf("deliveries channel closed")
				}
				out <- m

			}
		}
	}

}
