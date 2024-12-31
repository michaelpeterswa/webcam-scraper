package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type WebcamScraperConfig struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"error"`

	CronSchedule   string            `env:"CRON_SCHEDULE" envDefault:"@every 30m"`
	OutputImageDir string            `env:"OUTPUT_IMAGE_DIR" envDefault:"/data"`
	Domains        map[string]string `env:"DOMAINS,required"`

	WSDOTAPIKey string `env:"WSDOT_API_KEY"`

	MetricsEnabled bool `env:"METRICS_ENABLED" envDefault:"true"`
	MetricsPort    int  `env:"METRICS_PORT" envDefault:"8081"`

	TracingEnabled    bool    `env:"TRACING_ENABLED" envDefault:"false"`
	TracingSampleRate float64 `env:"TRACING_SAMPLERATE" envDefault:"0.01"`
	TracingService    string  `env:"TRACING_SERVICE" envDefault:"katalog-agent"`
	TracingVersion    string  `env:"TRACING_VERSION"`
}

func NewConfig() (*WebcamScraperConfig, error) {
	var cfg WebcamScraperConfig

	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}
