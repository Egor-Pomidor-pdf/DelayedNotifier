package model

import "time"

type CreateNotifyRequest struct {
    Recipient   string    `json:"recipient" db:"recipient"`    // email, telegram id и т.д.
    Channel     string    `json:"channel" db:"channel"`      // email, telegram
    Message     string    `json:"message" db:"message"`      // текст уведомления
    ScheduledAt time.Time `json:"scheduled_at" db:"scheduled_at"` // время отправки
}