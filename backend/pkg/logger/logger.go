package logger

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type Config struct {
	Level       string
	Format      string
	Output      string
	ServiceName string
	TimeFormat  string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func DefaultConfig() *Config {
	return &Config{
		Level:       getEnv("LOG_LEVEL", "info"),
		Format:      getEnv("LOG_FORMAT", "json"),
		Output:      getEnv("LOG_OUTPUT", "stdout"),
		ServiceName: getEnv("SERVICE_NAME", "industrix"),
		TimeFormat:  "2006-01-02T15:04:05.000Z07:00",
	}
}

type Logger struct {
	*zerolog.Logger
	serviceName string
}

func New(serviceName string) *Logger {
	cfg := DefaultConfig()
	cfg.ServiceName = serviceName
	return NewWithConfig(cfg)
}

func NewWithConfig(cfg *Config) *Logger {
	zerolog.TimeFieldFormat = cfg.TimeFormat
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var output io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stderr":
		output = os.Stderr
	default:
		output = os.Stdout
	}

	var logger zerolog.Logger
	if strings.ToLower(cfg.Format) == "console" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: output}).With().Timestamp().Caller().Logger()
	} else {
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

func (l *Logger) WithService() *Logger {
	newLogger := l.Logger.With().Str("service", l.serviceName).Logger()
	return &Logger{
		Logger:      &newLogger,
		serviceName: l.serviceName,
	}
}

func (l *Logger) WithTraceID(traceID string) *Logger {
	newLogger := l.Logger.With().Str("trace_id", traceID).Logger()
	return &Logger{
		Logger:      &newLogger,
		serviceName: l.serviceName,
	}
}

func TraceFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value("trace_id").(string); ok {
		return traceID
	}
	return ""
}

func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, "trace_id", traceID)
}

var globalLogger *Logger

func init() {
	globalLogger = New("global")
}

func Default() *Logger {
	return globalLogger
}

func SetGlobal(logger *Logger) {
	globalLogger = logger
}
