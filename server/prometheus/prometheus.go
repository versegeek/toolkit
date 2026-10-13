// Package prometheus serves Prometheus metrics from the default registry on a
// dedicated port as a run provider.
//
//	g.MustInit(prometheus.New(nil))
//	g.MustRun() // curl http://localhost:9090/metrics
//
// The server can be switched off with PROMETHEUS_ENABLED=false.
package prometheus

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/versegeek/toolkit/internal/httpserver"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Server.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewServer)

// Server serves the metrics endpoint as a run provider.
type Server struct {
	provider.AbstractRunProvider

	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	srv *http.Server
}

var (
	_ provider.RunProvider = (*Server)(nil)
	_ provider.Enabler     = (*Server)(nil)
)

// New returns a Server that is not initialised yet.
func New(cfg *Config) *Server {
	return &Server{Config: cfg}
}

// NewServer returns an initialised Server and a cleanup that closes it.
func NewServer(cfg *Config) (*Server, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Enabled reports whether the server is switched on. It reports true while
// the configuration has not been loaded yet.
func (s *Server) Enabled() bool {
	return s.Config == nil || s.Config.Enabled
}

// Init prepares the server unless it is disabled.
func (s *Server) Init() error {
	if s.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		s.Config = cfg
	}
	if !s.Config.Enabled {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle(s.Config.Endpoint, promhttp.Handler())
	s.srv = httpserver.New(s.Config.addr(), mux)
	return nil
}

// Run listens on the configured address and serves until Close. A disabled
// server returns at once.
func (s *Server) Run() error {
	if s.Config != nil && !s.Config.Enabled {
		return nil
	}
	if s.srv == nil {
		return errors.New("prometheus: run failed: server is not initialized")
	}
	if err := httpserver.Serve(s.srv, s.SetRunning); err != nil {
		return fmt.Errorf("prometheus: serve failed: %w", err)
	}
	return nil
}

// Close shuts the server down gracefully within Config.ShutdownTimeout.
func (s *Server) Close() error {
	if s.srv == nil {
		return nil
	}
	if err := httpserver.Shutdown(s.srv, s.Config.ShutdownTimeout); err != nil {
		return fmt.Errorf("prometheus: shutdown failed: %w", err)
	}
	return nil
}
