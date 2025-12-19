package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
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
		db:       db,
		strategy: strategy,
	}
}

func (r *StoreRepository) CreateNotify(ctx context.Context, notify *model.Notification) error {
	query := `INSERT INTO notifier_db.public.notifications (id, recipient, channel, message, scheduled_at)
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
	query := `SELECT * FROM notifier_db.public.notifications
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
		return nil,fmt.Errorf("error scan in get: %w", err)
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

func (r *StoreRepository) GetAllNotifies(ctx context.Context) ([]*model.Notification, error) {
	query := `SELECT 
                id,
                recipient,
                channel,
                message,
                scheduled_at,
                status,
                tries,
                last_error
              FROM notifier_db.public.notifications
              ORDER BY scheduled_at DESC`

	rows, err := r.db.QueryWithRetry(ctx, r.strategy, query)
	if err != nil {
		return nil, fmt.Errorf("error selecting all notifications from postgres: %w", err)
	}
	defer rows.Close()

	var result []*model.Notification

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
			return nil, fmt.Errorf("error scan in GetAllNotifies: %w", err)
		}

		uuid, _ := types.NewUUID(id)


		channelValid, _ := internaltypes.NotificationChannelFromString(channel)
		

		recipientValid, _ := internaltypes.NewSendTo(types.NewAnyText(recipient), channelValid)
		

		result = append(result, &model.Notification{
			ID:          &uuid,
			Recipient:   recipientValid,
			Channel:     channelValid,
			Message:     message,
			ScheduledAt: scheduledAt,
			Status:      status,
			Tries:       tries,
			LastError:   lastError,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error in GetAllNotifies: %w", err)
	}

	return result, nil
}


func (r *StoreRepository) FetchFromDb(ctx context.Context, needToSendTime time.Time) ([]*model.Notification, error) {
	query := `
    SELECT id, recipient, channel, message, scheduled_at, status, tries, last_error
    FROM notifier_db.public.notifications
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

	defer func(rows *sql.Rows) {
		if err = rows.Close(); err != nil {
			zlog.Logger.Error().Err(err).Msg("couldn't close postgres rows when fetching")
		}
	}(rows)

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
		recipientToValid = internaltypes.RecipientFromString(recipient)

		var UUID types.UUID
		UUID, err = types.NewUUID(id)
		if err != nil {
			zlog.Logger.Error().Err(fmt.Errorf("invalid uuid in postgres: %w", err))
			continue
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

func (r *StoreRepository) UpdateNotification(ctx context.Context, n *model.Notification) error {
    // SQL-запрос на обновление записи по id
    query := `
        UPDATE notifier_db.public.notifications
        SET recipient = $1,
            channel = $2,
            message = $3,
            scheduled_at = $4,
            status = $5,
            tries = $6,
            last_error = $7,
            updated_at = now()
        WHERE id = $8
    `

    // Выполняем запрос
    res, err := r.db.ExecWithRetry(
        ctx,
        r.strategy,
        query,
        n.Recipient.String(),  // преобразуем в строку
        n.Channel.String(),
        n.Message,
        n.ScheduledAt,
        n.Status,
        n.Tries,
        n.LastError,
        n.ID.String(),
    )
    if err != nil {
        return fmt.Errorf("failed to update notification: %w", err)
    }

    // Проверяем, сколько строк реально было обновлено
    rowsAffected, err := res.RowsAffected()
    if err != nil {
        return fmt.Errorf("couldn't get number of rows affected: %w", err)
    }

    // Если запись с таким ID не найдена
    if rowsAffected == 0 {
        return fmt.Errorf("notification not found")
    }

    return nil
}

func (r *StoreRepository) MarkAsSent(ctx context.Context, ids []*types.UUID) error {
	if len(ids) == 0 {
        // Нечего обновлять — выходим без ошибки
        return nil
    }


	idNumsList := make([]string, len(ids))
	for i := range idNumsList {
		idNumsList[i] = "$" + strconv.Itoa(i+1)
	}
	query := fmt.Sprintf(`UPDATE notifier_db.public.notifications SET status = 'sent' where id IN (%s)`, strings.Join(idNumsList, ","))

	idStrsList := make([]any, len(ids))
	for i, v := range ids {
		idStrsList[i] = v.String()
	}

	_, err := r.db.ExecWithRetry(ctx, r.strategy, query, idStrsList...)
	if err != nil {
		return fmt.Errorf("error marking %d notifications as sent: %w", len(ids), err)
	}

	return nil
}

func (r *StoreRepository) DeleteNotification(ctx context.Context, id types.UUID) error {
	// SQL-запрос на удаление по id
	query := `DELETE FROM notifier_db.public.notifications WHERE id = $1`

	// Выполняем запрос через ExecWithRetry
	res, err := r.db.ExecWithRetry(ctx, r.strategy, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	// Проверяем, сколько строк реально было удалено
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("couldn't get number of rows affected: %w", err)
	}

	// Если запись с таким ID не найдена
	if rowsAffected == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}