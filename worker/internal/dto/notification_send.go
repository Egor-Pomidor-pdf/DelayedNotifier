package dto

import (
	"fmt"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/model"
)

type NotificationSendBody struct {
	ID          int  `json:"id" db:"id"`                           // PRIMARY KEY, BIGSERIAL
	Recipient   string  `json:"recipient" db:"recipient"`             // email, telegram id и т.д.
	Channel     string  `json:"channel" db:"channel"`                 // email, telegram
	Message     string  `json:"message" db:"message"`                 // текст уведомления
	ScheduledAt string  `json:"scheduled_at" db:"scheduled_at"`       // время отправки в формате RFC3339
	Status      string  `json:"status" db:"status"`                   // pending / sent / cancelled / failed
	Tries       int  `json:"tries" db:"tries"`                     // количество попыток отправки
	LastError   *string `json:"last_error,omitempty" db:"last_error"` // текст последней ошибки (может быть NULL)
}

func NotificationModelFromSendDTO(dto *NotificationSendBody) (*model.Notification, error) {
	// ID
	

	// ScheduledAt
	scheduledAt, err := time.Parse(time.RFC3339, dto.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled_at: %w", err)
	}

	// Tries


	return &model.Notification{
		ID:          int64(dto.ID),
		Recipient:   dto.Recipient,
		Channel:     dto.Channel,
		Message:     dto.Message,
		ScheduledAt: scheduledAt,
		Status:      dto.Status,
		Tries:       dto.Tries,
		LastError:   dto.LastError,
	}, nil
}
