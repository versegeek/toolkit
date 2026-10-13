// Package ent provides an ent.io driver Provider.
//
// Supports MySQL, PostgreSQL, and SQLite via the DB_DRIVER config key.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(ent.New(nil))  // nil → reads from env/config.yaml
//	c.MustRun()
//
// Wire usage (Pattern B) – add ent.ProviderSet to wire.Build():
//
//	wire.Build(ent.NewConfigFromEnv, ent.NewClient, ...)
//
// Driver selection: set DB_DRIVER to "mysql" (default), "postgres", or "sqlite".
//   - MySQL:    uses MYSQL_* environment variables
//   - PostgreSQL: uses POSTGRES_* environment variables
//   - SQLite:   uses SQLITE_DSN (file path or ":memory:")
//
// After Init(), pass the driver to your generated ent client:
//
//	entClient := ent.NewClient(ent.Driver(client.Driver()))
package ent

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/config"
	"github.com/versegeek/toolkit/provider"
	_ "modernc.org/sqlite"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds connection parameters for the ent driver.
// Set Driver to choose the database backend; defaults to "mysql".
type Config struct {
	// Driver selects the database backend: "mysql", "postgres", or "sqlite".
	Driver string

	// DSN is an optional full connection string. When set it overrides all
	// individual host/port/user fields for the selected driver.
	DSN string

	// MySQL / PostgreSQL shared fields
	Host     string
	Port     int
	Username string
	Password string
	Database string

	// MySQL-specific
	Charset   string
	ParseTime bool
	Loc       string

	// PostgreSQL-specific
	SSLMode string

	// Connection pool (not used for sqlite)
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// NewConfigFromEnv reads database configuration from the environment.
//
// DB_DRIVER selects the backend (default "mysql").
// MySQL keys are prefixed with MYSQL_, PostgreSQL keys with POSTGRES_,
// and SQLite uses SQLITE_DSN. Keys are shared with the gorm provider.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("DB_DRIVER", "sqlite")

	// MySQL defaults
	v.SetDefault("MYSQL_HOST", "127.0.0.1")
	v.SetDefault("MYSQL_PORT", 3306)
	v.SetDefault("MYSQL_USERNAME", "root")
	v.SetDefault("MYSQL_PASSWORD", "")
	v.SetDefault("MYSQL_DATABASE", "")
	v.SetDefault("MYSQL_CHARSET", "utf8mb4")
	v.SetDefault("MYSQL_PARSE_TIME", true)
	v.SetDefault("MYSQL_LOC", "Local")
	v.SetDefault("MYSQL_MAX_IDLE_CONNS", 10)
	v.SetDefault("MYSQL_MAX_OPEN_CONNS", 100)
	v.SetDefault("MYSQL_CONN_MAX_LIFETIME", 1*time.Hour)

	// PostgreSQL defaults
	v.SetDefault("POSTGRES_HOST", "127.0.0.1")
	v.SetDefault("POSTGRES_PORT", 5432)
	v.SetDefault("POSTGRES_USERNAME", "postgres")
	v.SetDefault("POSTGRES_PASSWORD", "")
	v.SetDefault("POSTGRES_DATABASE", "")
	v.SetDefault("POSTGRES_SSLMODE", "disable")
	v.SetDefault("POSTGRES_MAX_IDLE_CONNS", 10)
	v.SetDefault("POSTGRES_MAX_OPEN_CONNS", 100)
	v.SetDefault("POSTGRES_CONN_MAX_LIFETIME", 1*time.Hour)

	driver := v.GetString("DB_DRIVER")

	cfg := &Config{Driver: driver}

	switch driver {
	case "postgres":
		cfg.DSN = v.GetString("POSTGRES_DSN")
		cfg.Host = v.GetString("POSTGRES_HOST")
		cfg.Port = v.GetInt("POSTGRES_PORT")
		cfg.Username = v.GetString("POSTGRES_USERNAME")
		cfg.Password = v.GetString("POSTGRES_PASSWORD")
		cfg.Database = v.GetString("POSTGRES_DATABASE")
		cfg.SSLMode = v.GetString("POSTGRES_SSLMODE")
		cfg.MaxIdleConns = v.GetInt("POSTGRES_MAX_IDLE_CONNS")
		cfg.MaxOpenConns = v.GetInt("POSTGRES_MAX_OPEN_CONNS")
		cfg.ConnMaxLifetime = v.GetDuration("POSTGRES_CONN_MAX_LIFETIME")
		logrus.WithFields(logrus.Fields{
			"host":     cfg.Host,
			"port":     cfg.Port,
			"database": cfg.Database,
		}).Debug("ent PostgreSQL config loaded")
	case "sqlite":
		cfg.DSN = v.GetString("SQLITE_DSN")
		if cfg.DSN == "" {
			cfg.DSN = ":memory:"
		}
		logrus.WithField("dsn", cfg.DSN).Debug("ent SQLite config loaded")
	default: // mysql
		cfg.DSN = v.GetString("MYSQL_DSN")
		cfg.Host = v.GetString("MYSQL_HOST")
		cfg.Port = v.GetInt("MYSQL_PORT")
		cfg.Username = v.GetString("MYSQL_USERNAME")
		cfg.Password = v.GetString("MYSQL_PASSWORD")
		cfg.Database = v.GetString("MYSQL_DATABASE")
		cfg.Charset = v.GetString("MYSQL_CHARSET")
		cfg.ParseTime = v.GetBool("MYSQL_PARSE_TIME")
		cfg.Loc = v.GetString("MYSQL_LOC")
		cfg.MaxIdleConns = v.GetInt("MYSQL_MAX_IDLE_CONNS")
		cfg.MaxOpenConns = v.GetInt("MYSQL_MAX_OPEN_CONNS")
		cfg.ConnMaxLifetime = v.GetDuration("MYSQL_CONN_MAX_LIFETIME")
		logrus.WithFields(logrus.Fields{
			"host":     cfg.Host,
			"port":     cfg.Port,
			"database": cfg.Database,
		}).Debug("ent MySQL config loaded")
	}

	return cfg
}

func (c *Config) buildDSN() string {
	if c.DSN != "" {
		return c.DSN
	}
	switch c.Driver {
	case "postgres":
		return fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			c.Host, c.Username, c.Password, c.Database, c.Port, c.SSLMode,
		)
	case "sqlite":
		// DSN is always "" here (non-empty was returned by the early guard above).
		// Fall back to in-memory database to match NewConfigFromEnv behaviour.
		return ":memory:"
	default: // mysql
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
			c.Username, c.Password,
			c.Host, c.Port,
			c.Database,
			c.Charset,
			c.ParseTime,
			c.Loc,
		)
	}
}

// ─── Client ───────────────────────────────────────────────────────────────────

// driverSpec holds the sql.Open driver name and ent dialect for a given backend.
type driverSpec struct {
	sqlName    string // name passed to sql.Open
	entDialect string // ent dialect constant
	hasPool    bool   // whether connection pool settings apply
}

var (
	driversMu sync.RWMutex
	// drivers is the built-in registry of supported ent backends.
	// Built-in drivers: "mysql", "postgres", "sqlite".
	drivers = map[string]driverSpec{
		"mysql":    {"mysql", dialect.MySQL, true},
		"postgres": {"postgres", dialect.Postgres, true},
		"sqlite":   {"sqlite", dialect.SQLite, false},
	}
)

// RegisterDriver registers a custom database backend for use with the ent provider.
// Call this before New() or NewClient() to add a backend not built into this package.
// Set hasPool to true if the driver supports sql.DB connection pool tuning.
//
//	ent.RegisterDriver("clickhouse", "clickhouse", "clickhouse", true)
func RegisterDriver(name, sqlName, entDialect string, hasPool bool) {
	driversMu.Lock()
	drivers[name] = driverSpec{sqlName, entDialect, hasPool}
	driversMu.Unlock()
}

// Client wraps an *entsql.Driver as a lifecycle-managed Provider.
// Obtain the driver via Driver() and pass it to your generated ent.NewClient.
type Client struct {
	provider.AbstractProvider

	Config  *Config
	drv     *entsql.Driver
	dialect string
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

// Driver returns the underlying *entsql.Driver. Only valid after Init().
func (c *Client) Driver() *entsql.Driver { return c.drv }

// Dialect returns the ent dialect constant for the configured driver.
func (c *Client) Dialect() string { return c.dialect }

// Init opens the database connection and wraps it in an ent driver.
func (c *Client) Init() error {
	driversMu.RLock()
	spec, ok := drivers[c.Config.Driver]
	driversMu.RUnlock()
	if !ok {
		return fmt.Errorf("ent: unsupported driver %q (built-in: mysql, postgres, sqlite)", c.Config.Driver)
	}

	db, err := sql.Open(spec.sqlName, c.Config.buildDSN())
	if err != nil {
		return fmt.Errorf("ent: failed to open sql.DB (%s): %w", c.Config.Driver, err)
	}

	if spec.hasPool {
		db.SetMaxIdleConns(c.Config.MaxIdleConns)
		db.SetMaxOpenConns(c.Config.MaxOpenConns)
		db.SetConnMaxLifetime(c.Config.ConnMaxLifetime)
	}

	c.dialect = spec.entDialect
	c.drv = entsql.OpenDB(c.dialect, db)
	return nil
}

// Close closes the underlying database connection.
func (c *Client) Close() error {
	if c.drv == nil {
		return nil
	}
	return c.drv.Close()
}
