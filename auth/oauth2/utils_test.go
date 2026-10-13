package oauth2

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// ─── ExtractBearerToken ───────────────────────────────────────────────────────

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		header    string
		wantToken string
		wantOK    bool
	}{
		{"valid Bearer", "Bearer eyJhbGciOiJIUzI1NiJ9", "eyJhbGciOiJIUzI1NiJ9", true},
		{"case insensitive", "bearer eyJhbGciOiJIUzI1NiJ9", "eyJhbGciOiJIUzI1NiJ9", true},
		{"no prefix", "eyJhbGciOiJIUzI1NiJ9", "", false},
		{"empty header", "", "", false},
		{"Bearer only", "Bearer ", "", false},
		{"Bearer with spaces", "Bearer   ", "", false},
		{"Basic auth", "Basic dXNlcjpwYXNz", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, ok := ExtractBearerToken(tt.header)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantToken, tok)
		})
	}
}

func TestExtractBearerTokenFromRequest(t *testing.T) {
	t.Run("with valid Bearer token", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer mytoken123")
		tok, ok := ExtractBearerTokenFromRequest(r)
		assert.True(t, ok)
		assert.Equal(t, "mytoken123", tok)
	})

	t.Run("without Authorization header", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		tok, ok := ExtractBearerTokenFromRequest(r)
		assert.False(t, ok)
		assert.Empty(t, tok)
	})
}

// ─── IsTokenExpired ───────────────────────────────────────────────────────────

func TestIsTokenExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		token *oauth2.Token
		grace time.Duration
		want  bool
	}{
		{
			name:  "nil token",
			token: nil,
			want:  true,
		},
		{
			name:  "empty access token",
			token: &oauth2.Token{AccessToken: ""},
			want:  true,
		},
		{
			name:  "zero expiry (no expiry set)",
			token: &oauth2.Token{AccessToken: "tok", Expiry: time.Time{}},
			want:  false,
		},
		{
			name:  "future expiry",
			token: &oauth2.Token{AccessToken: "tok", Expiry: now.Add(1 * time.Hour)},
			grace: 0,
			want:  false,
		},
		{
			name:  "past expiry",
			token: &oauth2.Token{AccessToken: "tok", Expiry: now.Add(-1 * time.Second)},
			grace: 0,
			want:  true,
		},
		{
			name:  "expiring within grace window",
			token: &oauth2.Token{AccessToken: "tok", Expiry: now.Add(10 * time.Second)},
			grace: 30 * time.Second,
			want:  true,
		},
		{
			name:  "expiry beyond grace window",
			token: &oauth2.Token{AccessToken: "tok", Expiry: now.Add(60 * time.Second)},
			grace: 30 * time.Second,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsTokenExpired(tt.token, tt.grace))
		})
	}
}

// ─── ParseJWTClaims ───────────────────────────────────────────────────────────

// makeJWT builds a fake JWT with the given payload claims (signature is not real).
func makeJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims)
	require.NoError(t, err)
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".fakesignature"
}

func TestParseJWTClaims(t *testing.T) {
	t.Run("valid JWT", func(t *testing.T) {
		tok := makeJWT(t, map[string]any{
			"sub": "user-123",
			"iss": "test-issuer",
			"exp": float64(9999999999),
		})
		claims, err := ParseJWTClaims(tok)
		require.NoError(t, err)
		assert.Equal(t, "user-123", claims["sub"])
		assert.Equal(t, "test-issuer", claims["iss"])
	})

	t.Run("wrong number of parts", func(t *testing.T) {
		_, err := ParseJWTClaims("onlyone")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "3 parts")
	})

	t.Run("invalid base64 payload", func(t *testing.T) {
		_, err := ParseJWTClaims("header.!!!invalid!!!.sig")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode")
	})

	t.Run("payload not JSON", func(t *testing.T) {
		badPayload := base64.RawURLEncoding.EncodeToString([]byte("not-json"))
		_, err := ParseJWTClaims("header." + badPayload + ".sig")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal")
	})
}

// ─── TokenSubject ─────────────────────────────────────────────────────────────

func TestTokenSubject(t *testing.T) {
	t.Run("valid sub claim", func(t *testing.T) {
		tok := makeJWT(t, map[string]any{"sub": "user-abc"})
		sub, err := TokenSubject(tok)
		require.NoError(t, err)
		assert.Equal(t, "user-abc", sub)
	})

	t.Run("missing sub claim", func(t *testing.T) {
		tok := makeJWT(t, map[string]any{"iss": "issuer"})
		_, err := TokenSubject(tok)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sub")
	})

	t.Run("malformed token", func(t *testing.T) {
		_, err := TokenSubject("bad.token")
		assert.Error(t, err)
	})
}

// ─── BuildAuthURL ─────────────────────────────────────────────────────────────

func TestBuildAuthURL(t *testing.T) {
	url := BuildAuthURL(
		"https://auth.example.com/oauth/authorize",
		"my-client-id",
		"https://app.example.com/callback",
		"random-state-xyz",
		[]string{"openid", "profile"},
	)

	assert.Contains(t, url, "https://auth.example.com/oauth/authorize")
	assert.Contains(t, url, "client_id=my-client-id")
	assert.Contains(t, url, "redirect_uri=")
	assert.Contains(t, url, "state=random-state-xyz")
	assert.Contains(t, url, "scope=")
	assert.Contains(t, url, "response_type=code")
}
