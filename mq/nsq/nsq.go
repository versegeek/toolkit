// Package nsq provides an NSQ producer Provider.
//
// NSQ separates publishing (Producer → nsqd) from consuming
// (Consumer → lookupd/nsqd). This package manages the Producer lifecycle.
// Consumers are topic+channel specific; create them directly using the
// LookupdAddresses from Config:
//
//	consumer, _ := gonsq.NewConsumer("topic", "channel", gonsq.NewConfig())
//	consumer.ConnectToNSQLookupds(client.Config.LookupdAddresses)
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(nsq.New(nil))  // nil → reads from env/config.yaml
//	c.MustRun()
//
// Wire usage (Pattern B) – add nsq.ProviderSet to wire.Build():
//
//	wire.Build(nsq.NewConfigFromEnv, nsq.NewClient, ...)
package nsq

import (
	"fmt"
	"strings"

	gonsq "github.com/nsqio/go-nsq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/internal/config"
	"github.com/versegeek/toolkit/provider"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds NSQ connection parameters.
type Config struct {
	// NsqdAddress is the nsqd TCP address for the Producer (host:port).
	NsqdAddress string

	// LookupdAddresses are the nsqlookupd HTTP addresses used for consumers.
	// Stored here for convenience; use when constructing gonsq.Consumer directly.
	LookupdAddresses []string
}

const defaultNsqdAddress = "127.0.0.1:4150"

// NewConfigFromEnv reads NSQ configuration from the environment.
//
//   - NSQ_NSQD_ADDRESS:      nsqd address for the producer (default 127.0.0.1:4150)
//   - NSQ_LOOKUPD_ADDRESSES: comma-separated list of nsqlookupd HTTP addresses
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("NSQ_NSQD_ADDRESS", defaultNsqdAddress)
	v.SetDefault("NSQ_LOOKUPD_ADDRESSES", "")

	cfg := &Config{
		NsqdAddress:      v.GetString("NSQ_NSQD_ADDRESS"),
		LookupdAddresses: splitCSV(v.GetString("NSQ_LOOKUPD_ADDRESSES")),
	}

	logrus.WithFields(logrus.Fields{
		"nsqd":    cfg.NsqdAddress,
		"lookupd": cfg.LookupdAddresses,
	}).Debug("NSQ config loaded")

	return cfg
}

// splitCSV splits a comma-separated string into a trimmed slice.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client wraps an *nsq.Producer as a lifecycle-managed Provider.
type Client struct {
	provider.AbstractProvider

	Config   *Config
	producer *gonsq.Producer
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

// Producer returns the underlying *nsq.Producer. Only valid after Init().
func (c *Client) Producer() *gonsq.Producer { return c.producer }

// Init creates the NSQ producer and pings nsqd.
func (c *Client) Init() error {
	p, err := gonsq.NewProducer(c.Config.NsqdAddress, gonsq.NewConfig())
	if err != nil {
		return fmt.Errorf("nsq: failed to create producer: %w", err)
	}
	if err := p.Ping(); err != nil {
		p.Stop() // clean up router goroutine started by NewProducer/connect
		return fmt.Errorf("nsq: failed to ping nsqd at %s: %w", c.Config.NsqdAddress, err)
	}
	c.producer = p
	return nil
}

// Close stops the NSQ producer.
func (c *Client) Close() error {
	if c.producer == nil {
		return nil
	}
	c.producer.Stop()
	return nil
}
