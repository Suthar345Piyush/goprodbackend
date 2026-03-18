// observability on the application in real time
// using new relic for observability and health checks
// using New Relic go APM (Application Performance monitoring)

package config

import (
	"fmt"
	"time"
)

// configuration for observability

type ObservabilityConfig struct {
	ServiceName  string             `koanf:"service_name" validate:"required"`
	Environment  string             `koanf:"environment" validate:"required"`
	Logging      LoggingConfig      `koanf:"logging" validate:"required"`
	NewRelic     NewRelicConfig     `koanf:"new_relic" validate:"required"`
	HealthChecks HealthChecksConfig `koanf:"health_checks" validate:"required"`
}

// logging config

type LoggingConfig struct {
	Level              string        `koanf:"level" validate:"required"`
	Format             string        `koanf:"format" validate:"required"`
	SlowQueryThreshold time.Duration `koanf:"slow_query_threshold"`
}

// new relic tool config

type NewRelicConfig struct {
	LicenseKey                string `koanf:"license_key" validate:"required"`
	AppLogForwardingEnabled   bool   `koanf:"app_log_forwarding_enabled"`
	DistributedTracingEnabled bool   `koanf:"distributed_tracing_enabled"`
	DebugLogging              bool   `koanf:"debug_logging"`
}

// health checks config

type HealthChecksConfig struct {
	Enabled  bool          `koanf:"enabled"`
	Interval time.Duration `koanf:"interval" validate:"min=1s"`
	Timeout  time.Duration `koanf:"timeout" validate:"min=1s"`
	Checks   []string      `koanf:"checks"`
}

// function for observability  , validation and log level

// function for observability config , setting it to default , taking struct parameters from  *ObservabilityConfig

func DefaultObservabilityConfig() *ObservabilityConfig {

	return &ObservabilityConfig{
		ServiceName: "boilerplate",
		Environment: "development",
		Logging: LoggingConfig{
			Level:              "info",
			Format:             "json",
			SlowQueryThreshold: 100 * time.Millisecond,
		},

		NewRelic: NewRelicConfig{
			LicenseKey:                "",
			AppLogForwardingEnabled:   true,
			DistributedTracingEnabled: true,
			DebugLogging:              false,
		},

		HealthChecks: HealthChecksConfig{
			Enabled:  true,
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
			Checks:   []string{"database", "redis"},
		},
	}
}

// validation function on the observability config

func (c *ObservabilityConfig) Validate() error {

	if c.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	//validating the log level

	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}

	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid logging level: %s (must be one of them: debug, info, warn, error)", c.Logging.Level)
	}

	// validating slow query threshold , which should be positive

	if c.Logging.SlowQueryThreshold < 0 {
		return fmt.Errorf("logging slow_query_threshold must be positive")
	}

	return nil

}

// getting log level function returning the log level  (production -> info , development -> debug)

func (c *ObservabilityConfig) GetLogLevel() string {

	switch c.Environment {

	case "production":
		if c.Logging.Level == "" {
			return "info"
		}

	case "development":
		if c.Logging.Level == "" {
			return "debug"
		}
	}

	return c.Logging.Level
}

//if environment is in the production

func (c *ObservabilityConfig) IsProduction() bool {
	return c.Environment == "production"
}
