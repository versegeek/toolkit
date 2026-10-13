// Package redis provides go-redis clients as providers: Client for a single
// server and ClusterClient for a Redis Cluster.
//
//	rdb := redis.New(nil)
//	g.MustInit(rdb)
//	err := rdb.Client().Set(ctx, "key", "value", 0).Err()
package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// pingTimeout bounds the connectivity check in Init, retries included.
const pingTimeout = 10 * time.Second

// Client is a single-server Redis client managed as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	client *redis.Client
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

// Client returns the go-redis client. It is nil before Init.
func (c *Client) Client() *redis.Client {
	return c.client
}

// Init connects and checks connectivity with a PING.
func (c *Client) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}
	if n := len(c.Config.Addrs); n != 1 {
		return fmt.Errorf("redis: Client needs exactly one address, got %d (use ClusterClient for a cluster)", n)
	}

	client := redis.NewClient(&redis.Options{
		Addr:         c.Config.Addrs[0],
		DB:           c.Config.DB,
		Username:     c.Config.Username,
		Password:     c.Config.Password,
		PoolSize:     c.Config.PoolSize,
		MinIdleConns: c.Config.MinIdleConns,
		MaxRetries:   c.Config.MaxRetries,
		DialTimeout:  c.Config.DialTimeout,
		ReadTimeout:  c.Config.ReadTimeout,
		WriteTimeout: c.Config.WriteTimeout,
	})
	if err := ping(client); err != nil {
		return errors.Join(fmt.Errorf("redis: ping %s failed: %w", c.Config.Addrs[0], err), client.Close())
	}

	c.client = client
	return nil
}

// Close closes the client.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

func ping(client redis.UniversalClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	return client.Ping(ctx).Err()
}
