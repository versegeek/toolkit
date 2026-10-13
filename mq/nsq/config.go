package nsq

import (
	"fmt"

	"github.com/versegeek/toolkit/config"
)

// Config holds the NSQ settings.
type Config struct {
	// NsqdAddress is the nsqd TCP address (host:port) the producer publishes to.
	NsqdAddress string
	// LookupdAddresses are the nsqlookupd HTTP addresses, kept here for the
	// consumers an application creates itself.
	LookupdAddresses []string
}

const defaultNsqdAddress = "127.0.0.1:4150"

// NewConfigFromEnv reads the settings from NSQ_NSQD_ADDRESS (default
// 127.0.0.1:4150) and NSQ_LOOKUPD_ADDRESSES (comma-separated).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("nsq: load config failed: %w", err)
	}
	v.SetDefault("NSQ_NSQD_ADDRESS", defaultNsqdAddress)
	v.SetDefault("NSQ_LOOKUPD_ADDRESSES", "")

	return &Config{
		NsqdAddress:      v.GetString("NSQ_NSQD_ADDRESS"),
		LookupdAddresses: config.StringSlice(v, "NSQ_LOOKUPD_ADDRESSES"),
	}, nil
}
