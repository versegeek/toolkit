package otel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"

	otelp "github.com/versegeek/toolkit/trace/otel"
)

func TestNewConfigFromEnv_Defaults(t *testing.T) {
	cfg := otelp.NewConfigFromEnv()
	assert.Equal(t, "unknown_service", cfg.ServiceName)
	assert.Equal(t, "otlp", cfg.Exporter)
	assert.Equal(t, 1.0, cfg.SamplingRatio)
	assert.NotZero(t, cfg.ShutdownTimeout)
}

func TestNew_NilConfig(t *testing.T) {
	c := otelp.New(nil)
	require.NotNil(t, c)
	assert.NotNil(t, c.Config)
}

func TestClient_Init_Stdout(t *testing.T) {
	c := otelp.New(&otelp.Config{
		ServiceName:   "test-svc",
		Exporter:      "stdout",
		SamplingRatio: 1.0,
	})
	require.NoError(t, c.Init())
	defer func() { require.NoError(t, c.Close()) }()

	// Global TracerProvider should be replaced.
	tp := otel.GetTracerProvider()
	assert.NotNil(t, tp)

	// Tracer() should work.
	tracer := c.Tracer("test")
	assert.NotNil(t, tracer)
}

func TestClient_Init_OTLP_Unreachable(t *testing.T) {
	// OTLP exporter creation itself should succeed even if the endpoint is down;
	// failures are surfaced during span export/flush, not during Init.
	c := otelp.New(&otelp.Config{
		ServiceName:  "test-svc",
		Exporter:     "otlp",
		OTLPEndpoint: "http://127.0.0.1:9999", // guaranteed unreachable
	})
	require.NoError(t, c.Init())
	// Close may return a flush error when the endpoint is unreachable — acceptable.
	_ = c.Close()
}

func TestClient_Init_Zipkin_Unreachable(t *testing.T) {
	c := otelp.New(&otelp.Config{
		ServiceName:    "test-svc",
		Exporter:       "zipkin",
		ZipkinEndpoint: "http://127.0.0.1:9998/api/v2/spans",
	})
	require.NoError(t, c.Init())
	// Close may return a flush error when the endpoint is unreachable — acceptable.
	_ = c.Close()
}

func TestClient_Close_WithoutInit(t *testing.T) {
	c := otelp.New(nil)
	// Close before Init should be a no-op.
	assert.NoError(t, c.Close())
}

func TestNewClient_WireCompatible(t *testing.T) {
	cfg := &otelp.Config{
		ServiceName:   "test-svc",
		Exporter:      "stdout",
		SamplingRatio: 1.0,
	}
	c, cleanup, err := otelp.NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, c)
	defer cleanup()

	assert.NotNil(t, c.Tracer("test"))
}

func TestClient_DoubleInit_ReturnsError(t *testing.T) {
	c := otelp.New(&otelp.Config{
		ServiceName:   "test-svc",
		Exporter:      "stdout",
		SamplingRatio: 1.0,
	})
	require.NoError(t, c.Init())
	defer func() { _ = c.Close() }()

	err := c.Init()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already initialized")
}

func TestClient_Sampling_ZeroRatio(t *testing.T) {
	c := otelp.New(&otelp.Config{
		ServiceName:   "test-svc",
		Exporter:      "stdout",
		SamplingRatio: 0.0, // drop-all sampler
	})
	require.NoError(t, c.Init())
	require.NoError(t, c.Close())
}
