package service

import (
	"context"


	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/ports"
)

type SendService struct {
	storageRepo   ports.FetcherRepository
	puvlisherRepo ports.PublisherRepository
}

func NewSendService(
	storageRepo ports.FetcherRepository,
	puvlisherRepo ports.PublisherRepository,
) *SendService {
	return &SendService{
		storageRepo:   storageRepo,
		puvlisherRepo: puvlisherRepo,
	}
}

func (s *SendService) SendBatch(ctx context.Context, notifycationsToSent []*model.Notification) error {
	
	err := s.puvlisherRepo.SendMany(ctx, notifycationsToSent)
	if err != nil {
		return err
	}
	return nil

}
