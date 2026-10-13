// Package otel provides an OpenTelemetry tracing Provider.
//
// Init creates a TracerProvider, registers it globally (otel.SetTracerProvider),
// and installs W3C TraceContext + Baggage propagators. Close flushes and shuts
// down the exporter gracefully.
//
// Exporter is selected by TRACE_EXPORTER:
//
//	otlp   → OTLP/HTTP (default, works with Jaeger ≥1.35, Grafana Tempo, SigNoz …)
//	zipkin → Zipkin HTTP (deprecated upstream; retained for legacy stacks)
//	stdout → human-readable output to stdout (development/debug)
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(otel.New(nil))
//	c.MustInit(redis.New(nil))
//	c.MustRun()
//
// Wire usage (Pattern B) – add otel.ProviderSet to wire.Build():
//
//	wire.Build(otel.NewConfigFromEnv, otel.NewClient, ...)
package otel

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/config"
	"github.com/versegeek/toolkit/provider"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds OpenTelemetry tracing configuration.
type Config struct {
	// ServiceName identifies the service in trace backends.
	ServiceName string
	// ServiceVersion is the version tag attached to all spans.
	ServiceVersion string
	// Exporter selects the trace exporter: "otlp" (default), "zipkin", "stdout".
	Exporter string
	// OTLPEndpoint is the base URL for the OTLP/HTTP collector, e.g. "http://localhost:4318".
	OTLPEndpoint string
	// ZipkinEndpoint is the Zipkin HTTP endpoint, e.g. "http://localhost:9411/api/v2/spans".
	ZipkinEndpoint string
	// SamplingRatio controls how many traces are sampled. 1.0 = always, 0.0 = never.
	SamplingRatio float64
	// ShutdownTimeout is the maximum time to wait for the exporter to flush on Close.
	ShutdownTimeout time.Duration
}

const (
	defaultExporter        = "otlp"
	defaultOTLPEndpoint    = "http://localhost:4318"
	defaultZipkinEndpoint  = "http://localhost:9411/api/v2/spans"
	defaultSamplingRatio   = 1.0
	defaultShutdownTimeout = 5 * time.Second
)

// NewConfigFromEnv reads tracing configuration from the environment.
//
// Keys: TRACE_SERVICE_NAME, TRACE_SERVICE_VERSION,
// TRACE_EXPORTER (otlp|zipkin|stdout),
// TRACE_OTLP_ENDPOINT, TRACE_ZIPKIN_ENDPOINT,
// TRACE_SAMPLING_RATIO (0.0–1.0), TRACE_SHUTDOWN_TIMEOUT.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("TRACE_SERVICE_NAME", "unknown_service")
	v.SetDefault("TRACE_SERVICE_VERSION", "")
	v.SetDefault("TRACE_EXPORTER", defaultExporter)
	v.SetDefault("TRACE_OTLP_ENDPOINT", defaultOTLPEndpoint)
	v.SetDefault("TRACE_ZIPKIN_ENDPOINT", defaultZipkinEndpoint)
	v.SetDefault("TRACE_SAMPLING_RATIO", defaultSamplingRatio)
	v.SetDefault("TRACE_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)

	return &Config{
		ServiceName:     v.GetString("TRACE_SERVICE_NAME"),
		ServiceVersion:  v.GetString("TRACE_SERVICE_VERSION"),
		Exporter:        v.GetString("TRACE_EXPORTER"),
		OTLPEndpoint:    v.GetString("TRACE_OTLP_ENDPOINT"),
		ZipkinEndpoint:  v.GetString("TRACE_ZIPKIN_ENDPOINT"),
		SamplingRatio:   v.GetFloat64("TRACE_SAMPLING_RATIO"),
		ShutdownTimeout: v.GetDuration("TRACE_SHUTDOWN_TIMEOUT"),
	}
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client wraps an OTel TracerProvider as a lifecycle-managed Provider.
// After Init(), use Client.Tracer(name) or the global otel.Tracer(name) to
// create spans.
type Client struct {
	provider.AbstractProvider

	Config   *Config
	provider *sdktrace.TracerProvider
}

var _ provider.Provider = (*Client)(nil)

// New creates a Client without initialising the TracerProvider.
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

// Tracer returns a named tracer backed by the initialised TracerProvider.
// Panics if called before Init().
func (c *Client) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return c.provider.Tracer(name, opts...)
}

// Init creates the exporter, builds the TracerProvider, and registers it
// globally together with W3C TraceContext + Baggage propagators.
// Returns an error if Init() has already been called on this instance.
func (c *Client) Init() error {
	if c.provider != nil {
		return fmt.Errorf("otel: already initialized; create a new Client instead of reusing one")
	}

	exp, err := c.buildExporter()
	if err != nil {
		return err
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(c.Config.ServiceName),
			semconv.ServiceVersionKey.String(c.Config.ServiceVersion),
		),
	)
	if err != nil {
		// Shut down the exporter so its connection is not leaked.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = exp.Shutdown(shutdownCtx)
		return fmt.Errorf("otel: failed to build resource: %w", err)
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.Config.SamplingRatio))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)
	c.provider = tp

	// Register globally so otel.Tracer() and otel.GetTracerProvider() work anywhere.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

// Close flushes pending spans and shuts down the TracerProvider.
func (c *Client) Close() error {
	if c.provider == nil {
		return nil
	}
	timeout := c.Config.ShutdownTimeout
	if timeout <= 0 {
		timeout = defaultShutdownTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := c.provider.Shutdown(ctx); err != nil {
		return fmt.Errorf("otel: shutdown error: %w", err)
	}
	return nil
}

// buildExporter constructs the trace exporter selected by Config.Exporter.
func (c *Client) buildExporter() (sdktrace.SpanExporter, error) {
	ctx := context.Background()
	switch c.Config.Exporter {
	case "stdout":
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("otel: stdout exporter error: %w", err)
		}
		return exp, nil

	case "zipkin":
		exp, err := zipkin.New(c.Config.ZipkinEndpoint)
		if err != nil {
			return nil, fmt.Errorf("otel: zipkin exporter error: %w", err)
		}
		return exp, nil

	default: // "otlp"
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpointURL(c.Config.OTLPEndpoint),
		}
		// Only disable TLS when the endpoint URL explicitly uses http://.
		if u, err := url.Parse(c.Config.OTLPEndpoint); err == nil && u.Scheme == "http" {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exp, err := otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("otel: OTLP exporter error: %w", err)
		}
		return exp, nil
	}
}
