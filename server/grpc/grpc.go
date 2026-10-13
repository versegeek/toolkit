// Package grpc provides a gRPC server as a run provider.
//
// Register services on Server after Init and before Run:
//
//	srv := grpc.New(nil)
//	g.MustInit(srv)
//	pb.RegisterGreeterServer(srv.Server, &greeter{})
//	g.MustRun()
//
// ProviderSet builds the server without extra options. To pass
// grpc.ServerOption values with Wire, write a constructor that returns
// provider.InitProvider(grpc.New(cfg, opts...)).
package grpc

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/wire"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Server.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewServer)

// Server is a gRPC server managed as a run provider.
type Server struct {
	provider.AbstractRunProvider

	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config
	// Server receives service registrations. Init creates it.
	Server *grpc.Server

	opts   []grpc.ServerOption
	health *health.Server
}

var _ provider.RunProvider = (*Server)(nil)

// New returns a Server that is not initialised yet; opts are passed to
// grpc.NewServer by Init.
func New(cfg *Config, opts ...grpc.ServerOption) *Server {
	return &Server{Config: cfg, opts: opts}
}

// NewServer returns an initialised Server and a cleanup that closes it.
func NewServer(cfg *Config) (*Server, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Init creates the gRPC server and registers the health and reflection
// services when enabled.
func (s *Server) Init() error {
	if s.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		s.Config = cfg
	}

	s.Server = grpc.NewServer(s.opts...)
	if s.Config.Health {
		s.health = health.NewServer()
		healthpb.RegisterHealthServer(s.Server, s.health)
	}
	if s.Config.Reflection {
		reflection.Register(s.Server)
	}
	return nil
}

// Run listens on the configured address and serves until Close.
func (s *Server) Run() error {
	if s.Server == nil {
		return errors.New("grpc: run failed: server is not initialized")
	}
	ln, err := net.Listen("tcp", s.Config.addr())
	if err != nil {
		return fmt.Errorf("grpc: listen failed: %w", err)
	}
	s.SetRunning(true)
	defer s.SetRunning(false)

	if err := s.Server.Serve(ln); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("grpc: serve failed: %w", err)
	}
	return nil
}

// Close marks the health service as not serving and stops the server
// gracefully. RPCs still running after Config.ShutdownTimeout are cancelled.
func (s *Server) Close() error {
	if s.Server == nil {
		return nil
	}
	if s.health != nil {
		s.health.Shutdown()
	}

	timeout := s.Config.ShutdownTimeout
	if timeout <= 0 {
		timeout = defaultShutdownTimeout
	}
	done := make(chan struct{})
	go func() {
		s.Server.GracefulStop()
		close(done)
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return nil
	case <-timer.C:
		s.Server.Stop()
		<-done
		return fmt.Errorf("grpc: graceful stop timed out after %s", timeout)
	}
}
