// Package nats provides a NATS (+ JetStream) connection Provider.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(nats.New(nil))  // nil → reads from env/config.yaml
//	c.MustRun()
//
// Wire usage (Pattern B) – add nats.ProviderSet to wire.Build():
//
//	wire.Build(nats.NewConfigFromEnv, nats.NewClient, ...)
//
// After Init(), use Conn() for raw NATS access and JetStream() for
// JetStream-enabled operations. JetStream() returns nil if the server
// does not support JetStream or EnableJetStream is false.
package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/internal/config"
	"github.com/versegeek/toolkit/provider"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds NATS connection parameters.
type Config struct {
	// URL is the NATS server URL, e.g. nats://localhost:4222
	// Multiple URLs can be comma-separated for cluster failover.
	URL string

	// Name is the optional client name shown in NATS server logs.
	Name string

	// ConnectTimeout is the timeout for the initial connection.
	ConnectTimeout time.Duration

	// ReconnectWait is the time to wait between reconnect attempts.
	ReconnectWait time.Duration

	// MaxReconnects is the maximum number of reconnect attempts (-1 = unlimited).
	MaxReconnects int

	// EnableJetStream controls whether JetStream context is initialized on Init.
	EnableJetStream bool
}

const (
	defaultNATSURL       = "nats://127.0.0.1:4222"
	defaultMaxReconnects = -1 // unlimited
)

var (
	defaultConnectTimeout = 5 * time.Second
	defaultReconnectWait  = 2 * time.Second
)

// NewConfigFromEnv reads NATS configuration from the environment.
// All keys are prefixed with NATS_.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("NATS_URL", defaultNATSURL)
	v.SetDefault("NATS_NAME", "")
	v.SetDefault("NATS_CONNECT_TIMEOUT", defaultConnectTimeout)
	v.SetDefault("NATS_RECONNECT_WAIT", defaultReconnectWait)
	v.SetDefault("NATS_MAX_RECONNECTS", defaultMaxReconnects)
	v.SetDefault("NATS_ENABLE_JETSTREAM", false)

	cfg := &Config{
		URL:             v.GetString("NATS_URL"),
		Name:            v.GetString("NATS_NAME"),
		ConnectTimeout:  v.GetDuration("NATS_CONNECT_TIMEOUT"),
		ReconnectWait:   v.GetDuration("NATS_RECONNECT_WAIT"),
		MaxReconnects:   v.GetInt("NATS_MAX_RECONNECTS"),
		EnableJetStream: v.GetBool("NATS_ENABLE_JETSTREAM"),
	}

	logrus.WithField("url", cfg.URL).Debug("NATS config loaded")

	return cfg
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client wraps a *nats.Conn (and optionally a jetstream.JetStream context)
// as a lifecycle-managed Provider.
type Client struct {
	provider.AbstractProvider

	Config *Config
	conn   *nats.Conn
	js     jetstream.JetStream
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

// Conn returns the underlying *nats.Conn. Only valid after Init().
func (c *Client) Conn() *nats.Conn { return c.conn }

// JetStream returns the JetStream context. Only non-nil after Init() when
// EnableJetStream is true and the server supports JetStream.
func (c *Client) JetStream() jetstream.JetStream { return c.js }

// Init establishes the NATS connection and optionally initializes JetStream.
func (c *Client) Init() error {
	opts := []nats.Option{
		nats.Timeout(c.Config.ConnectTimeout),
		nats.ReconnectWait(c.Config.ReconnectWait),
		nats.MaxReconnects(c.Config.MaxReconnects),
	}
	if c.Config.Name != "" {
		opts = append(opts, nats.Name(c.Config.Name))
	}

	conn, err := nats.Connect(c.Config.URL, opts...)
	if err != nil {
		return fmt.Errorf("nats: failed to connect: %w", err)
	}

	if c.Config.EnableJetStream {
		js, err := jetstream.New(conn)
		if err != nil {
			// Close the connection to avoid leaking it.
			_ = conn.Drain()
			return fmt.Errorf("nats: failed to create JetStream context: %w", err)
		}
		c.js = js
	}

	// Assign only after all initialisation succeeds.
	c.conn = conn
	return nil
}

// Close drains and closes the NATS connection.
func (c *Client) Close() error {
	if c.conn == nil || c.conn.IsClosed() {
		return nil
	}
	return c.conn.Drain()
}
