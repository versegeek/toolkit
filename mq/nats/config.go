package nats

import (
	"fmt"
	"time"

	"github.com/versegeek/toolkit/config"
)

// Config holds the NATS connection settings.
type Config struct {
	// URL is the server URL; separate several with commas for failover, as
	// in nats://a:4222,nats://b:4222.
	URL string
	// Name identifies the connection in the server's monitoring.
	Name string
	// ConnectTimeout bounds each connection attempt.
	ConnectTimeout time.Duration
	// ReconnectWait is the pause between reconnection attempts.
	ReconnectWait time.Duration
	// MaxReconnects limits reconnection attempts; -1 means no limit.
	MaxReconnects int
	// DrainTimeout bounds the drain performed by Close.
	DrainTimeout time.Duration
	// EnableJetStream creates a JetStream context during Init.
	EnableJetStream bool
}

const (
	defaultURL            = "nats://127.0.0.1:4222"
	defaultConnectTimeout = 5 * time.Second
	defaultReconnectWait  = 2 * time.Second
	defaultMaxReconnects  = -1
	defaultDrainTimeout   = 30 * time.Second
)

// NewConfigFromEnv reads the settings from NATS_URL (default
// nats://127.0.0.1:4222), NATS_NAME, NATS_CONNECT_TIMEOUT (default 5s),
// NATS_RECONNECT_WAIT (default 2s), NATS_MAX_RECONNECTS (default -1),
// NATS_DRAIN_TIMEOUT (default 30s) and NATS_ENABLE_JETSTREAM (default false).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("nats: load config failed: %w", err)
	}
	v.SetDefault("NATS_URL", defaultURL)
	v.SetDefault("NATS_NAME", "")
	v.SetDefault("NATS_CONNECT_TIMEOUT", defaultConnectTimeout)
	v.SetDefault("NATS_RECONNECT_WAIT", defaultReconnectWait)
	v.SetDefault("NATS_MAX_RECONNECTS", defaultMaxReconnects)
	v.SetDefault("NATS_DRAIN_TIMEOUT", defaultDrainTimeout)
	v.SetDefault("NATS_ENABLE_JETSTREAM", false)

	return &Config{
		URL:             v.GetString("NATS_URL"),
		Name:            v.GetString("NATS_NAME"),
		ConnectTimeout:  v.GetDuration("NATS_CONNECT_TIMEOUT"),
		ReconnectWait:   v.GetDuration("NATS_RECONNECT_WAIT"),
		MaxReconnects:   v.GetInt("NATS_MAX_RECONNECTS"),
		DrainTimeout:    v.GetDuration("NATS_DRAIN_TIMEOUT"),
		EnableJetStream: v.GetBool("NATS_ENABLE_JETSTREAM"),
	}, nil
}
