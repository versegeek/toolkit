package mongo

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/versegeek/toolkit/config"
)

// Config holds the MongoDB connection settings.
type Config struct {
	// URI is the connection string, such as mongodb://user:pass@host:27017.
	URI string
	// Database names the database returned by Client.DB; empty means none.
	Database string
	// Timeout bounds the connectivity check in Init and the disconnect in
	// Close.
	Timeout time.Duration
	// The pool and timing settings below keep the driver default when zero.
	MaxPoolSize            uint64
	MinPoolSize            uint64
	MaxConnIdleTime        time.Duration
	HeartbeatInterval      time.Duration
	ServerSelectionTimeout time.Duration
}

const (
	defaultURI                    = "mongodb://localhost:27017"
	defaultTimeout                = 10 * time.Second
	defaultMaxPoolSize            = 100
	defaultMinPoolSize            = 0
	defaultMaxConnIdleTime        = 30 * time.Second
	defaultHeartbeatInterval      = 10 * time.Second
	defaultServerSelectionTimeout = 10 * time.Second
)

// NewConfigFromEnv reads the settings from MONGODB_URI (default
// mongodb://localhost:27017), MONGODB_DATABASE, MONGODB_TIMEOUT (default 10s),
// MONGODB_MAX_POOL_SIZE (default 100), MONGODB_MIN_POOL_SIZE (default 0),
// MONGODB_MAX_CONN_IDLE_TIME (default 30s), MONGODB_HEARTBEAT_INTERVAL
// (default 10s) and MONGODB_SERVER_SELECTION_TIMEOUT (default 10s).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("mongo: load config failed: %w", err)
	}
	v.SetDefault("MONGODB_URI", defaultURI)
	v.SetDefault("MONGODB_DATABASE", "")
	v.SetDefault("MONGODB_TIMEOUT", defaultTimeout)
	v.SetDefault("MONGODB_MAX_POOL_SIZE", defaultMaxPoolSize)
	v.SetDefault("MONGODB_MIN_POOL_SIZE", defaultMinPoolSize)
	v.SetDefault("MONGODB_MAX_CONN_IDLE_TIME", defaultMaxConnIdleTime)
	v.SetDefault("MONGODB_HEARTBEAT_INTERVAL", defaultHeartbeatInterval)
	v.SetDefault("MONGODB_SERVER_SELECTION_TIMEOUT", defaultServerSelectionTimeout)

	return &Config{
		URI:                    v.GetString("MONGODB_URI"),
		Database:               v.GetString("MONGODB_DATABASE"),
		Timeout:                v.GetDuration("MONGODB_TIMEOUT"),
		MaxPoolSize:            v.GetUint64("MONGODB_MAX_POOL_SIZE"),
		MinPoolSize:            v.GetUint64("MONGODB_MIN_POOL_SIZE"),
		MaxConnIdleTime:        v.GetDuration("MONGODB_MAX_CONN_IDLE_TIME"),
		HeartbeatInterval:      v.GetDuration("MONGODB_HEARTBEAT_INTERVAL"),
		ServerSelectionTimeout: v.GetDuration("MONGODB_SERVER_SELECTION_TIMEOUT"),
	}, nil
}

// clientOptions applies URI and the non-zero settings.
func (c *Config) clientOptions() *options.ClientOptions {
	opts := options.Client().ApplyURI(c.URI)
	if c.MaxPoolSize > 0 {
		opts.SetMaxPoolSize(c.MaxPoolSize)
	}
	if c.MinPoolSize > 0 {
		opts.SetMinPoolSize(c.MinPoolSize)
	}
	if c.MaxConnIdleTime > 0 {
		opts.SetMaxConnIdleTime(c.MaxConnIdleTime)
	}
	if c.HeartbeatInterval > 0 {
		opts.SetHeartbeatInterval(c.HeartbeatInterval)
	}
	if c.ServerSelectionTimeout > 0 {
		opts.SetServerSelectionTimeout(c.ServerSelectionTimeout)
	}
	return opts
}
