package nats_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	natsprovider "github.com/versegeek/toolkit/mq/nats"
)

func TestConfigDefaults(t *testing.T) {
	cfg := natsprovider.NewConfigFromEnv()

	require.NotNil(t, cfg)
	assert.Equal(t, "nats://127.0.0.1:4222", cfg.URL)
	assert.Equal(t, "", cfg.Name)
	assert.Equal(t, 5*time.Second, cfg.ConnectTimeout)
	assert.Equal(t, 2*time.Second, cfg.ReconnectWait)
	assert.Equal(t, -1, cfg.MaxReconnects)
	assert.False(t, cfg.EnableJetStream)
}

func TestConfigOverride(t *testing.T) {
	t.Setenv("NATS_URL", "nats://nats-cluster:4222")
	t.Setenv("NATS_NAME", "my-service")
	t.Setenv("NATS_ENABLE_JETSTREAM", "true")
	t.Setenv("NATS_MAX_RECONNECTS", "5")

	cfg := natsprovider.NewConfigFromEnv()
	assert.Equal(t, "nats://nats-cluster:4222", cfg.URL)
	assert.Equal(t, "my-service", cfg.Name)
	assert.True(t, cfg.EnableJetStream)
	assert.Equal(t, 5, cfg.MaxReconnects)
}

func TestNew_NilConfig(t *testing.T) {
	client := natsprovider.New(nil)
	require.NotNil(t, client)
	assert.Equal(t, "nats://127.0.0.1:4222", client.Config.URL)
}

func TestNew_WithConfig(t *testing.T) {
	cfg := &natsprovider.Config{URL: "nats://custom:4222"}
	client := natsprovider.New(cfg)
	assert.Equal(t, cfg, client.Config)
}
