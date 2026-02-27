package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zlog zerolog.Logger
}

func New(serviceName string) *Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	zlog := zerolog.New(output).With().Timestamp().Str("service", serviceName).Logger()
	return &Logger{zlog: zlog}
}

func (l *Logger) Info() *zerolog.Event {
	return l.zlog.Info()
}

func (l *Logger) Error() *zerolog.Event {
	return l.zlog.Error()
}

func (l *Logger) Debug() *zerolog.Event {
	return l.zlog.Debug()
}

func (l *Logger) Warn() *zerolog.Event {
	return l.zlog.Warn()
}

func (l *Logger) Fatal() *zerolog.Event {
	return l.zlog.Fatal()
}

func (l *Logger) WithTraceID(ctx context.Context) *Logger {
	traceID, ok := ctx.Value("trace_id").(string)
	if !ok {
		return l
	}
	newLogger := l.zlog.With().Str("trace_id", traceID).Logger()
	return &Logger{zlog: newLogger}
}
