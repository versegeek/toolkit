package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalString_Map(t *testing.T) {
	s := MarshalString(map[string]int{"a": 1})
	assert.Equal(t, `{"a":1}`, s)
}

func TestMarshalString_Nil(t *testing.T) {
	assert.Equal(t, "null", MarshalString(nil))
}

func TestMarshalBytes_RoundTrip(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	b := MarshalBytes(payload{Name: "test"})
	assert.Equal(t, `{"name":"test"}`, string(b))
}

func TestMarshalString_Panics_OnUnsupported(t *testing.T) {
	assert.Panics(t, func() {
		MarshalString(make(chan int)) // channels are not JSON-serializable
	})
}

func TestUnmarshalString_Success(t *testing.T) {
	var m map[string]int
	require.NoError(t, UnmarshalString(`{"x":42}`, &m))
	assert.Equal(t, 42, m["x"])
}

func TestUnmarshalString_InvalidJSON(t *testing.T) {
	var m map[string]int
	assert.Error(t, UnmarshalString(`not json`, &m))
}
