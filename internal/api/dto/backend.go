package dto

import "time"

type PaymentWebhookPayload struct {
	Reference string  `json:"reference"`
	Status    string  `json:"status"` // "success" | "failed" | "refund"
	PhoneID   string  `json:"phoneId"`
	Email     string  `json:"email"`
	Reason    *string `json:"reason,omitempty"`
}

type Event struct {
	ID             int32     `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Date           time.Time `json:"date"`
	StartTime      time.Time `json:"startTime"`
	EndTime        time.Time `json:"endTime"`
	AgeRestriction *int      `json:"ageRestriction,omitempty"`
	Venue          string    `json:"venue"`
	Address        string    `json:"address"`
	Poster         string    `json:"poster"`
}

type TicketTier struct {
	Name                    string     `json:"name"`
	ID                      int32      `json:"id"`
	Price                   float64    `json:"price"`
	Discount                bool       `json:"discount"`
	DiscountPrice           *float64   `json:"discountPrice,omitempty"`
	DiscountExpiration      *time.Time `json:"discountExpiration,omitempty"`
	NumberOfDiscountTickets *int       `json:"numberOfDiscountTickets,omitempty"`
	DiscountStatus          *string    `json:"discountStatus,omitempty"` // "ACTIVE" | "ENDED"
	Benefits                *string    `json:"benefits,omitempty"`
	TotalNumberOfTickets    int        `json:"totalNumberOfTickets"`
	SoldOut                 bool       `json:"soldOut"`
}
