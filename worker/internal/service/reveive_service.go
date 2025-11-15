package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/config"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/ports"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/repository"
	"github.com/rabbitmq/amqp091-go"
)

type ReceiveService struct {
	consumerRepository ports.ConsumerRpositoryInterface
	senderRepository   ports.SenderRepositoryInterface
}

func NewRabbitService(consumerRepository *repository.RabbitRepository, senderRepository *repository.SenderRepository) *ReceiveService {
	return &ReceiveService{consumerRepository: consumerRepository, senderRepository: senderRepository}
}

func (s *ReceiveService) Start(ctx context.Context, rabbitCfg config.RabbitMQConfig) error {
	msgs, err := s.consumerRepository.СonsumeMsg(ctx)
	if err != nil {
		return fmt.Errorf("cannot start consumer: %w", err)

	}
	workerCtx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.worker(workerCtx, msgs, rabbitCfg)
	}()

	select {
	case <-ctx.Done():
		cancel()
		wg.Wait()
		return ctx.Err()
	case _, ok := <-msgs:
		if !ok {
			cancel()
			wg.Wait()
			return errors.New("message channel closed unexpectedly")
		}
	}
	wg.Wait()
	return nil
}


func (s * ReceiveService) worker(ctx context.Context, msgs <-chan amqp091.Delivery, rabbitCfg config.RabbitMQConfig) {
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
