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
	var object *model.Notification

	go s.serveHeap(ctx)

out:
	for {
		select {
		case <-ctx.Done():
			break out
		case object = <-objects:
			// надо добавить провекру, что такой канал есть в мапе
			s.heapMutex.Lock()
			heap.Push(s.notificationHeap, object)
			s.heapMutex.Unlock()
		}
	}
	return s.receiver.StopReceiving()
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
					fmt.Println("error", err)
				} else {
					fmt.Println("success")
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
		return fmt.Errorf("do not send %w", err)
	}
	return nil

}
