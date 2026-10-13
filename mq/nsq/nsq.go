// Package nsq provides an NSQ producer as a provider.
//
// Consumers are bound to a topic and channel, so applications create them
// directly, using the lookupd addresses from the same configuration:
//
//	p := nsq.New(nil)
//	g.MustInit(p)
//	err := p.Producer().Publish("orders", data)
//
//	consumer, _ := gonsq.NewConsumer("orders", "billing", gonsq.NewConfig())
//	err = consumer.ConnectToNSQLookupds(p.Config.LookupdAddresses)
package nsq

import (
	"fmt"
	"log/slog"

	"github.com/google/wire"
	gonsq "github.com/nsqio/go-nsq"

	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// Client is an NSQ producer managed as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	producer *gonsq.Producer
}

var _ provider.Provider = (*Client)(nil)

// New returns a Client that is not connected yet.
func New(cfg *Config) *Client {
	return &Client{Config: cfg}
}

// NewClient returns a connected Client and a cleanup that closes it.
func NewClient(cfg *Config) (*Client, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Producer returns the NSQ producer. It is nil before Init.
func (c *Client) Producer() *gonsq.Producer {
	return c.producer
}

// Init creates the producer, routes its warnings and errors to slog, and
// checks that nsqd is reachable.
func (c *Client) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}

	p, err := gonsq.NewProducer(c.Config.NsqdAddress, gonsq.NewConfig())
	if err != nil {
		return fmt.Errorf("nsq: create producer failed: %w", err)
	}
	p.SetLogger(slogWriter{}, gonsq.LogLevelWarning)
	if err := p.Ping(); err != nil {
		p.Stop()
		return fmt.Errorf("nsq: ping nsqd at %s failed: %w", c.Config.NsqdAddress, err)
	}

	c.producer = p
	return nil
}

// Close stops the producer after its in-flight publishes complete.
func (c *Client) Close() error {
	if c.producer == nil {
		return nil
	}
	c.producer.Stop()
	return nil
}

// slogWriter forwards go-nsq log lines, already filtered to warnings and
// errors, to slog.
type slogWriter struct{}

func (slogWriter) Output(_ int, s string) error {
	slog.Warn(s, "component", "nsq")
	return nil
}
