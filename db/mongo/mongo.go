// Package mongo provides a MongoDB client Provider.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(mongo.New(nil))  // nil → reads from env/config.yaml
//	c.MustRun()
//
// Wire usage (Pattern B) – add mongo.ProviderSet to wire.Build():
//
//	wire.Build(mongo.NewConfigFromEnv, mongo.NewClient, ...)
package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/internal/config"
	"github.com/versegeek/toolkit/provider"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds the MongoDB connection parameters.
type Config struct {
	URI                    string
	Database               string
	Timeout                time.Duration
	MaxPoolSize            uint64
	MinPoolSize            uint64
	MaxConnIdleTime        time.Duration
	HeartbeatInterval      time.Duration
	ServerSelectionTimeout time.Duration
}

const (
	defaultURI                = "mongodb://localhost:27017"
	defaultDatabase           = ""
	defaultMaxPoolSize uint64 = 16
	defaultMinPoolSize uint64 = 1
)

var (
	defaultTimeout                = 20 * time.Second
	defaultMaxConnIdleTime        = 30 * time.Second
	defaultHeartbeatInterval      = 15 * time.Second
	defaultServerSelectionTimeout = 10 * time.Second
)

// Exported for testing.
var (
	DefaultTimeout                = defaultTimeout
	DefaultServerSelectionTimeout = defaultServerSelectionTimeout
)

// MaskURI is exported for testing. See maskURI for details.
var MaskURI = maskURI

// NewConfigFromEnv reads MongoDB configuration from the environment.
// All keys are prefixed with MONGODB_.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("MONGODB_URI", defaultURI)
	v.SetDefault("MONGODB_DATABASE", defaultDatabase)
	v.SetDefault("MONGODB_TIMEOUT", defaultTimeout)
	v.SetDefault("MONGODB_MAX_POOL_SIZE", defaultMaxPoolSize)
	v.SetDefault("MONGODB_MIN_POOL_SIZE", defaultMinPoolSize)
	v.SetDefault("MONGODB_MAX_CONN_IDLE_TIME", defaultMaxConnIdleTime)
	v.SetDefault("MONGODB_HEARTBEAT_INTERVAL", defaultHeartbeatInterval)
	v.SetDefault("MONGODB_SERVER_SELECTION_TIMEOUT", defaultServerSelectionTimeout)

	cfg := &Config{
		URI:                    v.GetString("MONGODB_URI"),
		Database:               v.GetString("MONGODB_DATABASE"),
		Timeout:                v.GetDuration("MONGODB_TIMEOUT"),
		MaxPoolSize:            v.GetUint64("MONGODB_MAX_POOL_SIZE"),
		MinPoolSize:            v.GetUint64("MONGODB_MIN_POOL_SIZE"),
		MaxConnIdleTime:        v.GetDuration("MONGODB_MAX_CONN_IDLE_TIME"),
		HeartbeatInterval:      v.GetDuration("MONGODB_HEARTBEAT_INTERVAL"),
		ServerSelectionTimeout: v.GetDuration("MONGODB_SERVER_SELECTION_TIMEOUT"),
	}

	logrus.WithFields(logrus.Fields{
		"uri":      maskURI(cfg.URI),
		"database": cfg.Database,
	}).Debug("MongoDB config loaded")

	return cfg
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client wraps a *mongo.Client as a lifecycle-managed Provider.
type Client struct {
	provider.AbstractProvider

	Config *Config
	client *mongo.Client
	db     *mongo.Database
}

var _ provider.Provider = (*Client)(nil)

// New creates a Client without connecting.
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

// Client returns the underlying *mongo.Client. Only valid after Init().
func (c *Client) Client() *mongo.Client { return c.client }

// DB returns the default *mongo.Database (Config.Database). Only valid after Init().
func (c *Client) DB() *mongo.Database { return c.db }

// Init connects to MongoDB and verifies connectivity with a ping.
func (c *Client) Init() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Config.Timeout)
	defer cancel()

	opts := options.Client().
		ApplyURI(c.Config.URI).
		SetMaxPoolSize(c.Config.MaxPoolSize).
		SetMinPoolSize(c.Config.MinPoolSize).
		SetMaxConnIdleTime(c.Config.MaxConnIdleTime).
		SetHeartbeatInterval(c.Config.HeartbeatInterval).
		SetServerSelectionTimeout(c.Config.ServerSelectionTimeout)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return fmt.Errorf("mongo: connect failed: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx) // clean up background topology goroutines
		return fmt.Errorf("mongo: ping failed: %w", err)
	}

	c.client = client
	if c.Config.Database != "" {
		c.db = client.Database(c.Config.Database)
	}

	return nil
}

// Close disconnects from MongoDB.
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), c.Config.Timeout)
	defer cancel()
	return c.client.Disconnect(ctx)
}

// maskURI replaces the password in a MongoDB URI for safe logging.
// mongodb://user:pass@host:port/db → mongodb://user:***@host:port/db
func maskURI(uri string) string {
	// Find "://" to skip the scheme.
	schemeEnd := strings.Index(uri, "://")
	if schemeEnd < 0 {
		return uri
	}
	authority := uri[schemeEnd+3:]
	// Find the last "@" to locate the userinfo boundary (handles "@" in passwords).
	atIdx := strings.LastIndex(authority, "@")
	if atIdx < 0 {
		return uri // no credentials
	}
	userinfo := authority[:atIdx]
	// Replace everything after the colon in userinfo with ***.
	colonIdx := strings.Index(userinfo, ":")
	if colonIdx < 0 {
		return uri // no password
	}
	masked := uri[:schemeEnd+3] + userinfo[:colonIdx+1] + "***" + uri[schemeEnd+3+atIdx:]
	return masked
}
