package rabbitmq

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── maskURL ─────────────────────────────────────────────────────────────────

func TestMaskURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "standard AMQP URL with password",
			input: "amqp://guest:secret@localhost:5672/",
			want:  "amqp://guest:***@localhost:5672/",
		},
		{
			name:  "no password",
			input: "amqp://guest@localhost:5672/",
			want:  "amqp://guest@localhost:5672/",
		},
		{
			name:  "empty password",
			input: "amqp://user:@localhost:5672/",
			want:  "amqp://user:***@localhost:5672/",
		},
		{
			name:  "URL without credentials",
			input: "amqp://localhost:5672/",
			want:  "amqp://localhost:5672/",
		},
		{
			name:  "AMQPS with password",
			input: "amqps://admin:topsecret@rabbit.example.com:5671/vhost",
			want:  "amqps://admin:***@rabbit.example.com:5671/vhost",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, maskURL(tt.input))
		})
	}
}

// ─── Config defaults ──────────────────────────────────────────────────────────

func TestConfigDefaults(t *testing.T) {
	cfg := NewConfigFromEnv()

	require.NotNil(t, cfg)
	assert.Equal(t, "amqp://guest:guest@127.0.0.1:5672/", cfg.URL)
	assert.Equal(t, 10*time.Second, cfg.Heartbeat)
	assert.Equal(t, uint16(0), cfg.ChannelMax)
}

func TestConfigOverride(t *testing.T) {
	t.Setenv("RABBITMQ_URL", "amqp://user:pass@rabbit:5672/myvhost")
	t.Setenv("RABBITMQ_HEARTBEAT", "30s")

	cfg := NewConfigFromEnv()
	assert.Equal(t, "amqp://user:pass@rabbit:5672/myvhost", cfg.URL)
	assert.Equal(t, 30*time.Second, cfg.Heartbeat)
}

// ─── New ─────────────────────────────────────────────────────────────────────

func TestNew_NilConfig(t *testing.T) {
	client := New(nil)
	require.NotNil(t, client)
	assert.Equal(t, "amqp://guest:guest@127.0.0.1:5672/", client.Config.URL)
}

func TestNew_WithConfig(t *testing.T) {
	cfg := &Config{URL: "amqp://custom:pass@host:5672/"}
	client := New(cfg)
	assert.Equal(t, cfg, client.Config)
}
