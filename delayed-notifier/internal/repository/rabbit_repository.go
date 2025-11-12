package repository

import (
	"context"
	"encoding/json"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	"github.com/wb-go/wbf/rabbitmq"
)

type RabbitRepository struct {
	publisher *rabbitmq.Publisher
}

func NewRabbitRepository(p *rabbitmq.Publisher) *RabbitRepository {
	return &RabbitRepository{
		publisher: p,
	}
}

func (p *RabbitRepository) SendMany(ctx context.Context, notifications []*model.Notification) error {
	// need add DLQ
	for _, notification := range notifications {
		body, err := json.Marshal(notification)
		if err != nil {
			return err
		}
		err = p.publisher.Publish(body, notification.Channel, "application/json")
		if err != nil {
			return err
		}

	}
	return nil

}
