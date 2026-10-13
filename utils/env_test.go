package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetenv_Set(t *testing.T) {
	t.Setenv("TEST_GETENV_KEY", "value123")
	assert.Equal(t, "value123", Getenv("TEST_GETENV_KEY", "default"))
}

func TestGetenv_Unset_ReturnsFallback(t *testing.T) {
	assert.Equal(t, "default", Getenv("TEST_GETENV_UNSET_XYZ", "default"))
}

func TestGetenv_Empty_ReturnsFallback(t *testing.T) {
	t.Setenv("TEST_GETENV_EMPTY", "")
	assert.Equal(t, "fallback", Getenv("TEST_GETENV_EMPTY", "fallback"))
}
