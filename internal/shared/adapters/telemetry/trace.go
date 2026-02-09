package telemetry

import (
	"context"
	"fmt"

	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// initTracerProvider creates and configures the SDK tracer provider
func initTracerProvider(ctx context.Context, cfg *config.OTelConfig) (*sdktrace.TracerProvider, error) {
	// Create resource with service information
	res, err := NewResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create span exporter based on configuration
	exporter, err := createSpanExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create span exporter: %w", err)
	}

	// Configure sampler
	// ParentBased respects upstream sampling decisions
	// TraceIDRatioBased samples based on configured rate
	sampler := sdktrace.ParentBased(
		sdktrace.TraceIDRatioBased(cfg.SampleRate),
	)

	// Create tracer provider with batch processor for performance
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	return tp, nil
}

// createSpanExporter creates the appropriate span exporter based on configuration
func createSpanExporter(ctx context.Context, cfg *config.OTelConfig) (sdktrace.SpanExporter, error) {
	switch cfg.ExporterType {
	case "otlp":
		return createOTLPExporter(ctx, cfg)
	case "console":
		return createConsoleExporter()
	default:
		return nil, fmt.Errorf("unsupported exporter type: %s (supported: otlp, console)", cfg.ExporterType)
	}
}

// createOTLPExporter creates an OTLP/gRPC span exporter
// Compatible with Jaeger, Tempo, and standard OTel collectors
func createOTLPExporter(ctx context.Context, cfg *config.OTelConfig) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
	}

	// Use insecure credentials for local development
	if cfg.OTLPInsecure {
		opts = append(opts,
			otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()),
			otlptracegrpc.WithDialOption(grpc.WithBlock()),
		)
	}

	return otlptracegrpc.New(ctx, opts...)
}

// createConsoleExporter creates a console (stdout) span exporter
// Useful for local development and debugging
func createConsoleExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
}
