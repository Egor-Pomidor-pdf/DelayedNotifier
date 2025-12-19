package repository

import (
	"context"
	"fmt"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/dto"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	rabbitpublisher "github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/rabbitProducer"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/dlq"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type RabbitRepository struct {
	publisher     *rabbitpublisher.Publisher
	retryStrategy retry.Strategy
}

func NewRabbitRepository(publisher *rabbitpublisher.Publisher, retryStrategy retry.Strategy) *RabbitRepository {
	return &RabbitRepository{
		publisher:     publisher,
		retryStrategy: retryStrategy,
	}
}

func (n *RabbitRepository) SendOne(ctx context.Context, notification *model.Notification) error {
	body, err := dto.ToSendFromDTO(notification)
	if err != nil {
		return fmt.Errorf("couldn't create body to send one: %w", err)
	}
	err = n.publisher.PublishWithRetry(ctx, body, n.routingKey(notification))
	if err != nil {
		return fmt.Errorf("couldn't send message to rabbitMQ: %w", err)
	}
	zlog.Logger.Debug().Msg("sent one message to rabbitMQ")
	return nil
}

func (p *RabbitRepository) SendMany(ctx context.Context, notifications []*model.Notification) *dlq.DLQ[*model.Notification] {
	DLQ := dlq.NewDLQ[*model.Notification](len(notifications) / 10)

	go func() {
		for _, notification := range notifications {

			body, err := dto.ToSendFromDTO(notification)

			if err != nil {
				zlog.Logger.Err(err).Msg("dont recreate in dto")
			}
			if err != nil {
				DLQ.Put(notification, fmt.Errorf("couldn't send message to rabbitMQ: %w", err))
			}
			err = p.publisher.PublishWithRetry(ctx, body, p.routingKey(notification))
			if err != nil {
				DLQ.Put(notification, fmt.Errorf("couldn't send message to rabbitMQ: %w", err))
			} else {
				zlog.Logger.Debug().Msg("SendMany sent message in batch to rabbitMQ")
			}

		}
		DLQ.Close()
	}()
	return DLQ

}

func (n *RabbitRepository) routingKey(notification *model.Notification) string {
	return notification.Channel.String()
}
