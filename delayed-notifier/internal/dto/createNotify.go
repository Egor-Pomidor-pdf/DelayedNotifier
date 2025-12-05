package dto

import (
	"encoding/json"
	"fmt"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
)

type NotificationToSend struct {
	ID          string `json:"id"`
	Recipient   string `json:"recipient" db:"recipient"`       // email, telegram id и т.д.
	Channel     string `json:"channel" db:"channel"`           // email, telegram
	Message     string `json:"message" db:"message"`           // текст уведомления
	ScheduledAt string `json:"scheduled_at" db:"scheduled_at"` // время отправки
}

func ToDTOFromModel(obj *model.Notification) *NotificationToSend {
	return &NotificationToSend{
		ID:          obj.ID.String(),
		Recipient:   obj.Recipient.String(),
		Channel:     obj.Channel.String(),
		Message:     obj.Message,
		ScheduledAt: obj.ScheduledAt.Format("2006-01-02T15:04:05Z07:00"), // ISO8601
	}
}


func ToSendFromDTO(obj *model.Notification) ([]byte, error) {
	data, err := json.Marshal(ToDTOFromModel(obj))
	if err != nil {
		return nil, fmt.Errorf("could not marshal NotificationToSend: %w", err)
	}
	return data, nil
}

