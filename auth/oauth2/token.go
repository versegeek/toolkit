package oauth2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// ExtractBearerToken extracts the Bearer token string from an Authorization
// header value (e.g. "Bearer eyJ..."). Returns the token and true on success.
func ExtractBearerToken(authHeader string) (string, bool) {
	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) || !strings.EqualFold(authHeader[:len(prefix)], prefix) {
		return "", false
	}
	tok := strings.TrimSpace(authHeader[len(prefix):])
	if tok == "" {
		return "", false
	}
	return tok, true
}

// ExtractBearerTokenFromRequest extracts the Bearer token from the
// Authorization header of an HTTP request.
func ExtractBearerTokenFromRequest(r *http.Request) (string, bool) {
	return ExtractBearerToken(r.Header.Get("Authorization"))
}

// IsTokenExpired reports whether the token has expired or will expire within
// the given grace duration. Use a non-zero grace to refresh proactively.
//
//	oauth2util.IsTokenExpired(token, 30*time.Second)
func IsTokenExpired(t *oauth2.Token, grace time.Duration) bool {
	if t == nil || t.AccessToken == "" {
		return true
	}
	if t.Expiry.IsZero() {
		return false // no expiry set — treat as valid
	}
	return time.Until(t.Expiry) <= grace
}

// ParseJWTClaims decodes the payload of a JWT access token without verifying
// the signature. Useful for reading standard claims (sub, exp, iss, etc.)
// in a trusted context where signature verification is done upstream
// (e.g. by the API gateway or a middleware).
//
// For production use, always verify the signature with a proper JWT library.
func ParseJWTClaims(tokenStr string) (map[string]any, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("oauth2: invalid JWT format (expected 3 parts, got %d)", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("oauth2: failed to decode JWT payload: %w", err)
	}

	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("oauth2: failed to unmarshal JWT claims: %w", err)
	}

	return claims, nil
}

// TokenSubject returns the "sub" claim from a JWT access token,
// or an error if the token is malformed or the claim is missing.
func TokenSubject(tokenStr string) (string, error) {
	claims, err := ParseJWTClaims(tokenStr)
	if err != nil {
		return "", err
	}
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", fmt.Errorf("oauth2: missing or empty 'sub' claim")
	}
	return sub, nil
}

// BuildAuthURL is a stateless helper that constructs an OAuth2 authorization
// URL without requiring an initialized Client. Useful in CLI tools or scripts.
func BuildAuthURL(authEndpoint, clientID, redirectURL, state string, scopes []string) string {
	cfg := &oauth2.Config{
		ClientID:    clientID,
		RedirectURL: redirectURL,
		Endpoint:    oauth2.Endpoint{AuthURL: authEndpoint},
		Scopes:      scopes,
	}
	return cfg.AuthCodeURL(state)
}
