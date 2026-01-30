package dto

type FunctionName int

const (
	FindEvents FunctionName = iota
	FindNearbyEvents
	FindTrendingEvents
	SelectEvent
	SelectTicketTier
	InitiateTicketPurchase
)

func (s FunctionName) String() string {
	funcNames := []string{
		"find_events",
		"find_nearby_events",
		"find_trending_events",
		"select_event",
		"select_ticket_tier",
		"initiate_ticket_purchase",
	}

	if s < 0 || int(s) >= len(funcNames) {
		return "unknown"
	}
	return funcNames[s]
}
