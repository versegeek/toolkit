package oauth2

import (
	"fmt"
	"time"

	"github.com/versegeek/toolkit/config"
)

// Config holds the OAuth2 client settings.
type Config struct {
	ClientID     string
	ClientSecret string
	// AuthURL is the authorization endpoint of the authorization code flow.
	AuthURL string
	// TokenURL is the token endpoint used by both flows.
	TokenURL string
	// RedirectURL is the callback registered with the authorization server.
	RedirectURL string
	Scopes      []string
	// UseClientCredentials enables the client credentials flow: Init fetches
	// a token and Token keeps it fresh.
	UseClientCredentials bool
	// Timeout bounds each request to the token endpoint.
	Timeout time.Duration
}

const defaultTimeout = 10 * time.Second

// NewConfigFromEnv reads the settings from OAUTH2_CLIENT_ID,
// OAUTH2_CLIENT_SECRET, OAUTH2_AUTH_URL, OAUTH2_TOKEN_URL,
// OAUTH2_REDIRECT_URL, OAUTH2_SCOPES (comma-separated),
// OAUTH2_USE_CLIENT_CREDENTIALS (default false) and OAUTH2_TIMEOUT (default
// 10s).
func NewConfigFromEnv() (*Config, error) {
	v, err := config.New()
	if err != nil {
		return nil, fmt.Errorf("oauth2: load config failed: %w", err)
	}
	v.SetDefault("OAUTH2_CLIENT_ID", "")
	v.SetDefault("OAUTH2_CLIENT_SECRET", "")
	v.SetDefault("OAUTH2_AUTH_URL", "")
	v.SetDefault("OAUTH2_TOKEN_URL", "")
	v.SetDefault("OAUTH2_REDIRECT_URL", "")
	v.SetDefault("OAUTH2_SCOPES", "")
	v.SetDefault("OAUTH2_USE_CLIENT_CREDENTIALS", false)
	v.SetDefault("OAUTH2_TIMEOUT", defaultTimeout)

	return &Config{
		ClientID:             v.GetString("OAUTH2_CLIENT_ID"),
		ClientSecret:         v.GetString("OAUTH2_CLIENT_SECRET"),
		AuthURL:              v.GetString("OAUTH2_AUTH_URL"),
		TokenURL:             v.GetString("OAUTH2_TOKEN_URL"),
		RedirectURL:          v.GetString("OAUTH2_REDIRECT_URL"),
		Scopes:               config.StringSlice(v, "OAUTH2_SCOPES"),
		UseClientCredentials: v.GetBool("OAUTH2_USE_CLIENT_CREDENTIALS"),
		Timeout:              v.GetDuration("OAUTH2_TIMEOUT"),
	}, nil
}
