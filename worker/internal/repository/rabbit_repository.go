package repository

import (
	"context"
	"fmt"

	rabbitconsumer "github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/rabbitConsumer"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitRepository struct {
	consumer *rabbitconsumer.Consumer
}

func NewRabbitRepository(consumer *rabbitconsumer.Consumer) *RabbitRepository {
	return &RabbitRepository{consumer: consumer}
}

func (r *RabbitRepository) СonsumeMsg(ctx context.Context) (<-chan amqp091.Delivery, error) {
	msgs, err := r.consumer.Channel.Consume(r.consumer.RabbitCfg.Queue, // имя очереди
		"",    // consumer — пустая строка, RabbitMQ сгенерирует уникальный тег
		false, // autoAck
		false, // exclusive
		false, // noLocal (не поддерживается RabbitMQ, оставляем false)
		false, // noWait
		nil,   // args
	)

	if err != nil {
		return nil, fmt.Errorf("failed to consume: %w", err)
	}

	return msgs, nil
}

// func (r *RabbitRepository) Start(ctx context.Context) error {
// 	for {
// 		_, err := r.consumeOnce(ctx)
// 		if err == nil {
// 			return nil
// 		}
// 		select {
// 		case <-ctx.Done():
// 			return nil
// 		default:
// 		}

// 		// добавить backoff
// 	}
// }
