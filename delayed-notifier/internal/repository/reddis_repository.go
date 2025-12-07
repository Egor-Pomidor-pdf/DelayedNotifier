package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/internal/model"
	"github.com/Egor-Pomidor-pdf/DelayedNotifier/delayed-notifier/pkg/types"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type RedisRepository struct {
	redisClient   *redis.Client
	retryStrategy retry.Strategy
	expiration    time.Duration
}

func NewRedisRepository(redisClient *redis.Client, retryStrategy retry.Strategy, expiration time.Duration) *RedisRepository {
	return &RedisRepository{redisClient: redisClient, retryStrategy: retryStrategy, expiration: expiration}
}

func (r *RedisRepository) SaveNotification(ctx context.Context, notify *model.Notification) error {
	var err error
	key := notify.ID.String()
	data, err := json.Marshal(notify)
	if err != nil {
		return fmt.Errorf("redis: marshal notification: %w", err)
	}
	err = r.redisClient.SetWithExpiration(ctx, key, data, r.expiration)
	if err != nil {
		return fmt.Errorf("redis: set key %s: %w", key, err)
	}
	return nil
}

func (r *RedisRepository) GetNotification(ctx context.Context, id types.UUID) (*model.Notification, error) {
	key := id.String()

	data, err := r.redisClient.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	var notification model.Notification

	if err = json.Unmarshal([]byte(data), &notification); err != nil {
		return nil, fmt.Errorf("redis: marshal notification: %w", err)
	}
	return &notification, nil
}

func (r *RedisRepository) DeleteNotification(ctx context.Context, id types.UUID) error {
	key := id.String()
	err := r.redisClient.Del(ctx, key)
	if err != nil {
		return fmt.Errorf("error deleting from redis notification (id '%s'): %w", id, err)
	}
	return nil
}
