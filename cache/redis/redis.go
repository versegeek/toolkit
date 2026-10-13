// Package redis provides a Redis client Provider.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(redis.New(nil))  // nil → reads from env/config.yaml
//	c.MustRun()
//
// Wire usage (Pattern B) – add redis.ProviderSet to wire.Build():
//
//	wire.Build(redis.NewConfigFromEnv, redis.NewClient, ...)
package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/config"
	"github.com/versegeek/toolkit/provider"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds the Redis connection parameters.
type Config struct {
	Address      string
	DB           int
	Username     string
	Password     string
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

const (
	defaultAddress      = "127.0.0.1:6379"
	defaultDB           = 0
	defaultUsername     = ""
	defaultPassword     = ""
	defaultPoolSize     = 50
	defaultMinIdleConns = 5
	defaultMaxRetries   = 3
)

var (
	defaultDialTimeout  = 3 * time.Second
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 5 * time.Second
)

// NewConfigFromEnv reads Redis configuration from environment variables
// (or a config.yaml file). All keys are prefixed with REDIS_.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("REDIS_ADDRESS", defaultAddress)
	v.SetDefault("REDIS_DB", defaultDB)
	v.SetDefault("REDIS_USERNAME", defaultUsername)
	v.SetDefault("REDIS_PASSWORD", defaultPassword)
	v.SetDefault("REDIS_POOL_SIZE", defaultPoolSize)
	v.SetDefault("REDIS_MIN_IDLE_CONNS", defaultMinIdleConns)
	v.SetDefault("REDIS_MAX_RETRIES", defaultMaxRetries)
	v.SetDefault("REDIS_DIAL_TIMEOUT", defaultDialTimeout)
	v.SetDefault("REDIS_READ_TIMEOUT", defaultReadTimeout)
	v.SetDefault("REDIS_WRITE_TIMEOUT", defaultWriteTimeout)

	cfg := &Config{
		Address:      v.GetString("REDIS_ADDRESS"),
		DB:           v.GetInt("REDIS_DB"),
		Username:     v.GetString("REDIS_USERNAME"),
		Password:     v.GetString("REDIS_PASSWORD"),
		PoolSize:     v.GetInt("REDIS_POOL_SIZE"),
		MinIdleConns: v.GetInt("REDIS_MIN_IDLE_CONNS"),
		MaxRetries:   v.GetInt("REDIS_MAX_RETRIES"),
		DialTimeout:  v.GetDuration("REDIS_DIAL_TIMEOUT"),
		ReadTimeout:  v.GetDuration("REDIS_READ_TIMEOUT"),
		WriteTimeout: v.GetDuration("REDIS_WRITE_TIMEOUT"),
	}

	logrus.WithFields(logrus.Fields{
		"address":  cfg.Address,
		"db":       cfg.DB,
		"poolSize": cfg.PoolSize,
	}).Debug("Redis config loaded")

	return cfg
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client wraps a go-redis client as a lifecycle-managed Provider.
type Client struct {
	provider.AbstractProvider

	Config *Config
	client *redis.Client
}

// Ensure compile-time interface satisfaction.
var _ provider.Provider = (*Client)(nil)

// New creates a Client (does not connect yet).
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

// Client returns the underlying *redis.Client. Only valid after Init().
func (c *Client) Client() *redis.Client { return c.client }

// Init creates and pings the Redis connection.
func (c *Client) Init() error {
	c.client = redis.NewClient(&redis.Options{
		Addr:         c.Config.Address,
		Username:     c.Config.Username,
		Password:     c.Config.Password,
		DB:           c.Config.DB,
		MaxRetries:   c.Config.MaxRetries,
		DialTimeout:  c.Config.DialTimeout,
		ReadTimeout:  c.Config.ReadTimeout,
		WriteTimeout: c.Config.WriteTimeout,
		PoolSize:     c.Config.PoolSize,
		MinIdleConns: c.Config.MinIdleConns,
	})

	return c.client.Ping(context.Background()).Err()
}

// Close releases the Redis connection.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}
