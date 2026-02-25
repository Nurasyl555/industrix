package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func init() {
	logger = zerolog.New(nil)
}

// RequestLogger logs method, path, status, latency, trace-id for every request
func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start timer
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		status := c.Response().StatusCode()

		// Get trace ID
		traceID := GetTraceID(c)
		if traceID == "" {
			traceID = "none"
		}

		// Log request details
		logger.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Dur("latency", latency).
			Str("trace_id", traceID).
			Str("ip", c.IP()).
			Str("user_agent", c.Get("User-Agent")).
			Int64("request_id", c.Response().Header.ID()).
			Msg("request completed")

		return err
	}
}

// RequestLoggerWithFields creates a logger with custom fields
func RequestLoggerWithFields() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		latency := time.Since(start)
		status := c.Response().StatusCode()

		// Build log fields
		fields := map[string]interface{}{
			"method":    c.Method(),
			"path":      c.Path(),
			"status":    status,
			"latency":   latency.Milliseconds(),
			"trace_id":  GetTraceID(c),
			"client_ip": c.IP(),
		}

		// Add user ID if authenticated
		if userID := GetUserID(c); userID != "" {
			fields["user_id"] = userID
		}

		// Log based on status
		if status >= 500 {
			logger.Error().Fields(fields).Msg("server error")
		} else if status >= 400 {
			logger.Warn().Fields(fields).Msg("client error")
		} else {
			logger.Info().Fields(fields).Msg("request success")
		}

		return err
	}
}

// GetLatency returns the latency in milliseconds as a string
func GetLatency(c *fiber.Ctx) string {
	if latency, ok := c.Locals("latency").(time.Duration); ok {
		return strconv.FormatInt(latency.Milliseconds(), 10)
	}
	return "0"
}
