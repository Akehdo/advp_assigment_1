package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCacheRepository struct {
	client *redis.Client
}

func NewRedisCacheRepository(redisURL string) (*RedisCacheRepository, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	return &RedisCacheRepository{
		client: redis.NewClient(options),
	}, nil
}

func (r *RedisCacheRepository) Close() error {
	return r.client.Close()
}

func (r *RedisCacheRepository) Get(ctx context.Context, key string, dest any) (bool, error) {
	payload, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}

		return false, err
	}

	if err := json.Unmarshal(payload, dest); err != nil {
		return false, err
	}

	return true, nil
}

func (r *RedisCacheRepository) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, payload, ttl).Err()
}

func (r *RedisCacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
