package receivers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/dto"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/model"
	rabbitconsumer "github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/rabbitConsumer"
	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/retry"
)

type RabbitMQReceiver struct {
	consumer      *rabbitconsumer.Consumer
	messages      chan amqp091.Delivery
	objectsChan   chan *model.Notification
	retryStrategy retry.Strategy
}

func NewRabbitMQReceiver(consumer *rabbitconsumer.Consumer, retryStrategy retry.Strategy) *RabbitMQReceiver {
	return &RabbitMQReceiver{
		consumer:      consumer,
		messages:      make(chan amqp091.Delivery),
		objectsChan:   make(chan *model.Notification),
		retryStrategy: retryStrategy,
	}
}

func (r *RabbitMQReceiver) StartReceiving(ctx context.Context) (chan *model.Notification, error) {
	go func() {
		err := r.consumer.ConsumeWithRetry(ctx, r.messages, r.retryStrategy)
		if err != nil {
			slog.Error("ConsumeWithRetry", "err:", err)
		}
	}()

	go func() {
		defer close(r.objectsChan)
		for delivery := range r.messages {
			data := delivery.Body
			object, err := r.processMessage(data)
			if err != nil {
				slog.Error("ConsumeWithRetry", "err:", err)
				continue
				// обработка ошибок далее реализую
			}
			r.objectsChan <- object
			delivery.Ack(false)

		}

	}()

	return r.objectsChan, nil
}

func (r * RabbitMQReceiver) StopReceiving() error {
	err := r.consumer.Chan.Close()
	if err != nil {
		return fmt.Errorf("can not close chan %w", err)
	}
	close(r.messages)
	return nil
}

func (r *RabbitMQReceiver) processMessage(delivery []byte) (*model.Notification, error) {
	var msg dto.NotificationSendBody

	if err := json.Unmarshal(delivery, &msg); err != nil {
		return nil, fmt.Errorf("bad message (bad json): %w", err)
	}

	notification, err := dto.NotificationModelFromSendDTO(&msg)
	if err != nil {
		return nil, fmt.Errorf("bad message (could't convert to model): %w", err)
	}

	return notification, nil
}
