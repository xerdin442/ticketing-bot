package service

import (
	"github.com/xerdin442/ticketing-bot/internal/cache"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
)

type Manager struct {
	Message *MessageService
}

func NewManager(s *secrets.Secrets, c *cache.Cache) *Manager {
	return &Manager{
		Message: NewMessageService(s, c),
	}
}
