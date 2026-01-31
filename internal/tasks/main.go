package tasks

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
	"github.com/xerdin442/ticketing-bot/internal/service"
)

type TasksClient interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type TaskHandler struct {
	env    *secrets.Secrets
	cache  *redis.Client
	gemini *service.GeminiService
}

func NewHandler(s *secrets.Secrets, r *redis.Client) *TaskHandler {
	return &TaskHandler{
		env:    s,
		cache:  r,
		gemini: service.NewGeminiService(s, r),
	}
}
