package otel

import (
	"context"
	"fmt"
	"net/url"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// newExporter builds the span exporter selected by cfg.Exporter.
func newExporter(cfg *Config) (sdktrace.SpanExporter, error) {
	switch cfg.Exporter {
	case "otlp", "":
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpointURL(cfg.OTLPEndpoint)}
		if u, err := url.Parse(cfg.OTLPEndpoint); err == nil && u.Scheme == "http" {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exp, err := otlptracehttp.New(context.Background(), opts...)
		if err != nil {
			return nil, fmt.Errorf("otel: create OTLP exporter failed: %w", err)
		}
		return exp, nil
	case "stdout":
		exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("otel: create stdout exporter failed: %w", err)
		}
		return exp, nil
	default:
		return nil, fmt.Errorf("otel: unknown exporter %q: want otlp or stdout", cfg.Exporter)
	}
}
