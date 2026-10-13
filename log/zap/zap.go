// Package zaplog provides a zap logging Provider.
//
// Init replaces the global zap logger (zap.L / zap.S) and optionally creates
// a structured per-component logger accessible via Client().Logger.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(zaplog.New(nil))   // nil → reads from env/config.yaml
//	c.MustInit(redis.New(nil))
//	c.MustRun()
//
// Wire usage (Pattern B) – add zaplog.ProviderSet to wire.Build():
//
//	wire.Build(zaplog.NewConfigFromEnv, zaplog.NewClient, ...)
package zaplog

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/internal/config"
	"github.com/versegeek/toolkit/provider"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds zap logger configuration.
type Config struct {
	// Level is the minimum enabled logging level (debug/info/warn/error/dpanic/panic/fatal).
	Level string
	// Encoding selects the output format: "json" (default) or "console".
	Encoding string
	// OutputPaths is a list of URLs or file paths to write logging output to.
	// Defaults to []string{"stderr"}.
	OutputPaths []string
	// Development puts the logger in development mode which changes the behaviour
	// of DPanicLevel and takes stacktraces more liberally.
	Development bool
	// DisableCaller suppresses filename:line in every log entry.
	DisableCaller bool
	// DisableStacktrace suppresses automatic stacktrace capturing.
	DisableStacktrace bool
}

const (
	defaultZapLevel      = "info"
	defaultZapEncoding   = "json"
	defaultDevelopment   = false
	defaultDisableCaller = false
)

// NewConfigFromEnv reads zap configuration from the environment.
//
// Keys: ZAP_LEVEL, ZAP_ENCODING (json|console), ZAP_OUTPUT (stderr|stdout|file path),
// ZAP_DEVELOPMENT, ZAP_DISABLE_CALLER, ZAP_DISABLE_STACKTRACE.
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("ZAP_LEVEL", defaultZapLevel)
	v.SetDefault("ZAP_ENCODING", defaultZapEncoding)
	v.SetDefault("ZAP_OUTPUT", "stderr")
	v.SetDefault("ZAP_DEVELOPMENT", defaultDevelopment)
	v.SetDefault("ZAP_DISABLE_CALLER", defaultDisableCaller)
	v.SetDefault("ZAP_DISABLE_STACKTRACE", false)

	return &Config{
		Level:             v.GetString("ZAP_LEVEL"),
		Encoding:          v.GetString("ZAP_ENCODING"),
		OutputPaths:       []string{v.GetString("ZAP_OUTPUT")},
		Development:       v.GetBool("ZAP_DEVELOPMENT"),
		DisableCaller:     v.GetBool("ZAP_DISABLE_CALLER"),
		DisableStacktrace: v.GetBool("ZAP_DISABLE_STACKTRACE"),
	}
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client wraps a *zap.Logger as a lifecycle-managed Provider.
// It also replaces the process-wide global logger (zap.L / zap.S) on Init.
type Client struct {
	provider.AbstractProvider

	Config *Config
	Logger *zap.Logger
}

var _ provider.Provider = (*Client)(nil)

// New creates a Client without building the logger.
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

// Init builds the zap logger from Config and replaces the global logger.
func (c *Client) Init() error {
	level, err := zapcore.ParseLevel(c.Config.Level)
	if err != nil {
		return fmt.Errorf("zap: invalid level %q: %w", c.Config.Level, err)
	}

	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       c.Config.Development,
		DisableCaller:     c.Config.DisableCaller,
		DisableStacktrace: c.Config.DisableStacktrace,
		Encoding:          c.Config.Encoding,
		EncoderConfig:     encoderConfig(c.Config.Development),
		OutputPaths:       c.Config.OutputPaths,
		ErrorOutputPaths:  []string{"stderr"},
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return fmt.Errorf("zap: failed to build logger: %w", err)
	}

	c.Logger = logger
	// Replace the global loggers so that zap.L() / zap.S() work everywhere.
	zap.ReplaceGlobals(logger)
	return nil
}

// Close flushes buffered log entries. Should always be deferred.
func (c *Client) Close() error {
	if c.Logger != nil {
		// Sync may return an error on some platforms (e.g., /dev/stderr on Linux);
		// we intentionally ignore it to avoid noisy shutdown logs.
		_ = c.Logger.Sync()
	}
	return nil
}

// encoderConfig returns sane encoder defaults for the given mode.
func encoderConfig(development bool) zapcore.EncoderConfig {
	if development {
		cfg := zap.NewDevelopmentEncoderConfig()
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return cfg
	}
	return zap.NewProductionEncoderConfig()
}
