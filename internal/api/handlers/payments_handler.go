package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/xerdin442/ticketing-bot/internal/api/dto"
	"github.com/xerdin442/ticketing-bot/internal/tasks"
)

func (h *RouteHandler) CheckPaymentStatus(c *gin.Context) {
	var body dto.PaymentWebhookPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Compute signature hash
	dataToSign := fmt.Sprintf("%q", body.Reference)
	hash := hmac.New(sha256.New, []byte(h.Env.BackendServiceApiKey))
	hash.Write([]byte(dataToSign))
	signature := hex.EncodeToString(hash.Sum(nil))

	// Verify webhook request signature
	receivedSignature := c.GetHeader("x-webhook-signature")
	if receivedSignature == "" {
		log.Warn().Msg("Payment notification missing signature header")
		c.Status(http.StatusUnauthorized)
		return
	}

	if signature != receivedSignature {
		log.Warn().Msg("Payment notification signature mismatch")
		c.Status(http.StatusUnauthorized)
		return
	}

	ctx := c.Request.Context()
	cacheKey := fmt.Sprintf("payment_notification:%s", body.Reference)

	// Verify notification ID to ensure idempotent procssing
	exists, err := h.Cache.Exists(ctx, cacheKey).Result()
	if err != nil {
		log.Error().Err(err).Msg("Error fetching idempotency key from cache")
		c.Status(http.StatusInternalServerError)
		return
	}

	if exists > 0 {
		log.Warn().Str("reference", body.Reference).Msg("Duplicate payment notification received")
		c.String(http.StatusOK, "Duplicate notification")
		return
	}

	// Add webhook payload to queue for processing
	task, err := tasks.NewPaymentWebhookTask(body)
	if err != nil {
		log.Error().Err(err).Msg("Error creating new payment task worker")
		c.Status(http.StatusInternalServerError)
		return
	}

	if _, err := h.TasksQueue.EnqueueContext(ctx, task); err != nil {
		log.Error().Err(err).Msg("Error adding payment webhook payload to queue for processing")
		c.Status(http.StatusInternalServerError)
		return
	}

	// Acknowledge receipt of webhook from backend service
	c.String(http.StatusOK, "Payment notification processed")
}
