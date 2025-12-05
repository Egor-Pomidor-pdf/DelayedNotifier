package ports

import (
	"context"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/dlq"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/types"
)

type FetcherRepository interface {
	FetchFromDb(ctx context.Context, needToSendTime time.Time) ([]*model.Notification, error)
	MarkAsSent(ctx context.Context, ids []*types.UUID) error
}

type PublisherRepository interface {
	SendMany(ctx context.Context, notifications []*model.Notification) *dlq.DLQ[*model.Notification]
	SendOne(ctx context.Context, notification *model.Notification) error
}
