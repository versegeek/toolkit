// Package config loads provider configuration with Viper from environment
// variables and an optional config.yaml file.
//
// Environment variables take precedence over the file. Provider packages use
// it in their NewConfigFromEnv functions; applications can use it for their
// own settings too.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// New returns a Viper instance prepared by LoadFromFile.
func New() (*viper.Viper, error) {
	v := viper.New()
	if err := LoadFromFile(v); err != nil {
		return nil, err
	}
	return v, nil
}

// LoadFromFile makes v read environment variables and the first config.yaml
// found in ".", "./config" or "./configs", relative to the working directory.
// Environment variables take precedence over the file. A missing file is not
// an error; an unreadable or malformed one is.
func LoadFromFile(v *viper.Viper) error {
	v.AutomaticEnv()

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./configs")

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return nil
		}
		return fmt.Errorf("config: read config file failed: %w", err)
	}
	return nil
}

// StringSlice returns the list stored under key. A YAML list is returned
// element by element; a plain string, such as an environment variable, is
// split on commas. Entries are trimmed and empty ones dropped.
func StringSlice(v *viper.Viper, key string) []string {
	var items []string
	switch v.Get(key).(type) {
	case []any, []string:
		items = v.GetStringSlice(key)
	default:
		items = strings.Split(v.GetString(key), ",")
	}

	out := make([]string, 0, len(items))
	for _, item := range items {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	return out
}
