package gorm

import (
	"fmt"
	"time"

	"github.com/versegeek/toolkit/config"
	"github.com/versegeek/toolkit/internal/sqlconn"
)

// Config holds the database settings.
type Config struct {
	// Driver selects the database: mysql, postgres, sqlite, clickhouse or a
	// name registered with RegisterDialector. It is required.
	Driver string
	// DSN is the complete data source name. When set, the connection fields
	// below are ignored. SQLite takes a file path or ":memory:".
	DSN string

	// Host defaults to 127.0.0.1.
	Host string
	// Port defaults to the driver's standard port.
	Port int
	// Username defaults to root (MySQL), postgres or default (ClickHouse).
	Username string
	Password string
	// Database is the database name, or the file path for SQLite.
	Database string
	// Params are extra DSN parameters in query-string form, such as
	// "sslmode=require". MySQL defaults to charset=utf8mb4&parseTime=true&loc=Local.
	Params string

	// MaxIdleConns, MaxOpenConns and ConnMaxLifetime limit the pool; zero
	// keeps the database/sql default.
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration

	// LogLevel is the GORM log level: silent, error, warn or info.
	LogLevel string
}

// NewConfigFromEnv reads the settings from DB_DRIVER, DB_DSN, DB_HOST,
// DB_PORT, DB_USERNAME, DB_PASSWORD, DB_DATABASE, DB_PARAMS,
// DB_MAX_IDLE_CONNS (default 10), DB_MAX_OPEN_CONNS (default 100),
// DB_CONN_MAX_LIFETIME (default 1h) and DB_LOG_LEVEL (default warn).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("gorm: load config failed: %w", err)
	}
	v.SetDefault("DB_LOG_LEVEL", "warn")
	s := sqlconn.Load(v)

	return &Config{
		Driver:          s.Driver,
		DSN:             s.DSN,
		Host:            s.Host,
		Port:            s.Port,
		Username:        s.Username,
		Password:        s.Password,
		Database:        s.Database,
		Params:          s.Params,
		MaxIdleConns:    s.MaxIdleConns,
		MaxOpenConns:    s.MaxOpenConns,
		ConnMaxLifetime: s.ConnMaxLifetime,
		LogLevel:        v.GetString("DB_LOG_LEVEL"),
	}, nil
}

func (c *Config) settings() sqlconn.Settings {
	return sqlconn.Settings{
		Driver:          c.Driver,
		DSN:             c.DSN,
		Host:            c.Host,
		Port:            c.Port,
		Username:        c.Username,
		Password:        c.Password,
		Database:        c.Database,
		Params:          c.Params,
		MaxIdleConns:    c.MaxIdleConns,
		MaxOpenConns:    c.MaxOpenConns,
		ConnMaxLifetime: c.ConnMaxLifetime,
	}
}
