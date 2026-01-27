package dto

import "google.golang.org/genai"

type ConversationState string

const (
	StateInitial            ConversationState = "initial"
	StateEventQuery         ConversationState = "event_query"
	StateEventSelected      ConversationState = "event_selected"
	StateTicketTierSelected ConversationState = "ticket_tier_selected"
	StateAwaitingPayment    ConversationState = "awaiting_payment"
	StateCompleted          ConversationState = "completed"
	StateResponseError      ConversationState = "response_error"
)

type ConversationContext struct {
	Content      *genai.Content    `json:"content"`
	CurrentState ConversationState `json:"current_state"`
}
