package main

import (
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xerdin442/ticketing-bot/internal/cache"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
	"github.com/xerdin442/ticketing-bot/internal/service"
)

type application struct {
	port       int
	env        *secrets.Secrets
	tasksQueue *asynq.Client
	cache      *cache.Cache
	services   *service.Manager
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
	cache := cache.New(env)
	svc := service.NewManager(env, cache)

	// Initialize task queue
	tasksQueue := asynq.NewClient(
		asynq.RedisClientOpt{
			Addr:     env.RedisAddr,
			Password: env.RedisPassword,
		},
	)

	app := &application{
		port:       env.Port,
		cache:      cache,
		tasksQueue: tasksQueue,
		env:        env,
		services:   svc,
	}

	// Start the http server
	if err := app.serve(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
