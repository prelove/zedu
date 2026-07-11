package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prelove/zedu/backend/internal/platform/logging"
)

func TestRedactionMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	middleware := logging.NewMiddleware(logger)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(`{"password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "super-secret-key")
	req.Header.Set("X-Correlation-ID", "corr-abc-123")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	logs := parseLogs(t, &buf)
	if len(logs) == 0 {
		t.Fatalf("expected access log records")
	}

	combined := combineLogs(logs)

	// Request and correlation IDs must be present.
	if !hasKey(combined, "request_id") {
		t.Errorf("log missing request_id")
	}
	if !hasKey(combined, "correlation_id") {
		t.Errorf("log missing correlation_id")
	}
	if combined["correlation_id"] != "corr-abc-123" {
		t.Errorf("expected correlation_id corr-abc-123, got %q", combined["correlation_id"])
	}

	// Sensitive values must not be present.
	for _, sensitive := range []string{"secret123", "super-secret-key"} {
		if strings.Contains(combinedString(logs), sensitive) {
			t.Errorf("log contains sensitive value %q", sensitive)
		}
	}

	// Request body must not be present in logs.
	if strings.Contains(combinedString(logs), `"password"`) || strings.Contains(combinedString(logs), `secret123`) {
		t.Errorf("log contains request body or password field")
	}
}

func TestRedactionLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewJSONLogger(&buf)

	logger.Info("user action",
		slog.String("email", "user@example.com"),
		slog.String("token", "bearer-token-value"),
		slog.String("name", "Alice"),
	)

	output := buf.String()
	if strings.Contains(output, "user@example.com") {
		t.Errorf("logger leaked email: %s", output)
	}
	if strings.Contains(output, "bearer-token-value") {
		t.Errorf("logger leaked token: %s", output)
	}
	if !strings.Contains(output, "Alice") {
		t.Errorf("logger should contain non-sensitive value Alice: %s", output)
	}
}

func TestRequestIDInContext(t *testing.T) {
	var captured string
	var found bool

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	middleware := logging.NewMiddleware(logger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, found = logging.RequestID(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !found {
		t.Fatalf("request_id not injected into context")
	}
	if captured == "" {
		t.Fatalf("request_id is empty")
	}
}

func parseLogs(t *testing.T, r io.Reader) []map[string]any {
	t.Helper()
	var logs []map[string]any
	dec := json.NewDecoder(r)
	for dec.More() {
		var rec map[string]any
		if err := dec.Decode(&rec); err != nil {
			t.Fatalf("decode log record: %v", err)
		}
		logs = append(logs, rec)
	}
	return logs
}

func combineLogs(logs []map[string]any) map[string]any {
	combined := make(map[string]any)
	for _, rec := range logs {
		for k, v := range rec {
			combined[k] = v
		}
	}
	return combined
}

func hasKey(m map[string]any, key string) bool {
	_, ok := m[key]
	return ok
}

func combinedString(logs []map[string]any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for _, rec := range logs {
		_ = enc.Encode(rec)
	}
	return buf.String()
}

// Prevent unused import of context in case RequestID signature changes.
var _ context.Context
