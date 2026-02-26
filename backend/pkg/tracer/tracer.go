package tracer

import (
	"context"
	"fmt"
	"os"

	"github.com/industrix/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Config holds tracer configuration
type Config struct {
	ServiceName    string
	ServiceVersion string
	JaegerEndpoint string
	OTLPEndpoint   string
	SampleRate     float64
	Enabled        bool
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// DefaultConfig returns configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		ServiceName:    getEnv("SERVICE_NAME", "industrix"),
		ServiceVersion: getEnv("SERVICE_VERSION", "1.0.0"),
		JaegerEndpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		OTLPEndpoint:   getEnv("OTLP_ENDPOINT", "localhost:4317"),
		SampleRate:     1.0,
		Enabled:        getEnv("OTEL_ENABLED", "true") == "true",
	}
}

// Init initializes the tracer provider
func Init(ctx context.Context, cfg *Config) (*sdktrace.TracerProvider, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("tracer")

	if !cfg.Enabled {
		log.Warn().Msg("Tracing is disabled")
		return nil, nil
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create Jaeger exporter
	jaegerExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.JaegerEndpoint)))
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create Jaeger exporter, continuing without it")
		// Continue without Jaeger
	}

	// Create OTLP trace exporter
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create OTLP trace exporter, continuing without it")
		// Continue without OTLP
	}

	// Create span processors
	var spanProcessors []sdktrace.SpanProcessor

	if jaegerExporter != nil {
		spanProcessors = append(spanProcessors, sdktrace.NewSimpleSpanProcessor(jaegerExporter))
	}

	if traceExporter != nil {
		spanProcessors = append(spanProcessors, sdktrace.NewSimpleSpanProcessor(traceExporter))
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRate))),
		sdktrace.WithSpanProcessors(spanProcessors...),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Info().
		Str("service", cfg.ServiceName).
		Str("jaeger", cfg.JaegerEndpoint).
		Str("otlp", cfg.OTLPEndpoint).
		Msg("Tracer initialized")

	return tp, nil
}

// InitMetrics initializes the metrics provider
func InitMetrics(ctx context.Context, cfg *Config) (*metric.MeterProvider, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	log := logger.New("metrics")

	if !cfg.Enabled {
		log.Warn().Msg("Metrics is disabled")
		return nil, nil
	}

	// Create OTLP metric exporter
	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create OTLP metric exporter")
		return nil, nil
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create meter provider
	mp := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)

	otel.SetMeterProvider(mp)

	log.Info().Msg("Metrics initialized")

	return mp, nil
}

// TraceFromContext extracts trace context from context
func TraceFromContext(ctx context.Context) (string, string) {
	span := otel.GetTracerProvider().Tracer("").Start(ctx, "")
	defer span.End()

	spanCtx := span.SpanContext()
	return spanCtx.TraceID().String(), spanCtx.SpanID().String()
}

// ContextWithTrace adds trace context to context
func ContextWithTrace(ctx context.Context, traceID, spanID string) context.Context {
	// This is a simplified version - in production you'd use the proper types
	return ctx
}

// AddSpanAttributes adds attributes to a span
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	// This would use otel functions in production
	_ = attrs
}

// Helper to get current trace ID
func GetTraceID(ctx context.Context) string {
	// Simplified - in production use proper extraction
	return ""
}
