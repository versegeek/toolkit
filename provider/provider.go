// Package provider defines the core interfaces for composable application
// components.
//
// Lifecycle of a Provider:
//
//	New()           → struct created, not connected
//	Init()          → connections opened, pool ready
//	go Run() [opt]  → background service started (RunProvider only)
//	...
//	Close()         → connections closed, resources released
//
// A Container (package container) manages Init/Run/Close in the correct order.
// Alternatively, wire the providers manually and call Init() yourself via InitProvider.
package provider

import (
	"fmt"
	"reflect"
	"sync/atomic"

	"github.com/sirupsen/logrus"
)

// Provider is the base interface every component must implement.
type Provider interface {
	// Init establishes connections and prepares resources.
	// Called once before the component is used.
	Init() error

	// Close releases resources. Called in reverse order on shutdown.
	Close() error
}

// RunProvider extends Provider with a blocking Run method.
// Implement this for long-running services (HTTP/gRPC servers, consumers, etc.).
//
// The Container's MustRun starts each RunProvider in its own goroutine
// after all Init() calls succeed.
type RunProvider interface {
	Provider

	// Run must block until the service stops (Shutdown, signal, error).
	// Return nil on clean shutdown, non-nil on unexpected failure.
	Run() error

	// IsRunning reports true after Run() has entered its steady state
	// (e.g. the HTTP server is listening). Call SetRunning(true) once ready.
	IsRunning() bool
}

// ─── Abstract base types ──────────────────────────────────────────────────────

// AbstractProvider is an embeddable no-op Provider.
// Embed it and override only the methods you need.
type AbstractProvider struct{}

func (p *AbstractProvider) Init() error  { return nil }
func (p *AbstractProvider) Close() error { return nil }

// AbstractRunProvider is an embeddable base for RunProvider.
// Call SetRunning(true) at the start of Run() once the service is ready.
type AbstractRunProvider struct {
	running atomic.Bool
}

func (p *AbstractRunProvider) Init() error { return nil }
func (p *AbstractRunProvider) Close() error {
	p.running.Store(false)
	return nil
}
func (p *AbstractRunProvider) IsRunning() bool   { return p.running.Load() }
func (p *AbstractRunProvider) SetRunning(v bool) { p.running.Store(v) }

// ─── Init helper ──────────────────────────────────────────────────────────────

// InitProvider calls p.Init() and returns (p, cleanup, nil) on success,
// or (p, nil, error) on failure. The returned cleanup calls p.Close().
//
// Use this in Wire-compatible constructors:
//
//	func NewClient(cfg *Config) (*Client, func(), error) {
//	    return provider.InitProvider(New(cfg))
//	}
//
// InitProvider does NOT start RunProvider goroutines. The Container's
// MustRun is solely responsible for launching goroutines.
func InitProvider[T Provider](p T) (T, func(), error) {
	name := Name(p)

	logrus.Debugf("%s initializing...", name)
	if err := p.Init(); err != nil {
		return p, nil, fmt.Errorf("%s init failed: %w", name, err)
	}
	logrus.Infof("%s initialized", name)

	cleanup := func() {
		logrus.Debugf("%s closing...", name)
		if err := p.Close(); err != nil {
			logrus.WithError(err).Warnf("%s close failed", name)
		} else {
			logrus.Debugf("%s closed", name)
		}
	}

	return p, cleanup, nil
}

// ─── Utilities ────────────────────────────────────────────────────────────────

// Name returns the fully-qualified type name of p (e.g. "redis.Client").
func Name(p any) string {
	t := reflect.TypeOf(p)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t.String()
}
