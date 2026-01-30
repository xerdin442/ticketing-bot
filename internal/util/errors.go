package util

import "errors"

var (
	ErrInvalidFunctionName        = errors.New("Invalid function name")
	ErrChatHistoryUpdateFailed    = errors.New("Failed to update chat history")
	ErrSetChatExpirationFailed    = errors.New("Failed to set expiration time of conversation context in chat history")
	ErrChatHistoryFetchFailed     = errors.New("Failed to fetch chat history from cache")
	ErrEmptyConversationHistory   = errors.New("Empty conversation history")
	ErrMissingFunctionCall        = errors.New("Missing function call in latest conversation context")
	ErrBackendRequestFailed       = errors.New("Failed to request context from backend service")
	ErrInvalidResponsePayload     = errors.New("Invalid response payload")
	ErrWahtsappApiRequestFailed   = errors.New("Failed to send request to Whatsapp Cloud API")
	ErrInvalidIncomingMessageType = errors.New("Invalid incoming message type received from Whatsapp Cloud API")
	ErrUnknownModelResponseType   = errors.New("Unknown model response type")
	ErrInvalidContextPayloadType  = errors.New("Invalid payload type received from backend service")
	ErrExpectedFunctionCall       = errors.New("Incorrect model response. Expected a function call")
	ErrWrongFunctionCall          = errors.New("Incorrect function call from model")
)
