// Package redact hides credentials in connection URLs before they are logged.
package redact

import "strings"

// URL returns raw with the password in its user info replaced by "xxxxx",
// for URLs such as amqp://user:secret@host/ or mongodb://user:secret@a,b/db.
// Input without a scheme or a password is returned unchanged. The last "@"
// is taken as the end of the user info so passwords containing "@" are
// hidden too.
func URL(raw string) string {
	scheme, rest, ok := strings.Cut(raw, "://")
	if !ok {
		return raw
	}
	at := strings.LastIndex(rest, "@")
	if at < 0 {
		return raw
	}
	user, _, hasPassword := strings.Cut(rest[:at], ":")
	if !hasPassword {
		return raw
	}
	return scheme + "://" + user + ":xxxxx" + rest[at:]
}
