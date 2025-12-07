package dto

import (

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
)

type NotificationFull struct {
	ID          string `json:"id"`
	Recipient   string `json:"recipient"`
	Channel     string `json:"channel"`
	Message     string `json:"message"`
	ScheduledAt string `json:"scheduled_at"`
	Status      string `json:"status"`
	Tries       string `json:"tries"`
	LastError   string `json:"last_error"`
}

func ToFullFromModelNotification(notify *model.Notification) *NotificationFull {
	return &NotificationFull{
		ID: notify.ID.String(),
		Recipient: notify.Recipient.String(),
		Channel: notify.Channel.String(),
		ScheduledAt: notify.ScheduledAt.String(),
		Status: notify.Status,
		Message: notify.Message,
		// Tries: strconv.Itoa(notify.Tries),
		// LastError: *notify.LastError,
	}
}