package logrus

import (
	"fmt"

	"github.com/versegeek/toolkit/config"
)

// Config holds the logger settings.
type Config struct {
	// Level is the minimum level: trace, debug, info, warn, error, fatal or
	// panic.
	Level string
	// Format is json or text.
	Format string
	// Output is stderr or stdout.
	Output string
	// ReportCaller adds the calling function and file to each entry.
	ReportCaller bool
}

// NewConfigFromEnv reads the settings from LOGRUS_LEVEL (default info),
// LOGRUS_FORMAT (default json), LOGRUS_OUTPUT (default stderr) and
// LOGRUS_REPORT_CALLER.
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("logrus: load config failed: %w", err)
	}
	v.SetDefault("LOGRUS_LEVEL", "info")
	v.SetDefault("LOGRUS_FORMAT", "json")
	v.SetDefault("LOGRUS_OUTPUT", "stderr")
	v.SetDefault("LOGRUS_REPORT_CALLER", false)

	return &Config{
		Level:        v.GetString("LOGRUS_LEVEL"),
		Format:       v.GetString("LOGRUS_FORMAT"),
		Output:       v.GetString("LOGRUS_OUTPUT"),
		ReportCaller: v.GetBool("LOGRUS_REPORT_CALLER"),
	}, nil
}
