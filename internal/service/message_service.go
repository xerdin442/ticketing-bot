package service

import (
	"github.com/redis/go-redis/v9"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
)

type MessageService struct {
	env     *secrets.Secrets
	context *ContextService
	gemini  *GeminiService
}

func NewMessageService(s *secrets.Secrets, r *redis.Client) *MessageService {
	return &MessageService{
		env:     s,
		context: NewContextService(s, r),
		gemini:  NewGeminiService(s, r),
	}
}
