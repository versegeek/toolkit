package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUUID(t *testing.T) {
	id := NewUUID()
	assert.Len(t, id, 36, "UUID should be 36 characters")
	assert.Equal(t, 4, strings.Count(id, "-"), "UUID should contain 4 hyphens")
}

func TestNewUUID_Unique(t *testing.T) {
	seen := make(map[string]struct{}, 100)
	for range 100 {
		id := NewUUID()
		_, dup := seen[id]
		require.False(t, dup, "NewUUID generated a duplicate: %s", id)
		seen[id] = struct{}{}
	}
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid v4", NewUUID(), true},
		{"valid explicit", "550e8400-e29b-41d4-a716-446655440000", true},
		{"empty string", "", false},
		{"random string", "not-a-uuid", false},
		{"missing hyphens", "550e8400e29b41d4a716446655440000", true}, // uuid.Parse accepts no-hyphen form
		{"too short", "550e8400-e29b-41d4", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidUUID(tt.input))
		})
	}
}

func TestNewUUIDs(t *testing.T) {
	const n = 50
	ids := NewUUIDs(n)
	require.Len(t, ids, n)

	seen := make(map[string]struct{}, n)
	for _, id := range ids {
		assert.True(t, IsValidUUID(id), "expected valid UUID, got %q", id)
		_, dup := seen[id]
		assert.False(t, dup, "duplicate UUID: %s", id)
		seen[id] = struct{}{}
	}
}

func TestNewUUIDs_Zero(t *testing.T) {
	assert.Empty(t, NewUUIDs(0))
}
