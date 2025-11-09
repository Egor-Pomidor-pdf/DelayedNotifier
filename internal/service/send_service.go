package service

import (
	"context"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/ports"
)

type SendService struct {
	storageRepo ports.StoreRepositoryEnterface
}

func (s * SendService) SendBatch(ctx context.Context, notifycationsToSent []*model.Notification) error {

}

func (s * SendService) LyfeCycle(ctx context.Context) {
	batch, err := s.storageRepo.FetchFromDb(ctx, time.Now().Add(6*time.Hour))

}