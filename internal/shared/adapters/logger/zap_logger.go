package logger

import (
	"context"
	"os"
	"time"

	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger is the Zap adapter implementation of the Logger port.
// It wraps zap.Logger and converts port Fields to zap.Fields.
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates a new Zap logger adapter based on the environment.
// If logProvider is non-nil, logs will be exported to OpenTelemetry in addition to stdout.
// This factory replaces the previous internal/shared/logger.NewLogger().
func NewZapLogger(cfg *config.Config, logProvider *sdklog.LoggerProvider) (ports.Logger, error) {
	var zapConfig zap.Config

	if cfg.Server.Env == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level - priority: LOG_LEVEL env var, then environment-based default
	if cfg.Server.LogLevel != "" {
		level, err := zapcore.ParseLevel(cfg.Server.LogLevel)
		if err != nil {
			// Invalid level, fall back to environment-based default
			if cfg.Server.Env == "production" {
				zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
			} else {
				zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
			}
		} else {
			zapConfig.Level = zap.NewAtomicLevelAt(level)
		}
	} else {
		// No LOG_LEVEL set, use environment-based default
		if cfg.Server.Env == "production" {
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		} else {
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		}
	}

	// Build stdout core
	encoder := zapcore.NewJSONEncoder(zapConfig.EncoderConfig)
	if cfg.Server.Env != "production" {
		encoder = zapcore.NewConsoleEncoder(zapConfig.EncoderConfig)
	}
	stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapConfig.Level)

	// Build OTel core if provider available
	var core zapcore.Core
	if logProvider != nil {
		otelCore := otelzap.NewCore(cfg.OTel.ServiceName,
			otelzap.WithLoggerProvider(logProvider),
		)
		// Tee both cores - logs go to both stdout and OTel
		core = zapcore.NewTee(stdoutCore, otelCore)
	} else {
		// Only stdout when OTel disabled
		core = stdoutCore
	}

	// Build logger with composed core
	zapOptions := []zap.Option{
		zap.AddCallerSkip(1), // Skip the adapter wrapper to show actual caller
	}

	// Add automatic stack trace capture if LOG_STACK_LEVEL is configured
	if cfg.Server.LogStackLevel != "" {
		stackLevel, err := zapcore.ParseLevel(cfg.Server.LogStackLevel)
		if err == nil {
			zapOptions = append(zapOptions, zap.AddStacktrace(stackLevel))
		}
		// If parsing fails, silently ignore - no automatic stack traces
	}

	logger := zap.New(core, zapOptions...)

	return &ZapLogger{logger: logger}, nil
}

// Debug logs at debug level with structured fields.
func (l *ZapLogger) Debug(ctx context.Context, msg string, fields ...ports.Field) {
	zapFields := l.convertFields(ctx, fields)
	l.logger.Debug(msg, zapFields...)
}

// Info logs at info level with structured fields.
func (l *ZapLogger) Info(ctx context.Context, msg string, fields ...ports.Field) {
	zapFields := l.convertFields(ctx, fields)
	l.logger.Info(msg, zapFields...)
}

// Warn logs at warn level with structured fields.
func (l *ZapLogger) Warn(ctx context.Context, msg string, fields ...ports.Field) {
	zapFields := l.convertFields(ctx, fields)
	l.logger.Warn(msg, zapFields...)
}

// Error logs at error level with structured fields.
func (l *ZapLogger) Error(ctx context.Context, msg string, fields ...ports.Field) {
	zapFields := l.convertFields(ctx, fields)
	l.logger.Error(msg, zapFields...)
}

// With creates a child logger with persistent fields attached.
func (l *ZapLogger) With(fields ...ports.Field) ports.Logger {
	zapFields := l.convertFields(context.Background(), fields)
	return &ZapLogger{
		logger: l.logger.With(zapFields...),
	}
}

// convertFields converts port Fields to zap.Fields and extracts context metadata.
// Automatically extracts trace_id and span_id from OpenTelemetry context for log correlation.
func (l *ZapLogger) convertFields(ctx context.Context, fields []ports.Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields)+2) // +2 for potential trace_id and span_id

	// Extract trace context from OpenTelemetry (if active)
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		spanCtx := span.SpanContext()
		zapFields = append(zapFields,
			zap.String("trace_id", spanCtx.TraceID().String()),
			zap.String("span_id", spanCtx.SpanID().String()),
		)
	}

	for _, f := range fields {
		zapFields = append(zapFields, l.convertField(f))
	}

	return zapFields
}

// convertField converts a single port Field to zap.Field.
func (l *ZapLogger) convertField(f ports.Field) zap.Field {
	switch f.Type {
	case ports.StringType:
		return zap.String(f.Key, f.Value.(string))
	case ports.IntType:
		return zap.Int(f.Key, f.Value.(int))
	case ports.Int64Type:
		return zap.Int64(f.Key, f.Value.(int64))
	case ports.Float64Type:
		return zap.Float64(f.Key, f.Value.(float64))
	case ports.BoolType:
		return zap.Bool(f.Key, f.Value.(bool))
	case ports.DurationType:
		// Handle both time.Duration and other duration types
		if d, ok := f.Value.(time.Duration); ok {
			return zap.Duration(f.Key, d)
		}
		return zap.Any(f.Key, f.Value)
	case ports.TimeType:
		// Handle both time.Time and other time types
		if t, ok := f.Value.(time.Time); ok {
			return zap.Time(f.Key, t)
		}
		return zap.Any(f.Key, f.Value)
	case ports.ErrorType:
		if err, ok := f.Value.(error); ok && err != nil {
			return zap.Error(err)
		}
		return zap.Skip()
	case ports.StackTraceType:
		// Capture and return stack trace when explicitly requested
		return zap.Stack(f.Key)
	case ports.AnyType:
		return zap.Any(f.Key, f.Value)
	default:
		return zap.Any(f.Key, f.Value)
	}
}

// Deprecated: Use NewZapLogger instead. This function is kept for backward compatibility
// during migration but will be removed in a future version.
func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	adapter, err := NewZapLogger(cfg, nil)
	if err != nil {
		return nil, err
	}
	// This is a hack to extract the underlying zap.Logger from the adapter
	// Only use this during migration - all new code should use ports.Logger interface
	return adapter.(*ZapLogger).logger, nil
}
