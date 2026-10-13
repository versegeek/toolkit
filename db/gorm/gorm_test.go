package gorm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ─── buildDSN ─────────────────────────────────────────────────────────────────

func TestBuildDSN_MySQL(t *testing.T) {
	cfg := &Config{
		Driver:    "mysql",
		Host:      "127.0.0.1",
		Port:      3306,
		Username:  "root",
		Password:  "secret",
		Database:  "mydb",
		Charset:   "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}
	dsn := cfg.buildDSN()
	assert.Equal(t, "root:secret@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=true&loc=Local", dsn)
}

func TestBuildDSN_Postgres(t *testing.T) {
	cfg := &Config{
		Driver:   "postgres",
		Host:     "pg-host",
		Port:     5432,
		Username: "pguser",
		Password: "pgpass",
		Database: "pgdb",
		SSLMode:  "disable",
	}
	dsn := cfg.buildDSN()
	assert.Equal(t, "host=pg-host user=pguser password=pgpass dbname=pgdb port=5432 sslmode=disable", dsn)
}

func TestBuildDSN_SQLite(t *testing.T) {
	cfg := &Config{
		Driver: "sqlite",
		DSN:    "/tmp/test.db",
	}
	assert.Equal(t, "/tmp/test.db", cfg.buildDSN())
}

func TestBuildDSN_SQLite_Memory(t *testing.T) {
	cfg := &Config{Driver: "sqlite", DSN: ":memory:"}
	assert.Equal(t, ":memory:", cfg.buildDSN())
}

func TestBuildDSN_ClickHouse(t *testing.T) {
	cfg := &Config{
		Driver:      "clickhouse",
		Host:        "ch-host",
		Port:        9000,
		Username:    "default",
		Password:    "",
		Database:    "analytics",
		DialTimeout: 10 * time.Second,
		ReadTimeout: 20 * time.Second,
	}
	dsn := cfg.buildDSN()
	assert.Contains(t, dsn, "clickhouse://default:@ch-host:9000/analytics")
	assert.Contains(t, dsn, "dial_timeout=10s")
	assert.Contains(t, dsn, "read_timeout=20s")
}

func TestBuildDSN_ExplicitDSN(t *testing.T) {
	// When DSN is explicitly set, it must be returned as-is regardless of driver.
	explicit := "custom://user:pass@host/db?opt=1"
	for _, driver := range []string{"mysql", "postgres", "sqlite", "clickhouse"} {
		t.Run(driver, func(t *testing.T) {
			cfg := &Config{Driver: driver, DSN: explicit}
			assert.Equal(t, explicit, cfg.buildDSN())
		})
	}
}

// ─── dialectors registry ──────────────────────────────────────────────────────

func TestDialectors_BuiltIn(t *testing.T) {
	for _, name := range []string{"mysql", "postgres", "sqlite", "clickhouse"} {
		t.Run(name, func(t *testing.T) {
			open, ok := dialectors[name]
			assert.True(t, ok, "built-in dialector %q should be registered", name)
			assert.NotNil(t, open)
		})
	}
}

func TestRegisterDialector(t *testing.T) {
	called := false
	stub := func(dsn string) gorm.Dialector { called = true; return nil }

	RegisterDialector("stub-driver", stub)
	t.Cleanup(func() { delete(dialectors, "stub-driver") })

	open, ok := dialectors["stub-driver"]
	require.True(t, ok)
	open("dsn")
	assert.True(t, called)
}

// ─── poolDrivers registry ─────────────────────────────────────────────────────

func TestPoolDrivers_BuiltIn(t *testing.T) {
	assert.True(t, poolDrivers["mysql"], "mysql should be pool-capable")
	assert.True(t, poolDrivers["postgres"], "postgres should be pool-capable")
	assert.True(t, poolDrivers["clickhouse"], "clickhouse should be pool-capable")
	assert.False(t, poolDrivers["sqlite"], "sqlite should NOT be pool-capable")
}

func TestRegisterPoolDriver(t *testing.T) {
	RegisterPoolDriver("custom-pool-driver")
	t.Cleanup(func() { delete(poolDrivers, "custom-pool-driver") })
	assert.True(t, poolDrivers["custom-pool-driver"])
}

// ─── Config defaults ──────────────────────────────────────────────────────────

func TestConfigDefaults_MySQL(t *testing.T) {
	t.Setenv("DB_DRIVER", "mysql")
	cfg := NewConfigFromEnv()

	assert.Equal(t, "mysql", cfg.Driver)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, 3306, cfg.Port)
	assert.Equal(t, "root", cfg.Username)
	assert.Equal(t, "utf8mb4", cfg.Charset)
	assert.True(t, cfg.ParseTime)
	assert.Equal(t, "Local", cfg.Loc)
	assert.Equal(t, 10, cfg.MaxIdleConns)
	assert.Equal(t, 100, cfg.MaxOpenConns)
	assert.Equal(t, logger.Warn, cfg.LogLevel)
}

func TestConfigDefaults_Postgres(t *testing.T) {
	t.Setenv("DB_DRIVER", "postgres")
	cfg := NewConfigFromEnv()

	assert.Equal(t, "postgres", cfg.Driver)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, 5432, cfg.Port)
	assert.Equal(t, "postgres", cfg.Username)
	assert.Equal(t, "disable", cfg.SSLMode)
}

func TestConfigDefaults_ClickHouse(t *testing.T) {
	t.Setenv("DB_DRIVER", "clickhouse")
	cfg := NewConfigFromEnv()

	assert.Equal(t, "clickhouse", cfg.Driver)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, 9000, cfg.Port)
	assert.Equal(t, "default", cfg.Username)
	assert.Equal(t, "default", cfg.Database)
	assert.Equal(t, 10*time.Second, cfg.DialTimeout)
	assert.Equal(t, 20*time.Second, cfg.ReadTimeout)
}

func TestConfigDefaults_SQLite(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	cfg := NewConfigFromEnv()
	assert.Equal(t, "sqlite", cfg.Driver)
}

func TestConfigOverride_EnvVar(t *testing.T) {
	t.Setenv("DB_DRIVER", "postgres")
	t.Setenv("POSTGRES_HOST", "prod-pg")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DATABASE", "proddb")

	cfg := NewConfigFromEnv()
	assert.Equal(t, "prod-pg", cfg.Host)
	assert.Equal(t, 5433, cfg.Port)
	assert.Equal(t, "proddb", cfg.Database)
}
