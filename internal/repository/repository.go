package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/model"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateNotify(ctx context.Context, notify model.Notification) error {
	query := `INSERT INTO notifications (recipient, channel, message, scheduled_at)
			  VALUES (:recipient, :channel, :message, :scheduled_at)`

	_, err := r.db.NamedExecContext(ctx, query, notify)
	if err != nil {
		return err
	}

	return nil

}
func (r *Repository) GetNotify(ctx context.Context, id int) (*model.Notification, error) {
	var (
		recipient   string
		channel     string
		message     string
		scheduledAt time.Time
		status      string
		tries       int
		lastError   *string
	)
	query := `SELECT * FROM notifications
			  WHERE id = :id`

	params := map[string]interface{}{"id": id}
	rows, err := r.db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(
			&recipient,
			&channel,
			&message,
			&scheduledAt,
			&status,
			&tries,
			&lastError,
		)
		if err != nil {
			return nil, err
		}
		return &model.Notification{
			Recipient:   recipient,
			Channel:     channel,
			Message:     message,
			ScheduledAt: scheduledAt,
			Status:      status,
			Tries:       tries,
			LastError:   lastError,
		}, nil
	} else {
		return nil, fmt.Errorf("notification not found")
	}
}

func (r *Repository) FetchFromDb(ctx context.Context, needToSendTime time.Time) ([]*model.Notification, error) {
	query := `
		SELECT id, recipient, channel, message, scheduled_at, status, tries, last_error
		FROM notifications
		WHERE scheduled_at <= :needToSendTime AND status = 'pending' AND tries <= 3
		ORDER BY scheduled_at
	`
	params := map[string]interface{}{"needToSendTime": needToSendTime}
	rows, err := r.db.NamedQueryContext(ctx, query, params)

	if err != nil {
		return nil, fmt.Errorf("error fetch", err)
	}
	defer rows.Close()

	result := []*model.Notification{}

	for rows.Next() {
		var (
			id          int64
			recipient   string
			channel     string
			message     string
			scheduledAt time.Time
			status      string
			tries       int
			lastError   *string
		)

		if err := rows.Scan(
			&id,
			&recipient,
			&channel,
			&message,
			&scheduledAt,
			&status,
			&tries,
			&lastError,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		result = append(result, &model.Notification{
			ID:          id,
			Recipient:   recipient,
			Channel:     channel,
			Message:     message,
			ScheduledAt: scheduledAt,
			Status:      status,
			Tries:       tries,
			LastError:   lastError,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning rows: %w", err)
	}

	return result, nil
}

// after relise
func (r *Repository) DeleteNotify(ctx context.Context, id int) error {
	query := `DELETE FROM notifications
			  WHERE id = :id`

	params := map[string]interface{}{"id": id}
	res, err := r.db.NamedExecContext(ctx, query, params)
	if err != nil {
		return err
	}

	// Проверяем, что хотя бы одна строка была затронута
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("notification with id %d not found", id)
	}

	return nil
}
