package dto

import "google.golang.org/genai"

type ConversationState int

const (
	StateInitial ConversationState = iota
	StateEventQuery
	StateEventSelected
	StateTicketTierSelected
	StateAwaitingPayment
	StateCompleted
	StateResponseError
)

func (s ConversationState) String() string {
	states := []string{
		"initial",
		"event_query",
		"event_selected",
		"ticket_tier_selected",
		"awaiting_payment",
		"completed",
		"response_error",
	}
	if s < 0 || int(s) >= len(states) {
		return "unknown"
	}
	return states[s]
}

type ConversationContext struct {
	Content      *genai.Content    `json:"content"`
	CurrentState ConversationState `json:"current_state"`
}
