# Ticketing Bot

A AI-powered assistant that condenses the entire ticketing experience (from event discovery to ticket purchasing) into a seamless, conversational flow on WhatsApp. [See demo here](https://x.com/xerdin_442/status/1997670636143935603?s=20)

## Features

* **Event Discovery:** Users can search for events using everyday language (e.g., "Show me tech conferences in Lagos next week" or "Find any music festivals").

* **Conversational Filtering:** Search by Title, Category, Date, or Location with real-time data retrieval from the backend service.

* **Tier Selection:** Interactive selection of ticket tiers handled entirely within the chat.

* **Secure Checkout:** Generates unique, secure checkout links for frictionless payments.

* **Dynamic Intelligence:** The bot is powered by Gemini 3. This enables the bot to handle complex queries, remember conversation context and provide human-like assistance.

* **State & Memory:** The bot uses a TTL-based Redis cache (6-hour expiry) to ensure the model stays updated during conversations.

* **Real-time Notifications:** Integrated with backend webhooks to notify users instantly about their payment status.
