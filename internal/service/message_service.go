package service

import (
	"github.com/xerdin442/ticketing-bot/internal/cache"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
)

type MessageService struct {
	env     *secrets.Secrets
	context *ContextService
	gemini  *GeminiService
}

func NewMessageService(s *secrets.Secrets, c *cache.Cache) *MessageService {
	return &MessageService{
		env:     s,
		context: NewContextService(s, c),
		gemini:  NewGeminiService(s, c),
	}
}
