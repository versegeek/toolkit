package redis

import (
	"fmt"
	"time"

	"github.com/versegeek/toolkit/config"
)

// Config holds the Redis connection settings. Zero values keep the go-redis
// defaults.
type Config struct {
	// Addrs are the server addresses (host:port). Client needs exactly one;
	// ClusterClient uses them as cluster seeds.
	Addrs []string
	// DB selects the database; ClusterClient ignores it.
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
	defaultPoolSize     = 50
	defaultMinIdleConns = 5
	defaultMaxRetries   = 3
	defaultDialTimeout  = 3 * time.Second
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 5 * time.Second
)

// NewConfigFromEnv reads the settings from REDIS_ADDRESS (comma-separated for
// a cluster, default 127.0.0.1:6379), REDIS_DB, REDIS_USERNAME,
// REDIS_PASSWORD, REDIS_POOL_SIZE (default 50), REDIS_MIN_IDLE_CONNS (default
// 5), REDIS_MAX_RETRIES (default 3), REDIS_DIAL_TIMEOUT (default 3s),
// REDIS_READ_TIMEOUT (default 5s) and REDIS_WRITE_TIMEOUT (default 5s).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("redis: load config failed: %w", err)
	}
	v.SetDefault("REDIS_ADDRESS", defaultAddress)
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_USERNAME", "")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_POOL_SIZE", defaultPoolSize)
	v.SetDefault("REDIS_MIN_IDLE_CONNS", defaultMinIdleConns)
	v.SetDefault("REDIS_MAX_RETRIES", defaultMaxRetries)
	v.SetDefault("REDIS_DIAL_TIMEOUT", defaultDialTimeout)
	v.SetDefault("REDIS_READ_TIMEOUT", defaultReadTimeout)
	v.SetDefault("REDIS_WRITE_TIMEOUT", defaultWriteTimeout)

	return &Config{
		Addrs:        config.StringSlice(v, "REDIS_ADDRESS"),
		DB:           v.GetInt("REDIS_DB"),
		Username:     v.GetString("REDIS_USERNAME"),
		Password:     v.GetString("REDIS_PASSWORD"),
		PoolSize:     v.GetInt("REDIS_POOL_SIZE"),
		MinIdleConns: v.GetInt("REDIS_MIN_IDLE_CONNS"),
		MaxRetries:   v.GetInt("REDIS_MAX_RETRIES"),
		DialTimeout:  v.GetDuration("REDIS_DIAL_TIMEOUT"),
		ReadTimeout:  v.GetDuration("REDIS_READ_TIMEOUT"),
		WriteTimeout: v.GetDuration("REDIS_WRITE_TIMEOUT"),
	}, nil
}
