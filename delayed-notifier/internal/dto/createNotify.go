package dto

// import (
// 	"time"

// 	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
// )

// type NotificationToCreate struct {
// 	Recipient   string `json:"recipient" db:"recipient"`       // email, telegram id и т.д.
// 	Channel     string `json:"channel" db:"channel"`           // email, telegram
// 	Message     string `json:"message" db:"message"`           // текст уведомления
// 	ScheduledAt string `json:"scheduled_at" db:"scheduled_at"` // время отправки
// }

// func (s NotificationToCreate) ToEntity() (*model.Notification, error) {
// 	ShAt, _ := time.Parse("2006-01-02 15:04:05", s.ScheduledAt)
// 	return &model.Notification{
// 		Recipient:   s.Recipient,
// 		Channel:     s.Channel,
// 		Message:     s.Message,
// 		ScheduledAt: ShAt,
// 		Status:      "pending",
// 		Tries:       0,
// 		LastError:   nil,
// 	}, nil
// }
