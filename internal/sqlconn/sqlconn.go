// Package sqlconn holds what the gorm and ent providers share: the DB_*
// configuration keys, DSN construction and connection pool tuning.
package sqlconn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

// Built-in driver names accepted in DB_DRIVER.
const (
	DriverMySQL      = "mysql"
	DriverPostgres   = "postgres"
	DriverSQLite     = "sqlite"
	DriverClickHouse = "clickhouse"
)

// PingTimeout bounds the connectivity check performed during Init.
const PingTimeout = 10 * time.Second

// Settings are the connection settings shared by the SQL providers.
type Settings struct {
	Driver          string
	DSN             string
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	Params          string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// Load reads the DB_* keys from v. DB_DRIVER has no default.
func Load(v *viper.Viper) Settings {
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_MAX_OPEN_CONNS", 100)
	v.SetDefault("DB_CONN_MAX_LIFETIME", time.Hour)

	return Settings{
		Driver:          v.GetString("DB_DRIVER"),
		DSN:             v.GetString("DB_DSN"),
		Host:            v.GetString("DB_HOST"),
		Port:            v.GetInt("DB_PORT"),
		Username:        v.GetString("DB_USERNAME"),
		Password:        v.GetString("DB_PASSWORD"),
		Database:        v.GetString("DB_DATABASE"),
		Params:          v.GetString("DB_PARAMS"),
		MaxIdleConns:    v.GetInt("DB_MAX_IDLE_CONNS"),
		MaxOpenConns:    v.GetInt("DB_MAX_OPEN_CONNS"),
		ConnMaxLifetime: v.GetDuration("DB_CONN_MAX_LIFETIME"),
	}
}

// BuildDSN returns DSN when it is set and otherwise builds one for Driver
// from the individual fields. Drivers other than the built-in ones need DSN.
func (s Settings) BuildDSN() (string, error) {
	if s.Driver == "" {
		return "", ErrNoDriver
	}
	if s.DSN != "" {
		return s.DSN, nil
	}

	switch s.Driver {
	case DriverSQLite:
		if s.Database == "" {
			return "", errors.New("sqlite needs DB_DSN or DB_DATABASE (a file path, or :memory:)")
		}
		return withParams(s.Database, s.Params), nil
	case DriverMySQL:
		c := mysql.NewConfig()
		c.User = s.username("root")
		c.Passwd = s.Password
		c.Net = "tcp"
		c.Addr = s.hostPort(3306)
		c.DBName = s.Database
		params := s.Params
		if params == "" {
			params = "charset=utf8mb4&parseTime=true&loc=Local"
		}
		return withParams(c.FormatDSN(), params), nil
	case DriverPostgres:
		return s.url("postgres", "postgres", 5432), nil
	case DriverClickHouse:
		return s.url("clickhouse", "default", 9000), nil
	default:
		return "", fmt.Errorf("driver %q needs DB_DSN", s.Driver)
	}
}

// ApplyPool applies the positive pool limits to db; zero keeps the
// database/sql default. An in-memory SQLite database exists only as long as
// its connection, so it gets exactly one connection that is never recycled.
func (s Settings) ApplyPool(db *sql.DB, dsn string) {
	if s.Driver == DriverSQLite && IsMemory(dsn) {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(0)
		db.SetConnMaxIdleTime(0)
		return
	}
	if s.MaxIdleConns > 0 {
		db.SetMaxIdleConns(s.MaxIdleConns)
	}
	if s.MaxOpenConns > 0 {
		db.SetMaxOpenConns(s.MaxOpenConns)
	}
	if s.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(s.ConnMaxLifetime)
	}
}

// ErrNoDriver reports that no driver was configured.
var ErrNoDriver = errors.New("no driver configured (set DB_DRIVER)")

// IsMemory reports whether an SQLite DSN names an in-memory database.
func IsMemory(dsn string) bool {
	return strings.HasPrefix(dsn, ":memory:") ||
		strings.HasPrefix(dsn, "file::memory:") ||
		strings.Contains(dsn, "mode=memory")
}

// Ping checks that db is reachable within PingTimeout.
func Ping(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), PingTimeout)
	defer cancel()
	return db.PingContext(ctx)
}

func (s Settings) username(def string) string {
	if s.Username == "" {
		return def
	}
	return s.Username
}

func (s Settings) hostPort(defPort int) string {
	host, port := s.Host, s.Port
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = defPort
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func (s Settings) url(scheme, defUser string, defPort int) string {
	u := url.URL{
		Scheme:   scheme,
		User:     url.UserPassword(s.username(defUser), s.Password),
		Host:     s.hostPort(defPort),
		Path:     "/" + s.Database,
		RawQuery: s.Params,
	}
	return u.String()
}

func withParams(dsn, params string) string {
	if params == "" {
		return dsn
	}
	sep := "?"
	if strings.Contains(dsn, "?") {
		sep = "&"
	}
	return dsn + sep + params
}
