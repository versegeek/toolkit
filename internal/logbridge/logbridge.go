// Package logbridge swaps the process-wide slog default logger for the log
// providers and restores it when they close.
package logbridge

import (
	"log"
	"log/slog"
)

// SetDefault installs l as the slog default logger. The returned function
// restores the previous slog logger together with the output and flags of the
// standard log package, which slog.SetDefault redirects.
func SetDefault(l *slog.Logger) (restore func()) {
	prev := slog.Default()
	prevWriter, prevFlags := log.Writer(), log.Flags()
	slog.SetDefault(l)
	return func() {
		slog.SetDefault(prev)
		log.SetOutput(prevWriter)
		log.SetFlags(prevFlags)
	}
}
