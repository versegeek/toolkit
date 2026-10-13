// Package zap provides a zap logger as a provider. Init also makes it the
// process-wide logger: zap.L and zap.S return it and the slog default logger,
// which the toolkit itself logs through, writes to it. Close flushes the
// logger and restores the previous globals.
//
// Initialise it before the other providers so that their logs use it:
//
//	g.MustInit(zaplog.New(nil))
//	g.MustInit(redis.New(nil))
package zap

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/google/wire"
	uberzap "go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"

	"github.com/versegeek/toolkit/internal/logbridge"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// Client is a zap logger managed as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	logger  *uberzap.Logger
	restore func()
}

var _ provider.Provider = (*Client)(nil)

// New returns a Client whose logger is not built yet.
func New(cfg *Config) *Client {
	return &Client{Config: cfg}
}

// NewClient returns an initialised Client and a cleanup that closes it.
func NewClient(cfg *Config) (*Client, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Logger returns the zap logger. It is nil before Init.
func (c *Client) Logger() *uberzap.Logger {
	return c.logger
}

// Init builds the logger and installs it as the zap global logger and the
// slog default logger.
func (c *Client) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}

	level, err := zapcore.ParseLevel(c.Config.Level)
	if err != nil {
		return fmt.Errorf("zap: invalid level %q: %w", c.Config.Level, err)
	}
	encoderConfig := uberzap.NewProductionEncoderConfig()
	if c.Config.Development {
		encoderConfig = uberzap.NewDevelopmentEncoderConfig()
	}
	outputs := c.Config.OutputPaths
	if len(outputs) == 0 {
		outputs = []string{"stderr"}
	}

	logger, err := uberzap.Config{
		Level:             uberzap.NewAtomicLevelAt(level),
		Development:       c.Config.Development,
		DisableCaller:     c.Config.DisableCaller,
		DisableStacktrace: c.Config.DisableStacktrace,
		Encoding:          c.Config.Encoding,
		EncoderConfig:     encoderConfig,
		OutputPaths:       outputs,
		ErrorOutputPaths:  []string{"stderr"},
	}.Build()
	if err != nil {
		return fmt.Errorf("zap: build logger failed: %w", err)
	}

	undoGlobals := uberzap.ReplaceGlobals(logger)
	restoreSlog := logbridge.SetDefault(slog.New(
		zapslog.NewHandler(logger.Core(), zapslog.WithCaller(!c.Config.DisableCaller)),
	))
	c.logger = logger
	c.restore = func() {
		restoreSlog()
		undoGlobals()
	}
	return nil
}

// Close flushes buffered entries and restores the previous global loggers.
func (c *Client) Close() error {
	if c.logger == nil {
		return nil
	}
	c.restore()
	if err := c.logger.Sync(); err != nil && !onlyConsoleSyncErrors(err) {
		return fmt.Errorf("zap: sync failed: %w", err)
	}
	return nil
}

// onlyConsoleSyncErrors reports whether err only holds failures to sync
// stdout or stderr. Terminals and pipes do not support fsync and fail with a
// platform-dependent error (ENOTTY, EINVAL or EBADF), which is harmless.
func onlyConsoleSyncErrors(err error) bool {
	errs := []error{err}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		errs = joined.Unwrap()
	}
	for _, e := range errs {
		var pathErr *fs.PathError
		if !errors.As(e, &pathErr) || (pathErr.Path != os.Stdout.Name() && pathErr.Path != os.Stderr.Name()) {
			return false
		}
	}
	return true
}
