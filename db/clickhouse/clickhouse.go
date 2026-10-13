// Package clickhouse provides a native ClickHouse client Provider.
//
// Uses the official clickhouse-go/v2 native protocol client, which supports
// batch inserts, columnar data, and ClickHouse-specific types unavailable
// through the standard database/sql interface.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(clickhouse.New(nil))  // nil → reads from env/config.yaml
//	c.MustRun()
//
// Wire usage (Pattern B) – add clickhouse.ProviderSet to wire.Build():
//
//	wire.Build(clickhouse.NewConfigFromEnv, clickhouse.NewClient, ...)
//
// After Init(), use Conn() for the native protocol connection:
//
//	conn := client.Conn()
//	batch, _ := conn.PrepareBatch(ctx, "INSERT INTO events")
package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	chgo "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/config"
	"github.com/versegeek/toolkit/provider"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds ClickHouse connection parameters.
type Config struct {
	// Addrs is the list of ClickHouse server addresses (host:port).
	// Native protocol default port is 9000; HTTP default is 8123.
	Addrs []string

	Username string
	Password string
	Database string

	DialTimeout     time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

const (
	defaultClickHouseAddr     = "127.0.0.1:9000"
	defaultClickHouseDB       = "default"
	defaultClickHouseUsername = "default"
	defaultMaxOpenConns       = 10
	defaultMaxIdleConns       = 5
)

var (
	defaultDialTimeout     = 30 * time.Second
	defaultConnMaxLifetime = 1 * time.Hour
)

// NewConfigFromEnv reads ClickHouse configuration from the environment.
// All keys are prefixed with CLICKHOUSE_.
//
// CLICKHOUSE_ADDRS accepts a comma-separated list of addresses, e.g.:
//
//	CLICKHOUSE_ADDRS=ch1:9000,ch2:9000
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("CLICKHOUSE_ADDRS", defaultClickHouseAddr)
	v.SetDefault("CLICKHOUSE_USERNAME", defaultClickHouseUsername)
	v.SetDefault("CLICKHOUSE_PASSWORD", "")
	v.SetDefault("CLICKHOUSE_DATABASE", defaultClickHouseDB)
	v.SetDefault("CLICKHOUSE_DIAL_TIMEOUT", defaultDialTimeout)
	v.SetDefault("CLICKHOUSE_MAX_OPEN_CONNS", defaultMaxOpenConns)
	v.SetDefault("CLICKHOUSE_MAX_IDLE_CONNS", defaultMaxIdleConns)
	v.SetDefault("CLICKHOUSE_CONN_MAX_LIFETIME", defaultConnMaxLifetime)

	cfg := &Config{
		Addrs:           splitCSV(v.GetString("CLICKHOUSE_ADDRS")),
		Username:        v.GetString("CLICKHOUSE_USERNAME"),
		Password:        v.GetString("CLICKHOUSE_PASSWORD"),
		Database:        v.GetString("CLICKHOUSE_DATABASE"),
		DialTimeout:     v.GetDuration("CLICKHOUSE_DIAL_TIMEOUT"),
		MaxOpenConns:    v.GetInt("CLICKHOUSE_MAX_OPEN_CONNS"),
		MaxIdleConns:    v.GetInt("CLICKHOUSE_MAX_IDLE_CONNS"),
		ConnMaxLifetime: v.GetDuration("CLICKHOUSE_CONN_MAX_LIFETIME"),
	}

	logrus.WithFields(logrus.Fields{
		"addrs":    cfg.Addrs,
		"database": cfg.Database,
	}).Debug("ClickHouse config loaded")

	return cfg
}

// splitCSV splits a comma-separated string into a trimmed slice,
// handling both env vars and config file values uniformly.
func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}

func (c *Config) options() *chgo.Options {
	return &chgo.Options{
		Addr: c.Addrs,
		Auth: chgo.Auth{
			Database: c.Database,
			Username: c.Username,
			Password: c.Password,
		},
		DialTimeout:     c.DialTimeout,
		MaxOpenConns:    c.MaxOpenConns,
		MaxIdleConns:    c.MaxIdleConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
	}
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client wraps a native clickhouse.Conn as a lifecycle-managed Provider.
type Client struct {
	provider.AbstractProvider

	Config *Config
	conn   driver.Conn
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

// Conn returns the underlying native clickhouse.Conn. Only valid after Init().
func (c *Client) Conn() driver.Conn { return c.conn }

// Init opens the native ClickHouse connection and verifies connectivity.
func (c *Client) Init() error {
	conn, err := chgo.Open(c.Config.options())
	if err != nil {
		return fmt.Errorf("clickhouse: failed to open connection: %w", err)
	}
	if err := conn.Ping(context.Background()); err != nil {
		return fmt.Errorf("clickhouse: ping failed: %w", err)
	}
	c.conn = conn
	return nil
}

// Close closes the ClickHouse connection.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
