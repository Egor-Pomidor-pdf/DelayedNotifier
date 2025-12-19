package service

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/config"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/model"
	notificationheap "github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/notificationHeap"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/ports"
	"github.com/wb-go/wbf/zlog"
)

type NotificationService struct {
	receiver         ports.NotificationReceiver
	channelToSender  ports.NotificationSender //нужно мапу сделать
	checkPeriod      time.Duration
	notificationHeap *notificationheap.NotificationHeap
	heapMutex        sync.RWMutex
}

func NewNotificationService(receiver ports.NotificationReceiver, channelToSender ports.NotificationSender, checkPeriod time.Duration) *NotificationService {
	notificationHeap := &notificationheap.NotificationHeap{}
	heap.Init(notificationHeap)
	return &NotificationService{
		receiver:         receiver,
		channelToSender:  channelToSender,
		checkPeriod:      checkPeriod,
		notificationHeap: notificationHeap,
		heapMutex:        sync.RWMutex{}}
}

func (s *NotificationService) Run(ctx context.Context, rabbitCfg config.RabbitMQConfig) error {
	objects, err := s.receiver.StartReceiving(ctx)
	if err != nil {
		return fmt.Errorf("cannot start consumer: %w", err)
	}
	zlog.Logger.Info().
		Str("queue", rabbitCfg.Queue).
		Msg("notification service started receiving messages")
	var object *model.Notification

	go s.serveHeap(ctx)

out:
	for {
		select {
		case <-ctx.Done():
			zlog.Logger.Info().
				Msg("notification service: context cancelled, stopping Run loop")
			break out
		case object = <-objects:
			// надо добавить провекру, что такой канал есть в мапе
			s.heapMutex.Lock()
			heap.Push(s.notificationHeap, object)
			s.heapMutex.Unlock()
		}
	}
	if err := s.receiver.StopReceiving(); err != nil {
		zlog.Logger.Error().
			Err(err).
			Msg("error while stopping receiver")
		return err
	}

	zlog.Logger.Info().
		Msg("notification service stopped receiving messages")

	return nil
}

func (s *NotificationService) serveHeap(ctx context.Context) {
	// step 1. Peek
	// step 2. Pop
	// step 3. Send

	ticker := time.NewTicker(s.checkPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			s.heapMutex.Lock()

			for s.notificationHeap.Len() > 0 {
				notificationToPublish := s.notificationHeap.Peek()
				if notificationToPublish == nil {
					break
				}

				if publicationTime := notificationToPublish.ScheduledAt; publicationTime.Add(s.checkPeriod).After(now) {
					break
				}

				notification := heap.Pop(s.notificationHeap).(*model.Notification)
				s.heapMutex.Unlock()

				if err := s.sendNotification(ctx, notification); err != nil {
					zlog.Logger.Error().
						Err(err).
						Str("id", notification.ID.String()).
						Time("scheduled_at", notification.ScheduledAt).
						Msg("failed to send notification")
				} else {
					zlog.Logger.Info().Msg("success send notification")
				}
				s.heapMutex.Lock()
			}
			s.heapMutex.Unlock()
		}
	}

}

func (s *NotificationService) sendNotification(ctx context.Context, notification *model.Notification) error {
	// надо добавить провекру, что такой канал есть в мапе
	err := s.channelToSender.Send(ctx, notification)
	if err != nil {
		zlog.Logger.Error().
			Err(err).
			Str("id", notification.ID.String()).
			Str("channel", string(notification.Channel.String())).
			Msg("failed to send notification via sender")
		return fmt.Errorf("do not send %w", err)
	}

	return nil

}
