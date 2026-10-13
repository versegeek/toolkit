package redis

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/wire"
	"github.com/redis/go-redis/v9"

	"github.com/versegeek/toolkit/provider"
)

// ClusterProviderSet provides a *Config from the environment and an
// initialised *ClusterClient. Use it instead of ProviderSet.
var ClusterProviderSet = wire.NewSet(NewConfigFromEnv, NewClusterClient)

// ClusterClient is a Redis Cluster client managed as a provider.
type ClusterClient struct {
	// Config holds the settings; Init loads them from the environment when
	// nil. Config.Addrs are the cluster seed nodes.
	Config *Config

	client *redis.ClusterClient
}

var _ provider.Provider = (*ClusterClient)(nil)

// NewCluster returns a ClusterClient that is not connected yet.
func NewCluster(cfg *Config) *ClusterClient {
	return &ClusterClient{Config: cfg}
}

// NewClusterClient returns a connected ClusterClient and a cleanup that
// closes it.
func NewClusterClient(cfg *Config) (*ClusterClient, func(), error) {
	return provider.InitProvider(NewCluster(cfg))
}

// Client returns the go-redis cluster client. It is nil before Init.
func (c *ClusterClient) Client() *redis.ClusterClient {
	return c.client
}

// Init connects to the cluster and checks connectivity with a PING.
func (c *ClusterClient) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}
	if len(c.Config.Addrs) == 0 {
		return errors.New("redis: ClusterClient needs at least one address (set REDIS_ADDRESS)")
	}

	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        c.Config.Addrs,
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
		return errors.Join(
			fmt.Errorf("redis: ping cluster %s failed: %w", strings.Join(c.Config.Addrs, ","), err),
			client.Close(),
		)
	}

	c.client = client
	return nil
}

// Close closes the client.
func (c *ClusterClient) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}
