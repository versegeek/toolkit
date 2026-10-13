package logrus

import (
	"context"
	"log/slog"
	"maps"

	"github.com/sirupsen/logrus"
)

// slogHandler is a slog.Handler that writes records to a logrus logger.
// Attributes become logrus fields; groups prefix their keys with "group.".
type slogHandler struct {
	logger *logrus.Logger
	fields logrus.Fields
	prefix string
}

func newSlogHandler(logger *logrus.Logger) *slogHandler {
	return &slogHandler{logger: logger, fields: logrus.Fields{}}
}

func (h *slogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.logger.IsLevelEnabled(logrusLevel(level))
}

func (h *slogHandler) Handle(_ context.Context, r slog.Record) error {
	fields := make(logrus.Fields, len(h.fields)+r.NumAttrs())
	maps.Copy(fields, h.fields)
	r.Attrs(func(a slog.Attr) bool {
		addAttr(fields, h.prefix, a)
		return true
	})

	entry := h.logger.WithFields(fields)
	if !r.Time.IsZero() {
		entry = entry.WithTime(r.Time)
	}
	entry.Log(logrusLevel(r.Level), r.Message)
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := *h
	h2.fields = maps.Clone(h.fields)
	for _, a := range attrs {
		addAttr(h2.fields, h.prefix, a)
	}
	return &h2
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := *h
	h2.prefix = h.prefix + name + "."
	return &h2
}

// addAttr stores a under prefix+key, flattening groups and skipping empty
// attributes as the slog.Handler contract requires.
func addAttr(fields logrus.Fields, prefix string, a slog.Attr) {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return
	}
	if a.Value.Kind() == slog.KindGroup {
		groupPrefix := prefix
		if a.Key != "" {
			groupPrefix += a.Key + "."
		}
		for _, ga := range a.Value.Group() {
			addAttr(fields, groupPrefix, ga)
		}
		return
	}
	fields[prefix+a.Key] = a.Value.Any()
}

func logrusLevel(level slog.Level) logrus.Level {
	switch {
	case level >= slog.LevelError:
		return logrus.ErrorLevel
	case level >= slog.LevelWarn:
		return logrus.WarnLevel
	case level >= slog.LevelInfo:
		return logrus.InfoLevel
	case level >= slog.LevelDebug:
		return logrus.DebugLevel
	default:
		return logrus.TraceLevel
	}
}
