package dto

type IncomingMessage struct {
	Context *struct {
		From string `json:"from"`
		ID   string `json:"id"`
	} `json:"context,omitempty"`
	From        string       `json:"from"`
	ID          string       `json:"id"`
	Timestamp   string       `json:"timestamp"`
	Type        string       `json:"type"` // "text" | "location" | "interactive"
	Text        *MessageBody `json:"text,omitempty"`
	Location    *Location    `json:"location,omitempty"`
	Interactive *Interactive `json:"interactive,omitempty"`
}

type MessageBody struct {
	Body string `json:"body"`
}

type Location struct {
	Address   *string `json:"address,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      *string `json:"name,omitempty"`
}

type Interactive struct {
	Type        string       `json:"type"`
	ButtonReply *ButtonReply `json:"button_reply,omitempty"`
}

type ButtonReply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
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

type ReplyInteractive struct {
	Type   string `json:"type"` // "location_request_message" | "button"
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
