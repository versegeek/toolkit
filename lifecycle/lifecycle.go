// Package lifecycle initialises, runs and closes providers in order.
//
// Manual composition adds providers to a Group, which runs every run provider
// and closes everything in reverse order on shutdown:
//
//	g := lifecycle.New()
//	defer g.MustClose()
//	g.MustInit(zaplog.New(nil))
//	g.MustInit(redis.New(nil))
//	srv := http.New(nil)
//	g.MustInit(srv)
//	srv.Router.GET("/ping", ping)
//	g.MustRun()
//
// Wire composition builds and initialises providers with an injector, then
// hands the run providers and the injector's cleanup to Run:
//
//	app, cleanup, err := InitApp()
//	if err != nil {
//		log.Fatal(err)
//	}
//	if err := lifecycle.Run(context.Background(), cleanup, app.HTTP); err != nil {
//		log.Fatal(err)
//	}
package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/versegeek/toolkit/provider"
)

// Run starts every enabled run provider in its own goroutine and blocks until
// ctx is done, the process receives SIGINT or SIGTERM, or a run provider
// fails. It then closes the run providers in reverse order, calls cleanup
// once and returns the run provider's error, if any.
//
// A second signal during shutdown terminates the process immediately.
func Run(ctx context.Context, cleanup func(), runners ...provider.RunProvider) error {
	return run(ctx, runners, func() error {
		if cleanup != nil {
			cleanup()
		}
		return nil
	})
}

// run implements Run and Group.Run; shutdown releases everything after the
// run providers have been stopped.
func run(ctx context.Context, runners []provider.RunProvider, shutdown func() error) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	errc := make(chan error, len(runners))
	started := make([]provider.RunProvider, 0, len(runners))
	for _, r := range runners {
		name := provider.Name(r)
		if !provider.IsEnabled(r) {
			slog.Info("provider disabled, not running", "provider", name)
			continue
		}
		started = append(started, r)
		go func() {
			slog.Info("provider running", "provider", name)
			if err := r.Run(); err != nil {
				errc <- fmt.Errorf("%s: run failed: %w", name, err)
			}
		}()
	}

	var runErr error
	select {
	case <-ctx.Done():
		slog.Info("shutting down", "cause", context.Cause(ctx))
	case runErr = <-errc:
		slog.Error("shutting down after a run provider failed", "error", runErr)
	}
	// Restore default signal handling so that a second signal exits at once.
	stop()

	for i := len(started) - 1; i >= 0; i-- {
		if err := started[i].Close(); err != nil {
			slog.Error("provider stop failed", "provider", provider.Name(started[i]), "error", err)
		}
	}
	return errors.Join(runErr, shutdown())
}
