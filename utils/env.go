package utils

import "os"

// Getenv returns the value of the environment variable named by key. If the
// variable is unset or empty, fallback is returned instead.
//
// Example:
//
//	addr := utils.Getenv("LISTEN_ADDR", ":8080")
func Getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
