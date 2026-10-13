package clickhouse

import (
	"fmt"
	"time"

	"github.com/versegeek/toolkit/config"
)

// Config holds the ClickHouse connection settings.
type Config struct {
	// Addrs are the server addresses (host:port) of the native protocol,
	// whose default port is 9000.
	Addrs    []string
	Username string
	Password string
	Database string
	// DialTimeout bounds each connection attempt and the ping in Init.
	DialTimeout     time.Duration
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

const (
	defaultAddr            = "127.0.0.1:9000"
	defaultUsername        = "default"
	defaultDatabase        = "default"
	defaultDialTimeout     = 10 * time.Second
	defaultMaxOpenConns    = 10
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = time.Hour
)

// NewConfigFromEnv reads the settings from CLICKHOUSE_ADDRS (comma-separated,
// default 127.0.0.1:9000), CLICKHOUSE_USERNAME (default "default"),
// CLICKHOUSE_PASSWORD, CLICKHOUSE_DATABASE (default "default"),
// CLICKHOUSE_DIAL_TIMEOUT (default 10s), CLICKHOUSE_MAX_OPEN_CONNS (default
// 10), CLICKHOUSE_MAX_IDLE_CONNS (default 5) and CLICKHOUSE_CONN_MAX_LIFETIME
// (default 1h).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("clickhouse: load config failed: %w", err)
	}
	v.SetDefault("CLICKHOUSE_ADDRS", defaultAddr)
	v.SetDefault("CLICKHOUSE_USERNAME", defaultUsername)
	v.SetDefault("CLICKHOUSE_PASSWORD", "")
	v.SetDefault("CLICKHOUSE_DATABASE", defaultDatabase)
	v.SetDefault("CLICKHOUSE_DIAL_TIMEOUT", defaultDialTimeout)
	v.SetDefault("CLICKHOUSE_MAX_OPEN_CONNS", defaultMaxOpenConns)
	v.SetDefault("CLICKHOUSE_MAX_IDLE_CONNS", defaultMaxIdleConns)
	v.SetDefault("CLICKHOUSE_CONN_MAX_LIFETIME", defaultConnMaxLifetime)

	return &Config{
		Addrs:           config.StringSlice(v, "CLICKHOUSE_ADDRS"),
		Username:        v.GetString("CLICKHOUSE_USERNAME"),
		Password:        v.GetString("CLICKHOUSE_PASSWORD"),
		Database:        v.GetString("CLICKHOUSE_DATABASE"),
		DialTimeout:     v.GetDuration("CLICKHOUSE_DIAL_TIMEOUT"),
		MaxOpenConns:    v.GetInt("CLICKHOUSE_MAX_OPEN_CONNS"),
		MaxIdleConns:    v.GetInt("CLICKHOUSE_MAX_IDLE_CONNS"),
		ConnMaxLifetime: v.GetDuration("CLICKHOUSE_CONN_MAX_LIFETIME"),
	}, nil
}
