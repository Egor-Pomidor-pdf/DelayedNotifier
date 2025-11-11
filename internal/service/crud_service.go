package service

import (
	"context"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/ports"
)

type CRUDService struct {
	storageRepo  ports.CRUDRepositoryInterface

}


func NewCrudService(
storageRepo ports.CRUDRepositoryInterface,
) *CRUDService{
	return &CRUDService{
		storageRepo: storageRepo,
	}
}

func (s *CRUDService) CreateNotification(ctx context.Context, notify *model.Notification) (*model.Notification, error) {
	_ = s.storageRepo.CreateNotify(ctx, notify)
	return notify, nil
}

func (s *CRUDService) GetNotifyStatus(ctx context.Context, id int) (*model.Notification, error) {
	notify, _ := s.storageRepo.GetNotify(ctx, id)
	return notify, nil
}