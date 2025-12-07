package dto

import (
	"fmt"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/internaltypes"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
)

type NotificationCreate struct {
	Recipient   string `json:"recipient" db:"recipient"`       // email, telegram id и т.д.
	Channel     string `json:"channel" db:"channel"`           // email, telegram
	Message     string `json:"message" db:"message"`           // текст уведомления
	ScheduledAt string `json:"scheduled_at" db:"scheduled_at"` // время отправки
}

func (b NotificationCreate) ToEnity() (*model.Notification, error) {
	var err error
	rec := internaltypes.RecipientFromString(b.Recipient)
	if err != nil {
		return nil, fmt.Errorf("incorrect 'recipment' '%s': %w", b.ScheduledAt, err)
	}

	var channel internaltypes.NotificationChannel
	channel, err = internaltypes.NotificationChannelFromString(b.Channel)
	if err != nil {
		return nil, fmt.Errorf("incorrect 'channel' '%s': %w", b.Channel, err)
	}
	shedAt, err := time.Parse(time.RFC3339, b.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("incorrect 'scheduled_at' '%s': %w", b.ScheduledAt, err)
	}

	return &model.Notification{
		Recipient:   rec,
		Channel:     channel,
		Message:     b.Message,
		ScheduledAt: shedAt,
	}, nil

}
