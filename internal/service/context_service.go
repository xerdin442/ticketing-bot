package service

import (
	"github.com/redis/go-redis/v9"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
)

type ContextService struct {
	env   *secrets.Secrets
	cache *redis.Client
}

func NewContextService(s *secrets.Secrets, r *redis.Client) *ContextService {
	return &ContextService{env: s, cache: r}
}
