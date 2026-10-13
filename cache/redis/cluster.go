package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/versegeek/toolkit/provider"
)

// ClusterClient wraps a go-redis ClusterClient as a lifecycle-managed Provider.
type ClusterClient struct {
	provider.AbstractProvider

	Config *Config
	client *redis.ClusterClient
}

var _ provider.Provider = (*ClusterClient)(nil)

// NewCluster creates a ClusterClient (does not connect yet).
// If config is nil, NewConfigFromEnv() is used.
func NewCluster(cfg *Config) *ClusterClient {
	if cfg == nil {
		cfg = NewConfigFromEnv()
	}
	return &ClusterClient{Config: cfg}
}

// NewClusterClient is the Wire-compatible constructor.
func NewClusterClient(cfg *Config) (*ClusterClient, func(), error) {
	return provider.InitProvider(NewCluster(cfg))
}

// Client returns the underlying *redis.ClusterClient. Only valid after Init().
func (c *ClusterClient) Client() *redis.ClusterClient { return c.client }

// Init creates and pings the cluster connection.
func (c *ClusterClient) Init() error {
	c.client = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        []string{c.Config.Address},
		Username:     c.Config.Username,
		Password:     c.Config.Password,
		MaxRetries:   c.Config.MaxRetries,
		DialTimeout:  c.Config.DialTimeout,
		ReadTimeout:  c.Config.ReadTimeout,
		WriteTimeout: c.Config.WriteTimeout,
		PoolSize:     c.Config.PoolSize,
		MinIdleConns: c.Config.MinIdleConns,
	})

	return c.client.Ping(context.Background()).Err()
}

// Close releases the cluster connection.
func (c *ClusterClient) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}
