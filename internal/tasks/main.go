package tasks

import (
	"github.com/hibiken/asynq"
	"github.com/xerdin442/ticketing-bot/internal/cache"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
	"github.com/xerdin442/ticketing-bot/internal/service"
)

type TasksClient interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type TaskHandler struct {
	env    *secrets.Secrets
	gemini *service.GeminiService
}

func NewHandler(s *secrets.Secrets, c *cache.Cache) *TaskHandler {
	return &TaskHandler{
		env:    s,
		gemini: service.NewGeminiService(s, c),
	}
}
