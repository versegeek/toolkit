package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandString_Length(t *testing.T) {
	for _, n := range []int{0, 1, 8, 32, 128} {
		assert.Len(t, RandString(n), n)
	}
}

func TestRandString_Unique(t *testing.T) {
	a, b := RandString(32), RandString(32)
	assert.NotEqual(t, a, b, "two 32-char random strings should differ")
}

func TestRandString_OnlyDefaultCharset(t *testing.T) {
	s := RandString(500)
	for _, c := range s {
		assert.True(t, strings.ContainsRune(defaultCharset, c), "unexpected char %q", c)
	}
}

func TestRandStringWith_CustomCharset(t *testing.T) {
	s := RandStringWith(100, "01")
	for _, c := range s {
		assert.True(t, c == '0' || c == '1', "unexpected char %q", c)
	}
}

func TestRandStringWith_EmptyCharset_Panics(t *testing.T) {
	assert.Panics(t, func() { RandStringWith(10, "") })
}

func TestRandBytes_Length(t *testing.T) {
	for _, n := range []int{0, 1, 16, 64} {
		require.Len(t, RandBytes(n), n)
	}
}

func TestRandBytes_Unique(t *testing.T) {
	a, b := RandBytes(32), RandBytes(32)
	assert.NotEqual(t, a, b)
}
