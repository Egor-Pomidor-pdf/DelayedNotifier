package ports

import (
	"context"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/model"
)

type CRUDRepositoryInterface interface {
	CreateNotify(ctx context.Context, notify *model.Notification) error
	GetNotify(ctx context.Context, id int) (*model.Notification, error)
	FetchFromDb(ctx context.Context, needToSendTime time.Time) ([]*model.Notification, error)
	DeleteNotify(ctx context.Context, id int) error
}

type CRUDServiceInterface interface {
	CreateNotification(ctx context.Context, model *model.Notification) (*model.Notification, error)
	GetNotifyStatus(ctx context.Context, id int) (*model.Notification, error) 

	//after relise
	DeleteNotify(ctx context.Context, id int) error
}
