package pprof

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/versegeek/toolkit/config"
)

// Config holds the profiling server settings.
type Config struct {
	// Enabled switches the server on.
	Enabled bool
	// Host is the interface to listen on. It defaults to 127.0.0.1 because the
	// profiles expose process internals; set it to "" or 0.0.0.0 to listen on
	// all interfaces.
	Host string
	// Port is the TCP port to listen on.
	Port int
	// Endpoint is the URL path prefix of the profiles.
	Endpoint string
	// ShutdownTimeout bounds the graceful shutdown performed by Close.
	ShutdownTimeout time.Duration
}

const (
	defaultHost            = "127.0.0.1"
	defaultPort            = 9999
	defaultEndpoint        = "/debug/pprof"
	defaultShutdownTimeout = 5 * time.Second
)

// NewConfigFromEnv reads the settings from PPROF_ENABLED (default true),
// PPROF_HOST (default 127.0.0.1), PPROF_PORT (default 9999), PPROF_ENDPOINT
// (default /debug/pprof) and PPROF_SHUTDOWN_TIMEOUT (default 5s).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("pprof: load config failed: %w", err)
	}
	v.SetDefault("PPROF_ENABLED", true)
	v.SetDefault("PPROF_HOST", defaultHost)
	v.SetDefault("PPROF_PORT", defaultPort)
	v.SetDefault("PPROF_ENDPOINT", defaultEndpoint)
	v.SetDefault("PPROF_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)

	return &Config{
		Enabled:         v.GetBool("PPROF_ENABLED"),
		Host:            v.GetString("PPROF_HOST"),
		Port:            v.GetInt("PPROF_PORT"),
		Endpoint:        v.GetString("PPROF_ENDPOINT"),
		ShutdownTimeout: v.GetDuration("PPROF_SHUTDOWN_TIMEOUT"),
	}, nil
}

func (c *Config) addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// endpoint returns Endpoint with exactly one leading and no trailing slash.
func (c *Config) endpoint() string {
	return "/" + strings.Trim(c.Endpoint, "/")
}
