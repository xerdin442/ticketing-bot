package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func (m *Middleware) CustomRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Generate and attach a request ID to the request context
		reqID := uuid.New().String()
		c.Header("X-Request-ID", reqID)
		ctx := log.With().Str("id", reqID).Logger().WithContext(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		c.Next() // Process request

		log.Info().
			Str("method", c.Request.Method).
			Int("status", c.Writer.Status()).
			Str("path", path).
			Str("query", query).
			Str("ip", c.ClientIP()).
			Dur("latency", time.Since(start)).
			Msg("Request processed")
	}
}
