package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/ports"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/dlq"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/types"
	"github.com/wb-go/wbf/zlog"
	"golang.org/x/sync/errgroup"
)

type SendService struct {
	fetchPeriod      time.Duration
	fetchMaxDiapason time.Duration

	storageFetcherRepo ports.FetcherRepository
	puvlisherRepo      ports.PublisherRepository
}

func NewSendService(
	storageRepo ports.FetcherRepository,
	puvlisherRepo ports.PublisherRepository,
) *SendService {
	return &SendService{
		storageFetcherRepo: storageRepo,
		puvlisherRepo:      puvlisherRepo,
	}
}

func (s *SendService) Run(ctx context.Context) {
	ticker := time.NewTicker(s.fetchPeriod)
	defer ticker.Stop()
	s.lifeCycle(ctx)

out:
	for {
		select {
		case <-ctx.Done():
			break out
		case <-ticker.C:
			s.lifeCycle(ctx)
		}
	}
}

func (s *SendService) QuickSend(ctx context.Context, obj *model.Notification) error {
	err := s.puvlisherRepo.SendOne(ctx, obj) // it calls retry inside!
	go func() {
		errMark := s.storageFetcherRepo.MarkAsSent(ctx, []*types.UUID{obj.ID})
		if errMark != nil {
			zlog.Logger.Error().Err(errMark).Msg("failed to mark as sent")
		}
	}()
	return err
}

func (s *SendService) SendBatch(ctx context.Context, notifycationsToSent []*model.Notification) error {
	DLQ := s.puvlisherRepo.SendMany(ctx, notifycationsToSent)

	var err error
	errGroup := &errgroup.Group{}
	errCount := 0

	go func() {
		ids := make([]*types.UUID, len(notifycationsToSent))
		for i, obj := range notifycationsToSent {
			ids[i] = obj.ID
		}
		err := s.storageFetcherRepo.MarkAsSent(ctx, ids)
		if err != nil {
			zlog.Logger.Error().Err(err).Msg("failed to mark as sent")
		}
	}()

	// remark failede msg ToDo

	for obj := range DLQ.Items() {
		errCount++
		errGroup.Go(func() error {
			return func(obj *dlq.Item[*model.Notification]) error {
				zlog.Logger.Error().
					Err(obj.Error()).
					Stringer("id", obj.Value().ID).
					Msg("failed to send object, trying to resend...")

				err := s.QuickSend(ctx, obj.Value())
				if err != nil {
					zlog.Logger.Error().
						Err(obj.Error()).
						Stringer("id", obj.Value().ID).
						Any("object", obj.Value()).
						Msg("failed to send object!")

					return err
				}
				zlog.Logger.Info().
					Stringer("id", obj.Value().ID).
					Msg("successfully sent object on second try!")

				return nil
			}(obj)
		})
	}

	err = errGroup.Wait()
	if err != nil {
		return fmt.Errorf("failed to send '%d' objects, example err: %w", errCount, err)
	}
	return nil
}

func (s *SendService) lifeCycle(ctx context.Context) {
	now := time.Now()

	dateTimeForSent := now.Add(s.fetchPeriod)
	batch, err := s.storageFetcherRepo.FetchFromDb(ctx, dateTimeForSent)
	if err != nil {
		zlog.Logger.Error().Err(fmt.Errorf("failed to fetch batch for sending: %w", err)).Msg("error in SenderService loop")
		return
	}
	zlog.Logger.Info().Int("amount", len(batch)).Stringer("max_publication_at", dateTimeForSent).Msg("fetched batch")
	err = s.SendBatch(ctx, batch)
	if err != nil {
		zlog.Logger.Error().Err(fmt.Errorf("failed to send batch: %w", err)).Msg("error in SenderService loop")
		return
	}
}
