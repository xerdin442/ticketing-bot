package secrets

import (
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
)

type Secrets struct {
	Port                             int
	Environment                      string
	RedisAddr                        string
	RedisPassword                    string
	WhatsappWebhookVerificationToken string
	WhatsappUserAccessToken          string
	WhatsappMessagingApiUrl          string
	WhatsappBusinessAccountId        string
	GeminiApiKey                     string
	BackendServiceApiKey             string
	BackendServiceUrl                string
}

func Load() *Secrets {
	return &Secrets{
		Port:                             GetInt("PORT"),
		Environment:                      GetStr("ENVIRONMENT"),
		RedisAddr:                        GetStr("REDIS_ADDR"),
		RedisPassword:                    GetStr("REDIS_PASSWORD"),
		WhatsappWebhookVerificationToken: GetStr("WHATSAPP_WEBHOOK_VERIFICATION_TOKEN"),
		WhatsappUserAccessToken:          GetStr("WHATSAPP_USER_ACCESS_TOKEN"),
		WhatsappMessagingApiUrl:          GetStr("WHATSAPP_MESSAGING_API_URL"),
		WhatsappBusinessAccountId:        GetStr("WHATSAPP_BUSINESS_ACCOUNT_ID"),
		GeminiApiKey:                     GetStr("GEMINI_API_KEY"),
		BackendServiceApiKey:             GetStr("BACKEND_SERVICE_API_KEY"),
		BackendServiceUrl:                GetStr("BACKEND_SERVICE_URL"),
	}
}

func GetStr(key string) string {
	value := os.Getenv(key)

	if value == "" {
		log.Fatal().Msgf("Missing environment variable: %s", key)
	}

	return value
}

func GetInt(key string) int {
	strValue := GetStr(key)

	intValue, err := strconv.ParseInt(strValue, 10, 64)
	if err != nil {
		log.Fatal().Err(err).Msg("Invalid string value")
	}

	return int(intValue)
}
