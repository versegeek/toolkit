// Package rabbitmq provides a RabbitMQ (AMQP 0-9-1) connection as a
// provider.
//
//	mq := rabbitmq.New(nil)
//	g.MustInit(mq)
//	ch, err := mq.Conn().Channel()
package rabbitmq

import (
	"fmt"

	"github.com/google/wire"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/versegeek/toolkit/internal/redact"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// Client is a RabbitMQ connection managed as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	conn *amqp.Connection
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

// Conn returns the AMQP connection for opening channels. It is nil before
// Init.
func (c *Client) Conn() *amqp.Connection {
	return c.conn
}

// Init opens the connection.
func (c *Client) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}

	amqpConfig := amqp.Config{
		Heartbeat:  c.Config.Heartbeat,
		ChannelMax: c.Config.ChannelMax,
	}
	if c.Config.DialTimeout > 0 {
		amqpConfig.Dial = amqp.DefaultDial(c.Config.DialTimeout)
	}
	conn, err := amqp.DialConfig(c.Config.URL, amqpConfig)
	if err != nil {
		return fmt.Errorf("rabbitmq: connect to %s failed: %w", redact.URL(c.Config.URL), err)
	}

	c.conn = conn
	return nil
}

// Close closes the connection and every channel opened on it.
func (c *Client) Close() error {
	if c.conn == nil || c.conn.IsClosed() {
		return nil
	}
	return c.conn.Close()
}
