// Package http provides a Gin HTTP server RunProvider.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(http.New(nil))
//	// Register routes on the server before MustRun:
//	// server := http.New(nil); c.MustInit(server); server.Router.GET("/", handler)
//	c.MustRun()
//
// Wire usage (Pattern B) – add http.ProviderSet to wire.Build():
//
//	wire.Build(http.NewConfigFromEnv, http.NewServer, ...)
package http

import (
	"context"
	"fmt"
	"net"
	nethttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/config"
	"github.com/versegeek/toolkit/provider"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds HTTP server parameters.
type Config struct {
	Host            string
	Port            int
	Mode            string        // gin mode: debug | release | test
	ShutdownTimeout time.Duration // graceful shutdown timeout
}

const (
	defaultHost            = ""
	defaultPort            = 8080
	defaultMode            = "release"
	defaultShutdownTimeout = 5 * time.Second
)

// NewConfigFromEnv reads HTTP server configuration from the environment.
// Keys: HTTP_HOST, HTTP_PORT, HTTP_MODE, HTTP_SHUTDOWN_TIMEOUT.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("HTTP_HOST", defaultHost)
	v.SetDefault("HTTP_PORT", defaultPort)
	v.SetDefault("HTTP_MODE", defaultMode)
	v.SetDefault("HTTP_SHUTDOWN_TIMEOUT", defaultShutdownTimeout)

	cfg := &Config{
		Host:            v.GetString("HTTP_HOST"),
		Port:            v.GetInt("HTTP_PORT"),
		Mode:            v.GetString("HTTP_MODE"),
		ShutdownTimeout: v.GetDuration("HTTP_SHUTDOWN_TIMEOUT"),
	}

	logrus.WithFields(logrus.Fields{
		"host": cfg.Host,
		"port": cfg.Port,
		"mode": cfg.Mode,
	}).Debug("HTTP server config loaded")

	return cfg
}

// ─── Server ───────────────────────────────────────────────────────────────────

// Server wraps a Gin engine as a lifecycle-managed RunProvider.
// Register routes on Router after Init() and before the Container starts.
type Server struct {
	provider.AbstractRunProvider

	Config *Config
	Router *gin.Engine
	srv    *nethttp.Server
}

var _ provider.RunProvider = (*Server)(nil)

// New creates a Server without initialising Gin.
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

// Init sets the Gin mode, creates the engine, and prepares the http.Server.
func (s *Server) Init() error {
	gin.SetMode(s.Config.Mode)
	s.Router = gin.New()
	s.Router.Use(gin.Recovery())

	addr := fmt.Sprintf("%s:%d", s.Config.Host, s.Config.Port)
	s.srv = &nethttp.Server{Addr: addr, Handler: s.Router}
	return nil
}

// Run binds the TCP port and starts serving. Blocks until stopped via Close().
// SetRunning(true) is called only after the port is successfully bound.
func (s *Server) Run() error {
	ln, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("http: listen on %s failed: %w", s.srv.Addr, err)
	}

	logrus.WithField("addr", s.srv.Addr).Info("HTTP server starting")
	s.SetRunning(true)

	if err := s.srv.Serve(ln); err != nethttp.ErrServerClosed {
		return err
	}
	return nil
}

// Close gracefully shuts down the HTTP server, waiting up to ShutdownTimeout
// for in-flight requests to complete.
func (s *Server) Close() error {
	if s.srv == nil {
		return s.AbstractRunProvider.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Config.ShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		logrus.WithError(err).Warn("HTTP server shutdown error")
	}

	return s.AbstractRunProvider.Close()
}
