// Package clickhouse provides a native ClickHouse connection
// (github.com/ClickHouse/clickhouse-go/v2) as a provider, with batch inserts
// and the ClickHouse column types that database/sql cannot express. To use
// ClickHouse through GORM instead, use package db/gorm with
// DB_DRIVER=clickhouse.
//
//	ch := clickhouse.New(nil)
//	g.MustInit(ch)
//	batch, err := ch.Conn().PrepareBatch(ctx, "INSERT INTO events")
package clickhouse

import (
	"context"
	"errors"
	"fmt"

	chgo "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/wire"

	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// Client is a native ClickHouse connection managed as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	conn driver.Conn
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

// Conn returns the native connection. It is nil before Init.
func (c *Client) Conn() driver.Conn {
	return c.conn
}

// Init opens the connection and checks connectivity within
// Config.DialTimeout.
func (c *Client) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}
	if len(c.Config.Addrs) == 0 {
		return errors.New("clickhouse: no server address configured (set CLICKHOUSE_ADDRS)")
	}

	conn, err := chgo.Open(&chgo.Options{
		Addr: c.Config.Addrs,
		Auth: chgo.Auth{
			Database: c.Config.Database,
			Username: c.Config.Username,
			Password: c.Config.Password,
		},
		DialTimeout:     c.Config.DialTimeout,
		MaxOpenConns:    c.Config.MaxOpenConns,
		MaxIdleConns:    c.Config.MaxIdleConns,
		ConnMaxLifetime: c.Config.ConnMaxLifetime,
	})
	if err != nil {
		return fmt.Errorf("clickhouse: open failed: %w", err)
	}

	timeout := c.Config.DialTimeout
	if timeout <= 0 {
		timeout = defaultDialTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := conn.Ping(ctx); err != nil {
		return errors.Join(fmt.Errorf("clickhouse: ping failed: %w", err), conn.Close())
	}

	c.conn = conn
	return nil
}

// Close closes the connection.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
