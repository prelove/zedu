package logging

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type ctxKey int

const requestIDKey ctxKey = iota

const correlationIDHeader = "X-Correlation-ID"

// RequestID returns the request ID stored in the context.
func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

// NewMiddleware returns an HTTP middleware that injects request/correlation IDs
// and emits structured, redacted access logs.
func NewMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now().UTC()
			requestID := generateID()
			correlationID := r.Header.Get(correlationIDHeader)
			if correlationID == "" {
				correlationID = requestID
			}

			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r = r.WithContext(ctx)

			logger.Info("request started",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("request_id", requestID),
				slog.String("correlation_id", correlationID),
			)

			rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rec, r)

			logger.Info("request completed",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.statusCode),
				slog.Duration("duration", time.Since(start)),
				slog.String("request_id", requestID),
				slog.String("correlation_id", correlationID),
			)
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rec *responseRecorder) WriteHeader(code int) {
	if rec.written {
		return
	}
	rec.statusCode = code
	rec.written = true
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *responseRecorder) Write(p []byte) (int, error) {
	if !rec.written {
		rec.WriteHeader(http.StatusOK)
	}
	return rec.ResponseWriter.Write(p)
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to a timestamp-based ID if random source fails.
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
