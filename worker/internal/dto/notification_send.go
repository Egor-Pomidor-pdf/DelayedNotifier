package dto

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/internaltypes"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/worker/pkg/types"

)

type NotificationToSend struct {
	ID          string `json:"id"`
	Recipient   string `json:"recipient" db:"recipient"`       // email, telegram id и т.д.
	Channel     string `json:"channel" db:"channel"`           // email, telegram
	Message     string `json:"message" db:"message"`           // текст уведомления
	ScheduledAt string `json:"scheduled_at" db:"scheduled_at"` // время отправки
}

func ToModelFromSend(send []byte) (*model.Notification, error) {
	obj, err := ToDTOFromSend(send)
	if err != nil {
		return nil, fmt.Errorf("invalid notification json: %w", err)
	}
	
	uuid, err := types.NewUUID(obj.ID)
	if err != nil {
		return nil, fmt.Errorf("novalid id in dto: %w", err)
	}
	
	rec := internaltypes.RecipientFromString(obj.Recipient)
	ch, err := internaltypes.NotificationChannelFromString(obj.Channel)
	if err != nil {
		return nil, fmt.Errorf("novalid Channel in dto: %w", err)
	}
	shAt, err := time.Parse("2006-01-02T15:04:05Z07:00", obj.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("novalid ScheduledAt in dto: %w", err)
	}
	return &model.Notification{
		ID:          &uuid,
		Recipient:   rec,
		Channel:     ch,
		Message:     obj.Message,
		ScheduledAt: shAt, // ISO8601
	}, nil
}


func ToDTOFromSend(send []byte) (*NotificationToSend, error) {
	var NotifySend NotificationToSend
	err := json.Unmarshal(send, &NotifySend)
	if err != nil {
		return nil, fmt.Errorf("could not marshal NotificationToSend: %w", err)
	}
	return &NotifySend, nil
}
