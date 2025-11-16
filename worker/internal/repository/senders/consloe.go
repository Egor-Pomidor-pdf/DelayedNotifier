package senders

import (
	"context"
	"fmt"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/model"
)

type ConsoleSender struct{}

func NewConsoleSender() *ConsoleSender {
	return &ConsoleSender{}
}

func (s *ConsoleSender) Send(ctx context.Context, notification *model.Notification) error {
	fmt.Printf(
		"Sending notification ID=%v to recipient=%s via channel=%s: %s\n",
		notification.ID,
		notification.Recipient,
		notification.Channel,
		notification.Message,
	)
	return nil
}