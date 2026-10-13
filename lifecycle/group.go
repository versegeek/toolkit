package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/versegeek/toolkit/provider"
)

// Group manages the lifecycle of a hand-assembled set of providers.
// Providers are initialised in the order they are added and closed in reverse
// order. The zero value is ready to use.
type Group struct {
	mu        sync.Mutex
	providers []provider.Provider
	ran       atomic.Bool
	closeOnce sync.Once
}

// New returns an empty Group.
func New() *Group {
	return &Group{}
}

// Init initialises p and adds it to the group. A disabled provider (see
// provider.Enabler) is skipped. When Init fails, providers added earlier stay
// in the group and are released by Close.
func (g *Group) Init(p provider.Provider) error {
	name := provider.Name(p)
	if !provider.IsEnabled(p) {
		slog.Info("provider disabled", "provider", name)
		return nil
	}
	if err := p.Init(); err != nil {
		return fmt.Errorf("%s: init failed: %w", name, err)
	}

	g.mu.Lock()
	g.providers = append(g.providers, p)
	g.mu.Unlock()

	slog.Info("provider initialized", "provider", name)
	return nil
}

// MustInit is like Init but panics if p fails to initialise.
func (g *Group) MustInit(p provider.Provider) {
	if err := g.Init(p); err != nil {
		panic(err)
	}
}

// Run runs every enabled run provider in the group and blocks until ctx is
// done, the process receives SIGINT or SIGTERM, or a run provider fails. It
// then closes the group and returns the run error joined with any close
// errors. Run may be called only once.
func (g *Group) Run(ctx context.Context) error {
	if !g.ran.CompareAndSwap(false, true) {
		return errors.New("lifecycle: Group.Run called more than once")
	}

	g.mu.Lock()
	runners := make([]provider.RunProvider, 0, len(g.providers))
	for _, p := range g.providers {
		if r, ok := p.(provider.RunProvider); ok {
			runners = append(runners, r)
		}
	}
	g.mu.Unlock()

	return run(ctx, runners, g.Close)
}

// MustRun is like Run with a background context but panics on error.
func (g *Group) MustRun() {
	if err := g.Run(context.Background()); err != nil {
		panic(err)
	}
}

// Close closes every provider in reverse order of initialisation, continuing
// past failures, and returns the errors joined. Only the first call closes
// anything; later calls return nil.
func (g *Group) Close() error {
	var err error
	g.closeOnce.Do(func() {
		g.mu.Lock()
		providers := g.providers
		g.mu.Unlock()

		var errs []error
		for i := len(providers) - 1; i >= 0; i-- {
			name := provider.Name(providers[i])
			if cerr := providers[i].Close(); cerr != nil {
				errs = append(errs, fmt.Errorf("%s: close failed: %w", name, cerr))
				continue
			}
			slog.Info("provider closed", "provider", name)
		}
		err = errors.Join(errs...)
	})
	return err
}

// MustClose is like Close but panics on error.
func (g *Group) MustClose() {
	if err := g.Close(); err != nil {
		panic(err)
	}
}
