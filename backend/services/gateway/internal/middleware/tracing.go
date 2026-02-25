package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	TraceIDHeader = "X-Trace-ID"
	TraceIDKey    = "traceID"
)

// InjectTraceID generates or propagates X-Trace-ID header
// Injects into downstream request headers and context
func InjectTraceID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check for existing trace ID in header
		traceID := c.Get(TraceIDHeader)

		// Generate new trace ID if not present
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Set trace ID in locals for this request
		c.Locals(TraceIDKey, traceID)

		// Set trace ID in response header
		c.Set(TraceIDHeader, traceID)

		// Add trace ID to downstream request headers
		c.Set(TraceIDHeader, traceID)

		return c.Next()
	}
}

// GetTraceID extracts trace ID from context
func GetTraceID(c *fiber.Ctx) string {
	if traceID, ok := c.Locals(TraceIDKey).(string); ok {
		return traceID
	}
	return c.Get(TraceIDHeader)
}

// InjectTraceIDContext propagates trace ID to downstream services via context
// This is used when making gRPC/HTTP calls to other services
func InjectTraceIDContext() fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := GetTraceID(c)
		if traceID != "" {
			// Add to headers for downstream HTTP calls
			c.Set(TraceIDHeader, traceID)
		}
		return c.Next()
	}
}
