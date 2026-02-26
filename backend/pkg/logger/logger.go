package logger

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Config holds logger configuration
type Config struct {
	Level       string
	Format      string
	Output      string
	ServiceName string
	TimeFormat  string
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
		Level:       getEnv("LOG_LEVEL", "info"),
		Format:      getEnv("LOG_FORMAT", "json"),
		Output:      getEnv("LOG_OUTPUT", "stdout"),
		ServiceName: getEnv("SERVICE_NAME", "industrix"),
		TimeFormat:  "2006-01-02T15:04:05.000Z07:00",
	}
}

// Logger wraps zerolog.Logger
type Logger struct {
	*zerolog.Logger
	serviceName string
}

// New creates a new logger
func New(serviceName string) *Logger {
	cfg := DefaultConfig()
	cfg.ServiceName = serviceName
	return NewWithConfig(cfg)
}

// NewWithConfig creates a new logger with custom config
func NewWithConfig(cfg *Config) *Logger {
	zerolog.TimeFieldFormat = cfg.TimeFormat
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var output io.Writer

	switch strings.ToLower(cfg.Output) {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		output = os.Stdout
	}

	var logger zerolog.Logger

	switch strings.ToLower(cfg.Format) {
	case "console":
		logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	default:
		logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	}

	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}
	logger = logger.Level(level)

	return &Logger{
		Logger:      &logger,
		serviceName: cfg.ServiceName,
	}
}

// With creates a sublogger with additional context
func (l *Logger) With() zerolog.Context {
	return l.Logger.With()
}

// Trace starts a trace level log
func (l *Logger) Trace() *zerolog.Event {
	return l.Logger.Trace()
}

// Debug starts a debug level log
func (l *Logger) Debug() *zerolog.Event {
	return l.Logger.Debug()
}

// Info starts an info level log
func (l *Logger) Info() *zerolog.Event {
	return l.Logger.Info()
}

// Warn starts a warn level log
func (l *Logger) Warn() *zerolog.Event {
	return l.Logger.Warn()
}

// Error starts an error level log
func (l *Logger) Error() *zerolog.Event {
	return l.Logger.Error()
}

// Fatal starts a fatal level log
func (l *Logger) Fatal() *zerolog.Event {
	return l.Logger.Fatal()
}

// Panic starts a panic level log
func (l *Logger) Panic() *zerolog.Event {
	return l.Logger.Panic()
}

// WithService adds service name to context
func (l *Logger) WithService() *zerolog.Event {
	return l.Logger.With().Str("service", l.serviceName).Logger()
}

// WithTraceID adds trace ID to context
func (l *Logger) WithTraceID(traceID string) *zerolog.Event {
	return l.Logger.With().Str("trace_id", traceID).Logger()
}

// TraceFromContext extracts trace ID from context
func TraceFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return ""
}

// ContextWithTraceID adds trace ID to context
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, "trace_id", traceID)
}

// Global logger instance
var globalLogger *Logger

func init() {
	globalLogger = New("global")
}

// Default returns the global logger
func Default() *Logger {
	return globalLogger
}

// SetGlobal sets the global logger
func SetGlobal(logger *Logger) {
	globalLogger = logger
}

// Hook for adding additional fields
type ServiceHook struct {
	ServiceName string
}

func (h ServiceHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Str("service", h.ServiceName)
}

// TimeHook adds timestamp
type TimeHook struct{}

func (h TimeHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Time("time", time.Now().UTC())
}

// ContextHook adds context values
type ContextHook struct{}

func (h ContextHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Dict("context", zerolog.Dict())
}
