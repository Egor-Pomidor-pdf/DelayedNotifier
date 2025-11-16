package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/config"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/ports"
)

type NotificationService struct {
	receiver ports.NotificationReceiver
	channelToSender   ports.NotificationSender
	checkPeriod time.Duration

}

func NewRabbitService(receiver ports.NotificationReceiver, channelToSender ports.NotificationSender, checkPeriod time.Duration) *NotificationService {
	return &NotificationService{receiver: receiver, channelToSender: channelToSender, checkPeriod: checkPeriod}
}

func (s *NotificationService) Run(ctx context.Context, rabbitCfg config.RabbitMQConfig) error {

	objects, err := s.receiver.StartReceiving(ctx)
	if err != nil {
		return fmt.Errorf("cannot start consumer: %w", err)

	}
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.worker(workerCtx, objects, rabbitCfg)
	}()

	select {
	case <-ctx.Done():
		cancel()
		wg.Wait()
		return ctx.Err()
	case _, ok := <-objects:
		if !ok {
			cancel()
			wg.Wait()
			return errors.New("message channel closed unexpectedly")
		}
	}
	wg.Wait()
	return nil
}


func (s * NotificationService) worker(ctx context.Context, msgs chan *model.Notification, rabbitCfg config.RabbitMQConfig) {
for {
	select {
	case <- ctx.Done():
		return
	case _, ok := <- msgs:
		if !ok {
			return
		} else {
			if rabbitCfg.AutoAck{
				//обработка и отправка сообщений
			}
		}

	}
}
}
