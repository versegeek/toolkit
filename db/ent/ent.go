// Package ent provides an ent SQL driver as a provider.
//
// Config.Driver (DB_DRIVER) selects mysql, postgres or sqlite; other backends
// can be added with RegisterDriver. PostgreSQL uses pgx and SQLite the
// pure-Go github.com/glebarez/go-sqlite driver. After Init, pass the driver to
// the generated ent client:
//
//	db := ent.New(nil)
//	g.MustInit(db)
//	client := entgen.NewClient(entgen.Driver(db.Driver()))
package ent

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/wire"

	"github.com/versegeek/toolkit/internal/sqlconn"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// Client is an ent SQL driver managed as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	drv *entsql.Driver
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

// Driver returns the ent driver. It is nil before Init.
func (c *Client) Driver() *entsql.Driver {
	return c.drv
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

	spec, err := lookupDriver(c.Config.Driver)
	if err != nil {
		return fmt.Errorf("ent: %w", err)
	}
	settings := c.Config.settings()
	dsn, err := settings.BuildDSN()
	if err != nil {
		return fmt.Errorf("ent: build dsn failed: %w", err)
	}

	db, err := sql.Open(spec.sqlDriver, dsn)
	if err != nil {
		return fmt.Errorf("ent: open %s failed: %w", c.Config.Driver, err)
	}
	settings.ApplyPool(db, dsn)
	if err := sqlconn.Ping(db); err != nil {
		return errors.Join(fmt.Errorf("ent: ping %s failed: %w", c.Config.Driver, err), db.Close())
	}

	c.drv = entsql.OpenDB(spec.dialect, db)
	slog.Debug("ent connected", "driver", c.Config.Driver, "host", c.Config.Host, "database", c.Config.Database)
	return nil
}

// Close closes the underlying connection pool.
func (c *Client) Close() error {
	if c.drv == nil {
		return nil
	}
	return c.drv.Close()
}
