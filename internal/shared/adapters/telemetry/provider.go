package telemetry

import (
	"context"
	"fmt"

	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Provider manages OpenTelemetry providers (traces, metrics, logs)
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	loggerProvider *sdklog.LoggerProvider
	enabled        bool
	logsEnabled    bool
}

// NewProvider creates and initializes a new OpenTelemetry provider
// Returns a no-op provider if cfg.Enabled is false (zero overhead)
func NewProvider(ctx context.Context, cfg *config.OTelConfig) (*Provider, error) {
	provider := &Provider{
		enabled:     cfg.Enabled,
		logsEnabled: cfg.Logs.Enabled,
	}

	// Initialize tracer provider if enabled
	if cfg.Enabled {
		// Initialize tracer provider
		tp, err := initTracerProvider(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
		}

		// Set global tracer provider
		otel.SetTracerProvider(tp)

		// Set global propagator for distributed tracing
		// TraceContext propagates W3C Trace Context (traceparent, tracestate headers)
		// Baggage propagates additional context across service boundaries
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))

		provider.tracerProvider = tp
	}

	// Initialize logger provider if enabled
	if cfg.Logs.Enabled {
		lp, err := initLoggerProvider(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize logger provider: %w", err)
		}

		provider.loggerProvider = lp
	}

	return provider, nil
}

// Shutdown gracefully shuts down the provider, flushing any remaining spans/metrics/logs
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error

	// Shutdown logger provider first (flush logs before traces)
	if p.logsEnabled && p.loggerProvider != nil {
		if err := p.loggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("logger provider shutdown: %w", err))
		}
	}

	// Shutdown tracer provider
	if p.enabled && p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer provider shutdown: %w", err))
		}
	}

	// Return combined error if any
	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// Tracer returns a tracer for the given instrumentation name
// Returns a no-op tracer if provider is disabled
func (p *Provider) Tracer(name string) trace.Tracer {
	if !p.enabled || p.tracerProvider == nil {
		return otel.Tracer(name)
	}
	return p.tracerProvider.Tracer(name)
}

// IsEnabled returns whether OpenTelemetry is enabled
func (p *Provider) IsEnabled() bool {
	return p.enabled
}

// LoggerProvider returns the SDK logger provider (can be nil if disabled)
func (p *Provider) LoggerProvider() *sdklog.LoggerProvider {
	return p.loggerProvider
}
