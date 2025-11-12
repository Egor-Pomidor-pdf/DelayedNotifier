package rabbitpublisher

import (
	"context"
	"fmt"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/config"
	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

type Publisher struct {
	conn      *amqp091.Connection
	channel   *amqp091.Channel
	exchange  string
	contentType string
	retryStrategy retry.Strategy
}


func NewRabbitProducer(ctx context.Context, rabbitCfg config.RabbitMQConfig, rabbitmqRetryStrategy retry.Strategy) (*Publisher, error) {
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

	// биндим очередь к exchange (routing key можно оставить пустым для direct)
	if err := ch.QueueBind(
		rabbitCfg.Queue,
		"", 
		rabbitCfg.Exchange,
		false,
		nil,
	); err != nil {
		return nil, fmt.Errorf("error binding queue '%s' to exchange: %w", rabbitCfg.Queue, err)
	}

	return &Publisher{
		conn:       conn,
		channel:    ch,
		exchange:   rabbitCfg.Exchange,
		contentType: "application/json",
		retryStrategy: rabbitmqRetryStrategy,
	}, nil
}
// PublishWithRetry публикует сообщение с ретраями
func (p *Publisher) PublishWithRetry(ctx context.Context, body []byte, routingKey string) error {
	return retry.DoContext(ctx, p.retryStrategy, func() error {
		return p.channel.PublishWithContext(ctx, p.exchange, routingKey, false, false, amqp091.Publishing{
			ContentType: p.contentType,
			Body:        body,
		})
	})
}

// Close закрывает канал и соединение
func (p *Publisher) Close() error {
	if p.channel != nil {
		err := p.channel.Close()
		if err != nil {
			return err
		}
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}