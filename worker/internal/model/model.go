package model

import "time"


type Notification struct {
	ID          int64     `json:"id" db:"id"`                     // PRIMARY KEY, BIGSERIAL
	Recipient   string    `json:"recipient" db:"recipient"`       // email, telegram id и т.д.
	Channel     string    `json:"channel" db:"channel"`           // email, telegram
	Message     string    `json:"message" db:"message"`           // текст уведомления
	ScheduledAt time.Time `json:"scheduled_at" db:"scheduled_at"` // время отправки
	Status      string    `json:"status" db:"status"`             // pending / sent / cancelled / failed
	Tries       int       `json:"tries" db:"tries"`               // количество попыток отправки
	LastError   *string   `json:"last_error,omitempty" db:"last_error"` // текст последней ошибки (может быть NULL)
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