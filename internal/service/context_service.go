package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/xerdin442/ticketing-bot/internal/api/dto"
	"github.com/xerdin442/ticketing-bot/internal/secrets"
	"github.com/xerdin442/ticketing-bot/internal/util"
	"google.golang.org/genai"
)

type ApiResponse struct {
	Events   []*dto.Event      `json:"events"`
	Tickets  []*dto.TicketTier `json:"tickets"`
	Checkout string            `json:"checkout"`
	Message  string            `json:"message"`
}

type ContextService struct {
	env        *secrets.Secrets
	cache      *redis.Client
	httpClient *http.Client
}

func NewContextService(s *secrets.Secrets, r *redis.Client) *ContextService {
	return &ContextService{
		env:        s,
		cache:      r,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *ContextService) sendRequest(method, path string, body io.Reader, errorMsg string) (ApiResponse, error) {
	// Configure request details
	baseUrl := s.env.BackendServiceUrl + "/api/whatsapp"
	req, err := http.NewRequest(method, baseUrl+path, body)
	if err != nil {
		log.Error().Err(err).Msg("Error configuring new HTTP request")
		return ApiResponse{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.env.BackendServiceApiKey)

	// Send request to backend service
	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Error sending request to backend service")
		return ApiResponse{}, err
	}
	defer resp.Body.Close()

	// Decode response body
	var result ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Error().Msg("Error decoding repsonse body: Invalid structure")
		return ApiResponse{}, util.ErrInvalidResponsePayload
	}

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Str("error_msg", result.Message).
			Msg(errorMsg)

		return ApiResponse{}, util.ErrBackendRequestFailed
	}

	return result, nil
}

func (s *ContextService) SelectEndpoint(funcCall *genai.FunctionCall, phoneId string) (map[string]any, error) {
	switch funcCall.Name {
	case dto.FindEvents.String():
		return s.FindEventsByFilters(funcCall.Args)
	case dto.FindTrendingEvents.String():
		return s.GetTrendingEvents()
	case dto.SelectEvent.String():
		return s.SelectEvent(funcCall.Args["eventId"])
	case dto.SelectTicketTier.String():
		return s.SelectTicketTier(funcCall.Args, phoneId)
	case dto.InitiateTicketPurchase.String():
		return s.InitiateTicketPurchase(funcCall.Args["email"], phoneId)
	default:
		log.Fatal().Msg("Error selecting endpoint in context service: Invalid function name")
		return nil, util.ErrInvalidFunctionName
	}
}

func (s *ContextService) FindEventsByFilters(args map[string]any) (map[string]any, error) {
	params := url.Values{}

	// Add filter as search params
	for key, value := range args {
		// Map 'numberOfQueries' parameter as 'page'
		param := key
		if key == "numberOfQueries" {
			param = "page"
		}

		switch v := value.(type) {
		case []string:
			for _, item := range v {
				params.Add(param, item)
			}
		case string:
			params.Set(param, v)
		default:
			if v != nil {
				params.Set(param, fmt.Sprintf("%v", v))
			}
		}
	}

	urlPath := "/events?" + params.Encode()
	response, err := s.sendRequest("GET", urlPath, nil, "Error finding events by filter")
	if err != nil {
		return nil, err
	}

	return map[string]any{"events": response.Events}, nil
}

func (s *ContextService) GetNearbyEvents(latitude, longitude int) (map[string]any, error) {
	urlPath := fmt.Sprintf("/events/nearby?latitude=%d&longitude=%d", latitude, longitude)
	response, err := s.sendRequest("GET", urlPath, nil, "Error fetching nearby events")
	if err != nil {
		return nil, err
	}

	return map[string]any{"events": response.Events}, nil
}

func (s *ContextService) GetTrendingEvents() (map[string]any, error) {
	errorMsg := "Error fetching all trending events"
	response, err := s.sendRequest("GET", "/events/trending", nil, errorMsg)
	if err != nil {
		return nil, err
	}

	return map[string]any{"events": response.Events}, nil
}

func (s *ContextService) SelectEvent(eventId any) (map[string]any, error) {
	urlPath := fmt.Sprintf("/events/%d/tickets", eventId)
	errorMsg := "Error fetching available ticket tiers for an event"

	response, err := s.sendRequest("GET", urlPath, nil, errorMsg)
	if err != nil {
		return nil, err
	}

	return map[string]any{"tickets": response.Tickets}, nil
}

func (s *ContextService) SelectTicketTier(args map[string]any, phoneId string) (map[string]any, error) {
	cacheKey := "ticket_purchase:" + util.CreateHashedKey(phoneId)

	if _, err := s.cache.Set(context.Background(), cacheKey, args, time.Hour*3).Result(); err != nil {
		log.Error().Err(err).Msg("Error storing ticket purchase details in cache")
		return nil, err
	}

	return map[string]any{"message": "Ticket purchase details stored in cache"}, nil
}

func (s *ContextService) InitiateTicketPurchase(email any, phoneId string) (map[string]any, error) {
	cacheKey := "ticket_purchase:" + util.CreateHashedKey(phoneId)

	cacheResult, err := s.cache.Get(context.Background(), cacheKey).Result()
	if err == redis.Nil {
		return map[string]any{
			"message": "Ticket purchase window has expired. Please restart the process",
		}, nil
	} else if err != nil {
		log.Error().Err(err).Msg("Error fetching ticket purchase details from cache")
		return nil, err
	}

	// Extract purchase details from cache result
	var details map[string]any
	if err := json.Unmarshal([]byte(cacheResult), &details); err != nil {
		return nil, err
	}

	// Configure request payload
	quantity, _ := strconv.Atoi(fmt.Sprintf("%v", details["quantity"]))
	payload := map[string]any{
		"tier":            details["tierName"],
		"quantity":        quantity,
		"email":           email,
		"whatsappPhoneId": phoneId,
	}

	body, _ := json.Marshal(payload)
	urlPath := fmt.Sprintf("/events/%v/tickets/purchase", details["eventId"])
	errorMsg := "Error generating checkout link for ticket purchase"

	response, err := s.sendRequest("POST", urlPath, bytes.NewBuffer(body), errorMsg)
	if err != nil {
		return nil, err
	}

	return map[string]any{"checkout": response.Checkout}, nil
}
