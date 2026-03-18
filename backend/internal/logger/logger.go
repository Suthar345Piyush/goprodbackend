package logger

import (
	"fmt"

	"github.com/Suthar345Piyush/goprodbackend/internal/config"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// logger service manages new relic integration and logger creation

type LoggerService struct {
	nrApp *newrelic.Application
}

// NewLoggerService creates a new logger service with new relic integration

func NewLoggerService(cfg *config.ObservabilityConfig) *LoggerService {

	service := &LoggerService{}

	// checking license key

	if cfg.NewRelic.LicenseKey == "" {
		fmt.Printf("New Relic license key not provided, skipping initialization")

		return service
	}

	// setting up new relic config from observability

	var configOptions = []newrelic.ConfigOption

	configOptions = append(configOptions,
		newrelic.ConfigAppName(cfg.ServiceName),
		newrelic.ConfigLicense(cfg.NewRelic.LicenseKey),
		newrelic.ConfigAppLogForwardingEnabled(cfg.NewRelic.AppLogForwardingEnabled),
		newrelic.ConfigDistributedTracerEnabled(cfg.NewRelic.DistributedTracingEnabled),
	)

}
