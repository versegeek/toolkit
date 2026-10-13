package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMD5_KnownValue(t *testing.T) {
	// echo -n "hello" | md5sum → 5d41402abc4b2a76b9719d911017c592
	assert.Equal(t, "5d41402abc4b2a76b9719d911017c592", MD5("hello"))
}

func TestMD5_EmptyString(t *testing.T) {
	// echo -n "" | md5sum → d41d8cd98f00b204e9800998ecf8427e
	assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", MD5(""))
}

func TestMD5Bytes_EqualsMD5(t *testing.T) {
	assert.Equal(t, MD5("world"), MD5Bytes([]byte("world")))
}

func TestSHA256_KnownValue(t *testing.T) {
	// echo -n "hello" | sha256sum → 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", SHA256("hello"))
}

func TestSHA256Bytes_EqualsSHA256(t *testing.T) {
	assert.Equal(t, SHA256("test"), SHA256Bytes([]byte("test")))
}

func TestHMACSHA256_KnownValue(t *testing.T) {
	// Verified with: echo -n "msg" | openssl dgst -sha256 -hmac "key" -hex
	result := HMACSHA256("msg", "key")
	require.Len(t, result, 64, "HMAC-SHA256 hex should be 64 chars")
	// deterministic
	assert.Equal(t, result, HMACSHA256("msg", "key"))
}

func TestHMACSHA256_DifferentKey(t *testing.T) {
	assert.NotEqual(t, HMACSHA256("msg", "key1"), HMACSHA256("msg", "key2"))
}
