package tracer

import (
	"context"
	"fmt"
	"os"

	"github.com/industrix/pkg/logger"
	"github.com/jaegertracing/jaeger-client-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type Config struct {
	ServiceName string
	AgentHost   string
	AgentPort   string
	Enabled     bool
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		ServiceName: getEnv("SERVICE_NAME", "industrix"),
		AgentHost:   getEnv("JAEGER_AGENT_HOST", "localhost"),
		AgentPort:   getEnv("JAEGER_AGENT_PORT", "6831"),
		Enabled:     getEnv("JAEGER_ENABLED", "true") == "true",
	}
}

// Init initializes the OpenTelemetry tracer with Jaeger exporter
func Init(ctx context.Context, cfg *Config) (*sdktrace.TracerProvider, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if !cfg.Enabled {
		log := logger.New("tracer")
		log.Info().Msg("Tracing disabled")
		return nil, nil
	}

	// Create Jaeger exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint(fmt.Sprintf("http://%s:14268/api/traces", cfg.AgentHost)),
		jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(cfg.AgentHost),
			jaeger.WithAgentPort(cfg.AgentPort),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Register the global tracer provider
	otel.SetTracerProvider(tp)

	// Set up text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log := logger.New("tracer")
	log.Info().
		Str("service", cfg.ServiceName).
		Str("agent", cfg.AgentHost).
		Msg("Tracer initialized")

	return tp, nil
}

// TraceFromContext extracts trace ID from context
func TraceFromContext(ctx context.Context) string {
	span := otel.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// StartSpan starts a new span
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, interface{}) {
	tracer := otel.Tracer("industrix")
	ctx, span := tracer.Start(ctx, name, attrs...)
	return ctx, span
}

// EndSpan ends a span
func EndSpan(span interface{}) {
	if s, ok := span.(interface{ End() }); ok {
		s.End()
	}
}

// GetTraceID extracts trace ID from context (convenience function)
func GetTraceID(ctx context.Context) string {
	return TraceFromContext(ctx)
}

// Helper to extract traceparent header for downstream requests
func InjectContext(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}
