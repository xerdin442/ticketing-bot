package middleware

import (
	"github.com/xerdin442/ticketing-bot/internal/cache"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
)

type Middleware struct {
	env   *secrets.Secrets
	cache *cache.Cache
}

func New(s *secrets.Secrets, c *cache.Cache) *Middleware {
	return &Middleware{env: s, cache: c}
}
