package ports

import (
	"context"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/model"
)


type NotificationReceiver interface {
	StartReceiving(ctx context.Context) (chan *model.Notification, error)
	StopReceiving() error
}

type NotificationSender interface {
	Send(ctx context.Context, notification *model.Notification) error
}