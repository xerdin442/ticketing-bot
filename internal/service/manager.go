package service

import (
	"github.com/redis/go-redis/v9"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
)

type Manager struct {
	Message *MessageService
}

func NewManager(s *secrets.Secrets, r *redis.Client) *Manager {
	return &Manager{
		Message: NewMessageService(s, r),
	}
}
