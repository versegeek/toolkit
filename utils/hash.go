package utils

import (
	"crypto/hmac"
	"crypto/md5" //nolint:gosec // MD5 is used for non-cryptographic checksums only
	"crypto/sha256"
	"encoding/hex"
)

// MD5 returns the hex-encoded MD5 checksum of s.
// Note: MD5 is not cryptographically secure. Use it only for checksums or
// cache keys, never for passwords or security-sensitive hashing.
func MD5(s string) string {
	return MD5Bytes([]byte(s))
}

// MD5Bytes returns the hex-encoded MD5 checksum of b.
func MD5Bytes(b []byte) string {
	h := md5.Sum(b) //nolint:gosec
	return hex.EncodeToString(h[:])
}

// SHA256 returns the hex-encoded SHA-256 digest of s.
func SHA256(s string) string {
	return SHA256Bytes([]byte(s))
}

// SHA256Bytes returns the hex-encoded SHA-256 digest of b.
func SHA256Bytes(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// HMACSHA256 returns the hex-encoded HMAC-SHA256 of message using key.
// Use this for signing tokens, webhook payloads, or API request signatures.
func HMACSHA256(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
