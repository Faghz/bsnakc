package telemetry

import (
	"github.com/elzestia/go-boilerplate/internal/shared/config"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// NewResource creates a resource with service information using semantic conventions
// This resource is shared across all telemetry signals (traces, logs)
func NewResource(cfg *config.OTelConfig) (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
}
