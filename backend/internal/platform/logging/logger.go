package logging

import (
	"context"
	"io"
	"log/slog"
	"strings"
)

// redactedValue is the replacement string for sensitive fields.
const redactedValue = "[REDACTED]"

// redactedKeys contains lowercase key names whose values must be masked.
var redactedKeys = map[string]struct{}{
	"password": {},
	"token":    {},
	"api_key":  {},
	"apikey":   {},
	"secret":   {},
	"email":    {},
}

// NewJSONLogger returns a structured JSON logger that redacts sensitive keys.
func NewJSONLogger(w io.Writer) *slog.Logger {
	return slog.New(NewRedactingHandler(slog.NewJSONHandler(w, nil)))
}

// NewRedactingHandler wraps a slog.Handler and masks sensitive attribute values.
func NewRedactingHandler(inner slog.Handler) slog.Handler {
	return &redactingHandler{inner: inner}
}

type redactingHandler struct {
	inner slog.Handler
}

func (h *redactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *redactingHandler) Handle(ctx context.Context, r slog.Record) error {
	newRecord := slog.Record{
		Time:    r.Time,
		Level:   r.Level,
		Message: r.Message,
		PC:      r.PC,
	}
	r.Attrs(func(a slog.Attr) bool {
		newRecord.AddAttrs(redactAttr(a))
		return true
	})
	return h.inner.Handle(ctx, newRecord)
}

func (h *redactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redacted := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		redacted[i] = redactAttr(a)
	}
	return &redactingHandler{inner: h.inner.WithAttrs(redacted)}
}

func (h *redactingHandler) WithGroup(name string) slog.Handler {
	return &redactingHandler{inner: h.inner.WithGroup(name)}
}

func redactAttr(a slog.Attr) slog.Attr {
	if _, ok := redactedKeys[strings.ToLower(a.Key)]; ok {
		return slog.String(a.Key, redactedValue)
	}
	if a.Value.Kind() == slog.KindGroup {
		groupAttrs := a.Value.Group()
		redacted := make([]slog.Attr, len(groupAttrs))
		for i, ga := range groupAttrs {
			redacted[i] = redactAttr(ga)
		}
		return slog.Attr{
			Key:   a.Key,
			Value: slog.GroupValue(redacted...),
		}
	}
	return a
}
