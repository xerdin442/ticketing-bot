package main

import (
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xerdin442/ticketing-bot/internal/api/handlers"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
	"github.com/xerdin442/ticketing-bot/internal/service"
)

type application struct {
	port int
	handlers.Base
}

func main() {
	// Initialize logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Load environment variables
	env := secrets.Load()

	// Improve readability of the logs in development
	if env.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	// Initialize cache and services
	cache := redis.NewClient(&redis.Options{
		Addr:     env.RedisAddr,
		Password: env.RedisPassword,
	})
	svc := service.NewManager(env, cache)

	// Initialize task queue
	tasksQueue := asynq.NewClient(
		asynq.RedisClientOpt{
			Addr:     env.RedisAddr,
			Password: env.RedisPassword,
		},
	)

	app := &application{
		port: env.Port,
		Base: handlers.Base{
			Env:        env,
			Cache:      cache,
			TasksQueue: tasksQueue,
			Services:   svc,
		},
	}

	// Start the http server
	if err := app.serve(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
