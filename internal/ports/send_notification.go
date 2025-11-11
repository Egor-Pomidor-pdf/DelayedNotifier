package ports

import (
	"context"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/model"
)

type FetcherRepository interface {
	FetchFromDb(ctx context.Context, needToSendTime time.Time) ([]*model.Notification, error)
	MarkAsSent(ctx context.Context, id int) error
}

type PublisherRepository interface {
	SendMany(ctx context.Context, notifications []*model.Notification) error
}
