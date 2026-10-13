// Package httpserver runs net/http servers for the server providers. It binds
// the port before reporting readiness and shuts servers down gracefully
// within a deadline.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	// DefaultReadHeaderTimeout limits how long a client may take to send the
	// request headers.
	DefaultReadHeaderTimeout = 10 * time.Second
	// DefaultShutdownTimeout is used when Shutdown is given no timeout.
	DefaultShutdownTimeout = 5 * time.Second
)

// New returns an *http.Server for addr and h with a read-header timeout.
func New(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: DefaultReadHeaderTimeout,
	}
}

// Serve binds srv.Addr, calls ready(true) once the listener is bound and
// ready(false) when serving stops, and serves until srv is shut down.
// A shutdown is reported as nil.
func Serve(srv *http.Server, ready func(bool)) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	ready(true)
	defer ready(false)

	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown stops srv gracefully, waiting up to timeout (DefaultShutdownTimeout
// when not positive) for active requests, then closes the connections that
// are left. It is safe to call more than once and before srv serves.
func Shutdown(srv *http.Server, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = DefaultShutdownTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return errors.Join(fmt.Errorf("graceful shutdown timed out after %s", timeout), srv.Close())
		}
		return err
	}
	return nil
}
