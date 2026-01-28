package handlers

import (
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/xerdin442/ticketing-bot/internal/service"
)

type RouteHandler struct {
	services   *service.Manager
	cache      *redis.Client
	tasksQueue *asynq.Client
}

func New(svc *service.Manager, r *redis.Client, q *asynq.Client) *RouteHandler {
	return &RouteHandler{services: svc, cache: r, tasksQueue: q}
}
