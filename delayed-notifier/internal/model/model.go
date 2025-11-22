package model

import (
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/internaltypes"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/types"
)

type Notification struct {
	ID          *types.UUID                            `json:"id" db:"id"`                           // PRIMARY KEY,
	Recipient   internaltypes.Recipient              `json:"recipient" db:"recipient"`             // email, telegram id и т.д.
	Channel     internaltypes.NotificationChannel `json:"channel" db:"channel"`                 // email, telegram
	Message     string                            `json:"message" db:"message"`                 // текст уведомления
	ScheduledAt time.Time                         `json:"scheduled_at" db:"scheduled_at"`       // время отправки
	Status      string                            `json:"status" db:"status"`                   // pending / sent / cancelled / failed
	Tries       int                               `json:"tries" db:"tries"`                     // количество попыток отправки
	LastError   *string                           `json:"last_error,omitempty" db:"last_error"` // текст последней ошибки (может быть NULL)
}

// type NotificationFromDB struct {
// 	Recipient   string    `json:"recipient" db:"recipient"`         // email, telegram id и т.д.
// 	Channel     string    `json:"channel" db:"channel"`             // email, telegram
// 	Message     string    `json:"message" db:"message"`             // текст уведомления
// 	ScheduledAt time.Time `json:"scheduled_at" db:"scheduled_at"`   // время отправки
// 	Status      string    `json:"status" db:"status"`               // pending / sent / cancelled / failed
// 	Tries       int       `json:"tries" db:"tries"`                 // количество попыток отправки
// 	LastError   *string   `json:"last_error,omitempty" db:"last_error"` // текст последней ошибки (может быть NULL)
// 	CreatedAt   time.Time `json:"created_at" db:"created_at"`
// 	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
// }
