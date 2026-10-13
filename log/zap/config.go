package zap

import (
	"fmt"

	"github.com/versegeek/toolkit/config"
)

// Config holds the logger settings.
type Config struct {
	// Level is the minimum level: debug, info, warn, error, dpanic, panic or
	// fatal.
	Level string
	// Encoding is json or console.
	Encoding string
	// OutputPaths are zap sinks such as stderr, stdout or file paths.
	OutputPaths []string
	// Development enables development mode: DPanic panics and stack traces
	// are attached from the warn level.
	Development bool
	// DisableCaller omits the calling file and line.
	DisableCaller bool
	// DisableStacktrace omits automatic stack traces.
	DisableStacktrace bool
}

// NewConfigFromEnv reads the settings from ZAP_LEVEL (default info),
// ZAP_ENCODING (default json), ZAP_OUTPUT (comma-separated, default stderr),
// ZAP_DEVELOPMENT, ZAP_DISABLE_CALLER and ZAP_DISABLE_STACKTRACE.
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("zap: load config failed: %w", err)
	}
	v.SetDefault("ZAP_LEVEL", "info")
	v.SetDefault("ZAP_ENCODING", "json")
	v.SetDefault("ZAP_OUTPUT", "stderr")
	v.SetDefault("ZAP_DEVELOPMENT", false)
	v.SetDefault("ZAP_DISABLE_CALLER", false)
	v.SetDefault("ZAP_DISABLE_STACKTRACE", false)

	return &Config{
		Level:             v.GetString("ZAP_LEVEL"),
		Encoding:          v.GetString("ZAP_ENCODING"),
		OutputPaths:       config.StringSlice(v, "ZAP_OUTPUT"),
		Development:       v.GetBool("ZAP_DEVELOPMENT"),
		DisableCaller:     v.GetBool("ZAP_DISABLE_CALLER"),
		DisableStacktrace: v.GetBool("ZAP_DISABLE_STACKTRACE"),
	}, nil
}
