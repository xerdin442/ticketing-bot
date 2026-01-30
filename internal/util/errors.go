package util

import "errors"

var (
	ErrInvalidFunctionName      = errors.New("Invalid function name")
	ErrChatHistoryUpdateFailed  = errors.New("Failed to update chat history")
	ErrSetChatExpirationFailed  = errors.New("Failed to set expiration time of conversation context in chat history")
	ErrChatHistoryFetchFailed   = errors.New("Failed to fetch chat history from cache")
	ErrEmptyConversationHistory = errors.New("Empty conversation history")
	ErrMissingFunctionCall      = errors.New("Missing function call in latest conversation context")
)
