package grpc

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/versegeek/toolkit/config"
)

// Config holds the gRPC server settings.
type Config struct {
	// Host is the interface to listen on; empty means all interfaces.
	Host string
	// Port is the TCP port to listen on.
	Port int
	// Health registers the standard gRPC health service.
	Health bool
	// Reflection registers the server reflection service used by tools such
	// as grpcurl.
	Reflection bool
	// ShutdownTimeout bounds the graceful stop performed by Close; remaining
	// RPCs are cancelled after it.
	ShutdownTimeout time.Duration
}

const (
	defaultPort            = 3000
	defaultShutdownTimeout = 5 * time.Second
)

// NewConfigFromEnv reads the settings from GRPC_HOST, GRPC_PORT (default
// 3000), GRPC_HEALTH_ENABLED (default true), GRPC_REFLECTION_ENABLED (default
// true) and GRPC_SHUTDOWN_TIMEOUT (default 5s).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("grpc: load config failed: %w", err)
	}
	v.SetDefault("GRPC_HOST", "")
	v.SetDefault("GRPC_PORT", defaultPort)
	v.SetDefault("GRPC_HEALTH_ENABLED", true)
	v.SetDefault("GRPC_REFLECTION_ENABLED", true)
	v.SetDefault("GRPC_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)

	return &Config{
		Host:            v.GetString("GRPC_HOST"),
		Port:            v.GetInt("GRPC_PORT"),
		Health:          v.GetBool("GRPC_HEALTH_ENABLED"),
		Reflection:      v.GetBool("GRPC_REFLECTION_ENABLED"),
		ShutdownTimeout: v.GetDuration("GRPC_SHUTDOWN_TIMEOUT"),
	}, nil
}

func (c *Config) addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
