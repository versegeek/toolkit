package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMust_Success(t *testing.T) {
	result := Must(42, nil)
	assert.Equal(t, 42, result)
}

func TestMust_Panics(t *testing.T) {
	err := errors.New("boom")
	assert.PanicsWithError(t, "boom", func() {
		Must(0, err)
	})
}

func TestMust0_Success(t *testing.T) {
	require.NotPanics(t, func() {
		Must0(nil)
	})
}

func TestMust0_Panics(t *testing.T) {
	err := errors.New("fatal")
	assert.PanicsWithError(t, "fatal", func() {
		Must0(err)
	})
}
