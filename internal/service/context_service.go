package service

import (
	"github.com/xerdin442/ticketing-bot/internal/cache"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
)

type ContextService struct {
	env   *secrets.Secrets
	cache *cache.Cache
}

func NewContextService(s *secrets.Secrets, c *cache.Cache) *ContextService {
	return &ContextService{env: s, cache: c}
}
