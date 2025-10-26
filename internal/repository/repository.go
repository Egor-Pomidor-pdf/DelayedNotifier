package repository

import (
	"context"
	"fmt"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/model"
	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db *sqlx.DB
}

type OrderRepositoryEnterface interface {
	CreateNotify(ctx context.Context, notify model.CreateNotifyRequest) (int, error)
	GetNotifyStatus(ctx context.Context, id int) (string, error)
	DeleteNotify(ctx context.Context, id int) error
}

func (r *OrderRepository) CreateNotify(ctx context.Context, notify model.CreateNotifyRequest) (int, error) {
	const op = "repository.CreateNotify"

	var id int
	query := `INSERT INTO notifications (recipient, channel, message, scheduled_at)
			  VALUES (:recipient, :channel, :message, :scheduled_at)
			  RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, query, notify)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (r *OrderRepository) GetNotifyStatus(ctx context.Context, id int) (string, error) {
	const op = "repository.GetNotifyStatus"
	var status string
	query := `SELECT status FROM notifications
			  WHERE id = :id`

	params := map[string]interface{}{"id": id}
	rows, err := r.db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&status)
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("notification not found")
	}
	return status, nil
}

func (r *OrderRepository) DeleteNotify(ctx context.Context, id int) error {
	const op = "repository.DeleteNotify"

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
