package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/xerdin442/ticketing-bot/internal/api/dto"
)

func (h *RouteHandler) VerifyWebhook(c *gin.Context) {
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")

	// Validate details of webhook verification request
	if mode != "" && token != "" {
		if mode == "subscribe" && token == h.Env.WhatsappWebhookVerificationToken {
			log.Info().Msg("Webhook verification successful!")
			c.String(http.StatusOK, challenge)

			return
		}
	}

	log.Error().Msg("Error verifying Whatsapp webhook token")

	c.String(http.StatusForbidden, "Verification failed: Invalid token")
}

func (h *RouteHandler) HandleIncomingMessage(c *gin.Context) {
	var payload dto.WebhookRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Error().Err(err).Msg("Error binding Whatsapp message webhook payload")
		c.Status(http.StatusBadRequest)

		return
	}

	if len(payload.Entry) == 0 {
		c.Status(http.StatusOK)
		return
	}

	entry := payload.Entry[0]

	// Verify Business Account ID in message payload
	if entry.ID != h.Env.WhatsappBusinessAccountId {
		log.Warn().Str("received_id", entry.ID).Msg("Error handling incoming message: Received webhook for unauthorized ID")
		c.Status(http.StatusOK)

		return
	}

	if len(entry.Changes) == 0 || len(entry.Changes[0].Value.Messages) == 0 {
		c.Status(http.StatusOK)
		return
	}

	message := entry.Changes[0].Value.Messages[0]

	err := h.Services.Message.HandleIncomingMessage(message)
	if err != nil {
		log.Error().Err(err).Msg("Error processing incoming message")
		c.Status(http.StatusOK)

		return
	}

	c.Status(http.StatusOK)
}
