// Package config provides shared Viper configuration loading for all providers.
package config

import "github.com/spf13/viper"

// LoadFromFile configures a Viper instance to read from a config file
// and environment variables.
//
// Search order (highest priority last, i.e. env wins):
//  1. config.yaml / config.yml in ".", "./config", "./configs"
//  2. Real environment variables (AutomaticEnv)
//
// Each provider calls this with its own viper.New() instance so that
// every provider gets an independent, correctly scoped Viper instance.
func LoadFromFile(v *viper.Viper) {
	v.AutomaticEnv()

	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("./configs")

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Ignore "file not found" – config file is always optional.
	_ = v.ReadInConfig()
}
