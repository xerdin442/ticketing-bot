package middleware

import (
	"github.com/redis/go-redis/v9"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
)

type Middleware struct {
	env   *secrets.Secrets
	cache *redis.Client
}

func New(s *secrets.Secrets, r *redis.Client) *Middleware {
	return &Middleware{env: s, cache: r}
}
