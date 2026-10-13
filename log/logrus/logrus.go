// Package logrus configures the logrus standard logger as a provider. Init
// also routes the slog default logger, which the toolkit itself logs
// through, into logrus; Close restores the previous slog logger.
//
// Initialise it before the other providers so that their logs use it:
//
//	g.MustInit(logruslog.New(nil))
//	g.MustInit(redis.New(nil))
package logrus

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/google/wire"
	"github.com/sirupsen/logrus"

	"github.com/versegeek/toolkit/internal/logbridge"
	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// Client configures the logrus standard logger as a provider.
type Client struct {
	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	restore func()
}

var _ provider.Provider = (*Client)(nil)

// New returns a Client that has not configured logrus yet.
func New(cfg *Config) *Client {
	return &Client{Config: cfg}
}

// NewClient returns an initialised Client and a cleanup that closes it.
func NewClient(cfg *Config) (*Client, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Logger returns the configured logrus standard logger.
func (c *Client) Logger() *logrus.Logger {
	return logrus.StandardLogger()
}

// Init applies the settings to the logrus standard logger and routes the
// slog default logger into it.
func (c *Client) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}

	level, err := logrus.ParseLevel(c.Config.Level)
	if err != nil {
		return fmt.Errorf("logrus: invalid level %q: %w", c.Config.Level, err)
	}
	var formatter logrus.Formatter
	switch c.Config.Format {
	case "json", "":
		formatter = &logrus.JSONFormatter{}
	case "text":
		formatter = &logrus.TextFormatter{FullTimestamp: true}
	default:
		return fmt.Errorf("logrus: invalid format %q: want json or text", c.Config.Format)
	}
	var output io.Writer
	switch c.Config.Output {
	case "stderr", "":
		output = os.Stderr
	case "stdout":
		output = os.Stdout
	default:
		return fmt.Errorf("logrus: invalid output %q: want stderr or stdout", c.Config.Output)
	}

	logger := logrus.StandardLogger()
	logger.SetLevel(level)
	logger.SetFormatter(formatter)
	logger.SetOutput(output)
	logger.SetReportCaller(c.Config.ReportCaller)

	c.restore = logbridge.SetDefault(slog.New(newSlogHandler(logger)))
	return nil
}

// Close restores the previous slog default logger.
func (c *Client) Close() error {
	if c.restore != nil {
		c.restore()
		c.restore = nil
	}
	return nil
}
