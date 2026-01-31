package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/xerdin442/ticketing-bot/internal/api/dto"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
	"github.com/xerdin442/ticketing-bot/internal/util"
	"google.golang.org/genai"
)

type GeminiService struct {
	env    *secrets.Secrets
	cache  *redis.Client
	client *genai.Client
}

func NewGeminiService(s *secrets.Secrets, r *redis.Client) *GeminiService {
	client, _ := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  s.GeminiApiKey,
		Backend: genai.BackendGeminiAPI,
	})

	return &GeminiService{
		env:    s,
		cache:  r,
		client: client,
	}
}

func (s *GeminiService) GetNextStateAfterFunctionCall(funcName string) (dto.ConversationState, error) {
	switch {
	case strings.Contains(funcName, "find"):
		return dto.StateEventQuery, nil
	case strings.Contains(funcName, "select_event"):
		return dto.StateEventSelected, nil
	case strings.Contains(funcName, "tier"):
		return dto.StateTicketTierSelected, nil
	case strings.Contains(funcName, "initiate"):
		return dto.StateAwaitingPayment, nil
	default:
		err := fmt.Errorf("Error getting next state after function call: Invalid function name")
		return dto.StateResponseError, err
	}
}

func (s *GeminiService) UpdateChatHistory(phoneId string, contextInfo *dto.ConversationContext) error {
	cacheKey := "chat_history:" + util.CreateHashedKey(phoneId)

	// Update chat history in cache
	if _, err := s.cache.RPush(context.Background(), cacheKey, contextInfo).Result(); err != nil {
		return fmt.Errorf("Error updating chat history")
	}

	// Clear stored contexts in chat history after 6 hours
	if err := s.cache.Expire(context.Background(), cacheKey, time.Hour*6).Err(); err != nil {
		return fmt.Errorf("Error setting expiration time of chat context")
	}

	return nil
}

func (s *GeminiService) GetChatHistory(phoneId string) ([]dto.ConversationContext, error) {
	cacheKey := "chat_history:" + util.CreateHashedKey(phoneId)
	result, err := s.cache.LRange(context.Background(), cacheKey, 0, -1).Result()

	if err != nil {
		return nil, fmt.Errorf("Error fetching chat history from cache")
	}

	chatHistory := make([]dto.ConversationContext, 0, len(result))

	for _, item := range result {
		var contextObj dto.ConversationContext

		if err := json.Unmarshal([]byte(item), &contextObj); err != nil {
			return nil, fmt.Errorf("Unmarshal error: Invalid conversation context retrieved from cache")
		}

		chatHistory = append(chatHistory, contextObj)
	}

	return chatHistory, nil
}

func (s *GeminiService) GenerateModelResponse(contents []*genai.Content) (*genai.GenerateContentResponse, error) {
	return s.client.Models.GenerateContent(
		context.Background(),
		"gemini-3-flash-preview",
		contents,
		&genai.GenerateContentConfig{
			Tools:             []*genai.Tool{util.RequiredTools},
			SystemInstruction: util.SystemInstructions,
		},
	)
}

func (s *GeminiService) ProcessUserMessage(phoneId string, userInput string) (any, error) {
	currentState := dto.StateInitial
	var contents []*genai.Content

	// Fetch current conversation history
	chatHistory, err := s.GetChatHistory(phoneId)
	if err != nil {
		return "", err
	}

	// Extract the current state and contents of the conversation history
	if len(chatHistory) > 0 {
		currentState = chatHistory[len(chatHistory)-1].CurrentState
		for _, h := range chatHistory {
			contents = append(contents, h.Content)
		}
	}

	// Configure the context to be passed to the model
	userContext := &genai.Content{Role: genai.RoleUser, Parts: []*genai.Part{{Text: userInput}}}
	contents = append(contents, userContext)

	// Generate model response
	resp, err := s.GenerateModelResponse(contents)
	if err != nil {
		s.UpdateChatHistory(phoneId, &dto.ConversationContext{
			Content:      &genai.Content{Role: genai.RoleModel, Parts: []*genai.Part{{Text: "Response generation error"}}},
			CurrentState: dto.StateResponseError,
		})

		log.Error().Err(err).Msg("Error generating response from Gemini API")

		return "Sorry, I am unable to process your request at the moment.", nil
	}

	var result any
	var modelPart *genai.Part
	part := resp.Candidates[0].Content.Parts[0]

	// Check the response if the model made a function call
	if part.FunctionCall != nil {
		// Determine the next conversation state based on the function call
		currentState, err = s.GetNextStateAfterFunctionCall(part.FunctionCall.Name)
		if err != nil {
			return "", err
		}

		result = part.FunctionCall
		modelPart = &genai.Part{FunctionCall: part.FunctionCall}
	} else {
		result = part.Text
		modelPart = &genai.Part{Text: part.Text}
	}

	// Add user input to conversation history
	s.UpdateChatHistory(phoneId, &dto.ConversationContext{
		Content:      userContext,
		CurrentState: currentState,
	})

	// Add model response to conversation history
	s.UpdateChatHistory(phoneId, &dto.ConversationContext{
		Content:      &genai.Content{Role: genai.RoleModel, Parts: []*genai.Part{modelPart}},
		CurrentState: currentState,
	})

	return result, nil
}

func (s *GeminiService) ProcessFunctionCall(phoneId string, apiContext map[string]any) (string, error) {
	// Fetch current conversation history
	chatHistory, err := s.GetChatHistory(phoneId)
	if err != nil {
		return "", err
	}

	if len(chatHistory) == 0 {
		return "", fmt.Errorf("Error processing function call: Empty conversation history")
	}

	// Extract the current state and contents of the conversation history
	currentState := chatHistory[len(chatHistory)-1].CurrentState
	var contents []*genai.Content
	for _, h := range chatHistory {
		contents = append(contents, h.Content)
	}

	// Retrieve details of last function call
	var lastFunctionCall *genai.FunctionCall
	for i := len(contents) - 1; i >= 0; i-- {
		if contents[i].Role == genai.RoleModel && contents[i].Parts[0].FunctionCall != nil {
			lastFunctionCall = contents[i].Parts[0].FunctionCall
			break
		}
	}

	if lastFunctionCall == nil {
		err := fmt.Errorf("Error processing function call: Missing function call in latest conversation context")
		return "", err
	}

	// Configure the context to be passed to the model
	functionContent := &genai.Content{
		Role: "system",
		Parts: []*genai.Part{{
			FunctionResponse: &genai.FunctionResponse{
				Name:     lastFunctionCall.Name,
				Response: apiContext, // Data from the backend service passed as context to the model
			},
		}},
	}
	contents = append(contents, functionContent)

	// Generate model response
	resp, err := s.GenerateModelResponse(contents)
	if err != nil {
		s.UpdateChatHistory(phoneId, &dto.ConversationContext{
			Content:      &genai.Content{Role: "model", Parts: []*genai.Part{{Text: "Response generation error"}}},
			CurrentState: dto.StateResponseError,
		})

		log.Error().Err(err).Msg("Error generating response from Gemini API")

		return "Sorry, I am unable to process your request at the moment.", nil
	}

	finalText := resp.Candidates[0].Content.Parts[0].Text

	// Add details of function call to conversation history
	s.UpdateChatHistory(phoneId, &dto.ConversationContext{
		Content:      functionContent,
		CurrentState: currentState,
	})

	// Add model's final response to conversation history
	s.UpdateChatHistory(phoneId, &dto.ConversationContext{
		Content:      &genai.Content{Role: genai.RoleModel, Parts: []*genai.Part{{Text: finalText}}},
		CurrentState: currentState,
	})

	return finalText, nil
}
