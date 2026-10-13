// Package provider defines the building blocks the toolkit manages.
//
// A provider owns one external resource or service. It is initialised, used
// and closed; a run provider also works in the foreground between Init and
// Close:
//
//	New → Init → [Run] → Close
//
// Package lifecycle drives these steps, either through a lifecycle.Group or
// through lifecycle.Run together with a Wire injector.
package provider

import "sync/atomic"

// Provider is a component that owns an external resource or service.
//
// Init opens connections or prepares resources. When it fails it must release
// anything it opened before returning the error. Close releases the resources
// and must be safe to call when Init was never called or failed.
type Provider interface {
	Init() error
	Close() error
}

// RunProvider is a Provider that also works in the foreground, such as an
// HTTP or gRPC server.
//
// Run blocks until the provider stops and returns nil when Close stopped it.
// Close must make Run return and must be safe to call more than once.
// IsRunning reports whether Run has reached its steady state, for example
// whether the listener is bound.
type RunProvider interface {
	Provider
	Run() error
	IsRunning() bool
}

// Enabler is implemented by providers that configuration can switch off.
//
// Enabled is checked before Init and again before Run; a provider that
// reports false is skipped at that step. A provider whose configuration is
// only loaded in Init should report true until then.
type Enabler interface {
	Enabled() bool
}

// IsEnabled reports whether p is enabled. Providers that do not implement
// Enabler are always enabled.
func IsEnabled(p Provider) bool {
	e, ok := p.(Enabler)
	return !ok || e.Enabled()
}

// AbstractProvider is an embeddable Provider whose Init and Close do nothing.
type AbstractProvider struct{}

// Init does nothing.
func (*AbstractProvider) Init() error { return nil }

// Close does nothing.
func (*AbstractProvider) Close() error { return nil }

// AbstractRunProvider is an embeddable base for run providers that tracks
// whether Run is in its steady state. Call SetRunning(true) once Run is ready
// and SetRunning(false) when it returns.
type AbstractRunProvider struct {
	running atomic.Bool
}

// Init does nothing.
func (*AbstractRunProvider) Init() error { return nil }

// Close marks the provider as not running.
func (p *AbstractRunProvider) Close() error {
	p.running.Store(false)
	return nil
}

// IsRunning reports whether the provider is running.
func (p *AbstractRunProvider) IsRunning() bool { return p.running.Load() }

// SetRunning records whether the provider is running.
func (p *AbstractRunProvider) SetRunning(running bool) { p.running.Store(running) }
