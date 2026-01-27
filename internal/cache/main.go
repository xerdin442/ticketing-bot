package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
)

type Cache struct {
	Redis *redis.Client
}

func New(s *secrets.Secrets) *Cache {
	client := redis.NewClient(&redis.Options{
		Addr:     s.RedisAddr,
		Password: s.RedisPassword,
	})

	return &Cache{Redis: client}
}

func (c *Cache) Set(ctx context.Context, key, value string, exp time.Time) error {
	return c.Redis.Set(ctx, key, value, time.Until(exp)).Err()
}
