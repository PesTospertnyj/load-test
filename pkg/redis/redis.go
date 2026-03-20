package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

func NewClient(ctx context.Context, addr, password string, db int) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info("Redis connection established successfully")
	return client, nil
}
