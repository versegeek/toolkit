// Package container provides a lightweight lifecycle container for managing
// provider Init → Run → Close sequences.
//
// Two usage patterns:
//
// Pattern A – Container manages everything:
//
//	c := container.New()
//	defer c.MustClose()
//	c.MustInit(redis.New(nil))
//	c.MustRun()
//
// Pattern B – Wire/manual composition, Container only handles signals:
//
//	app, cleanup, err := InitApp()  // Wire-generated or hand-written
//	defer cleanup()
//	container.HandleSignals(cleanup)
package container

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/versegeek/toolkit/provider"
)

// Provider and RunProvider are re-exported for convenience so callers only
// need to import "container" rather than both "container" and "provider".
type Provider = provider.Provider
type RunProvider = provider.RunProvider

// Container is an ordered collection of Providers.
// Providers are initialised in insertion order and closed in reverse (LIFO).
//
// Each Container instance has its own sync.Once guards — multiple Container
// instances can coexist safely within the same process.
type Container struct {
	providers []Provider
	runOnce   sync.Once
	closeOnce sync.Once
}

// New creates an empty Container.
func New() *Container {
	return &Container{
		providers: make([]Provider, 0),
	}
}

// Add appends one or more already-created providers to the Container.
// Providers are not initialised here — use MustInit for that.
// Returns the Container for fluent chaining.
func (c *Container) Add(providers ...Provider) *Container {
	c.providers = append(c.providers, providers...)
	return c
}

// MustInit calls Init() on the provider and appends it to the Container.
// Panics on error.
func (c *Container) MustInit(p Provider) {
	name := provider.Name(p)
	logrus.Debugf("%s initializing...", name)

	if err := p.Init(); err != nil {
		logrus.WithError(err).Panicf("error during %s initialization", name)
	}

	c.providers = append(c.providers, p)
	logrus.Infof("%s initialized", name)
}

// MustRun launches goroutines for all RunProviders, then blocks until a
// termination signal (SIGINT / SIGTERM / SIGQUIT) is received.
// On signal, MustClose is called automatically.
// MustRun is idempotent per Container instance.
func (c *Container) MustRun() {
	c.runOnce.Do(func() {
		for _, p := range c.providers {
			if rp, ok := p.(RunProvider); ok {
				go c.launch(rp)
			}
		}
		c.waitForSignal()
	})
}

// MustClose shuts down all providers in LIFO order.
// Each provider's Close() is always called regardless of previous errors.
// If any Close() calls fail, the errors are logged and MustClose panics
// with a combined error after all providers have been closed.
// MustClose is idempotent per Container instance.
func (c *Container) MustClose() {
	c.closeOnce.Do(func() {
		var errs []error
		for i := len(c.providers) - 1; i >= 0; i-- {
			name := provider.Name(c.providers[i])
			logrus.Debugf("%s closing...", name)

			if err := c.providers[i].Close(); err != nil {
				logrus.WithError(err).Errorf("%s failed to close", name)
				errs = append(errs, fmt.Errorf("%s: %w", name, err))
			} else {
				logrus.Debugf("%s closed", name)
			}
		}
		if len(errs) > 0 {
			logrus.Panicf("MustClose: %d provider(s) failed to close: %v", len(errs), errors.Join(errs...))
		}
	})
}

// launch starts a RunProvider in a goroutine and logs any run error.
func (c *Container) launch(rp RunProvider) {
	name := provider.Name(rp)
	logrus.Debugf("%s launching...", name)

	if err := rp.Run(); err != nil {
		logrus.WithError(err).Errorf("%s run error", name)
	}
}

// waitForSignal blocks until SIGINT/SIGTERM/SIGQUIT, then calls MustClose.
func (c *Container) waitForSignal() {
	ch := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(ch)

	go func() {
		sig := <-ch
		logrus.Infof("received %s – shutting down...", sig)
		c.MustClose()
		close(done)
	}()

	<-done
}

// HandleSignals blocks until a termination signal is received, then calls
// shutdownFn. Use this in Pattern B where providers are initialised via Wire
// or a hand-written InitApp and only signal handling is delegated here.
func HandleSignals(shutdownFn func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer signal.Stop(ch)

	sig := <-ch
	logrus.Infof("received %s – shutting down...", sig)
	shutdownFn()
}
