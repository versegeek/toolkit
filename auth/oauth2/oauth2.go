// Package oauth2 provides an OAuth2 client (golang.org/x/oauth2) as a
// provider, for both the authorization code flow and the client credentials
// flow. Every request to the token endpoint is bounded by Config.Timeout.
//
// Authorization code flow (user login):
//
//	url := client.AuthURL(state)
//	token, err := client.Exchange(ctx, code)
//	httpClient, err := client.HTTPClient(ctx, token)
//
// Client credentials flow (service to service), with the token cached and
// refreshed before it expires:
//
//	token, err := client.Token()
//	httpClient, err := client.HTTPClient(ctx, nil)
package oauth2

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/wire"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/versegeek/toolkit/provider"
)

// ProviderSet provides a *Config from the environment and an initialised
// *Client.
var ProviderSet = wire.NewSet(NewConfigFromEnv, NewClient)

// errNoClientCredentials is returned when a client credentials operation is
// used without Config.UseClientCredentials.
var errNoClientCredentials = errors.New("oauth2: client credentials flow is not enabled (set UseClientCredentials)")

// Client is an OAuth2 client managed as a provider.
type Client struct {
	provider.AbstractProvider

	// Config holds the settings; Init loads them from the environment when
	// nil.
	Config *Config

	oauthConfig *oauth2.Config
	httpClient  *http.Client
	tokens      oauth2.TokenSource
}

var _ provider.Provider = (*Client)(nil)

// New returns a Client that is not initialised yet.
func New(cfg *Config) *Client {
	return &Client{Config: cfg}
}

// NewClient returns an initialised Client and a cleanup that closes it.
func NewClient(cfg *Config) (*Client, func(), error) {
	return provider.InitProvider(New(cfg))
}

// Init prepares both flows. With the client credentials flow it fetches the
// first token, which also checks the credentials.
func (c *Client) Init() error {
	if c.Config == nil {
		cfg, err := NewConfigFromEnv()
		if err != nil {
			return err
		}
		c.Config = cfg
	}

	timeout := c.Config.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	c.httpClient = &http.Client{Timeout: timeout}
	c.oauthConfig = &oauth2.Config{
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
		cc := &clientcredentials.Config{
			ClientID:     c.Config.ClientID,
			ClientSecret: c.Config.ClientSecret,
			TokenURL:     c.Config.TokenURL,
			Scopes:       c.Config.Scopes,
		}
		// The context is kept by the token source for every refresh.
		tokens := cc.TokenSource(c.withHTTPClient(context.Background()))
		if _, err := tokens.Token(); err != nil {
			return fmt.Errorf("oauth2: fetch client credentials token failed: %w", err)
		}
		c.tokens = tokens
	}
	return nil
}

// AuthURL returns the authorization endpoint URL to redirect a user to.
// state must be an unguessable value that the callback verifies.
func (c *Client) AuthURL(state string, opts ...oauth2.AuthCodeOption) string {
	return c.oauthConfig.AuthCodeURL(state, opts...)
}

// Exchange trades an authorization code from the callback for a token.
func (c *Client) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	token, err := c.oauthConfig.Exchange(c.withHTTPClient(ctx), code, opts...)
	if err != nil {
		return nil, fmt.Errorf("oauth2: exchange code failed: %w", err)
	}
	return token, nil
}

// RefreshToken trades a refresh token for a new token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token, err := c.oauthConfig.TokenSource(c.withHTTPClient(ctx), &oauth2.Token{RefreshToken: refreshToken}).Token()
	if err != nil {
		return nil, fmt.Errorf("oauth2: refresh token failed: %w", err)
	}
	return token, nil
}

// Token returns the client credentials token, fetching a new one only when
// the cached token has expired.
func (c *Client) Token() (*oauth2.Token, error) {
	if c.tokens == nil {
		return nil, errNoClientCredentials
	}
	token, err := c.tokens.Token()
	if err != nil {
		return nil, fmt.Errorf("oauth2: fetch client credentials token failed: %w", err)
	}
	return token, nil
}

// TokenSource returns the caching client credentials token source, or nil
// when that flow is not enabled.
func (c *Client) TokenSource() oauth2.TokenSource {
	return c.tokens
}

// HTTPClient returns an *http.Client that authorises every request with
// token, refreshing it when it expires. A nil token uses the client
// credentials flow.
func (c *Client) HTTPClient(ctx context.Context, token *oauth2.Token) (*http.Client, error) {
	ctx = c.withHTTPClient(ctx)
	if token != nil {
		return c.oauthConfig.Client(ctx, token), nil
	}
	if c.tokens == nil {
		return nil, errNoClientCredentials
	}
	return oauth2.NewClient(ctx, c.tokens), nil
}

// withHTTPClient makes token requests made with ctx use the client's
// timeout-bounded HTTP client unless ctx already carries one.
func (c *Client) withHTTPClient(ctx context.Context) context.Context {
	if ctx.Value(oauth2.HTTPClient) != nil {
		return ctx
	}
	return context.WithValue(ctx, oauth2.HTTPClient, c.httpClient)
}
