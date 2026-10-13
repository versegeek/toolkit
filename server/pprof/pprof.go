// Package pprof provides a Go pprof profiling endpoint RunProvider.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(pprof.New(nil))
//	c.MustRun()  // exposes profiling at PPROF_ENDPOINT (default: /debug/pprof:9999)
//
// Wire usage (Pattern B) – add pprof.ProviderSet to wire.Build():
//
//	wire.Build(pprof.NewConfigFromEnv, pprof.NewServer, ...)
package pprof

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/config"
	"github.com/versegeek/toolkit/provider"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds pprof server parameters.
type Config struct {
	Enabled  bool
	Port     int
	Endpoint string
}

const (
	defaultEnabled  = true
	defaultPort     = 9999
	defaultEndpoint = "/debug/pprof"
)

// NewConfigFromEnv reads pprof configuration from the environment.
// Keys: PPROF_ENABLED, PPROF_PORT, PPROF_ENDPOINT.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("PPROF_ENABLED", defaultEnabled)
	v.SetDefault("PPROF_PORT", defaultPort)
	v.SetDefault("PPROF_ENDPOINT", defaultEndpoint)

	cfg := &Config{
		Enabled:  v.GetBool("PPROF_ENABLED"),
		Port:     v.GetInt("PPROF_PORT"),
		Endpoint: v.GetString("PPROF_ENDPOINT"),
	}

	logrus.WithFields(logrus.Fields{
		"enabled":  cfg.Enabled,
		"port":     cfg.Port,
		"endpoint": cfg.Endpoint,
	}).Debug("PProf config loaded")

	return cfg
}

// ─── Server ───────────────────────────────────────────────────────────────────

// Server exposes Go pprof profiling endpoints as a lifecycle-managed RunProvider.
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

// Run starts the pprof HTTP server. Blocks until stopped.
// If Enabled is false, returns immediately.
func (s *Server) Run() error {
	if !s.Config.Enabled {
		logrus.Info("PProf server disabled")
		return nil
	}

	addr := fmt.Sprintf(":%d", s.Config.Port)
	ep := s.Config.Endpoint

	mux := http.NewServeMux()
	mux.HandleFunc(ep+"/", pprof.Index)
	mux.HandleFunc(ep+"/cmdline", pprof.Cmdline)
	mux.HandleFunc(ep+"/profile", pprof.Profile)
	mux.HandleFunc(ep+"/symbol", pprof.Symbol)
	mux.HandleFunc(ep+"/trace", pprof.Trace)

	s.srv = &http.Server{Addr: addr, Handler: mux}
	s.SetRunning(true)

	logrus.WithFields(logrus.Fields{
		"addr":     addr,
		"endpoint": ep,
	}).Info("PProf server starting")

	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Close shuts down the pprof HTTP server.
func (s *Server) Close() error {
	if !s.Config.Enabled || s.srv == nil {
		return s.AbstractRunProvider.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		logrus.WithError(err).Warn("PProf server shutdown error")
	}

	return s.AbstractRunProvider.Close()
}
