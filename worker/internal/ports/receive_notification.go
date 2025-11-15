package ports

import (
	"context"

	"github.com/rabbitmq/amqp091-go"
)


type ConsumerRpositoryInterface interface {
	СonsumeMsg(ctx context.Context) (<-chan amqp091.Delivery, error)
}

type SenderRepositoryInterface interface {

}