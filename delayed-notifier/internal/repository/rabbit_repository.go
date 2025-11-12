package repository

import (
	"context"
	"encoding/json"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	rabbitpublisher "github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/rabbitProducer"
)

type RabbitRepository struct {
	publisher *rabbitpublisher.Publisher
}

func NewRabbitRepository(publisher *rabbitpublisher.Publisher) *RabbitRepository {
	return &RabbitRepository{
		publisher: publisher,
	}
}

func (p *RabbitRepository) SendMany(ctx context.Context, notifications []*model.Notification) error {
	// need add DLQ
	for _, notification := range notifications {
		body, err := json.Marshal(notification)
		if err != nil {
			return err
		}
		err = p.publisher.PublishWithRetry(ctx, body, notification.Channel)
		if err != nil {
			return err
		}

	}
	return nil

}
