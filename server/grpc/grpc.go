// Package grpc provides a gRPC server RunProvider.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	server := grpc.New(nil)
//	c.MustInit(server)
//	// Register gRPC services on server.Server before MustRun:
//	// pb.RegisterMyServiceServer(server.Server, &myServiceImpl{})
//	c.MustRun()
//
// Wire usage (Pattern B) – add grpc.ProviderSet to wire.Build():
//
//	wire.Build(grpc.NewConfigFromEnv, grpc.NewServer, ...)
package grpc

import (
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/internal/config"
	"github.com/versegeek/toolkit/provider"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds gRPC server parameters.
type Config struct {
	Port          int
	EnableHealth  bool // Register gRPC health check endpoint.
	EnableReflect bool // Register gRPC server reflection (useful for grpcurl).
}

const (
	defaultPort          = 3000
	defaultEnableHealth  = true
	defaultEnableReflect = true
)

// NewConfigFromEnv reads gRPC server configuration from the environment.
// Keys: GRPC_PORT, GRPC_HEALTH_ENABLED, GRPC_REFLECTION_ENABLED.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("GRPC_PORT", defaultPort)
	v.SetDefault("GRPC_HEALTH_ENABLED", defaultEnableHealth)
	v.SetDefault("GRPC_REFLECTION_ENABLED", defaultEnableReflect)

	cfg := &Config{
		Port:          v.GetInt("GRPC_PORT"),
		EnableHealth:  v.GetBool("GRPC_HEALTH_ENABLED"),
		EnableReflect: v.GetBool("GRPC_REFLECTION_ENABLED"),
	}

	logrus.WithFields(logrus.Fields{
		"port":       cfg.Port,
		"health":     cfg.EnableHealth,
		"reflection": cfg.EnableReflect,
	}).Debug("gRPC server config loaded")

	return cfg
}

// ─── Server ───────────────────────────────────────────────────────────────────

// Server wraps a *grpc.Server as a lifecycle-managed RunProvider.
// Register gRPC services on Server after Init() and before the Container starts.
type Server struct {
	provider.AbstractRunProvider

	Config   *Config
	Server   *grpc.Server
	listener net.Listener
}

var _ provider.RunProvider = (*Server)(nil)

// New creates a Server without initialising the gRPC server.
// If config is nil, NewConfigFromEnv() is used.
func New(cfg *Config, opts ...grpc.ServerOption) *Server {
	if cfg == nil {
		cfg = NewConfigFromEnv()
	}
	return &Server{
		Config: cfg,
		Server: grpc.NewServer(opts...),
	}
}

// NewServer is the Wire-compatible constructor (no extra ServerOptions).
func NewServer(cfg *Config) (*Server, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Init registers optional health and reflection endpoints.
func (s *Server) Init() error {
	if s.Config.EnableHealth {
		grpc_health_v1.RegisterHealthServer(s.Server, health.NewServer())
		logrus.Debug("gRPC health endpoint registered")
	}
	if s.Config.EnableReflect {
		reflection.Register(s.Server)
		logrus.Debug("gRPC reflection registered")
	}
	return nil
}

// Run creates a TCP listener and starts the gRPC server. Blocks until stopped.
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.Config.Port)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("grpc: listen on %s failed: %w", addr, err)
	}
	s.listener = ln
	s.SetRunning(true)

	logrus.WithField("addr", addr).Info("gRPC server starting")
	return s.Server.Serve(ln)
}

// Close performs a graceful shutdown of the gRPC server.
func (s *Server) Close() error {
	s.Server.GracefulStop()
	return s.AbstractRunProvider.Close()
}
