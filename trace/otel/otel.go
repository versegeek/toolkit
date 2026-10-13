// Package otel provides OpenTelemetry tracing as a provider.
//
// Init builds a TracerProvider with the exporter selected by Config.Exporter
// (OTLP over HTTP by default, or stdout), installs it and the W3C trace
// context and baggage propagators as the OpenTelemetry globals, so
// otel.Tracer works anywhere. Close flushes pending spans and restores the
// previous globals.
//
//	g.MustInit(otel.New(nil))
//	ctx, span := otelapi.Tracer("orders").Start(ctx, "place-order")
//	defer span.End()
package otel

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/wire"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// Client is an OpenTelemetry TracerProvider managed as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	tp      *sdktrace.TracerProvider
	restore func()
}

var _ provider.Provider = (*Client)(nil)

// New returns a Client that is not initialised yet.
func New(cfg *Config) *Client {
	return &Client{Config: cfg}
}

// NewClient returns an initialised Client and a cleanup that closes it.
func NewClient(cfg *Config) (*Client, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Tracer returns a named tracer from this Client's TracerProvider, or from
// the global one before Init.
func (c *Client) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	if c.tp == nil {
		return otel.Tracer(name, opts...)
	}
	return c.tp.Tracer(name, opts...)
}

// Init builds the exporter and TracerProvider and installs them as the
// OpenTelemetry globals.
func (c *Client) Init() error {
	if c.tp != nil {
		return errors.New("otel: already initialized")
	}
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}

	attrs := []attribute.KeyValue{semconv.ServiceName(c.Config.ServiceName)}
	if c.Config.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersion(c.Config.ServiceVersion))
	}
	res, err := resource.Merge(resource.Default(), resource.NewWithAttributes(semconv.SchemaURL, attrs...))
	if err != nil {
		return fmt.Errorf("otel: build resource failed: %w", err)
	}
	exp, err := newExporter(c.Config)
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(c.Config.SamplingRatio))),
	)

	prevTP, prevPropagator := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	c.tp = tp
	c.restore = func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevPropagator)
	}
	return nil
}

// Close flushes pending spans within Config.ShutdownTimeout, shuts the
// TracerProvider down and restores the previous globals.
func (c *Client) Close() error {
	if c.tp == nil {
		return nil
	}
	c.restore()

	timeout := c.Config.ShutdownTimeout
	if timeout <= 0 {
		timeout = defaultShutdownTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := c.tp.Shutdown(ctx); err != nil {
		return fmt.Errorf("otel: shutdown failed: %w", err)
	}
	return nil
}
