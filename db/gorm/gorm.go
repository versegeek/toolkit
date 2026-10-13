// Package gorm provides a GORM database client Provider.
//
// Supports MySQL, PostgreSQL, SQLite, and ClickHouse via the DB_DRIVER config key.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(gorm.New(nil))  // nil → reads from env/config.yaml
//	c.MustRun()
//
// Wire usage (Pattern B) – add gorm.ProviderSet to wire.Build():
//
//	wire.Build(gorm.NewConfigFromEnv, gorm.NewClient, ...)
//
// Driver selection: set DB_DRIVER to "sqlite" (default), "mysql", "postgres", or "clickhouse".
//   - SQLite:      uses SQLITE_DSN (file path or ":memory:")
//   - MySQL:       uses MYSQL_* environment variables
//   - PostgreSQL:  uses POSTGRES_* environment variables
//   - ClickHouse:  uses CLICKHOUSE_* environment variables
package gorm

import (
	"fmt"
	"sync"
	"time"

	gormsqlite "github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/internal/config"
	"github.com/versegeek/toolkit/provider"
	gormclickhouse "gorm.io/driver/clickhouse"
	gormmysql "gorm.io/driver/mysql"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds GORM connection parameters.
// Set Driver to choose the database backend; defaults to "sqlite".
type Config struct {
	// Driver selects the database backend: "sqlite", "mysql", "postgres", or "clickhouse".
	Driver string

	// DSN is an optional full connection string. When set it overrides all
	// individual host/port/user fields for the selected driver.
	DSN string

	// MySQL / PostgreSQL / ClickHouse shared fields
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

	// ClickHouse-specific
	DialTimeout time.Duration
	ReadTimeout time.Duration

	// Connection pool
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration

	LogLevel logger.LogLevel
}

// NewConfigFromEnv reads database configuration from the environment.
//
// DB_DRIVER selects the backend (default "sqlite").
// SQLite uses SQLITE_DSN, MySQL keys are prefixed with MYSQL_,
// PostgreSQL keys with POSTGRES_, and ClickHouse keys with CLICKHOUSE_.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("DB_DRIVER", "sqlite")
	v.SetDefault("MYSQL_LOG_LEVEL", int(logger.Warn))

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

	// ClickHouse defaults
	v.SetDefault("CLICKHOUSE_HOST", "127.0.0.1")
	v.SetDefault("CLICKHOUSE_PORT", 9000)
	v.SetDefault("CLICKHOUSE_USERNAME", "default")
	v.SetDefault("CLICKHOUSE_PASSWORD", "")
	v.SetDefault("CLICKHOUSE_DATABASE", "default")
	v.SetDefault("CLICKHOUSE_DIAL_TIMEOUT", 10*time.Second)
	v.SetDefault("CLICKHOUSE_READ_TIMEOUT", 20*time.Second)
	v.SetDefault("CLICKHOUSE_MAX_IDLE_CONNS", 5)
	v.SetDefault("CLICKHOUSE_MAX_OPEN_CONNS", 10)
	v.SetDefault("CLICKHOUSE_CONN_MAX_LIFETIME", 1*time.Hour)

	driver := v.GetString("DB_DRIVER")

	cfg := &Config{
		Driver:   driver,
		LogLevel: logger.LogLevel(v.GetInt("MYSQL_LOG_LEVEL")),
	}

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
		}).Debug("GORM PostgreSQL config loaded")
	case "clickhouse":
		cfg.DSN = v.GetString("CLICKHOUSE_DSN")
		cfg.Host = v.GetString("CLICKHOUSE_HOST")
		cfg.Port = v.GetInt("CLICKHOUSE_PORT")
		cfg.Username = v.GetString("CLICKHOUSE_USERNAME")
		cfg.Password = v.GetString("CLICKHOUSE_PASSWORD")
		cfg.Database = v.GetString("CLICKHOUSE_DATABASE")
		cfg.DialTimeout = v.GetDuration("CLICKHOUSE_DIAL_TIMEOUT")
		cfg.ReadTimeout = v.GetDuration("CLICKHOUSE_READ_TIMEOUT")
		cfg.MaxIdleConns = v.GetInt("CLICKHOUSE_MAX_IDLE_CONNS")
		cfg.MaxOpenConns = v.GetInt("CLICKHOUSE_MAX_OPEN_CONNS")
		cfg.ConnMaxLifetime = v.GetDuration("CLICKHOUSE_CONN_MAX_LIFETIME")
		logrus.WithFields(logrus.Fields{
			"host":     cfg.Host,
			"port":     cfg.Port,
			"database": cfg.Database,
		}).Debug("GORM ClickHouse config loaded")
	case "sqlite":
		cfg.DSN = v.GetString("SQLITE_DSN")
		if cfg.DSN == "" {
			cfg.DSN = ":memory:"
		}
		logrus.WithField("dsn", cfg.DSN).Debug("GORM SQLite config loaded")
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
		}).Debug("GORM MySQL config loaded")
	}

	return cfg
}

// buildDSN returns the effective DSN for the configured driver.
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
	case "clickhouse":
		return fmt.Sprintf(
			"clickhouse://%s:%s@%s:%d/%s?dial_timeout=%s&read_timeout=%s",
			c.Username, c.Password, c.Host, c.Port, c.Database,
			c.DialTimeout, c.ReadTimeout,
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

// Client wraps a *gorm.DB as a lifecycle-managed Provider.
type Client struct {
	provider.AbstractProvider

	Config *Config
	db     *gorm.DB
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

var (
	dialectorsMu sync.RWMutex
	// dialectors maps a driver name to its GORM dialector constructor.
	// Built-in drivers: "sqlite", "mysql", "postgres", "clickhouse".
	dialectors = map[string]func(string) gorm.Dialector{
		"mysql":      gormmysql.Open,
		"postgres":   gormpostgres.Open,
		"sqlite":     gormsqlite.Open,
		"clickhouse": gormclickhouse.Open,
	}

	poolDriversMu sync.RWMutex
	// poolDrivers is the set of drivers that support sql.DB connection pool tuning.
	// SQLite is file-based and does not benefit from pool settings.
	poolDrivers = map[string]bool{
		"mysql":      true,
		"postgres":   true,
		"clickhouse": true,
	}
)

// RegisterDialector registers a custom GORM dialector factory under the given
// driver name. Call this before New() or NewClient() to add a backend not
// built into this package (e.g. "sqlserver").
//
//	gorm.RegisterDialector("sqlserver", sqlserver.Open)
func RegisterDialector(name string, open func(string) gorm.Dialector) {
	dialectorsMu.Lock()
	dialectors[name] = open
	dialectorsMu.Unlock()
}

// RegisterPoolDriver marks a custom driver as pool-capable so that
// MaxIdleConns / MaxOpenConns / ConnMaxLifetime are applied on Init().
func RegisterPoolDriver(name string) {
	poolDriversMu.Lock()
	poolDrivers[name] = true
	poolDriversMu.Unlock()
}

// NewClient is the Wire-compatible constructor.
func NewClient(cfg *Config) (*Client, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Client returns the underlying *gorm.DB. Only valid after Init().
func (c *Client) Client() *gorm.DB { return c.db }

// Init opens the database connection and configures the connection pool.
func (c *Client) Init() error {
	dialectorsMu.RLock()
	open, ok := dialectors[c.Config.Driver]
	dialectorsMu.RUnlock()
	if !ok {
		return fmt.Errorf("gorm: unsupported driver %q (supported: mysql, postgres, sqlite, clickhouse)", c.Config.Driver)
	}

	db, err := gorm.Open(open(c.Config.buildDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(c.Config.LogLevel),
	})
	if err != nil {
		return fmt.Errorf("gorm: failed to open %s connection: %w", c.Config.Driver, err)
	}

	poolDriversMu.RLock()
	hasPool := poolDrivers[c.Config.Driver]
	poolDriversMu.RUnlock()
	if hasPool {
		sqlDB, err := db.DB()
		if err != nil {
			// Close the opened connection to avoid leaking it.
			if rawDB, e := db.DB(); e == nil {
				_ = rawDB.Close()
			}
			return fmt.Errorf("gorm: failed to get sql.DB: %w", err)
		}
		sqlDB.SetMaxIdleConns(c.Config.MaxIdleConns)
		sqlDB.SetMaxOpenConns(c.Config.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(c.Config.ConnMaxLifetime)
	}

	c.db = db
	return nil
}

// Close closes the underlying database connection.
func (c *Client) Close() error {
	if c.db == nil {
		return nil
	}
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
