package bootstrap

import (
	"context"
	"log"
	"strings"

	"github.com/elzestia/go-boilerplate/internal/shared/adapters/telemetry"
	"github.com/elzestia/go-boilerplate/internal/shared/config"
)

// initOTel initializes the OpenTelemetry provider
// Returns a no-op provider if OTel is disabled (zero overhead)
func initOTel(cfg *config.Config) (*telemetry.Provider, error) {
	if !cfg.OTel.Enabled && !cfg.OTel.Logs.Enabled {
		log.Println("OpenTelemetry is disabled")
		return telemetry.NewProvider(context.Background(), &cfg.OTel)
	}

	var features []string
	if cfg.OTel.Enabled {
		features = append(features, "traces")
	}
	if cfg.OTel.Logs.Enabled {
		features = append(features, "logs")
	}

	log.Printf("Initializing OpenTelemetry (%s, service: %s)...",
		strings.Join(features, "+"), cfg.OTel.ServiceName)

	provider, err := telemetry.NewProvider(context.Background(), &cfg.OTel)
	if err != nil {
		return nil, err
	}

	if cfg.OTel.Enabled {
		log.Printf("OpenTelemetry traces initialized (exporter: %s, sampling: %.0f%%)",
			cfg.OTel.ExporterType, cfg.OTel.SampleRate*100)
	}
	if cfg.OTel.Logs.Enabled {
		log.Printf("OpenTelemetry logs initialized (exporter: %s, endpoint: %s)",
			cfg.OTel.Logs.ExporterType, cfg.OTel.GetLogsEndpoint())
	}

	return provider, nil
}
