package utils

import "encoding/json"

// MarshalString serializes v to a JSON string. It panics if v cannot be
// marshaled (e.g., contains a channel or function). For production code where
// marshal errors must be handled gracefully, use encoding/json directly.
func MarshalString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic("utils: MarshalString: " + err.Error())
	}
	return string(b)
}

// MarshalBytes serializes v to a JSON byte slice. Panics on marshal error.
func MarshalBytes(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic("utils: MarshalBytes: " + err.Error())
	}
	return b
}

// UnmarshalString deserializes the JSON string s into v. Returns an error if
// s is not valid JSON or does not match the shape of v.
func UnmarshalString(s string, v any) error {
	return json.Unmarshal([]byte(s), v)
}
