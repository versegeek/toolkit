package provider

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
)

// InitProvider initialises p and returns it with a cleanup that closes it,
// matching the (T, func(), error) shape of a Wire provider:
//
//	func NewClient(cfg *Config) (*Client, func(), error) {
//		return provider.InitProvider(New(cfg))
//	}
//
// The cleanup closes p at most once, however often it is called. A disabled
// provider (see Enabler) is returned without Init and with a cleanup that does
// nothing. InitProvider never calls Run; package lifecycle does.
func InitProvider[T Provider](p T) (T, func(), error) {
	name := Name(p)
	if !IsEnabled(p) {
		slog.Info("provider disabled", "provider", name)
		return p, func() {}, nil
	}
	if err := p.Init(); err != nil {
		var zero T
		return zero, nil, fmt.Errorf("%s: init failed: %w", name, err)
	}
	slog.Info("provider initialized", "provider", name)

	var once sync.Once
	cleanup := func() {
		once.Do(func() {
			if err := p.Close(); err != nil {
				slog.Error("provider close failed", "provider", name, "error", err)
				return
			}
			slog.Info("provider closed", "provider", name)
		})
	}
	return p, cleanup, nil
}

// Name returns the type name of p without a pointer prefix, such as
// "redis.Client".
func Name(p any) string {
	return strings.TrimPrefix(fmt.Sprintf("%T", p), "*")
}
