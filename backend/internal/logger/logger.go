package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Suthar345Piyush/goprodbackend/internal/config"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
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

	// adding debug logging only if explicitly enabled

	if cfg.NewRelic.DebugLogging {
		configOptions = append(configOptions, newrelic.ConfigDebugLogger(os.Stdout))
	}

	app, err := newrelic.NewApplication(configOptions...)

	if err != nil {
		fmt.Printf("failed to initialize New relic: %v\n", err)
		return service
	}

	service.nrApp = app
	fmt.Printf("New Relic initialized for app: %s\n", cfg.ServiceName)

	return service

}

//Shutdown , shuts down the New Relic service

func (ls *LoggerService) Shutdown() {

	if ls.nrApp != nil {
		ls.nrApp.Shutdown(10 * time.Second)
	}

}

// GetApplication returns the new relic application instance , which can be called outside the package

func (ls *LoggerService) GetApplication() *newrelic.Application {
	return ls.nrApp
}

// new logger creates a new logger with specified level and environment (backward compatibility)

func NewLogger(level string, isProd bool) zerolog.Logger {
	return NewLoggerWithService(&config.ObservabilityConfig{
		Logging: config.LoggingConfig{
			Level: level,
		},
		Environment: func() string {
			if isProd {
				return "production"
			}
			return "development"
		}(),
	}, nil)
}

// NewLoggerWithConfig creates a logger with full config (backward compatibility)

func NewLoggerWithConfig(cfg *config.ObservabilityConfig) zerolog.Logger {
	return NewLoggerWithService(cfg, nil)
}

// NewLoggerWithService creates a logger with full config and logger service

func NewLoggerWithService(cfg *config.ObservabilityConfig, loggerService *LoggerService) zerolog.Logger {

	var logLevel zerolog.Level

	level := cfg.GetLogLevel()

	switch level {

	case "debug":
		logLevel = zerolog.DebugLevel

	case "info":
		logLevel = zerolog.InfoLevel

	case "warn":
		logLevel = zerolog.WarnLevel

	case "error":
		logLevel = zerolog.ErrorLevel

	default:
		logLevel = zerolog.InfoLevel
	}

	// don't set global level - let each logger have it's own level
	// the ErrorStackMarshaler extracts stack from err , if their any present , (stack trace marshaling)

	zerolog.TimeFieldFormat = "2006-01-02 15:04:05"
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var writer io.Writer

	// setup base writer

	var baseWriter io.Writer

	if cfg.IsProduction() && cfg.Logging.Format == "json" {

		// in production , write to stdout

		baseWriter = os.Stdout

		// wrap with new relic zerologWriter for log forwarding in production

		if loggerService != nil && loggerService.nrApp != nil {
			nrWriter := zerologWriter.New(baseWriter, loggerService.nrApp)
			writer = nrWriter
		} else {
			writer = baseWriter
		}
	} else {

		// dev mode - using console writer

		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}
		writer = consoleWriter

	}

	// New Relic log forwarding is now handled automatically by zerologWriter integration

	logger := zerolog.New(writer).Level(logLevel).With().Timestamp().Str("service", cfg.ServiceName).Str("environment", cfg.Environment).Logger()

	// including stack traces for errors in development

	if !cfg.IsProduction() {
		logger = logger.With().Stack().Logger()
	}

	return logger

}
