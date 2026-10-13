package nsq_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	nsqprovider "github.com/versegeek/toolkit/mq/nsq"
)

func TestConfigDefaults(t *testing.T) {
	cfg := nsqprovider.NewConfigFromEnv()

	require.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1:4150", cfg.NsqdAddress)
	assert.Empty(t, cfg.LookupdAddresses)
}

func TestConfigOverride(t *testing.T) {
	t.Setenv("NSQ_NSQD_ADDRESS", "nsqd-host:4150")
	t.Setenv("NSQ_LOOKUPD_ADDRESSES", "lookupd1:4161,lookupd2:4161")

	cfg := nsqprovider.NewConfigFromEnv()
	assert.Equal(t, "nsqd-host:4150", cfg.NsqdAddress)
	assert.Equal(t, []string{"lookupd1:4161", "lookupd2:4161"}, cfg.LookupdAddresses)
}

func TestNew_NilConfig(t *testing.T) {
	client := nsqprovider.New(nil)
	require.NotNil(t, client)
	assert.Equal(t, "127.0.0.1:4150", client.Config.NsqdAddress)
}

func TestNew_WithConfig(t *testing.T) {
	cfg := &nsqprovider.Config{NsqdAddress: "custom:4150"}
	client := nsqprovider.New(cfg)
	assert.Equal(t, cfg, client.Config)
}
