package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
	"github.com/xerdin442/ticketing-bot/internal/tasks"
)

func main() {
	// Load environment variables
	env := secrets.Load()

	// Initialize cache
	cacheOpts, err := redis.ParseURL(env.RedisUri)
	if err != nil {
		log.Fatal().Msg("Invalid Redis connection URL")
	}

	cache := redis.NewClient(cacheOpts)

	// Improve readability of the logs in development
	if env.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	queueOpts, err := asynq.ParseRedisURI(env.RedisUri)
	if err != nil {
		log.Fatal().Msg("Invalid Redis connection URL")
	}

	// Initialize the worker server
	srv := asynq.NewServer(
		queueOpts,
		asynq.Config{Concurrency: 10},
	)

	// Define tasks handlers
	h := tasks.NewHandler(env, cache)

	mux := asynq.NewServeMux()
	mux.HandleFunc("payment_queue", h.HandlePaymentWebhookTask)

	// Start the worker server
	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatal().Err(err).Msg("Worker initialization failed")
		}
	}()

	log.Info().Msg("Task worker is running...")

	// Keep the server running unless interrupted
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// Graceful shutdown
	srv.Shutdown()
	log.Warn().Msg("Shutdown complete.")
}
