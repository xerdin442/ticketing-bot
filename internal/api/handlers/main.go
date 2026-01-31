package handlers

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
	"github.com/xerdin442/ticketing-bot/internal/service"
)

type Base struct {
	Services   *service.Manager
	Env        *secrets.Secrets
	Cache      *redis.Client
	TasksQueue *asynq.Client
}

type RouteHandler struct {
	Base
}

func New(b Base) *RouteHandler {
	return &RouteHandler{Base: b}
}
