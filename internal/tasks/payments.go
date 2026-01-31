package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"github.com/xerdin442/ticketing-bot/internal/api/dto"
	"google.golang.org/genai"
)

func ptr(s string) *string {
	return &s
}

func NewPaymentWebhookTask(data dto.PaymentWebhookPayload) (*asynq.Task, error) {
	payload, _ := json.Marshal(data)
	return asynq.NewTask("payment_queue", payload), nil
}

func (h *TaskHandler) HandlePaymentWebhookTask(ctx context.Context, t *asynq.Task) error {
	var p dto.PaymentWebhookPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Error().Err(err).Msg("Error parsing payment webhook payload: Invalid structure")
		return err
	}

	// Update conversation history with payment status
	apiContext := map[string]any{
		"status": p.Status,
		"email":  p.Email,
		"reason": p.Reason,
	}

	// Add function result to conversation history
	functionResult := &dto.ConversationContext{
		Content: &genai.Content{
			Role: "system",
			Parts: []*genai.Part{{
				FunctionResponse: &genai.FunctionResponse{
					Name:     dto.InitiateTicketPurchase.String(),
					Response: apiContext,
				},
			}},
		},
		CurrentState: dto.StateCompleted,
	}
	h.gemini.UpdateChatHistory(p.PhoneID, functionResult)

	// Fetch current conversation history
	chatHistory, err := h.gemini.GetChatHistory(p.PhoneID)
	if err != nil {
		return err
	}

	var contents []*genai.Content
	for _, context := range chatHistory {
		contents = append(contents, context.Content)
	}

	// Generate response from model
	var modelResponse string
	response, err := h.gemini.GenerateModelResponse(contents)
	if err != nil {
		modelResponse = "Your payment is being processed."
	}
	modelResponse = response.Candidates[0].Content.Parts[0].Text

	// Configure request payload
	payload := dto.MessageRequestPayload{
		MessagingProduct: "whatsapp",
		RecipientType:    ptr("individual"),
		To:               &p.PhoneID,
		Type:             ptr("text"),
		Text: &dto.ReplyText{
			PreviewURL: true,
			Body:       modelResponse,
		},
	}

	// Configure request details
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", h.env.WhatsappMessagingApiUrl, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Msg("Error configuring new HTTP request")
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+h.env.WhatsappUserAccessToken)

	// Send request to backend service
	httpClient := http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Error sending request to Whatsapp Cloud API")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Int("status_code", resp.StatusCode).Msg("Error sending payment confirmation to user")
		return err
	}

	// Add model response to conversation history
	modelContext := &dto.ConversationContext{
		Content: &genai.Content{
			Role: genai.RoleModel,
			Parts: []*genai.Part{{
				Text: modelResponse,
			}},
		},
		CurrentState: dto.StateCompleted,
	}
	h.gemini.UpdateChatHistory(p.PhoneID, modelContext)

	// Store notification ID in Redis to prevent duplicate processing
	cacheKey := fmt.Sprintf("payment_notification:%s", p.Reference)
	_, cacheErr := h.cache.Set(context.Background(), cacheKey, "processed", 24*time.Hour).Result()
	if cacheErr != nil {
		log.Error().Err(cacheErr).Msg("Error storing payment webhook reference in cache")
		return cacheErr
	}

	return nil
}
