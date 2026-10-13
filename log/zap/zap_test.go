package zaplog_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	zaplog "github.com/versegeek/toolkit/log/zap"
	"go.uber.org/zap"
)

func TestNewConfigFromEnv_Defaults(t *testing.T) {
	cfg := zaplog.NewConfigFromEnv()
	assert.Equal(t, "info", cfg.Level)
	assert.Equal(t, "json", cfg.Encoding)
	assert.Equal(t, []string{"stderr"}, cfg.OutputPaths)
	assert.False(t, cfg.Development)
}

func TestNew_NilConfig(t *testing.T) {
	c := zaplog.New(nil)
	require.NotNil(t, c)
	assert.NotNil(t, c.Config)
}

func TestClient_Init_JSON(t *testing.T) {
	c := zaplog.New(&zaplog.Config{
		Level:       "debug",
		Encoding:    "json",
		OutputPaths: []string{"stderr"},
	})
	require.NoError(t, c.Init())
	assert.NotNil(t, c.Logger)
	// Global logger must be replaced.
	assert.NotNil(t, zap.L())
}

func TestClient_Init_Console(t *testing.T) {
	c := zaplog.New(&zaplog.Config{
		Level:       "warn",
		Encoding:    "console",
		OutputPaths: []string{"stderr"},
		Development: true,
	})
	require.NoError(t, c.Init())
	assert.NotNil(t, c.Logger)
}

func TestClient_Init_InvalidLevel(t *testing.T) {
	c := zaplog.New(&zaplog.Config{
		Level:       "not-a-level",
		Encoding:    "json",
		OutputPaths: []string{"stderr"},
	})
	err := c.Init()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid level")
}

func TestClient_Close_WithoutInit(t *testing.T) {
	c := zaplog.New(nil)
	// Close before Init should not panic or error.
	assert.NoError(t, c.Close())
}

func TestClient_Close_AfterInit(t *testing.T) {
	c := zaplog.New(&zaplog.Config{
		Level:       "info",
		Encoding:    "json",
		OutputPaths: []string{"stderr"},
	})
	require.NoError(t, c.Init())
	assert.NoError(t, c.Close())
}

func TestNewClient_WireCompatible(t *testing.T) {
	cfg := &zaplog.Config{
		Level:       "info",
		Encoding:    "json",
		OutputPaths: []string{"stderr"},
	}
	c, cleanup, err := zaplog.NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, c)
	defer cleanup()

	assert.NotNil(t, c.Logger)
}
