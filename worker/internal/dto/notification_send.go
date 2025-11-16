package dto

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/model"
)

type NotificationSendBody struct {
	ID          string  `json:"id" db:"id"`                           // PRIMARY KEY, BIGSERIAL
	Recipient   string  `json:"recipient" db:"recipient"`             // email, telegram id и т.д.
	Channel     string  `json:"channel" db:"channel"`                 // email, telegram
	Message     string  `json:"message" db:"message"`                 // текст уведомления
	ScheduledAt string  `json:"scheduled_at" db:"scheduled_at"`       // время отправки в формате RFC3339
	Status      string  `json:"status" db:"status"`                   // pending / sent / cancelled / failed
	Tries       string  `json:"tries" db:"tries"`                     // количество попыток отправки
	LastError   *string `json:"last_error,omitempty" db:"last_error"` // текст последней ошибки (может быть NULL)
}

func NotificationModelFromSendDTO(dto *NotificationSendBody) (*model.Notification, error) {
	// ID
	id, err := strconv.ParseInt(dto.ID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	// ScheduledAt
	scheduledAt, err := time.Parse("2006-01-02 15:04:05", dto.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("invalid scheduled_at: %w", err)
	}

	// Tries
	tries, err := strconv.Atoi(dto.Tries)
	if err != nil {
		return nil, fmt.Errorf("invalid tries: %w", err)
	}

	return &model.Notification{
		ID:          id,
		Recipient:   dto.Recipient,
		Channel:     dto.Channel,
		Message:     dto.Message,
		ScheduledAt: scheduledAt,
		Status:      dto.Status,
		Tries:       tries,
		LastError:   dto.LastError,
	}, nil
}
