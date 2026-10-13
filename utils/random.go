package utils

import (
	"crypto/rand"
	"math/big"
)

const defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandString returns a cryptographically random alphanumeric string of length n.
func RandString(n int) string {
	return RandStringWith(n, defaultCharset)
}

// RandStringWith returns a cryptographically random string of length n drawn
// from charset. It panics if charset is empty or n < 0.
func RandStringWith(n int, charset string) string {
	if len(charset) == 0 {
		panic("utils: RandStringWith: charset must not be empty")
	}
	b := make([]byte, n)
	size := big.NewInt(int64(len(charset)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, size)
		if err != nil {
			panic("utils: RandStringWith: " + err.Error())
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}

// RandBytes returns n cryptographically random bytes.
func RandBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("utils: RandBytes: " + err.Error())
	}
	return b
}
