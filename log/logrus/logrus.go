// Package logprovider provides a logrus logging Provider.
//
// Init configures the global logrus logger. Use this as the first provider
// so that all subsequent Init logs use the configured format and level.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(logprovider.New(nil))  // nil → reads from env/config.yaml
//	c.MustInit(redis.New(nil))
//	c.MustRun()
//
// Wire usage (Pattern B) – add logprovider.ProviderSet to wire.Build():
//
//	wire.Build(logprovider.NewConfigFromEnv, logprovider.NewClient, ...)
package logprovider

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/config"
	"github.com/versegeek/toolkit/provider"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds logrus configuration.
type Config struct {
	Level        logrus.Level
	Formatter    logrus.Formatter
	Output       io.Writer
	ReportCaller bool
}

const (
	defaultLevel        = "info"
	defaultFormatter    = "json"
	defaultOutput       = "stderr"
	defaultReportCaller = false
)

// NewConfigFromEnv reads logrus configuration from the environment.
// Keys: LOGRUS_LEVEL, LOGRUS_FORMATTER (json|text|text_color), LOGRUS_OUTPUT (stderr|stdout), LOGRUS_REPORT_CALLER.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("LOGRUS_LEVEL", defaultLevel)
	v.SetDefault("LOGRUS_FORMATTER", defaultFormatter)
	v.SetDefault("LOGRUS_OUTPUT", defaultOutput)
	v.SetDefault("LOGRUS_REPORT_CALLER", defaultReportCaller)

	level, err := logrus.ParseLevel(v.GetString("LOGRUS_LEVEL"))
	if err != nil {
		level = logrus.InfoLevel
	}

	var formatter logrus.Formatter
	switch v.GetString("LOGRUS_FORMATTER") {
	case "text":
		formatter = &logrus.TextFormatter{}
	case "text_color":
		formatter = &logrus.TextFormatter{ForceColors: true, FullTimestamp: true}
	default: // "json"
		formatter = &logrus.JSONFormatter{}
	}

	var output io.Writer
	switch v.GetString("LOGRUS_OUTPUT") {
	case "stdout":
		output = os.Stdout
	default: // "stderr"
		output = os.Stderr
	}

	return &Config{
		Level:        level,
		Formatter:    formatter,
		Output:       output,
		ReportCaller: v.GetBool("LOGRUS_REPORT_CALLER"),
	}
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client configures the global logrus logger as a lifecycle-managed Provider.
type Client struct {
	provider.AbstractProvider
	Config *Config
}

var _ provider.Provider = (*Client)(nil)

// New creates a Client.
// If config is nil, NewConfigFromEnv() is used.
func New(cfg *Config) *Client {
	if cfg == nil {
		cfg = NewConfigFromEnv()
	}
	return &Client{Config: cfg}
}

// NewClient is the Wire-compatible constructor.
func NewClient(cfg *Config) (*Client, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Init applies the configuration to the global logrus logger.
func (c *Client) Init() error {
	logrus.SetLevel(c.Config.Level)
	logrus.SetFormatter(c.Config.Formatter)
	logrus.SetOutput(c.Config.Output)
	logrus.SetReportCaller(c.Config.ReportCaller)
	return nil
}
