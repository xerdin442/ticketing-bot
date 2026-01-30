package util

import (
	"github.com/xerdin442/ticketing-bot/internal/api/dto"
	"google.golang.org/genai"
)

var findEventsByFilters = &genai.FunctionDeclaration{
	Name: dto.FindEvents.String(),
	Description: `Retrieves a list of upcoming events based on the filters (i.e. title, location, categories or date) provided by the user.
    Only call this function when the user has provided any of the filters.
    User can provide multiple filters to help the function return more accurate search results.`,
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"eventTitle": {
				Type: genai.TypeString,
				Description: `The full or partial name of the event the user is looking for (e.g. "Burna Boy Homecoming Concert", "Devfest 2025").
          If it is a partial name, the function retrieves a list of events that match the search string`,
			},
			"location": {
				Type:        genai.TypeString,
				Description: `The town, city, or state where the event is taking place.`,
			},
			"venue": {
				Type:        genai.TypeString,
				Description: `The venue where the event is taking place.`,
			},
			"startDate": {
				Type: genai.TypeString,
				Description: `The start date of the event in ISO format: YYYY-MM-DD.
          If the user says "next week", return the ISO date string of the upcoming Monday.
          If the user says "next month", return the ISO date string of the first day of next month.
          If the user says "weekend", return the ISO date string of the upcoming Friday.
          Follow this process if the user provides other date values in grammatical phrases (e.g "within the week", "tomorrow", "a week from now", etc.).
          If the user does not provide a "year" value in the request, default to the year of the current date.
          If the date provided by the user is less than the current date, request for a valid date value.
          `,
			},
			"endDate": {
				Type: genai.TypeString,
				Description: `The end date of the event in ISO format: YYYY-MM-DD.
          This value is only required if the user provides a date range with a start and end (e.g "next month", "over the weekend").
          If the user says "next week", return the ISO date string of the upcoming Saturday.
          If the user says "next month", return the ISO date string of the last day of next month.
          If the user says "weekend", return the ISO date string of the upcoming Sunday.
          `,
			},
			"categories": {
				Type:        genai.TypeArray,
				Description: "The category of the event to search for. Must be one of the enumerated values. User can select multiple categories",
				Items: &genai.Schema{
					Type: genai.TypeString,
					Enum: []string{
						"TECH",
						"HEALTH",
						"MUSIC",
						"COMEDY",
						"NIGHTLIFE",
						"ART",
						"FASHION",
						"SPORTS",
						"BUSINESS",
						"CONFERENCE",
						"OTHER",
					},
				},
			},
			"numberOfQueries": {
				Type: genai.TypeNumber,
				Description: `Acts as a cursor to paginate the results of this function call when it is called consecutively
          with the same parameters to retrieve more events. Default is 1 for the first call.
          Value increments by 1 for each consecutive call. Resets to default value when another function is called`,
			},
		},
		Required: []string{"numberOfQueries"},
	},
}

var findNearbyEvents = &genai.FunctionDeclaration{
	Name:        dto.FindNearbyEvents.String(),
	Description: "Retrieves a list of upcoming events happening close to the user",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}

var findTrendingEvents = &genai.FunctionDeclaration{
	Name:        dto.FindTrendingEvents.String(),
	Description: "Retrieves a list of the most popular and trending events",
	Parameters: &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	},
}

var selectEvent = &genai.FunctionDeclaration{
	Name: dto.SelectEvent.String(),
	Description: `Returns a list of ticket tiers available for the specific event selected by the user.
    This function is called after the user has selected a specific event from a list of options presented to them.`,
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"eventId": {
				Type:        genai.TypeNumber,
				Description: "The ID of the event selected by the user",
			},
		},
		Required: []string{"eventId"},
	},
}

var selectTicketTier = &genai.FunctionDeclaration{
	Name:        dto.SelectTicketTier.String(),
	Description: "Stores the name of the selected ticket tier and the purchase quantity",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"eventId": {
				Type: genai.TypeNumber,
				Description: `The ID of the event the ticket tier belongs to.
          Must match the ID of the specific event earlier selected by the user.`,
			},
			"tierName": {
				Type: genai.TypeString,
				Description: `The name of the ticket tier the user intends to purchase.
          This must match the names of the ticket tiers available in the selected event.`,
			},
			"quantity": {
				Type:        genai.TypeNumber,
				Description: "The number of tickets the user intends to purchase in the selected tier",
			},
		},
		Required: []string{"eventId", "tierName", "quantity"},
	},
}

var initiateTicketPurchase = &genai.FunctionDeclaration{
	Name:        dto.InitiateTicketPurchase.String(),
	Description: "Generates a secure checkout link to initiate purchase of the selected tickets",
	Parameters: &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"email": {
				Type:        genai.TypeString,
				Description: "Valid email address of the user, required to generate checkout link",
			},
		},
		Required: []string{"email"},
	},
}

var RequiredTools = &genai.Tool{
	FunctionDeclarations: []*genai.FunctionDeclaration{
		findEventsByFilters,
		findNearbyEvents,
		findTrendingEvents,
		selectEvent,
		selectTicketTier,
		initiateTicketPurchase,
	},
}
