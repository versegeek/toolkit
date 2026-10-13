// Package http provides a Gin HTTP server as a run provider.
//
// Register routes on Router after Init and before Run:
//
//	srv := http.New(nil)
//	g.MustInit(srv)
//	srv.Router.GET("/ping", ping)
//	g.MustRun()
//
// With Wire, add ProviderSet to the injector and register the routes in the
// constructor that receives the *Server.
package http

import (
	"errors"
	"fmt"
	nethttp "net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"github.com/versegeek/toolkit/internal/httpserver"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Server.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewServer)

// Server is a Gin HTTP server managed as a run provider.
type Server struct {
	provider.AbstractRunProvider

	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config
	// Router receives routes and middleware. Init creates it.
	Router *gin.Engine

	srv *nethttp.Server
}

var _ provider.RunProvider = (*Server)(nil)

// New returns a Server that is not initialised yet.
func New(cfg *Config) *Server {
	return &Server{Config: cfg}
}

// NewServer returns an initialised Server and a cleanup that closes it.
func NewServer(cfg *Config) (*Server, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Init creates the Gin engine with panic recovery and prepares the server.
func (s *Server) Init() error {
	if s.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		s.Config = cfg
	}
	switch s.Config.Mode {
	case gin.DebugMode, gin.ReleaseMode, gin.TestMode:
	default:
		return fmt.Errorf("http: invalid mode %q: want debug, release or test", s.Config.Mode)
	}

	gin.SetMode(s.Config.Mode)
	s.Router = gin.New()
	s.Router.Use(gin.Recovery())
	s.srv = httpserver.New(s.Config.addr(), s.Router)
	return nil
}

// Run listens on the configured address and serves until Close.
func (s *Server) Run() error {
	if s.srv == nil {
		return errors.New("http: run failed: server is not initialized")
	}
	if err := httpserver.Serve(s.srv, s.SetRunning); err != nil {
		return fmt.Errorf("http: serve failed: %w", err)
	}
	return nil
}

// Close shuts the server down gracefully within Config.ShutdownTimeout.
func (s *Server) Close() error {
	if s.srv == nil {
		return nil
	}
	if err := httpserver.Shutdown(s.srv, s.Config.ShutdownTimeout); err != nil {
		return fmt.Errorf("http: shutdown failed: %w", err)
	}
	return nil
}
