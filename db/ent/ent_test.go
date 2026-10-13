package ent

import (
	"testing"

	"entgo.io/ent/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t,
		"root:secret@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.buildDSN(),
	)
}

func TestBuildDSN_Postgres(t *testing.T) {
	cfg := &Config{
		Driver:   "postgres",
		Host:     "pg-host",
		Port:     5432,
		Username: "pguser",
		Password: "pgpass",
		Database: "pgdb",
		SSLMode:  "require",
	}
	assert.Equal(t,
		"host=pg-host user=pguser password=pgpass dbname=pgdb port=5432 sslmode=require",
		cfg.buildDSN(),
	)
}

func TestBuildDSN_SQLite(t *testing.T) {
	cfg := &Config{Driver: "sqlite", DSN: "/var/data/app.db"}
	assert.Equal(t, "/var/data/app.db", cfg.buildDSN())
}

func TestBuildDSN_ExplicitDSN_OverridesDriver(t *testing.T) {
	explicit := "mysql://user:pass@host/db"
	for _, drv := range []string{"mysql", "postgres", "sqlite"} {
		t.Run(drv, func(t *testing.T) {
			cfg := &Config{Driver: drv, DSN: explicit}
			assert.Equal(t, explicit, cfg.buildDSN())
		})
	}
}

// ─── drivers registry ────────────────────────────────────────────────────────

func TestDrivers_BuiltIn(t *testing.T) {
	tests := []struct {
		name       string
		sqlName    string
		entDialect string
		hasPool    bool
	}{
		{"mysql", "mysql", dialect.MySQL, true},
		{"postgres", "postgres", dialect.Postgres, true},
		{"sqlite", "sqlite", dialect.SQLite, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, ok := drivers[tt.name]
			require.True(t, ok, "built-in driver %q should be registered", tt.name)
			assert.Equal(t, tt.sqlName, spec.sqlName)
			assert.Equal(t, tt.entDialect, spec.entDialect)
			assert.Equal(t, tt.hasPool, spec.hasPool)
		})
	}
}

func TestRegisterDriver(t *testing.T) {
	RegisterDriver("custom-db", "custom", "custom_dialect", true)
	t.Cleanup(func() { delete(drivers, "custom-db") })

	spec, ok := drivers["custom-db"]
	require.True(t, ok)
	assert.Equal(t, "custom", spec.sqlName)
	assert.Equal(t, "custom_dialect", spec.entDialect)
	assert.True(t, spec.hasPool)
}

func TestRegisterDriver_NoPool(t *testing.T) {
	RegisterDriver("file-db", "filedrv", "file_dialect", false)
	t.Cleanup(func() { delete(drivers, "file-db") })

	spec, ok := drivers["file-db"]
	require.True(t, ok)
	assert.False(t, spec.hasPool)
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
	assert.Equal(t, 10, cfg.MaxIdleConns)
	assert.Equal(t, 100, cfg.MaxOpenConns)
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

func TestConfigDefaults_SQLite(t *testing.T) {
	t.Setenv("DB_DRIVER", "sqlite")
	cfg := NewConfigFromEnv()
	assert.Equal(t, "sqlite", cfg.Driver)
}
