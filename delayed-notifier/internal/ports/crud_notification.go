package ports

import (
	"context"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/types"
)

type CRUDStoreRepositoryInterface interface {
	CreateNotify(ctx context.Context, notify *model.Notification) error
	GetNotify(ctx context.Context, id types.UUID) (*model.Notification, error)
	FetchFromDb(ctx context.Context, needToSendTime time.Time) ([]*model.Notification, error)
	DeleteNotification(ctx context.Context, id types.UUID) error
}

type CRUDRedisRepositoryInterface interface {
	SaveNotification(ctx context.Context, notify *model.Notification) error
	GetNotification(ctx context.Context, id types.UUID) (*model.Notification, error)
	DeleteNotification(ctx context.Context, id types.UUID) error 
}

type CRUDServiceInterface interface {
	CreateNotification(ctx context.Context, model *model.Notification) (*model.Notification, error)
	GetNotification(ctx context.Context, id types.UUID) (*model.Notification, error) 
	DeleteNotification(ctx context.Context, id types.UUID) error
}
