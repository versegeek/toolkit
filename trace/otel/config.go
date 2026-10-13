package otel

import (
	"fmt"
	"time"

	"github.com/versegeek/toolkit/config"
)

// Config holds the tracing settings.
type Config struct {
	// ServiceName identifies the service in the tracing backend.
	ServiceName string
	// ServiceVersion is attached to every span when set.
	ServiceVersion string
	// Exporter selects where spans go: otlp (OTLP over HTTP) or stdout.
	Exporter string
	// OTLPEndpoint is the collector base URL, such as http://localhost:4318.
	// An http:// URL disables TLS.
	OTLPEndpoint string
	// SamplingRatio is the fraction of new traces sampled, from 0 to 1;
	// child spans follow their parent's decision.
	SamplingRatio float64
	// ShutdownTimeout bounds flushing the pending spans in Close.
	ShutdownTimeout time.Duration
}

const (
	defaultServiceName     = "unknown_service"
	defaultExporter        = "otlp"
	defaultOTLPEndpoint    = "http://localhost:4318"
	defaultSamplingRatio   = 1.0
	defaultShutdownTimeout = 5 * time.Second
)

// NewConfigFromEnv reads the settings from TRACE_SERVICE_NAME (default
// unknown_service), TRACE_SERVICE_VERSION, TRACE_EXPORTER (default otlp),
// TRACE_OTLP_ENDPOINT (default http://localhost:4318), TRACE_SAMPLING_RATIO
// (default 1) and TRACE_SHUTDOWN_TIMEOUT (default 5s).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("otel: load config failed: %w", err)
	}
	v.SetDefault("TRACE_SERVICE_NAME", defaultServiceName)
	v.SetDefault("TRACE_SERVICE_VERSION", "")
	v.SetDefault("TRACE_EXPORTER", defaultExporter)
	v.SetDefault("TRACE_OTLP_ENDPOINT", defaultOTLPEndpoint)
	v.SetDefault("TRACE_SAMPLING_RATIO", defaultSamplingRatio)
	v.SetDefault("TRACE_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)

	return &Config{
		ServiceName:     v.GetString("TRACE_SERVICE_NAME"),
		ServiceVersion:  v.GetString("TRACE_SERVICE_VERSION"),
		Exporter:        v.GetString("TRACE_EXPORTER"),
		OTLPEndpoint:    v.GetString("TRACE_OTLP_ENDPOINT"),
		SamplingRatio:   v.GetFloat64("TRACE_SAMPLING_RATIO"),
		ShutdownTimeout: v.GetDuration("TRACE_SHUTDOWN_TIMEOUT"),
	}, nil
}
