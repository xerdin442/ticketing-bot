package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/xerdin442/ticketing-bot/internal/api/dto"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
	"github.com/xerdin442/ticketing-bot/internal/util"
	"google.golang.org/genai"
)

type MessageService struct {
	env        *secrets.Secrets
	context    *ContextService
	gemini     *GeminiService
	httpClient *http.Client
}

func NewMessageService(s *secrets.Secrets, r *redis.Client) *MessageService {
	return &MessageService{
		env:        s,
		context:    NewContextService(s, r),
		gemini:     NewGeminiService(s, r),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func ptr(s string) *string {
	return &s
}

func (s *MessageService) sendRequest(body io.Reader, errorMsg string) error {
	// Configure request details
	req, err := http.NewRequest("POST", s.env.WhatsappMessagingApiUrl, body)
	if err != nil {
		return fmt.Errorf("Error configuring new HTTP request. %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.env.WhatsappUserAccessToken)

	// Send request to backend service
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request to Whatsapp Cloud API. %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s. Status code: %d", errorMsg, resp.StatusCode)
	}

	log.Info().Msg("Reply sent successfully!")
	return nil
}

func (s *MessageService) markMessageAsRead(messageId string) error {
	// Configure request payload
	payload := dto.MessageRequestPayload{
		MessagingProduct: "whatsapp",
		Status:           ptr("read"),
		MessageID:        &messageId,
		TypingIndicator: &struct {
			Type string "json:\"type\""
		}{Type: "text"},
	}

	body, _ := json.Marshal(payload)
	return s.sendRequest(bytes.NewBuffer(body), "Error marking message as read")
}

func (s *MessageService) sendLocationRequest(phoneId, messageId string) error {
	// Configure request payload
	payload := dto.MessageRequestPayload{
		MessagingProduct: "whatsapp",
		RecipientType:    ptr("individual"),
		To:               &phoneId,
		Type:             ptr("interactive"),
		Context: &struct {
			MessageID string "json:\"message_id\""
		}{MessageID: messageId},
		Interactive: &dto.ReplyInteractive{
			Type: dto.LocationRequestReply,
			Body: struct {
				Text string "json:\"text\""
			}{
				Text: "To help us find nearby events, please share your location.",
			},
			Action: dto.ReplyInteractiveAction{
				Name: ptr("send_location"),
			},
		},
	}

	// Mark previous message as read
	if err := s.markMessageAsRead(messageId); err != nil {
		return err
	}

	body, _ := json.Marshal(payload)
	return s.sendRequest(bytes.NewBuffer(body), "Error sending location request message")
}

func (s *MessageService) sendInteractiveBtnMessage(phoneId string, event *dto.Event) error {
	// Configure request payload
	payload := dto.MessageRequestPayload{
		MessagingProduct: "whatsapp",
		RecipientType:    ptr("individual"),
		To:               &phoneId,
		Type:             ptr("interactive"),
		Interactive: &dto.ReplyInteractive{
			Type: dto.ButtonInteractiveReply,
			Header: &dto.ReplyInteractiveHeader{
				Type: "image",
				Image: struct {
					Link string "json:\"link\""
				}{
					Link: event.Poster,
				},
			},
			Body: struct {
				Text string "json:\"text\""
			}{
				Text: fmt.Sprintf("%v\n\nDate: %v", strings.ToUpper(event.Title), util.FormatDate(event.Date)),
			},
			Action: dto.ReplyInteractiveAction{
				Buttons: []dto.ReplyInteractiveButton{
					{
						Type: "reply",
						Reply: struct {
							ID    string "json:\"id\""
							Title string "json:\"title\""
						}{
							ID:    fmt.Sprintf("I want to attend event with ID: %d", event.ID),
							Title: "Select",
						},
					},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)
	return s.sendRequest(bytes.NewBuffer(body), "Error sending interactive button message")
}

func (s *MessageService) sendEventsList(phoneId, messageId, funcName string, ctx map[string]any, events []*dto.Event) error {
	functionResult := &dto.ConversationContext{
		Content: &genai.Content{
			Role: "system",
			Parts: []*genai.Part{{
				FunctionResponse: &genai.FunctionResponse{
					Name:     funcName,
					Response: ctx,
				},
			}},
		},
		CurrentState: dto.StateEventQuery,
	}

	// Add function result to conversation history
	s.gemini.UpdateChatHistory(phoneId, functionResult)

	// Mark previous message as read
	if err := s.markMessageAsRead(messageId); err != nil {
		return err
	}

	// Send list of events (trending or filter search results) to user
	for _, e := range events {
		if err := s.sendInteractiveBtnMessage(phoneId, e); err != nil {
			log.Error().Err(err).Msgf("Failed to send event with ID: %d", e.ID)
			continue
		}

		// Add small delay to prevent rate limiting
		time.Sleep(300 * time.Millisecond)
	}

	return nil
}

func (s *MessageService) HandleIncomingMessage(message dto.IncomingMessage) error {
	senderId := message.From
	messageId := message.ID

	switch message.Type {
	case dto.TextMessageType:
		var finalResponse string

		// Process incoming message from user
		userInput := message.Text.Body
		firstResponse, err := s.gemini.ProcessUserMessage(senderId, userInput)
		if err != nil {
			return err
		}

		switch v := firstResponse.(type) {
		case string:
			// Model responds directly with text (initial welcome message or follow-up question)
			text, _ := firstResponse.(string)
			finalResponse = text
		case *genai.FunctionCall:
			// Model makes a function call (requires context from backend service)
			functionCall := v

			// Retrieve data from backend service to be used as context
			if functionCall.Name == dto.FindNearbyEvents.String() {
				// Send a location request to the user to get coordinates
				s.sendLocationRequest(senderId, messageId)
				return nil
			}

			apiContext, err := s.context.SelectEndpoint(functionCall, senderId)
			if err != nil {
				return err
			}

			if strings.HasPrefix(functionCall.Name, "find_") {
				events, ok := apiContext["events"].([]*dto.Event)
				if !ok {
					return fmt.Errorf("Invalid payload type received from backend service")
				}

				// Send interactive buttton messages for users to select from if context is a non-empty list of events
				if len(events) > 0 {
					if err := s.sendEventsList(senderId, messageId, functionCall.Name, apiContext, events); err != nil {
						return err
					}

					return nil
				}

				// Update function call with empty events search result
				secondResponse, err := s.gemini.ProcessFunctionCall(senderId, apiContext)
				if err != nil {
					return err
				}

				finalResponse = secondResponse
			}
		default:
			return fmt.Errorf("Error processing user input: Unknown model response type")
		}

		// Configure request payload
		payload := dto.MessageRequestPayload{
			MessagingProduct: "whatsapp",
			RecipientType:    ptr("individual"),
			To:               &senderId,
			Type:             ptr("text"),
			Context: &struct {
				MessageID string "json:\"message_id\""
			}{MessageID: messageId},
			Text: &dto.ReplyText{
				PreviewURL: true,
				Body:       finalResponse,
			},
		}

		// Mark previous message as read
		if err := s.markMessageAsRead(messageId); err != nil {
			return err
		}

		body, _ := json.Marshal(payload)
		errorMsg := "Error handling text message webhook from Whatsapp Cloud API"
		return s.sendRequest(bytes.NewBuffer(body), errorMsg)
	case dto.LocationMessageType:
		// Extract coordinates from location message
		latitude := message.Location.Latitude
		longitude := message.Location.Longitude

		apiContext, err := s.context.GetNearbyEvents(int(latitude), int(longitude))
		if err != nil {
			return err
		}

		events, ok := apiContext["events"].([]*dto.Event)
		if !ok {
			return fmt.Errorf("Invalid payload type received from backend service")
		}

		// Send list of nearby events to user
		if len(events) > 0 {
			funcName := dto.FindNearbyEvents.String()
			if err := s.sendEventsList(senderId, messageId, funcName, apiContext, events); err != nil {
				return err
			}

			return nil
		}

		// Update function call with empty result
		resp, err := s.gemini.ProcessFunctionCall(senderId, apiContext)
		if err != nil {
			return err
		}

		// Configure request payload
		payload := dto.MessageRequestPayload{
			MessagingProduct: "whatsapp",
			RecipientType:    ptr("individual"),
			To:               &senderId,
			Type:             ptr("text"),
			Context: &struct {
				MessageID string "json:\"message_id\""
			}{MessageID: messageId},
			Text: &dto.ReplyText{
				PreviewURL: true,
				Body:       resp,
			},
		}

		// Mark previous message as read
		if err := s.markMessageAsRead(messageId); err != nil {
			return err
		}

		body, _ := json.Marshal(payload)
		errorMsg := "Error handling location message webhook from Whatsapp Cloud API"
		return s.sendRequest(bytes.NewBuffer(body), errorMsg)
	case dto.InteractiveMessageType:
		// Extract details of user's selection and pass as context to model
		userInput := message.Interactive.ButtonReply.ID
		firstResponse, err := s.gemini.ProcessUserMessage(senderId, userInput)
		if err != nil {
			return err
		}

		// Verify that model's response is a function call
		switch v := firstResponse.(type) {
		case string:
			return fmt.Errorf("Incorrect model response. Expected a function call")
		case *genai.FunctionCall:
			// Verify details of function call
			if v.Name != dto.SelectEvent.String() {
				return fmt.Errorf("Error verifying function call from model. Expected %s, received: %s", dto.SelectEvent, v.Name)
			}

			apiContext, err := s.context.SelectEndpoint(v, senderId)
			if err != nil {
				return err
			}

			// Process function call
			resp, err := s.gemini.ProcessFunctionCall(senderId, apiContext)
			if err != nil {
				return err
			}

			// Configure request payload
			payload := dto.MessageRequestPayload{
				MessagingProduct: "whatsapp",
				RecipientType:    ptr("individual"),
				To:               &senderId,
				Type:             ptr("text"),
				Context: &struct {
					MessageID string "json:\"message_id\""
				}{MessageID: messageId},
				Text: &dto.ReplyText{
					PreviewURL: true,
					Body:       resp,
				},
			}

			// Mark previous message as read
			if err := s.markMessageAsRead(messageId); err != nil {
				return err
			}

			body, _ := json.Marshal(payload)
			errorMsg := "Error handling interactive message webhook from Whatsapp Cloud API"
			return s.sendRequest(bytes.NewBuffer(body), errorMsg)
		default:
		}

		return nil
	default:
		return fmt.Errorf("Invalid incoming message type received from Whatsapp Cloud API")
	}
}
