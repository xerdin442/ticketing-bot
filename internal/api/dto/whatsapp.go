package dto

type IncomingMessageType string

const (
	TextMessageType        IncomingMessageType = "text"
	LocationMessageType    IncomingMessageType = "location"
	InteractiveMessageType IncomingMessageType = "interactive"
)

type IncomingMessage struct {
	Context *struct {
		From string `json:"from"`
		ID   string `json:"id"`
	} `json:"context,omitempty"`
	From      string              `json:"from"`
	ID        string              `json:"id"`
	Timestamp string              `json:"timestamp"`
	Type      IncomingMessageType `json:"type"`
	Text      *struct {
		Body string `json:"body"`
	} `json:"text,omitempty"`
	Location    *LocationMessage    `json:"location,omitempty"`
	Interactive *InteractiveMessage `json:"interactive,omitempty"`
}

type LocationMessage struct {
	Address   *string `json:"address,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      *string `json:"name,omitempty"`
}

type InteractiveMessage struct {
	Type        string `json:"type"`
	ButtonReply *struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"button_reply,omitempty"`
}

type WebhookRequest struct {
	Entry []struct {
		ID      string `json:"id"`
		Changes []struct {
			Value struct {
				Messages []IncomingMessage `json:"messages"`
			} `json:"value"`
			Field string `json:"field"`
		} `json:"changes"`
	} `json:"entry"`
}

type MessageReplyPayload struct {
	MessagingProduct string `json:"messaging_product"` // "whatsapp"
	RecipientType    string `json:"recipient_type"`    // "individual"
	To               string `json:"to"`
	Type             string `json:"type"` // "text" | "interactive"
	Context          *struct {
		MessageID string `json:"message_id"`
	} `json:"context,omitempty"`
	Text        *ReplyText        `json:"text,omitempty"`
	Interactive *ReplyInteractive `json:"interactive,omitempty"`
}

type ReplyText struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}

type ReplyInteractiveType string

const (
	LocationRequestReply   ReplyInteractiveType = "location_request_message"
	ButtonInteractiveReply ReplyInteractiveType = "button"
)

type ReplyInteractive struct {
	Type   ReplyInteractiveType `json:"type"`
	Header *struct {
		Type  string `json:"type"`
		Image struct {
			Link string `json:"link"`
		} `json:"image"`
	} `json:"header,omitempty"`
	Body struct {
		Text string `json:"text"`
	} `json:"body"`
	Action struct {
		Name    *string `json:"name,omitempty"`
		Buttons []struct {
			Type  string `json:"type"` // "reply"
			Reply struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"reply"`
		} `json:"buttons,omitempty"`
	} `json:"action"`
}
