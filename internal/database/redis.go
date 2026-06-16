package database

import (
	"context"
	"fmt"

	"github.com/kodokbakar/pylon/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("redis url is required")
	}

	options, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse redis config: %w", err)
	}

	if cfg.Password != "" {
		options.Password = cfg.Password
	}

	options.DB = cfg.DB

	client := redis.NewClient(options)

	if err := client.Ping(ctx).Err(); err != nil {
		if closeErr := client.Close(); closeErr != nil {
			return nil, fmt.Errorf("ping redis: %w; close redis client: %v", err, closeErr)
		}
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
