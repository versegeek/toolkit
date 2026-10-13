// Package nats provides a NATS connection, optionally with JetStream, as a
// provider.
//
//	nc := nats.New(nil)
//	g.MustInit(nc)
//	err := nc.Conn().Publish("orders", data)
//
// Close drains the connection, so in-flight messages of active subscriptions
// are processed before it returns.
package nats

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/wire"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/versegeek/toolkit/internal/redact"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// Client is a NATS connection managed as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	conn   *nats.Conn
	js     jetstream.JetStream
	closed chan struct{}
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

// Conn returns the NATS connection. It is nil before Init.
func (c *Client) Conn() *nats.Conn {
	return c.conn
}

// JetStream returns the JetStream context, or nil unless
// Config.EnableJetStream is set and Init succeeded.
func (c *Client) JetStream() jetstream.JetStream {
	return c.js
}

// Init connects and, when enabled, creates the JetStream context.
func (c *Client) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}

	closed := make(chan struct{})
	var closeOnce sync.Once
	opts := []nats.Option{
		nats.MaxReconnects(c.Config.MaxReconnects),
		nats.ClosedHandler(func(*nats.Conn) { closeOnce.Do(func() { close(closed) }) }),
	}
	if c.Config.Name != "" {
		opts = append(opts, nats.Name(c.Config.Name))
	}
	if c.Config.ConnectTimeout > 0 {
		opts = append(opts, nats.Timeout(c.Config.ConnectTimeout))
	}
	if c.Config.ReconnectWait > 0 {
		opts = append(opts, nats.ReconnectWait(c.Config.ReconnectWait))
	}
	if c.Config.DrainTimeout > 0 {
		opts = append(opts, nats.DrainTimeout(c.Config.DrainTimeout))
	}

	conn, err := nats.Connect(c.Config.URL, opts...)
	if err != nil {
		return fmt.Errorf("nats: connect to %s failed: %w", redact.URL(c.Config.URL), err)
	}
	if c.Config.EnableJetStream {
		js, err := jetstream.New(conn)
		if err != nil {
			conn.Close()
			return fmt.Errorf("nats: create JetStream context failed: %w", err)
		}
		c.js = js
	}

	c.conn = conn
	c.closed = closed
	return nil
}

// Close drains the connection and waits until it is closed, at most
// Config.DrainTimeout plus a second.
func (c *Client) Close() error {
	if c.conn == nil || c.conn.IsClosed() {
		return nil
	}
	if err := c.conn.Drain(); err != nil {
		c.conn.Close()
		return fmt.Errorf("nats: drain failed: %w", err)
	}

	wait := c.Config.DrainTimeout
	if wait <= 0 {
		wait = nats.GetDefaultOptions().DrainTimeout
	}
	timer := time.NewTimer(wait + time.Second)
	defer timer.Stop()
	select {
	case <-c.closed:
		return nil
	case <-timer.C:
		c.conn.Close()
		return errors.New("nats: drain did not finish in time")
	}
}
