// Package mongo provides a MongoDB client (go.mongodb.org/mongo-driver/v2)
// as a provider.
//
//	m := mongo.New(nil)
//	g.MustInit(m)
//	users := m.DB().Collection("users")
package mongo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/versegeek/toolkit/internal/redact"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// Client is a MongoDB client managed as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	client *mongo.Client
	db     *mongo.Database
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

// Client returns the MongoDB client. It is nil before Init.
func (c *Client) Client() *mongo.Client {
	return c.client
}

// DB returns the database named by Config.Database, or nil when none is
// configured or before Init.
func (c *Client) DB() *mongo.Database {
	return c.db
}

// Init connects and checks connectivity within Config.Timeout.
func (c *Client) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}

	client, err := mongo.Connect(c.Config.clientOptions())
	if err != nil {
		return fmt.Errorf("mongo: connect to %s failed: %w", redact.URL(c.Config.URI), err)
	}

	ctx, cancel := c.timeoutContext()
	defer cancel()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		dctx, dcancel := c.timeoutContext()
		defer dcancel()
		return errors.Join(
			fmt.Errorf("mongo: ping %s failed: %w", redact.URL(c.Config.URI), err),
			client.Disconnect(dctx),
		)
	}

	c.client = client
	if c.Config.Database != "" {
		c.db = client.Database(c.Config.Database)
	}
	slog.Debug("mongo connected", "uri", redact.URL(c.Config.URI), "database", c.Config.Database)
	return nil
}

// Close disconnects within Config.Timeout.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	ctx, cancel := c.timeoutContext()
	defer cancel()
	return c.client.Disconnect(ctx)
}

func (c *Client) timeoutContext() (context.Context, context.CancelFunc) {
	timeout := c.Config.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return context.WithTimeout(context.Background(), timeout)
}
