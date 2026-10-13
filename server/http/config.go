package http

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/versegeek/toolkit/config"
)

// Config holds the HTTP server settings.
type Config struct {
	// Host is the interface to listen on; empty means all interfaces.
	Host string
	// Port is the TCP port to listen on.
	Port int
	// Mode is the Gin mode: debug, release or test.
	Mode string
	// ShutdownTimeout bounds the graceful shutdown performed by Close.
	ShutdownTimeout time.Duration
}

const (
	defaultPort            = 8080
	defaultMode            = gin.ReleaseMode
	defaultShutdownTimeout = 5 * time.Second
)

// NewConfigFromEnv reads the settings from HTTP_HOST, HTTP_PORT (default
// 8080), HTTP_MODE (default release) and HTTP_SHUTDOWN_TIMEOUT (default 5s).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("http: load config failed: %w", err)
	}
	v.SetDefault("HTTP_HOST", "")
	v.SetDefault("HTTP_PORT", defaultPort)
	v.SetDefault("HTTP_MODE", defaultMode)
	v.SetDefault("HTTP_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)

	return &Config{
		Host:            v.GetString("HTTP_HOST"),
		Port:            v.GetInt("HTTP_PORT"),
		Mode:            v.GetString("HTTP_MODE"),
		ShutdownTimeout: v.GetDuration("HTTP_SHUTDOWN_TIMEOUT"),
	}, nil
}

func (c *Config) addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}
