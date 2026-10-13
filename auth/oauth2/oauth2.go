// Package oauth2 provides an OAuth2 client Provider with token lifecycle management.
//
// Supports both Authorization Code flow (user-facing) and Client Credentials
// flow (server-to-server). Tokens are automatically refreshed on expiry.
//
// Standalone usage (Pattern A, no Wire):
//
//	c := container.New()
//	c.MustInit(oauth2.New(nil))
//	c.MustRun()
//
// Wire usage (Pattern B) – add oauth2.ProviderSet to wire.Build():
//
//	wire.Build(oauth2.NewConfigFromEnv, oauth2.NewClient, ...)
//
// # Authorization Code flow (user login)
//
//	// 1. Redirect user to authorization URL
//	url := client.AuthURL("random-state")
//
//	// 2. Exchange code from callback for a token
//	token, err := client.Exchange(ctx, code)
//
//	// 3. Make authenticated HTTP requests
//	httpClient := client.HTTPClient(ctx, token)
//
// # Client Credentials flow (server-to-server)
//
//	// Token is fetched and refreshed automatically
//	token, err := client.Token(ctx)
package oauth2

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/versegeek/toolkit/internal/config"
	"github.com/versegeek/toolkit/provider"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// ─── Config ──────────────────────────────────────────────────────────────────

// Config holds OAuth2 client parameters.
type Config struct {
	ClientID     string
	ClientSecret string

	// AuthURL is the authorization endpoint (used for Authorization Code flow).
	AuthURL string

	// TokenURL is the token endpoint (used by both flows).
	TokenURL string

	// RedirectURL is the callback URL registered with the provider.
	RedirectURL string

	// Scopes is the list of requested OAuth2 scopes.
	Scopes []string

	// UseClientCredentials enables server-to-server token management.
	// When true, Init() sets up an auto-refreshing token source.
	UseClientCredentials bool
}

// NewConfigFromEnv reads OAuth2 configuration from the environment.
// All keys are prefixed with OAUTH2_.
//
//   - OAUTH2_CLIENT_ID, OAUTH2_CLIENT_SECRET
//   - OAUTH2_AUTH_URL, OAUTH2_TOKEN_URL, OAUTH2_REDIRECT_URL
//   - OAUTH2_SCOPES (comma-separated, e.g. "openid,profile,email")
//   - OAUTH2_USE_CLIENT_CREDENTIALS (bool, default false)
func NewConfigFromEnv() *Config {
	v := viper.New()
	config.LoadFromFile(v)

	v.SetDefault("OAUTH2_CLIENT_ID", "")
	v.SetDefault("OAUTH2_CLIENT_SECRET", "")
	v.SetDefault("OAUTH2_AUTH_URL", "")
	v.SetDefault("OAUTH2_TOKEN_URL", "")
	v.SetDefault("OAUTH2_REDIRECT_URL", "")
	v.SetDefault("OAUTH2_SCOPES", "")
	v.SetDefault("OAUTH2_USE_CLIENT_CREDENTIALS", false)

	cfg := &Config{
		ClientID:             v.GetString("OAUTH2_CLIENT_ID"),
		ClientSecret:         v.GetString("OAUTH2_CLIENT_SECRET"),
		AuthURL:              v.GetString("OAUTH2_AUTH_URL"),
		TokenURL:             v.GetString("OAUTH2_TOKEN_URL"),
		RedirectURL:          v.GetString("OAUTH2_REDIRECT_URL"),
		Scopes:               splitScopes(v.GetString("OAUTH2_SCOPES")),
		UseClientCredentials: v.GetBool("OAUTH2_USE_CLIENT_CREDENTIALS"),
	}

	logrus.WithFields(logrus.Fields{
		"clientID":             cfg.ClientID,
		"tokenURL":             cfg.TokenURL,
		"useClientCredentials": cfg.UseClientCredentials,
	}).Debug("OAuth2 config loaded")

	return cfg
}

func splitScopes(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	scopes := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			scopes = append(scopes, t)
		}
	}
	return scopes
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client is a lifecycle-managed OAuth2 client.
// It wraps golang.org/x/oauth2 and provides token management for both
// Authorization Code and Client Credentials flows.
type Client struct {
	provider.AbstractProvider

	Config   *Config
	oauthCfg *oauth2.Config
	ccCfg    *clientcredentials.Config // non-nil only for client credentials flow
}

var _ provider.Provider = (*Client)(nil)

// New creates a Client without initializing.
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

// Init configures the OAuth2 client. For Client Credentials flow, it verifies
// that the credentials are valid by fetching the first token eagerly.
func (c *Client) Init() error {
	c.oauthCfg = &oauth2.Config{
		ClientID:     c.Config.ClientID,
		ClientSecret: c.Config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  c.Config.AuthURL,
			TokenURL: c.Config.TokenURL,
		},
		RedirectURL: c.Config.RedirectURL,
		Scopes:      c.Config.Scopes,
	}

	if c.Config.UseClientCredentials {
		c.ccCfg = &clientcredentials.Config{
			ClientID:     c.Config.ClientID,
			ClientSecret: c.Config.ClientSecret,
			TokenURL:     c.Config.TokenURL,
			Scopes:       c.Config.Scopes,
		}
		// Eagerly verify credentials are valid.
		if _, err := c.ccCfg.Token(context.Background()); err != nil {
			return fmt.Errorf("oauth2: client credentials token fetch failed: %w", err)
		}
	}

	return nil
}

// ─── Authorization Code flow ──────────────────────────────────────────────────

// AuthURL returns the URL to redirect users to for authorization.
// state should be a random, unguessable string to prevent CSRF.
func (c *Client) AuthURL(state string, opts ...oauth2.AuthCodeOption) string {
	return c.oauthCfg.AuthCodeURL(state, opts...)
}

// Exchange converts an authorization code (from the OAuth2 callback) into a token.
func (c *Client) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return c.oauthCfg.Exchange(ctx, code, opts...)
}

// RefreshToken exchanges a refresh token for a new access token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	t := &oauth2.Token{RefreshToken: refreshToken}
	src := c.oauthCfg.TokenSource(ctx, t)
	newToken, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("oauth2: token refresh failed: %w", err)
	}
	return newToken, nil
}

// HTTPClient returns an *http.Client that automatically attaches the Bearer
// token to every request. Pass nil token to use client credentials.
func (c *Client) HTTPClient(ctx context.Context, t *oauth2.Token) (*http.Client, error) {
	if t == nil {
		if c.ccCfg == nil {
			return nil, fmt.Errorf("oauth2: no token provided and client credentials not configured")
		}
		return oauth2.NewClient(ctx, c.ccCfg.TokenSource(ctx)), nil
	}
	return c.oauthCfg.Client(ctx, t), nil
}

// ─── Client Credentials flow ──────────────────────────────────────────────────

// Token fetches a valid access token for the Client Credentials flow using ctx.
// A new token is fetched if the cached one is expired; ctx controls the timeout.
// Only available when UseClientCredentials is true.
func (c *Client) Token(ctx context.Context) (*oauth2.Token, error) {
	if c.ccCfg == nil {
		return nil, fmt.Errorf("oauth2: client credentials not configured (set UseClientCredentials=true)")
	}
	return c.ccCfg.Token(ctx)
}
