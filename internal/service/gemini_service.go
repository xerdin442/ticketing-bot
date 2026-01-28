package service

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/xerdin442/ticketing-bot/internal/api/dto"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
	"github.com/xerdin442/ticketing-bot/internal/util"
	"google.golang.org/genai"
)

type GeminiService struct {
	env    *secrets.Secrets
	cache  *redis.Client
	client *genai.Client
}

func NewGeminiService(s *secrets.Secrets, r *redis.Client) *GeminiService {
	client, _ := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  s.GeminiApiKey,
		Backend: genai.BackendGeminiAPI,
	})

	return &GeminiService{
		env:    s,
		cache:  r,
		client: client,
	}
}

func (s *GeminiService) GetNextStateAfterFunctionCall(funcName string) (dto.ConversationState, error) {
	switch {
	case strings.Contains(funcName, "find"):
		return dto.StateEventQuery, nil
	case strings.Contains(funcName, "select_event"):
		return dto.StateEventSelected, nil
	case strings.Contains(funcName, "tier"):
		return dto.StateTicketTierSelected, nil
	case strings.Contains(funcName, "initiate"):
		return dto.StateAwaitingPayment, nil
	default:
		return dto.StateResponseError, util.ErrInvalidFunctionName
	}
}
