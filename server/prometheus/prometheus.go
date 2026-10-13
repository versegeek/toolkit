// Package prometheus provides a Prometheus metrics endpoint RunProvider.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(prometheus.New(nil))
//	c.MustRun()  // exposes metrics at PROMETHEUS_ENDPOINT (default: /metrics:9090)
//
// Wire usage (Pattern B) – add prometheus.ProviderSet to wire.Build():
//
//	wire.Build(prometheus.NewConfigFromEnv, prometheus.NewServer, ...)
package prometheus

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/config"
	"github.com/versegeek/toolkit/provider"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds Prometheus server parameters.
type Config struct {
	Enabled  bool
	Port     int
	Endpoint string
}

const (
	defaultEnabled  = true
	defaultPort     = 9090
	defaultEndpoint = "/metrics"
)

// NewConfigFromEnv reads Prometheus configuration from the environment.
// Keys: PROMETHEUS_ENABLED, PROMETHEUS_PORT, PROMETHEUS_ENDPOINT.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("PROMETHEUS_ENABLED", defaultEnabled)
	v.SetDefault("PROMETHEUS_PORT", defaultPort)
	v.SetDefault("PROMETHEUS_ENDPOINT", defaultEndpoint)

	cfg := &Config{
		Enabled:  v.GetBool("PROMETHEUS_ENABLED"),
		Port:     v.GetInt("PROMETHEUS_PORT"),
		Endpoint: v.GetString("PROMETHEUS_ENDPOINT"),
	}

	logrus.WithFields(logrus.Fields{
		"enabled":  cfg.Enabled,
		"port":     cfg.Port,
		"endpoint": cfg.Endpoint,
	}).Debug("Prometheus config loaded")

	return cfg
}

// ─── Server ───────────────────────────────────────────────────────────────────

// Server exposes Prometheus metrics as a lifecycle-managed RunProvider.
type Server struct {
	provider.AbstractRunProvider

	Config *Config
	srv    *http.Server
}

var _ provider.RunProvider = (*Server)(nil)

// New creates a Server.
// If config is nil, NewConfigFromEnv() is used.
func New(cfg *Config) *Server {
	if cfg == nil {
		cfg = NewConfigFromEnv()
	}
	return &Server{Config: cfg}
}

// NewServer is the Wire-compatible constructor.
func NewServer(cfg *Config) (*Server, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Run starts the metrics HTTP server. Blocks until stopped.
// If Enabled is false, returns immediately.
func (s *Server) Run() error {
	if !s.Config.Enabled {
		logrus.Info("Prometheus server disabled")
		return nil
	}

	addr := fmt.Sprintf(":%d", s.Config.Port)

	mux := http.NewServeMux()
	mux.Handle(s.Config.Endpoint, promhttp.Handler())
	s.srv = &http.Server{Addr: addr, Handler: mux}
	s.SetRunning(true)

	logrus.WithFields(logrus.Fields{
		"addr":     addr,
		"endpoint": s.Config.Endpoint,
	}).Info("Prometheus server starting")

	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Close shuts down the metrics HTTP server.
func (s *Server) Close() error {
	if !s.Config.Enabled || s.srv == nil {
		return s.AbstractRunProvider.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		logrus.WithError(err).Warn("Prometheus server shutdown error")
	}

	return s.AbstractRunProvider.Close()
}
