package telemetry

import (
	"context"
	"fmt"

	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// initLoggerProvider creates and configures the SDK logger provider
func initLoggerProvider(ctx context.Context, cfg *config.OTelConfig) (*sdklog.LoggerProvider, error) {
	// Create resource with service information
	res, err := NewResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create log exporter based on configuration
	exporter, err := createLogExporter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	// Create logger provider with batch processor for performance
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	return lp, nil
}

// createLogExporter creates the appropriate log exporter based on configuration
func createLogExporter(ctx context.Context, cfg *config.OTelConfig) (sdklog.Exporter, error) {
	switch cfg.Logs.ExporterType {
	case "otlp":
		return createOTLPLogExporter(ctx, cfg)
	case "console":
		return createConsoleLogExporter()
	default:
		return nil, fmt.Errorf("unsupported log exporter type: %s (supported: otlp, console)", cfg.Logs.ExporterType)
	}
}

// createOTLPLogExporter creates an OTLP/gRPC log exporter
// Compatible with SigNoz and standard OTel collectors
func createOTLPLogExporter(ctx context.Context, cfg *config.OTelConfig) (sdklog.Exporter, error) {
	endpoint := cfg.GetLogsEndpoint()
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(endpoint),
	}

	// Use insecure credentials for local development
	if cfg.OTLPInsecure {
		opts = append(opts,
			otlploggrpc.WithTLSCredentials(insecure.NewCredentials()),
			otlploggrpc.WithDialOption(grpc.WithBlock()),
		)
	}

	return otlploggrpc.New(ctx, opts...)
}

// createConsoleLogExporter creates a console (stdout) log exporter
// Useful for local development and debugging
func createConsoleLogExporter() (sdklog.Exporter, error) {
	return stdoutlog.New(
		stdoutlog.WithPrettyPrint(),
	)
}
