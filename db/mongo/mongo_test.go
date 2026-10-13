package mongo_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mongoprovider "github.com/versegeek/toolkit/db/mongo"
)

func TestNewConfigFromEnv_Defaults(t *testing.T) {
	cfg := mongoprovider.NewConfigFromEnv()
	require.NotNil(t, cfg)
	assert.Equal(t, "mongodb://localhost:27017", cfg.URI)
	assert.NotZero(t, cfg.MaxPoolSize)
	assert.NotZero(t, cfg.Timeout)
}

func TestNew_NilConfig(t *testing.T) {
	c := mongoprovider.New(nil)
	require.NotNil(t, c)
	assert.NotNil(t, c.Config)
}

func TestClient_Close_WithoutInit(t *testing.T) {
	c := mongoprovider.New(nil)
	assert.NoError(t, c.Close())
}

func TestNewClient_WireCompatible(t *testing.T) {
	cfg := &mongoprovider.Config{
		URI:                    "mongodb://localhost:27017",
		Timeout:                mongoprovider.DefaultTimeout,
		MaxPoolSize:            1,
		MinPoolSize:            1,
		MaxConnIdleTime:        30 * time.Second,
		HeartbeatInterval:      10 * time.Second,
		ServerSelectionTimeout: mongoprovider.DefaultServerSelectionTimeout,
	}
	c, cleanup, err := mongoprovider.NewClient(cfg)
	if err != nil {
		t.Skipf("MongoDB not reachable, skipping: %v", err)
	}
	require.NotNil(t, c)
	defer cleanup()
}

func TestClient_Init_PingFailure_NoLeak(t *testing.T) {
	cfg := &mongoprovider.Config{
		URI:                    "mongodb://127.0.0.1:27099",
		Timeout:                2 * time.Second,
		MaxPoolSize:            1,
		MinPoolSize:            1,
		MaxConnIdleTime:        5 * time.Second,
		HeartbeatInterval:      1 * time.Second,
		ServerSelectionTimeout: 1 * time.Second,
	}
	c := mongoprovider.New(cfg)
	err := c.Init()
	assert.Error(t, err)
	assert.NoError(t, c.Close())
}

func TestMaskURI(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"mongodb://localhost:27017", "mongodb://localhost:27017"},
		{"mongodb://user:secret@host:27017/db", "mongodb://user:***@host:27017/db"},
		{"mongodb://user:p%40ss@host:27017/db", "mongodb://user:***@host:27017/db"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, mongoprovider.MaskURI(tc.in), tc.in)
	}
}
