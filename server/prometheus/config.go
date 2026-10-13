package prometheus

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/versegeek/toolkit/config"
)

// Config holds the metrics server settings.
type Config struct {
	// Enabled switches the server on.
	Enabled bool
	// Host is the interface to listen on; empty means all interfaces.
	Host string
	// Port is the TCP port to listen on.
	Port int
	// Endpoint is the URL path of the metrics.
	Endpoint string
	// ShutdownTimeout bounds the graceful shutdown performed by Close.
	ShutdownTimeout time.Duration
}

const (
	defaultPort            = 9090
	defaultEndpoint        = "/metrics"
	defaultShutdownTimeout = 5 * time.Second
)

// NewConfigFromEnv reads the settings from PROMETHEUS_ENABLED (default true),
// PROMETHEUS_HOST, PROMETHEUS_PORT (default 9090), PROMETHEUS_ENDPOINT
// (default /metrics) and PROMETHEUS_SHUTDOWN_TIMEOUT (default 5s).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("prometheus: load config failed: %w", err)
	}
	v.SetDefault("PROMETHEUS_ENABLED", true)
	v.SetDefault("PROMETHEUS_HOST", "")
	v.SetDefault("PROMETHEUS_PORT", defaultPort)
	v.SetDefault("PROMETHEUS_ENDPOINT", defaultEndpoint)
	v.SetDefault("PROMETHEUS_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)

	return &Config{
		Enabled:         v.GetBool("PROMETHEUS_ENABLED"),
		Host:            v.GetString("PROMETHEUS_HOST"),
		Port:            v.GetInt("PROMETHEUS_PORT"),
		Endpoint:        v.GetString("PROMETHEUS_ENDPOINT"),
		ShutdownTimeout: v.GetDuration("PROMETHEUS_SHUTDOWN_TIMEOUT"),
	}, nil
}

func (c *Config) addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
