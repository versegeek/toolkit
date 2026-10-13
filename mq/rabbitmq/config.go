package rabbitmq

import (
	"fmt"
	"time"

	"github.com/versegeek/toolkit/config"
)

// Config holds the RabbitMQ connection settings.
type Config struct {
	// URL is the AMQP URL, such as amqp://user:pass@host:5672/vhost.
	URL string
	// Heartbeat is the connection heartbeat interval.
	Heartbeat time.Duration
	// ChannelMax limits the channels per connection; zero means the server
	// maximum.
	ChannelMax uint16
	// DialTimeout bounds the TCP connection and handshake in Init.
	DialTimeout time.Duration
}

const (
	defaultURL         = "amqp://guest:guest@127.0.0.1:5672/"
	defaultHeartbeat   = 10 * time.Second
	defaultDialTimeout = 30 * time.Second
)

// NewConfigFromEnv reads the settings from RABBITMQ_URL (default
// amqp://guest:guest@127.0.0.1:5672/), RABBITMQ_HEARTBEAT (default 10s),
// RABBITMQ_CHANNEL_MAX (default 0) and RABBITMQ_DIAL_TIMEOUT (default 30s).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: load config failed: %w", err)
	}
	v.SetDefault("RABBITMQ_URL", defaultURL)
	v.SetDefault("RABBITMQ_HEARTBEAT", defaultHeartbeat)
	v.SetDefault("RABBITMQ_CHANNEL_MAX", 0)
	v.SetDefault("RABBITMQ_DIAL_TIMEOUT", defaultDialTimeout)

	return &Config{
		URL:         v.GetString("RABBITMQ_URL"),
		Heartbeat:   v.GetDuration("RABBITMQ_HEARTBEAT"),
		ChannelMax:  v.GetUint16("RABBITMQ_CHANNEL_MAX"),
		DialTimeout: v.GetDuration("RABBITMQ_DIAL_TIMEOUT"),
	}, nil
}
