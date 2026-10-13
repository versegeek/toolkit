// Package gorm provides a GORM database handle as a provider.
//
// Config.Driver (DB_DRIVER) selects mysql, postgres, sqlite or clickhouse;
// other databases can be added with RegisterDialector. SQLite uses the
// pure-Go github.com/glebarez/sqlite driver, so no cgo is needed.
//
//	db := gorm.New(&gorm.Config{Driver: "sqlite", DSN: "app.db"})
//	g.MustInit(db)
//	db.Client().AutoMigrate(&User{})
//
// GORM logs through the slog default logger at Init, so initialise the log
// provider first.
package gorm

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/wire"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/versegeek/toolkit/internal/sqlconn"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// Client is a GORM database handle managed as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	db *gorm.DB
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

// Client returns the GORM handle. It is nil before Init.
func (c *Client) Client() *gorm.DB {
	return c.db
}

// Init opens the database, applies the pool limits and checks connectivity.
func (c *Client) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}

	open, err := dialector(c.Config.Driver)
	if err != nil {
		return fmt.Errorf("gorm: %w", err)
	}
	level, err := logLevel(c.Config.LogLevel)
	if err != nil {
		return err
	}
	settings := c.Config.settings()
	dsn, err := settings.BuildDSN()
	if err != nil {
		return fmt.Errorf("gorm: build dsn failed: %w", err)
	}

	db, err := gorm.Open(open(dsn), &gorm.Config{
		Logger: logger.NewSlogLogger(slog.Default(), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  level,
		}),
		DisableAutomaticPing: true,
	})
	if err != nil {
		return fmt.Errorf("gorm: open %s failed: %w", c.Config.Driver, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("gorm: get sql.DB failed: %w", err)
	}
	settings.ApplyPool(sqlDB, dsn)
	if err := sqlconn.Ping(sqlDB); err != nil {
		return errors.Join(fmt.Errorf("gorm: ping %s failed: %w", c.Config.Driver, err), sqlDB.Close())
	}

	c.db = db
	slog.Debug("gorm connected", "driver", c.Config.Driver, "host", c.Config.Host, "database", c.Config.Database)
	return nil
}

// Close closes the underlying connection pool.
func (c *Client) Close() error {
	if c.db == nil {
		return nil
	}
	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("gorm: get sql.DB failed: %w", err)
	}
	return sqlDB.Close()
}

func logLevel(name string) (logger.LogLevel, error) {
	switch name {
	case "silent":
		return logger.Silent, nil
	case "error":
		return logger.Error, nil
	case "warn", "":
		return logger.Warn, nil
	case "info":
		return logger.Info, nil
	default:
		return 0, fmt.Errorf("gorm: invalid log level %q: want silent, error, warn or info", name)
	}
}
