package utils

import "github.com/google/uuid"

// NewUUID returns a random (version 4) UUID in its canonical string form.
func NewUUID() string {
	return uuid.NewString()
}

// IsValidUUID reports whether u parses as a UUID.
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

// NewUUIDs returns n random UUIDs.
func NewUUIDs(n int) []string {
	uuids := make([]string, n)
	for i := range n {
		uuids[i] = NewUUID()
	}
	return uuids
}
