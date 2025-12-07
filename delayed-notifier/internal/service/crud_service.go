package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/ports"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/types"
	"github.com/wb-go/wbf/zlog"
	"golang.org/x/sync/errgroup"
)

type SignalFunc func(ctx context.Context, notify *model.Notification) error

type CRUDService struct {
	storageRepo  ports.CRUDStoreRepositoryInterface
	redisRepo    ports.CRUDRedisRepositoryInterface
	funcOnCreate SignalFunc
}

func NewCrudService(
	storageRepo ports.CRUDStoreRepositoryInterface,
	redisRepo ports.CRUDRedisRepositoryInterface,
	funcOnCreate SignalFunc,
) *CRUDService {
	return &CRUDService{
		storageRepo:  storageRepo,
		redisRepo:    redisRepo,
		funcOnCreate: funcOnCreate,
	}
}

func (s *CRUDService) CreateNotification(ctx context.Context, notify *model.Notification) (*model.Notification, error) {
	uuid := types.GenerateUUID()
	notify.ID = &uuid

	err := s.storageRepo.CreateNotify(ctx, notify)
	if err != nil {
		return nil, fmt.Errorf("notification storage failed to create: %v", err)
	}

	s.trySaveInCache(ctx, notify)

	if s.funcOnCreate != nil {
		go func(notify *model.Notification) {
			err := s.funcOnCreate(ctx, notify)
			if err != nil {
				zlog.Logger.Error().Err(err).Msg(fmt.Sprintf("error in funcOnCreate %v", s.funcOnCreate))
			}
		}(notify)
	}
	zlog.Logger.Info().Msg("success create notification")
	return notify, nil
}

func (s *CRUDService) GetNotification(ctx context.Context, id types.UUID) (*model.Notification, error) {
	result, err := s.getObjectFromCache(ctx, id)
	if err != nil {
		result, err = s.getObjectFromStorage(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("error getting object from storage: %w", err)
		}

		if result != nil {
			s.trySaveInCache(ctx, result)
		}
	}
	zlog.Logger.Info().Msg("success get notification")
	return result, err
}

func (s *CRUDService) DeleteNotification(ctx context.Context, id types.UUID) error {
	object, err := s.GetNotification(ctx, id)
	if err != nil {
		return fmt.Errorf("error checking object existence: %w", err)
	}

	if object == nil || object.ID == nil {
		return errors.New("notification not found")
	}

	var errGroup errgroup.Group

	errGroup.Go(func() error {
		return s.storageRepo.DeleteNotification(ctx, id)
	})
	errGroup.Go(func() error {
		return s.redisRepo.DeleteNotification(ctx, id)
	})
	errGroup.Wait()
	zlog.Logger.Info().Msg("success delete notification")
	return nil
}

func (s *CRUDService) getObjectFromStorage(ctx context.Context, id types.UUID) (*model.Notification, error) {
	return s.storageRepo.GetNotify(ctx, id)
}

func (s *CRUDService) getObjectFromCache(ctx context.Context, id types.UUID) (*model.Notification, error) {
	return s.redisRepo.GetNotification(ctx, id)
}

func (s *CRUDService) trySaveInCache(ctx context.Context, model *model.Notification) {
	go func() {
		err := s.redisRepo.SaveNotification(ctx, model)
		if err != nil {
			zlog.Logger.Error().Err(fmt.Errorf("error saving in cache: %w", err))
		}
	}()
}
