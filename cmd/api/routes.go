package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xerdin442/ticketing-bot/internal/api/handlers"
	"github.com/xerdin442/ticketing-bot/internal/api/middleware"
)

func (app *application) routes() http.Handler {
	r := gin.New()
	m := middleware.New(app.Base)
	h := handlers.New(app.Base)

	r.Use(m.CustomRequestLogger())
	r.Use(m.RateLimiters()...)
	r.Use(gin.Recovery())

	message := r.Group("/messages")
	{
		message.GET("/webhook", h.VerifyWebhook)
		message.POST("/webhook", h.HandleIncomingMessage)
	}

	r.POST("/payments/callback", h.CheckPaymentStatus)

	return r
}
