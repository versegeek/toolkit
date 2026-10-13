package clickhouse_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	chprovider "github.com/versegeek/toolkit/db/clickhouse"
)

func TestConfigDefaults(t *testing.T) {
	cfg := chprovider.NewConfigFromEnv()

	require.NotNil(t, cfg)
	assert.Equal(t, []string{"127.0.0.1:9000"}, cfg.Addrs)
	assert.Equal(t, "default", cfg.Username)
	assert.Equal(t, "", cfg.Password)
	assert.Equal(t, "default", cfg.Database)
	assert.Equal(t, 30*time.Second, cfg.DialTimeout)
	assert.Equal(t, 10, cfg.MaxOpenConns)
	assert.Equal(t, 5, cfg.MaxIdleConns)
	assert.Equal(t, 1*time.Hour, cfg.ConnMaxLifetime)
}

func TestConfigOverride(t *testing.T) {
	t.Setenv("CLICKHOUSE_ADDRS", "ch1:9000,ch2:9000")
	t.Setenv("CLICKHOUSE_USERNAME", "admin")
	t.Setenv("CLICKHOUSE_DATABASE", "metrics")

	cfg := chprovider.NewConfigFromEnv()
	assert.Equal(t, []string{"ch1:9000", "ch2:9000"}, cfg.Addrs)
	assert.Equal(t, "admin", cfg.Username)
	assert.Equal(t, "metrics", cfg.Database)
}

func TestNew_UsesDefaultConfig(t *testing.T) {
	client := chprovider.New(nil)
	require.NotNil(t, client)
	require.NotNil(t, client.Config)
	assert.Equal(t, "default", client.Config.Database)
}

func TestNew_UsesProvidedConfig(t *testing.T) {
	cfg := &chprovider.Config{
		Addrs:    []string{"ch-custom:9000"},
		Username: "user1",
		Database: "db1",
	}
	client := chprovider.New(cfg)
	assert.Equal(t, cfg, client.Config)
}
