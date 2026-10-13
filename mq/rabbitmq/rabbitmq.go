// Package rabbitmq provides a RabbitMQ connection Provider.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(rabbitmq.New(nil))  // nil → reads from env/config.yaml
//	c.MustRun()
//
// Wire usage (Pattern B) – add rabbitmq.ProviderSet to wire.Build():
//
//	wire.Build(rabbitmq.NewConfigFromEnv, rabbitmq.NewClient, ...)
//
// After Init(), use Conn() to obtain a *amqp091.Connection for channel creation.
package rabbitmq

import (
	"fmt"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/internal/config"
	"github.com/versegeek/toolkit/provider"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds RabbitMQ connection parameters.
type Config struct {
	// URL is the AMQP connection URL, e.g. amqp://guest:guest@localhost:5672/
	URL string

	// Heartbeat is the connection heartbeat interval.
	Heartbeat time.Duration

	// ChannelMax is the max number of channels (0 = no limit).
	ChannelMax uint16
}

const (
	defaultURL        = "amqp://guest:guest@127.0.0.1:5672/"
	defaultChannelMax = 0
)

var defaultHeartbeat = 10 * time.Second

// NewConfigFromEnv reads RabbitMQ configuration from the environment.
// All keys are prefixed with RABBITMQ_.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("RABBITMQ_URL", defaultURL)
	v.SetDefault("RABBITMQ_HEARTBEAT", defaultHeartbeat)
	v.SetDefault("RABBITMQ_CHANNEL_MAX", defaultChannelMax)

	cfg := &Config{
		URL:        v.GetString("RABBITMQ_URL"),
		Heartbeat:  v.GetDuration("RABBITMQ_HEARTBEAT"),
		ChannelMax: uint16(v.GetUint("RABBITMQ_CHANNEL_MAX")),
	}

	logrus.WithField("url", maskURL(cfg.URL)).Debug("RabbitMQ config loaded")

	return cfg
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client wraps an *amqp091.Connection as a lifecycle-managed Provider.
type Client struct {
	provider.AbstractProvider

	Config *Config
	conn   *amqp.Connection
}

var _ provider.Provider = (*Client)(nil)

// New creates a Client without connecting.
// If config is nil, NewConfigFromEnv() is used.
func New(cfg *Config) *Client {
	if cfg == nil {
		cfg = NewConfigFromEnv()
	}
	return &Client{Config: cfg}
}

// NewClient is the Wire-compatible constructor.
func NewClient(cfg *Config) (*Client, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Conn returns the underlying *amqp091.Connection. Only valid after Init().
func (c *Client) Conn() *amqp.Connection { return c.conn }

// Init establishes the AMQP connection.
func (c *Client) Init() error {
	conn, err := amqp.DialConfig(c.Config.URL, amqp.Config{
		Heartbeat:  c.Config.Heartbeat,
		ChannelMax: c.Config.ChannelMax,
	})
	if err != nil {
		return fmt.Errorf("rabbitmq: failed to connect: %w", err)
	}
	c.conn = conn
	return nil
}

// Close closes the AMQP connection.
func (c *Client) Close() error {
	if c.conn == nil || c.conn.IsClosed() {
		return nil
	}
	return c.conn.Close()
}

// maskURL replaces the password in an AMQP URL for safe logging.
// Uses the last "@" to correctly handle passwords containing "@".
// amqp://user:pass@host:port/vhost → amqp://user:***@host:port/vhost
func maskURL(rawURL string) string {
	schemeEnd := strings.Index(rawURL, "://")
	if schemeEnd < 0 {
		return rawURL
	}
	authority := rawURL[schemeEnd+3:]
	atIdx := strings.LastIndex(authority, "@")
	if atIdx < 0 {
		return rawURL
	}
	userinfo := authority[:atIdx]
	colonIdx := strings.Index(userinfo, ":")
	if colonIdx < 0 {
		return rawURL
	}
	return rawURL[:schemeEnd+3] + userinfo[:colonIdx+1] + "***" + rawURL[schemeEnd+3+atIdx:]
}
