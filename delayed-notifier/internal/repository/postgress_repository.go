package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/internaltypes"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/types"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type StoreRepository struct {
	db       *dbpg.DB
	strategy retry.Strategy
}

func NewRepository(db *dbpg.DB, strategy retry.Strategy) *StoreRepository {
	return &StoreRepository{
		db: db,
		strategy: strategy,

	}
}

func (r *StoreRepository) CreateNotify(ctx context.Context, notify model.Notification) error {
	query := `INSERT INTO notifications (id, recipient, channel, message, scheduled_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecWithRetry(
		ctx,
		r.strategy,
		query,
		notify.ID.String(),
		notify.Recipient.String(),
		notify.Channel.String(),
		notify.Message,
		notify.ScheduledAt.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *StoreRepository) GetNotify(ctx context.Context, id types.UUID) (*model.Notification, error) {
	query := `SELECT * FROM notifications
			  WHERE id = :id`

	var (
		recipient   string
		channel     string
		message     string
		scheduledAt time.Time
		status      string
		tries       int
		lastError   *string
	)

	rows, err := r.db.QueryRowWithRetry(ctx, r.strategy, query, id.String())
	if err != nil {
		return nil, fmt.Errorf("error select by id in postgres: %w", err)
	}

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

	var channelValid internaltypes.NotificationChannel
	channelValid, err = internaltypes.NotificationChannelFromString(channel)
	if err != nil {
		return nil, fmt.Errorf("invalid channel in postgres: %w", err)
	}

	var recipientToValid internaltypes.Recipient
	recipientToValid, err = internaltypes.NewSendTo(types.NewAnyText(recipient), channelValid)
	if err != nil {
		return nil, fmt.Errorf("invalid send_to in postgres: %w", err)
	}

	return &model.Notification{
		Recipient:   recipientToValid,
		Channel:     channelValid,
		Message:     message,
		ScheduledAt: scheduledAt,
		Status:      status,
		Tries:       tries,
		LastError:   lastError,
	}, nil
}

func (r *StoreRepository) FetchFromDb(ctx context.Context, needToSendTime time.Time) ([]*model.Notification, error) {
	query := `
    SELECT id, recipient, channel, message, scheduled_at, status, tries, last_error
    FROM notifications
    WHERE scheduled_at <= $1 AND status = 'pending' AND tries <= 3
    ORDER BY scheduled_at
`
	rows, err := r.db.QueryWithRetry(ctx, r.strategy, query, needToSendTime)

	if err != nil {
		return nil, fmt.Errorf("error fetch %w", err)
	}
	if rows == nil {
		zlog.Logger.Warn().Msg("QueryWithRetry returned nil rows, returning empty result")
		return []*model.Notification{}, nil
	}
	defer rows.Close()

	result := []*model.Notification{}

	for rows.Next() {
		var (
			id          string
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

		var channelValid internaltypes.NotificationChannel
		channelValid, err = internaltypes.NotificationChannelFromString(channel)
		if err != nil {
			zlog.Logger.Error().Err(fmt.Errorf("invalid channel in postgres: %w", err))
			continue
		}

		var recipientToValid internaltypes.Recipient
		recipientToValid, err = internaltypes.NewSendTo(types.NewAnyText(recipient), channelValid)
		if err != nil {
			return nil, fmt.Errorf("invalid send_to in postgres: %w", err)
		}

		var UUID types.UUID
		UUID, err = types.NewUUID(id)
		if err != nil {
			return nil, err
		}

		result = append(result, &model.Notification{
			ID:          &UUID,
			Recipient:   recipientToValid,
			Channel:     channelValid,
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

func (r *StoreRepository) UpdateNotiy(ctx context.Context, newNotify model.Notification) error {
	return nil
}

func (r *StoreRepository) MarkAsSent(ctx context.Context, id int) error {
	return nil
}

// func (r *StoreRepository) DeleteNotify(ctx context.Context, id int) error {
// 	query := `DELETE FROM notifications
// 			  WHERE id = :id`

// 	params := map[string]interface{}{"id": id}
// 	res, err := r.db.NamedExecContext(ctx, query, params)
// 	if err != nil {
// 		return err
// 	}

// 	// Проверяем, что хотя бы одна строка была затронута
// 	affected, err := res.RowsAffected()
// 	if err != nil {
// 		return err
// 	}
// 	if affected == 0 {
// 		return fmt.Errorf("notification with id %d not found", id)
// 	}

// 	return nil
// }
