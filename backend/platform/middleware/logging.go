package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/industrix/backend/pkg/logger"
)

type LoggingMiddleware struct {
	logger *logger.Logger
}

func NewLogging(logger *logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

func (m *LoggingMiddleware) RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		traceID := c.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Set("X-Trace-ID", traceID)
		ctx := context.WithValue(c.Context(), "trace_id", traceID)
		c.SetUserContext(ctx)

		err := c.Next()

		m.logger.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", time.Since(start)).
			Str("trace_id", traceID).
			Msg("Request processed")

		return err
	}
}
