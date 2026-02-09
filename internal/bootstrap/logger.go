package bootstrap

import (
	"github.com/elzestia/go-boilerplate/internal/shared/adapters/logger"
	"github.com/elzestia/go-boilerplate/internal/shared/adapters/telemetry"
	"github.com/elzestia/go-boilerplate/internal/shared/application/ports"
	"github.com/elzestia/go-boilerplate/internal/shared/config"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func initLogger(cfg *config.Config, otelProvider *telemetry.Provider) (ports.Logger, error) {
	var logProvider *sdklog.LoggerProvider
	if otelProvider != nil {
		logProvider = otelProvider.LoggerProvider()
	}
	return logger.NewZapLogger(cfg, logProvider)
}
