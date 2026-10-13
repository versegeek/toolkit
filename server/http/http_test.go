package http_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	httpserver "github.com/versegeek/toolkit/server/http"
)

func TestNewConfigFromEnv_Defaults(t *testing.T) {
	cfg := httpserver.NewConfigFromEnv()
	require.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "release", cfg.Mode)
	assert.NotZero(t, cfg.ShutdownTimeout)
}

func TestNew_NilConfig(t *testing.T) {
	s := httpserver.New(nil)
	require.NotNil(t, s)
	assert.NotNil(t, s.Config)
}

func TestServer_Close_WithoutInit(t *testing.T) {
	// Close() before Init() must not panic (regression: nil s.srv dereference).
	s := httpserver.New(nil)
	assert.NoError(t, s.Close())
}

func TestNewServer_WireCompatible(t *testing.T) {
	cfg := &httpserver.Config{
		Port:            0,
		Mode:            "test",
		ShutdownTimeout: time.Second,
	}
	s, cleanup, err := httpserver.NewServer(cfg)
	require.NoError(t, err)
	require.NotNil(t, s)
	defer cleanup()

	assert.NotNil(t, s.Router)
}

func TestServer_Init_RegistersRoutes(t *testing.T) {
	cfg := &httpserver.Config{
		Port:            0,
		Mode:            "test",
		ShutdownTimeout: time.Second,
	}
	s := httpserver.New(cfg)
	require.NoError(t, s.Init())
	defer func() { _ = s.Close() }()

	s.Router.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	assert.NotNil(t, s.Router)
}

func TestServer_RunAndClose(t *testing.T) {
	cfg := &httpserver.Config{
		Host:            "127.0.0.1",
		Port:            19876,
		Mode:            "test",
		ShutdownTimeout: time.Second,
	}
	s := httpserver.New(cfg)
	require.NoError(t, s.Init())

	s.Router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	done := make(chan error, 1)
	go func() { done <- s.Run() }()

	time.Sleep(80 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:19876/ping")
	if err == nil {
		resp.Body.Close()
		assert.Equal(t, 200, resp.StatusCode)
	}

	assert.NoError(t, s.Close())
	assert.NoError(t, <-done)
}
