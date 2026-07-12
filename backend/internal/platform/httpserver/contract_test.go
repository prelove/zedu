package httpserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
)

func TestWriteSuccessEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	httpserver.WriteSuccess(rr, http.StatusOK, map[string]any{"role": "OWNER"})

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	code, ok := resp["code"]
	if !ok || code != float64(0) {
		t.Fatalf("expected code 0, got %v", resp["code"])
	}

	data, ok := resp["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %v", resp["data"])
	}
	if data["role"] != "OWNER" {
		t.Fatalf("expected role OWNER in data, got %v", data["role"])
	}
}

func TestWriteErrorEnvelope(t *testing.T) {
	rr := httptest.NewRecorder()
	httpserver.WriteError(rr, http.StatusUnauthorized, 40101, "AUTH_REQUIRED", "req-123")

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if resp["code"] != float64(40101) {
		t.Fatalf("expected code 40101, got %v", resp["code"])
	}
	if resp["message"] != "AUTH_REQUIRED" {
		t.Fatalf("expected message AUTH_REQUIRED, got %v", resp["message"])
	}
	if resp["requestId"] != "req-123" {
		t.Fatalf("expected requestId req-123, got %v", resp["requestId"])
	}
}

func TestWriteErrorEnvelopeWithRequestIDFromContext(t *testing.T) {
	rr := httptest.NewRecorder()

	// Without a request ID in context, the error envelope should still have requestId field.
	httpserver.WriteError(rr, http.StatusForbidden, httpserver.CodeForbidden, "FORBIDDEN", "")

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := resp["requestId"]; !ok {
		t.Fatalf("error envelope must always have requestId field")
	}
}

func TestErrorCodesMapping(t *testing.T) {
	tests := []struct {
		code     httpserver.ErrorCode
		httpStat int
	}{
		{httpserver.CodeUnauth, http.StatusUnauthorized},
		{httpserver.CodeLoginFailed, http.StatusUnauthorized},
		{httpserver.CodeLocked, http.StatusUnauthorized},
		{httpserver.CodeForbidden, http.StatusForbidden},
		{httpserver.CodeNotFound, http.StatusNotFound},
		{httpserver.CodeConflict, http.StatusConflict},
		{httpserver.CodeInvalidState, http.StatusUnprocessableEntity},
	}

	for _, tc := range tests {
		rr := httptest.NewRecorder()
		httpserver.WriteError(rr, tc.httpStat, tc.code, "KEY", "req")
		if rr.Code != tc.httpStat {
			t.Fatalf("code %d: expected HTTP %d, got %d", tc.code, tc.httpStat, rr.Code)
		}
	}
}

func TestUnauthenticatedRequestReturns40101(t *testing.T) {
	// Open a temp DB and migrate so the middleware can query user_account.
	tmpDir := t.TempDir()
	dsn := "file:" + filepath.Join(tmpDir, "test.db")
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	// Create a mux with an auth-protected route to test the middleware.
	mux := http.NewServeMux()
	authMW := httpserver.AuthMiddleware("test-secret-at-least-32-chars-long", db)
	mux.Handle("GET /auth/me", authMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpserver.WriteSuccess(w, http.StatusOK, map[string]any{"ok": true})
	})))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// /auth/me requires authentication — without token, must return 40101.
	resp, err := http.Get(srv.URL + "/auth/me")
	if err != nil {
		t.Fatalf("GET /auth/me: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if body["code"] != float64(40101) {
		t.Fatalf("expected code 40101, got %v", body["code"])
	}
	if _, ok := body["requestId"]; !ok {
		t.Fatalf("error response must include requestId")
	}
}

func TestSuccessEnvelopeContentType(t *testing.T) {
	rr := httptest.NewRecorder()
	httpserver.WriteSuccess(rr, http.StatusOK, map[string]any{"ok": true})

	ct := rr.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("expected application/json Content-Type, got %q", ct)
	}
}
