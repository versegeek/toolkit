// Package pprof serves Go runtime profiles (net/http/pprof) on a dedicated
// port as a run provider.
//
//	g.MustInit(pprof.New(nil))
//	g.MustRun() // go tool pprof http://127.0.0.1:9999/debug/pprof/profile
//
// The server listens on 127.0.0.1 by default and can be switched off with
// PPROF_ENABLED=false.
package pprof

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/google/wire"

	"github.com/versegeek/toolkit/internal/httpserver"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Server.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewServer)

// standardPrefix is the path net/http/pprof expects for named profiles.
const standardPrefix = "/debug/pprof/"

// Server serves the profiling endpoints as a run provider.
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

	profiles := http.NewServeMux()
	profiles.HandleFunc(standardPrefix, pprof.Index)
	profiles.HandleFunc(standardPrefix+"cmdline", pprof.Cmdline)
	profiles.HandleFunc(standardPrefix+"profile", pprof.Profile)
	profiles.HandleFunc(standardPrefix+"symbol", pprof.Symbol)
	profiles.HandleFunc(standardPrefix+"trace", pprof.Trace)

	// pprof.Index only serves named profiles under /debug/pprof/, so requests
	// under a custom endpoint are rewritten to that prefix.
	endpoint := s.Config.endpoint()
	mux := http.NewServeMux()
	mux.HandleFunc(endpoint+"/", func(w http.ResponseWriter, r *http.Request) {
		r2 := r.Clone(r.Context())
		r2.URL.Path = standardPrefix + strings.TrimPrefix(r.URL.Path, endpoint+"/")
		r2.URL.RawPath = ""
		profiles.ServeHTTP(w, r2)
	})

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
		return errors.New("pprof: run failed: server is not initialized")
	}
	if err := httpserver.Serve(s.srv, s.SetRunning); err != nil {
		return fmt.Errorf("pprof: serve failed: %w", err)
	}
	return nil
}

// Close shuts the server down gracefully within Config.ShutdownTimeout.
func (s *Server) Close() error {
	if s.srv == nil {
		return nil
	}
	if err := httpserver.Shutdown(s.srv, s.Config.ShutdownTimeout); err != nil {
		return fmt.Errorf("pprof: shutdown failed: %w", err)
	}
	return nil
}
