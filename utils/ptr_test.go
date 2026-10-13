package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Ptr ---

func TestPtr_Int(t *testing.T) {
	v := 42
	p := Ptr(v)
	require.NotNil(t, p)
	assert.Equal(t, v, *p)
}

func TestPtr_String(t *testing.T) {
	p := Ptr("hello")
	require.NotNil(t, p)
	assert.Equal(t, "hello", *p)
}

func TestPtr_Duration(t *testing.T) {
	d := 5 * time.Second
	p := Ptr(d)
	require.NotNil(t, p)
	assert.Equal(t, d, *p)
}

func TestPtr_IndependentCopy(t *testing.T) {
	v := 1
	p := Ptr(v)
	v = 99
	assert.Equal(t, 1, *p, "Ptr should capture value at call time, not track mutations")
}

// --- Val ---

func TestVal_NonNil(t *testing.T) {
	v := 7
	assert.Equal(t, 7, Val(&v))
}

func TestVal_Nil_ReturnsZero(t *testing.T) {
	var p *int
	assert.Equal(t, 0, Val(p))
}

func TestVal_NilString_ReturnsEmpty(t *testing.T) {
	var p *string
	assert.Equal(t, "", Val(p))
}

// --- PtrOr ---

func TestPtrOr_NonNil(t *testing.T) {
	v := 42
	assert.Equal(t, 42, PtrOr(&v, 99))
}

func TestPtrOr_Nil_ReturnsFallback(t *testing.T) {
	var p *int
	assert.Equal(t, 99, PtrOr(p, 99))
}

func TestPtrOr_NilDuration_ReturnsFallback(t *testing.T) {
	var p *time.Duration
	assert.Equal(t, 30*time.Second, PtrOr(p, 30*time.Second))
}
