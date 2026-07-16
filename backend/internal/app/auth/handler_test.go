package auth_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	appauth "github.com/prelove/zedu/backend/internal/app/auth"
	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/platform/logging"
)

// testServer sets up a full HTTP server with auth routes backed by a migrated SQLite DB.
type testServer struct {
	db      *sql.DB
	srv     *httptest.Server
	jwtSec  string
	handler http.Handler
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	tmpDir := t.TempDir()
	dsn := "file:" + filepath.Join(tmpDir, "test.db")

	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	jwtSec := "test-jwt-secret-for-m2"
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	handler := appauth.NewHandler(db, jwtSec, logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	wrapped := logging.NewMiddleware(logger)(mux)

	srv := httptest.NewServer(wrapped)

	ts := &testServer{db: db, srv: srv, jwtSec: jwtSec, handler: wrapped}
	t.Cleanup(func() {
		srv.Close()
		db.Close()
	})
	return ts
}

// createTestUser inserts a user directly into the DB and returns the user ID.
func createTestUser(t *testing.T, db *sql.DB, username, password, role, status string) int64 {
	t.Helper()
	hash, err := hashForTest(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	res, err := db.Exec(`INSERT INTO user_account (username, password_hash, role, display_name, status) VALUES (?, ?, ?, ?, ?)`,
		username, hash, role, username, status)
	if err != nil {
		t.Fatalf("insert user %s: %v", username, err)
	}
	id, _ := res.LastInsertId()
	return id
}

// hashForTest uses the auth package's bcrypt hashing.
func hashForTest(password string) (string, error) {
	return auth.HashPassword(password)
}

// doRequest performs an HTTP request and returns status code and parsed body.
func doRequest(t *testing.T, method, url string, body any, cookie *http.Cookie) (int, map[string]any, *http.Response) {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var parsed map[string]any
	_ = json.Unmarshal(data, &parsed)
	return resp.StatusCode, parsed, resp
}

// doRequestWithToken performs an HTTP request with a Bearer token.
func doRequestWithToken(t *testing.T, method, url, token string, body any) (int, map[string]any, *http.Response) {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var parsed map[string]any
	_ = json.Unmarshal(data, &parsed)
	return resp.StatusCode, parsed, resp
}

// extractRefreshCookie finds the refresh token cookie from a response.
func extractRefreshCookie(resp *http.Response) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == appauth.RefreshCookieName {
			return c
		}
	}
	return nil
}

// ==================== Login Tests ====================

func TestLoginSuccess(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	code, body, resp := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)

	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", code, body)
	}
	if body["code"] != float64(0) {
		t.Fatalf("expected code 0, got %v", body["code"])
	}

	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %v", body["data"])
	}
	if data["accessToken"] == nil || data["accessToken"] == "" {
		t.Fatalf("expected non-empty accessToken")
	}
	if data["role"] != "OWNER" {
		t.Fatalf("expected role OWNER, got %v", data["role"])
	}

	// Refresh cookie must be set with correct attributes.
	cookie := extractRefreshCookie(resp)
	if cookie == nil {
		t.Fatalf("expected refresh cookie to be set")
	}
	if !cookie.HttpOnly {
		t.Fatalf("refresh cookie must be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("refresh cookie must be SameSite=Strict, got %v", cookie.SameSite)
	}
}

func TestLoginUnknownUserAndWrongPasswordSameResponse(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// Unknown username.
	code1, body1, _ := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "nonexistent", "password": "Pass1234"}, nil)

	// Wrong password.
	code2, body2, _ := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "wrongpass"}, nil)

	if code1 != code2 {
		t.Fatalf("unknown user and wrong password must have same HTTP status: %d vs %d", code1, code2)
	}
	if body1["code"] != body2["code"] {
		t.Fatalf("unknown user and wrong password must have same error code: %v vs %v", body1["code"], body2["code"])
	}
	if body1["code"] != float64(40102) {
		t.Fatalf("expected code 40102, got %v", body1["code"])
	}
	if body1["message"] != body2["message"] {
		t.Fatalf("unknown user and wrong password must have same message: %v vs %v", body1["message"], body2["message"])
	}
}

func TestLoginLockoutAfterFiveFailures(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// First 4 failed attempts return 40102 (wrong password, not yet locked).
	for i := 0; i < 4; i++ {
		code, body, _ := doRequest(t, "POST", ts.srv.URL+"/auth/login",
			map[string]string{"username": "owner1", "password": "wrong"}, nil)
		if code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", i+1, code)
		}
		if body["code"] != float64(40102) {
			t.Fatalf("attempt %d: expected code 40102, got %v", i+1, body["code"])
		}
	}

	// 5th failed attempt triggers the lock and returns 40103 (account locked).
	code, body, _ := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "wrong"}, nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("attempt 5: expected 401, got %d", code)
	}
	if body["code"] != float64(40103) {
		t.Fatalf("attempt 5: expected code 40103 (locked), got %v", body["code"])
	}

	// Sixth attempt (even with correct password) must be locked.
	code, body, _ = doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for locked account, got %d", code)
	}
	if body["code"] != float64(40103) {
		t.Fatalf("expected code 40103 for lockout, got %v", body["code"])
	}
}

func TestLoginDisabledAccountRejected(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "disabled1", "Pass1234", "OPERATOR", "DISABLED")

	code, body, _ := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "disabled1", "password": "Pass1234"}, nil)

	if code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for disabled account, got %d", code)
	}
	if body["code"] != float64(40102) {
		t.Fatalf("expected code 40102 for disabled account login, got %v", body["code"])
	}
}

// ==================== Refresh Tests ====================

func TestRefreshTokenRotation(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// Login to get refresh cookie.
	_, _, resp := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)

	cookie := extractRefreshCookie(resp)
	if cookie == nil {
		t.Fatalf("no refresh cookie after login")
	}

	// Refresh should return a new access token and a new refresh cookie.
	code, body, resp2 := doRequest(t, "POST", ts.srv.URL+"/auth/refresh", nil, cookie)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", code, body)
	}
	if body["code"] != float64(0) {
		t.Fatalf("expected code 0, got %v", body["code"])
	}
	data := body["data"].(map[string]any)
	if data["accessToken"] == nil || data["accessToken"] == "" {
		t.Fatalf("expected non-empty accessToken after refresh")
	}

	newCookie := extractRefreshCookie(resp2)
	if newCookie == nil {
		t.Fatalf("expected new refresh cookie after rotation")
	}
	if newCookie.Value == cookie.Value {
		t.Fatalf("refresh token must rotate: new cookie same as old")
	}
}

func TestRefreshOldTokenReplayFails(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	_, _, resp := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)

	oldCookie := extractRefreshCookie(resp)
	if oldCookie == nil {
		t.Fatalf("no refresh cookie after login")
	}

	// First refresh succeeds and rotates.
	_, _, resp2 := doRequest(t, "POST", ts.srv.URL+"/auth/refresh", nil, oldCookie)
	newCookie := extractRefreshCookie(resp2)
	if newCookie == nil {
		t.Fatalf("no new cookie after refresh")
	}

	// Replay of the old cookie must fail.
	code, body, _ := doRequest(t, "POST", ts.srv.URL+"/auth/refresh", nil, oldCookie)
	if code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for replayed old refresh token, got %d", code)
	}
	if body["code"] != float64(40101) {
		t.Fatalf("expected code 40101 for replayed token, got %v", body["code"])
	}
}

func TestLogoutRevokesRefresh(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	_, loginBody, resp := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)

	cookie := extractRefreshCookie(resp)
	accessToken := loginBody["data"].(map[string]any)["accessToken"].(string)

	// Logout with access token AND refresh cookie (both are needed).
	req, _ := http.NewRequest("POST", ts.srv.URL+"/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.AddCookie(cookie)
	logoutResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("logout request: %v", err)
	}
	logoutResp.Body.Close()
	if logoutResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for logout, got %d", logoutResp.StatusCode)
	}

	// Refresh after logout must fail.
	code2, body2, _ := doRequest(t, "POST", ts.srv.URL+"/auth/refresh", nil, cookie)
	if code2 != http.StatusUnauthorized {
		t.Fatalf("expected 401 for refresh after logout, got %d", code2)
	}
	if body2["code"] != float64(40101) {
		t.Fatalf("expected code 40101 after logout, got %v", body2["code"])
	}
}

func TestDisabledAccountRefreshFails(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	createTestUser(t, ts.db, "op1", "Pass1234", "OPERATOR", "ACTIVE")

	// Login as operator.
	_, _, resp := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "op1", "password": "Pass1234"}, nil)
	cookie := extractRefreshCookie(resp)
	if cookie == nil {
		t.Fatalf("no refresh cookie")
	}

	// Owner disables operator.
	ownerToken := loginAndGetToken(t, ts, "owner1", "Pass1234")
	_, _, _ = doRequestWithToken(t, "POST", ts.srv.URL+"/users/2/disable", ownerToken, nil)

	// Refresh with disabled account's cookie must fail.
	code, body, _ := doRequest(t, "POST", ts.srv.URL+"/auth/refresh", nil, cookie)
	if code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for disabled account refresh, got %d", code)
	}
	if body["code"] != float64(40101) {
		t.Fatalf("expected code 40101, got %v", body["code"])
	}
}

// ==================== RBAC Tests ====================

func TestUnauthenticatedBusinessRequestReturns40101(t *testing.T) {
	ts := newTestServer(t)

	code, body, _ := doRequestWithToken(t, "GET", ts.srv.URL+"/auth/me", "", nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}
	if body["code"] != float64(40101) {
		t.Fatalf("expected code 40101, got %v", body["code"])
	}
}

func TestOperatorDeniedOwnerOnlyAPI(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	createTestUser(t, ts.db, "op1", "Pass1234", "OPERATOR", "ACTIVE")

	opToken := loginAndGetToken(t, ts, "op1", "Pass1234")

	// Operator tries to list users: Owner-only.
	code, body, _ := doRequestWithToken(t, "GET", ts.srv.URL+"/users", opToken, nil)
	if code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", code)
	}
	if body["code"] != float64(40301) {
		t.Fatalf("expected code 40301, got %v", body["code"])
	}
}

func TestOwnerCanListUsers(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	createTestUser(t, ts.db, "op1", "Pass1234", "OPERATOR", "ACTIVE")

	ownerToken := loginAndGetToken(t, ts, "owner1", "Pass1234")

	code, body, _ := doRequestWithToken(t, "GET", ts.srv.URL+"/users", ownerToken, nil)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", code, body)
	}
	if body["code"] != float64(0) {
		t.Fatalf("expected code 0, got %v", body["code"])
	}
}

func TestAuthMeReturnsCurrentUser(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	token := loginAndGetToken(t, ts, "owner1", "Pass1234")

	code, body, _ := doRequestWithToken(t, "GET", ts.srv.URL+"/auth/me", token, nil)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	data := body["data"].(map[string]any)
	if data["username"] != "owner1" {
		t.Fatalf("expected username owner1, got %v", data["username"])
	}
	if data["role"] != "OWNER" {
		t.Fatalf("expected role OWNER, got %v", data["role"])
	}
}

// ==================== Logging Redaction Tests ====================

func TestLoginLogDoesNotLeakPasswordOrToken(t *testing.T) {
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

	var logBuf bytes.Buffer
	logger := slog.New(logging.NewRedactingHandler(slog.NewJSONHandler(&logBuf, nil)))

	handler := appauth.NewHandler(db, "test-secret", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	wrapped := logging.NewMiddleware(logger)(mux)
	srv := httptest.NewServer(wrapped)
	defer srv.Close()

	createTestUser(t, db, "owner1", "SecretPass123", "OWNER", "ACTIVE")

	_, _, _ = doRequest(t, "POST", srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "SecretPass123"}, nil)

	logOutput := logBuf.String()
	if strings.Contains(logOutput, "SecretPass123") {
		t.Fatalf("log leaked password: %s", logOutput)
	}

	// Login again to get a token and check it doesn't appear in logs.
	_, body, _ := doRequest(t, "POST", srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "SecretPass123"}, nil)
	accessToken := body["data"].(map[string]any)["accessToken"].(string)

	logOutput2 := logBuf.String()
	if strings.Contains(logOutput2, accessToken) {
		t.Fatalf("log leaked access token: %s", logOutput2)
	}
}

// ==================== Helpers ====================

func loginAndGetToken(t *testing.T, ts *testServer, username, password string) string {
	t.Helper()
	_, body, _ := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": username, "password": password}, nil)
	data := body["data"].(map[string]any)
	return data["accessToken"].(string)
}

// Suppress unused import warnings for context and sync.
var (
	_ context.Context
	_ sync.WaitGroup
	_ time.Duration
	_ = fmt.Sprintf
)

// ==================== P1.2: Disabled account access token immediately rejected ====================

func TestDisabledAccountAccessTokenImmediatelyRejected(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	createTestUser(t, ts.db, "op1", "Pass1234", "OPERATOR", "ACTIVE")

	// Operator logs in and gets access token.
	opToken := loginAndGetToken(t, ts, "op1", "Pass1234")

	// Verify token works before disable.
	code, _, _ := doRequestWithToken(t, "GET", ts.srv.URL+"/auth/me", opToken, nil)
	if code != http.StatusOK {
		t.Fatalf("expected 200 before disable, got %d", code)
	}

	// Owner disables operator.
	ownerToken := loginAndGetToken(t, ts, "owner1", "Pass1234")
	code, _, _ = doRequestWithToken(t, "POST", ts.srv.URL+"/users/2/disable", ownerToken, nil)
	if code != http.StatusOK {
		t.Fatalf("expected 200 for disable, got %d", code)
	}

	// Operator's existing access token must now return 40101 immediately.
	code, body, _ := doRequestWithToken(t, "GET", ts.srv.URL+"/auth/me", opToken, nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after disable, got %d", code)
	}
	if body["code"] != float64(40101) {
		t.Fatalf("expected code 40101 after disable, got %v", body["code"])
	}
}

// ==================== P1.3: Concurrent refresh only one succeeds ====================

func TestConcurrentRefreshOnlyOneSucceeds(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// Login to get refresh cookie.
	_, _, resp := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)
	cookie := extractRefreshCookie(resp)
	if cookie == nil {
		t.Fatalf("no refresh cookie after login")
	}

	// Fire multiple concurrent refresh requests with the same cookie.
	const goroutines = 5
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	failures := 0

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("POST", ts.srv.URL+"/auth/refresh", nil)
			req.AddCookie(cookie)
			r, err := http.DefaultClient.Do(req)
			if err != nil {
				mu.Lock()
				failures++
				mu.Unlock()
				return
			}
			r.Body.Close()
			mu.Lock()
			if r.StatusCode == http.StatusOK {
				successes++
			} else {
				failures++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()

	if successes != 1 {
		t.Fatalf("expected exactly 1 concurrent refresh success, got %d (failures=%d)", successes, failures)
	}
	if failures != goroutines-1 {
		t.Fatalf("expected %d concurrent refresh failures, got %d", goroutines-1, failures)
	}
}

// ==================== P1.4: Password rules + Operator-only creation ====================

func TestCreateUserRejectsWeakPassword(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	ownerToken := loginAndGetToken(t, ts, "owner1", "Pass1234")

	// Too short.
	code, body, _ := doRequestWithToken(t, "POST", ts.srv.URL+"/users", ownerToken,
		map[string]string{"username": "op2", "password": "Ab1", "role": "OPERATOR"})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for short password, got %d", code)
	}
	if body["code"] != float64(42201) {
		t.Fatalf("expected code 42201, got %v", body["code"])
	}

	// No digit.
	code, body, _ = doRequestWithToken(t, "POST", ts.srv.URL+"/users", ownerToken,
		map[string]string{"username": "op2", "password": "abcdefgh", "role": "OPERATOR"})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for password with no digit, got %d", code)
	}

	// No letter.
	code, body, _ = doRequestWithToken(t, "POST", ts.srv.URL+"/users", ownerToken,
		map[string]string{"username": "op2", "password": "12345678", "role": "OPERATOR"})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for password with no letter, got %d", code)
	}
}

func TestCreateUserRejectsOwnerRole(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	ownerToken := loginAndGetToken(t, ts, "owner1", "Pass1234")

	code, body, _ := doRequestWithToken(t, "POST", ts.srv.URL+"/users", ownerToken,
		map[string]string{"username": "owner2", "password": "Pass1234", "role": "OWNER"})
	if code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for creating OWNER via API, got %d", code)
	}
	if body["code"] != float64(42201) {
		t.Fatalf("expected code 42201, got %v", body["code"])
	}
}

func TestCreateUserOperatorSuccess(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	ownerToken := loginAndGetToken(t, ts, "owner1", "Pass1234")

	code, body, _ := doRequestWithToken(t, "POST", ts.srv.URL+"/users", ownerToken,
		map[string]string{"username": "op2", "password": "Pass1234", "role": "OPERATOR"})
	if code != http.StatusCreated {
		t.Fatalf("expected 201 for valid operator creation, got %d body=%v", code, body)
	}
	if body["code"] != float64(0) {
		t.Fatalf("expected code 0, got %v", body["code"])
	}
	data := body["data"].(map[string]any)
	if data["role"] != "OPERATOR" {
		t.Fatalf("expected role OPERATOR, got %v", data["role"])
	}
}

func TestCreateOperatorAuditDetailEscapesUsername(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	token := loginAndGetToken(t, ts, "owner1", "Pass1234")
	username := "operator\"\\line\nnext"

	code, body, _ := doRequestWithToken(t, "POST", ts.srv.URL+"/users", token,
		map[string]string{"username": username, "password": "Pass1234", "role": "OPERATOR"})
	if code != http.StatusCreated || body == nil || body["code"] != float64(0) {
		t.Fatalf("expected successful operator creation, got code=%d body=%v", code, body)
	}

	logs := queryAuditLogs(t, ts.db, "CREATE_OPERATOR")
	if len(logs) != 1 {
		t.Fatalf("expected 1 CREATE_OPERATOR audit row, got %d", len(logs))
	}

	var detail map[string]string
	if err := json.Unmarshal([]byte(logs[0].DetailJSON), &detail); err != nil {
		t.Fatalf("audit detail_json must be valid JSON: %v; detail=%q", err, logs[0].DetailJSON)
	}
	if len(detail) != 2 || detail["action"] != "create_operator" || detail["username"] != username {
		t.Fatalf("unexpected audit detail: %#v", detail)
	}
	if strings.Contains(logs[0].DetailJSON, "Pass1234") {
		t.Fatalf("audit detail must not contain submitted password: %q", logs[0].DetailJSON)
	}
}

// ==================== P2.2: Log includes request ID, unknown user no username ====================

func TestLoginLogIncludesRequestIDAndNoUsernameForUnknownUser(t *testing.T) {
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

	var logBuf bytes.Buffer
	logger := slog.New(logging.NewRedactingHandler(slog.NewJSONHandler(&logBuf, nil)))

	handler := appauth.NewHandler(db, "test-secret-at-least-32-chars", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	wrapped := logging.NewMiddleware(logger)(mux)
	srv := httptest.NewServer(wrapped)
	defer srv.Close()

	// Login with unknown username.
	_, _, _ = doRequest(t, "POST", srv.URL+"/auth/login",
		map[string]string{"username": "nonexistentuser123", "password": "Pass1234"}, nil)

	logOutput := logBuf.String()

	// Log must not contain the username (sensitive identifier for unknown user).
	if strings.Contains(logOutput, "nonexistentuser123") {
		t.Fatalf("log leaked unknown username: %s", logOutput)
	}

	// Log must contain request_id.
	if !strings.Contains(logOutput, "request_id") {
		t.Fatalf("log must include request_id: %s", logOutput)
	}
}

// ==================== P0.1: No default JWT secret in production ====================

func TestValidateSecretFailsOnEmpty(t *testing.T) {
	if err := auth.ValidateSecret(""); err == nil {
		t.Fatalf("ValidateSecret must fail on empty secret")
	}
	if err := auth.ValidateSecret("short"); err == nil {
		t.Fatalf("ValidateSecret must fail on short secret")
	}
}

// ==================== P1.7: Concurrent login lockout regression ====================

// newConcurrentTestServer creates a test server with MaxOpenConns > 1 so that
// concurrent HTTP requests can execute DB operations in parallel.
// PRAGMAs are set via DSN query parameters so every pooled connection inherits them.
func newConcurrentTestServer(t *testing.T) *testServer {
	t.Helper()
	tmpDir := t.TempDir()
	dsn := "file:" + filepath.Join(tmpDir, "test.db") +
		"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(10)

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	jwtSec := "test-jwt-secret-for-m2-concurrent"
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	handler := appauth.NewHandler(db, jwtSec, logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	wrapped := logging.NewMiddleware(logger)(mux)

	srv := httptest.NewServer(wrapped)

	ts := &testServer{db: db, srv: srv, jwtSec: jwtSec, handler: wrapped}
	t.Cleanup(func() {
		srv.Close()
		db.Close()
	})
	return ts
}

func TestConcurrentLoginLockout(t *testing.T) {
	ts := newConcurrentTestServer(t)
	createTestUser(t, ts.db, "victim", "Pass1234", "OWNER", "ACTIVE")

	// Fire 5 concurrent wrong-password login requests.
	const goroutines = 5
	var wg sync.WaitGroup
	type result struct {
		statusCode int
		code       float64
	}
	results := make([]result, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req, _ := http.NewRequest("POST", ts.srv.URL+"/auth/login",
				bytes.NewReader([]byte(`{"username":"victim","password":"wrongpass1"}`)))
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				results[idx] = result{statusCode: 0, code: 0}
				return
			}
			defer resp.Body.Close()
			data, _ := io.ReadAll(resp.Body)
			var parsed map[string]any
			_ = json.Unmarshal(data, &parsed)
			code, _ := parsed["code"].(float64)
			results[idx] = result{statusCode: resp.StatusCode, code: code}
		}(i)
	}
	wg.Wait()

	// All 5 should return 401 (either 40102 or 40103 for the one that triggers lockout).
	for i, r := range results {
		if r.statusCode != http.StatusUnauthorized {
			t.Fatalf("request %d: expected 401, got %d", i, r.statusCode)
		}
		if r.code != float64(40102) && r.code != float64(40103) {
			t.Fatalf("request %d: expected code 40102 or 40103, got %v", i, r.code)
		}
	}

	// Verify the account is now locked: login_fail_count must be exactly 5
	// and locked_until must be set.
	var failCount int
	var lockedUntil *time.Time
	err := ts.db.QueryRow(
		`SELECT login_fail_count, locked_until FROM user_account WHERE username = 'victim'`,
	).Scan(&failCount, &lockedUntil)
	if err != nil {
		t.Fatalf("query user: %v", err)
	}
	if failCount != goroutines {
		t.Fatalf("expected login_fail_count=%d, got %d", goroutines, failCount)
	}
	if lockedUntil == nil {
		t.Fatalf("expected locked_until to be set after %d concurrent failures", goroutines)
	}

	// A subsequent login with the CORRECT password must still be rejected
	// with 40103 (account locked).
	code, body, _ := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "victim", "password": "Pass1234"}, nil)
	if code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for locked account, got %d", code)
	}
	if body["code"] != float64(40103) {
		t.Fatalf("expected code 40103 for locked account, got %v", body["code"])
	}
}

// ==================== P1.7b: Login fail-count update failure does not return 40102 ====================

// TestLoginFailCountUpdateErrorDoesNotReturn40102 specifically tests that when
// the atomic UPDATE ... RETURNING fails (DB error), the response is NOT 40102
// (which would falsely indicate "wrong password" while the counter was never
// incremented). It uses a custom *sql.DB wrapper that intercepts QueryRowContext
// to fail only on UPDATE statements.
func TestLoginFailCountUpdateErrorDoesNotReturn40102(t *testing.T) {
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

	createTestUser(t, db, "victim", "Pass1234", "OWNER", "ACTIVE")

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	// Wrap the DB so that QueryRowContext fails when the query contains "UPDATE".
	// The closedDB provides a *sql.Row whose Scan always errors.
	closedDB, _ := sql.Open("sqlite", "file:"+filepath.Join(tmpDir, "closed.db"))
	closedDB.Close()
	failingDB := &failingUpdateDB{db: db, closedDB: closedDB}

	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-m2-failcount", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	wrapped := logging.NewMiddleware(logger)(mux)
	srv := httptest.NewServer(wrapped)
	defer srv.Close()

	// Login with wrong password. The SELECT succeeds (not an UPDATE),
	// password verification fails, then the atomic UPDATE ... RETURNING
	// fails because failingDB intercepts it.
	code, body, _ := doRequest(t, "POST", srv.URL+"/auth/login",
		map[string]string{"username": "victim", "password": "wrongpass1"}, nil)

	// Must NOT be 40102 (LOGIN_FAILED): that would mask the DB error.
	if code == http.StatusUnauthorized && body != nil && body["code"] == float64(40102) {
		t.Fatalf("must not return 40102 when fail-count UPDATE fails; got code=40102 body=%v", body)
	}

	// Must be 500 with code 50002 (database error).
	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when fail-count UPDATE fails, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got body=%v", body)
	}

	// Must not issue a token.
	if body != nil {
		if data, ok := body["data"].(map[string]any); ok {
			if _, hasToken := data["accessToken"]; hasToken {
				t.Fatalf("must not issue access token when fail-count UPDATE fails")
			}
		}
	}

	// The log must contain the error with request_id.
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "login fail count update error") {
		t.Fatalf("expected 'login fail count update error' in log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "request_id") {
		t.Fatalf("expected request_id in error log, got: %s", logOutput)
	}

	// Verify the fail count was NOT incremented (because the UPDATE failed).
	var failCount int
	err = db.QueryRow(`SELECT login_fail_count FROM user_account WHERE username = 'victim'`).Scan(&failCount)
	if err != nil {
		t.Fatalf("query fail count: %v", err)
	}
	if failCount != 0 {
		t.Fatalf("fail count must remain 0 when UPDATE failed, got %d", failCount)
	}
}

// failingUpdateDB wraps *sql.DB and returns an error for any QueryRowContext
// call whose query contains "UPDATE". This simulates a DB write failure while
// allowing SELECTs to pass through. The error is produced by querying a
// separate closed *sql.DB, whose Row.Scan always returns an error.
type failingUpdateDB struct {
	db       *sql.DB
	closedDB *sql.DB
}

func (f *failingUpdateDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	tx, err := f.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (f *failingUpdateDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.db.ExecContext(ctx, query, args...)
}

func (f *failingUpdateDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if strings.Contains(strings.ToUpper(query), "UPDATE") {
		// Return a Row from a closed DB; Scan will fail with "sql: database is closed".
		return f.closedDB.QueryRowContext(ctx, "SELECT 1")
	}
	return f.db.QueryRowContext(ctx, query, args...)
}

func (f *failingUpdateDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.db.QueryContext(ctx, query, args...)
}

// ==================== P2.2: Refresh error logs include request_id (real DB error) ====================

func TestRefreshErrorLogIncludesRequestID(t *testing.T) {
	tmpDir := t.TempDir()
	dsn := "file:" + filepath.Join(tmpDir, "test.db")

	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	// Create a user and a valid refresh session so the cookie is real.
	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	handler := appauth.NewHandler(db, "test-secret-at-least-32-chars-long", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	wrapped := logging.NewMiddleware(logger)(mux)
	srv := httptest.NewServer(wrapped)

	// Login to get a real refresh cookie.
	_, _, loginResp := doRequest(t, "POST", srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)
	cookie := extractRefreshCookie(loginResp)
	if cookie == nil {
		t.Fatalf("no refresh cookie after login")
	}

	// Close the DB so the refresh query will fail with a real DB error.
	// This hits the `h.logger.Error("refresh query error", ...)` path.
	db.Close()

	// Call refresh with the real cookie; the DB query will fail.
	req, _ := http.NewRequest("POST", srv.URL+"/auth/refresh", nil)
	req.AddCookie(cookie)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("refresh request: %v", err)
	}
	resp.Body.Close()

	srv.Close()

	logOutput := logBuf.String()

	// The Refresh handler's own error log must be present.
	if !strings.Contains(logOutput, "refresh query error") {
		t.Fatalf("expected 'refresh query error' in log (Refresh's own logger.Error), got: %s", logOutput)
	}

	// The error log must include request_id.
	if !strings.Contains(logOutput, "request_id") {
		t.Fatalf("expected request_id in refresh error log, got: %s", logOutput)
	}

	// The log must NOT contain the raw refresh token value.
	if cookie != nil && strings.Contains(logOutput, cookie.Value) {
		t.Fatalf("log must not leak the refresh token value: %s", logOutput)
	}
}

// ==================== r4: Auth bypass removal regression tests ====================

// TestAuthBypassRemovedFailingDBRejectsUnauthenticatedME verifies that when
// the Handler uses a failingUpdateDB wrapper for business SQL, the AuthMiddleware
// still uses the real *sql.DB passed to MountRoutes and rejects unauthenticated
// requests to GET /auth/me with 401 / 40101.
func TestAuthBypassRemovedFailingDBRejectsUnauthenticatedME(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	closedDB, _ := sql.Open("sqlite", "file:"+filepath.Join(tmpDir, "closed.db"))
	closedDB.Close()
	failingDB := &failingUpdateDB{db: db, closedDB: closedDB}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r4", logger)
	mux := httpserver.New()
	// Pass the REAL db to MountRoutes for AuthMiddleware.
	mux = appauth.MountRoutes(mux, handler, db)
	wrapped := logging.NewMiddleware(logger)(mux)
	srv := httptest.NewServer(wrapped)
	defer srv.Close()

	// No Authorization header must get 401 / 40101.
	req, _ := http.NewRequest("GET", srv.URL+"/auth/me", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /auth/me: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated GET /auth/me, got %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var body map[string]any
	_ = json.Unmarshal(data, &body)
	if body["code"] != float64(40101) {
		t.Fatalf("expected code 40101, got %v", body["code"])
	}
}

// TestAuthBypassRemovedFailingDBRejectsUnauthenticatedUsers verifies that when
// the Handler uses a failingUpdateDB wrapper for business SQL, the AuthMiddleware
// still uses the real *sql.DB passed to MountRoutes and rejects unauthenticated
// requests to GET /users with 401 / 40101.
func TestAuthBypassRemovedFailingDBRejectsUnauthenticatedUsers(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	closedDB, _ := sql.Open("sqlite", "file:"+filepath.Join(tmpDir, "closed.db"))
	closedDB.Close()
	failingDB := &failingUpdateDB{db: db, closedDB: closedDB}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r4", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	wrapped := logging.NewMiddleware(logger)(mux)
	srv := httptest.NewServer(wrapped)
	defer srv.Close()

	// No Authorization header must get 401 / 40101.
	req, _ := http.NewRequest("GET", srv.URL+"/users", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /users: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated GET /users, got %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var body map[string]any
	_ = json.Unmarshal(data, &body)
	if body["code"] != float64(40101) {
		t.Fatalf("expected code 40101, got %v", body["code"])
	}
}

// TestAuthBypassRemovedFailingDBAuthenticatedMESucceeds verifies that when
// the Handler uses a failingUpdateDB wrapper for business SQL, the AuthMiddleware
// still uses the real *sql.DB passed to MountRoutes and accepts a valid Bearer
// token for GET /auth/me, returning 200. This proves the middleware is genuinely
// checking the DB, not blindly rejecting all requests.
func TestAuthBypassRemovedFailingDBAuthenticatedMESucceeds(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	closedDB, _ := sql.Open("sqlite", "file:"+filepath.Join(tmpDir, "closed.db"))
	closedDB.Close()
	failingDB := &failingUpdateDB{db: db, closedDB: closedDB}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r4", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	wrapped := logging.NewMiddleware(logger)(mux)
	srv := httptest.NewServer(wrapped)
	defer srv.Close()

	// Login to get a real access token (Login uses failingDB for business SQL,
	// but the SELECT and password verify don't hit UPDATE, so login succeeds).
	token := loginAndGetTokenWith(t, srv, "owner1", "Pass1234")
	if token == "" {
		t.Fatalf("login failed: no token returned")
	}

	// GET /auth/me with valid Bearer token must get 200.
	code, body, _ := doRequestWithToken(t, "GET", srv.URL+"/auth/me", token, nil)
	if code != http.StatusOK {
		t.Fatalf("expected 200 for authenticated GET /auth/me, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(0) {
		t.Fatalf("expected code 0, got body=%v", body)
	}
}

// loginAndGetTokenWith logs in against the given server and returns the access token.
func loginAndGetTokenWith(t *testing.T, srv *httptest.Server, username, password string) string {
	t.Helper()
	req, _ := http.NewRequest("POST", srv.URL+"/auth/login",
		bytes.NewReader([]byte(`{"username":"`+username+`","password":"`+password+`"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var body map[string]any
	_ = json.Unmarshal(data, &body)
	if d, ok := body["data"].(map[string]any); ok {
		if token, ok := d["accessToken"].(string); ok {
			return token
		}
	}
	return ""
}

// ==================== r5: Audit, transaction, and error contract tests ====================

// failingTxDB wraps *sql.DB and makes BeginTx return an error, simulating
// a transaction-level failure. All other methods pass through.
type failingTxDB struct {
	db *sql.DB
}

func (f *failingTxDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	return nil, fmt.Errorf("simulated BeginTx failure")
}
func (f *failingTxDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.db.ExecContext(ctx, query, args...)
}
func (f *failingTxDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.db.QueryRowContext(ctx, query, args...)
}
func (f *failingTxDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.db.QueryContext(ctx, query, args...)
}

// failingInsertDB wraps *sql.DB and fails any INSERT query (returns error
// from ExecContext or makes QueryRowContext return a closed-DB Row).
// SELECTs and UPDATEs pass through. This simulates a write failure inside
// a transaction (e.g. audit log INSERT fails).
type failingInsertDB struct {
	db       *sql.DB
	closedDB *sql.DB
}

func (f *failingInsertDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	tx, err := f.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &failingInsertTx{tx: tx}, nil
}
func (f *failingInsertDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if strings.Contains(strings.ToUpper(query), "INSERT") {
		return nil, fmt.Errorf("simulated INSERT failure")
	}
	return f.db.ExecContext(ctx, query, args...)
}
func (f *failingInsertDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if strings.Contains(strings.ToUpper(query), "INSERT") {
		return f.closedDB.QueryRowContext(ctx, "SELECT 1")
	}
	return f.db.QueryRowContext(ctx, query, args...)
}
func (f *failingInsertDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.db.QueryContext(ctx, query, args...)
}

// failingInsertTx wraps *sql.Tx and fails any INSERT query inside the transaction.
type failingInsertTx struct {
	tx *sql.Tx
}

func (f *failingInsertTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if strings.Contains(strings.ToUpper(query), "INSERT") {
		return nil, fmt.Errorf("simulated INSERT failure in tx")
	}
	return f.tx.ExecContext(ctx, query, args...)
}
func (f *failingInsertTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.tx.QueryRowContext(ctx, query, args...)
}
func (f *failingInsertTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.tx.QueryContext(ctx, query, args...)
}
func (f *failingInsertTx) Commit() error {
	return f.tx.Commit()
}
func (f *failingInsertTx) Rollback() error {
	return f.tx.Rollback()
}

// failingCommitDB wraps *sql.DB and makes any UPDATE to refresh_session fail
// (simulating revoke failure inside a transaction).
type failingSessionUpdateDB struct {
	db       *sql.DB
	closedDB *sql.DB
}

func (f *failingSessionUpdateDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	tx, err := f.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &failingSessionUpdateTx{tx: tx}, nil
}
func (f *failingSessionUpdateDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if strings.Contains(query, "refresh_session") && strings.Contains(strings.ToUpper(query), "UPDATE") {
		return nil, fmt.Errorf("simulated session update failure")
	}
	return f.db.ExecContext(ctx, query, args...)
}
func (f *failingSessionUpdateDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.db.QueryRowContext(ctx, query, args...)
}
func (f *failingSessionUpdateDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.db.QueryContext(ctx, query, args...)
}

// failingSessionUpdateTx wraps *sql.Tx and fails any UPDATE to refresh_session
// inside the transaction.
type failingSessionUpdateTx struct {
	tx *sql.Tx
}

func (f *failingSessionUpdateTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if strings.Contains(query, "refresh_session") && strings.Contains(strings.ToUpper(query), "UPDATE") {
		return nil, fmt.Errorf("simulated session update failure in tx")
	}
	return f.tx.ExecContext(ctx, query, args...)
}
func (f *failingSessionUpdateTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.tx.QueryRowContext(ctx, query, args...)
}
func (f *failingSessionUpdateTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.tx.QueryContext(ctx, query, args...)
}
func (f *failingSessionUpdateTx) Commit() error {
	return f.tx.Commit()
}
func (f *failingSessionUpdateTx) Rollback() error {
	return f.tx.Rollback()
}

// opLogRow holds a row from operation_log for test assertions.
type opLogRow struct {
	ID           int64
	OperatorID   int64
	OperatorName string
	Action       string
	TargetType   string
	TargetID     int64
	DetailJSON   string
	RequestID    string
}

// queryAuditLogs reads all operation_log rows for a given action.
func queryAuditLogs(t *testing.T, db *sql.DB, action string) []opLogRow {
	t.Helper()
	rows, err := db.Query(`SELECT id, operator_id, operator_name, action, target_type, target_id, detail_json, request_id FROM operation_log WHERE action = ? ORDER BY id`, action)
	if err != nil {
		t.Fatalf("query audit logs: %v", err)
	}
	defer rows.Close()
	var result []opLogRow
	for rows.Next() {
		var r opLogRow
		var opName, detail sql.NullString
		var targetID sql.NullInt64
		if err := rows.Scan(&r.ID, &r.OperatorID, &opName, &r.Action, &r.TargetType, &targetID, &detail, &r.RequestID); err != nil {
			t.Fatalf("scan audit log: %v", err)
		}
		r.OperatorName = opName.String
		r.TargetID = targetID.Int64
		r.DetailJSON = detail.String
		result = append(result, r)
	}
	return result
}

// assertNoSensitiveInAudit checks that detail_json does not contain
// password, hash, token, or Authorization values.
func assertNoSensitiveInAudit(t *testing.T, detail string) {
	t.Helper()
	lower := strings.ToLower(detail)
	for _, bad := range []string{"password", "hash", "token", "authorization", "bearer"} {
		if strings.Contains(lower, bad) {
			t.Fatalf("audit detail contains sensitive word %q: %s", bad, detail)
		}
	}
}

// TestOperationLogHasRequestIDColumn verifies the migration added request_id TEXT NOT NULL.
func TestOperationLogHasRequestIDColumn(t *testing.T) {
	db := newMigratedDB(t)
	defer db.Close()

	// Use NULL operator_id (FK allows NULL) to avoid bcrypt cost.
	// Insert must fail without request_id (NOT NULL constraint).
	_, err := db.Exec(`INSERT INTO operation_log (action, target_type) VALUES ('TEST', 'USER')`)
	if err == nil {
		t.Fatalf("expected NOT NULL constraint failure when omitting request_id")
	}

	// Insert with request_id must succeed.
	_, err = db.Exec(`INSERT INTO operation_log (action, target_type, request_id) VALUES ('TEST', 'USER', 'req-123')`)
	if err != nil {
		t.Fatalf("insert with request_id failed: %v", err)
	}

	// Verify the row has the correct request_id.
	var reqID string
	err = db.QueryRow(`SELECT request_id FROM operation_log WHERE action = 'TEST'`).Scan(&reqID)
	if err != nil {
		t.Fatalf("query request_id: %v", err)
	}
	if reqID != "req-123" {
		t.Fatalf("expected req-123, got %s", reqID)
	}
}

// TestLoginSuccessWritesAuditLog verifies that a successful login writes
// an audit row with request_id and no sensitive fields.
func TestLoginSuccessWritesAuditLog(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	code, body, resp := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", code, body)
	}

	rid := resp.Header.Get("X-Request-Id")
	if rid == "" {
		// Try to extract from the response body's requestId field.
		if body != nil {
			if r, ok := body["requestId"].(string); ok {
				rid = r
			}
		}
	}

	logs := queryAuditLogs(t, ts.db, "LOGIN")
	if len(logs) != 1 {
		t.Fatalf("expected 1 LOGIN audit row, got %d", len(logs))
	}
	if logs[0].RequestID == "" {
		t.Fatalf("audit row has empty request_id")
	}
	if logs[0].OperatorID == 0 {
		t.Fatalf("audit row has no operator_id")
	}
	if logs[0].TargetType != "USER" {
		t.Fatalf("expected target_type USER, got %s", logs[0].TargetType)
	}
	assertNoSensitiveInAudit(t, logs[0].DetailJSON)
}

// TestRefreshSuccessWritesAuditLog verifies that a successful refresh writes
// an audit row with request_id and no sensitive fields.
func TestRefreshSuccessWritesAuditLog(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	_, _, resp := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)
	cookie := extractRefreshCookie(resp)
	if cookie == nil {
		t.Fatalf("no refresh cookie after login")
	}

	code, body, _ := doRequestWithCookie(t, "POST", ts.srv.URL+"/auth/refresh", cookie)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", code, body)
	}

	logs := queryAuditLogs(t, ts.db, "REFRESH")
	if len(logs) != 1 {
		t.Fatalf("expected 1 REFRESH audit row, got %d", len(logs))
	}
	if logs[0].RequestID == "" {
		t.Fatalf("audit row has empty request_id")
	}
	assertNoSensitiveInAudit(t, logs[0].DetailJSON)
}

// TestLogoutSuccessWritesAuditLog verifies that a successful logout writes
// an audit row with request_id and no sensitive fields.
func TestLogoutSuccessWritesAuditLog(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	token := loginAndGetToken(t, ts, "owner1", "Pass1234")
	_, _, resp := doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)
	cookie := extractRefreshCookie(resp)

	code, body, _ := doRequestWithTokenAndCookie(t, "POST", ts.srv.URL+"/auth/logout", token, cookie)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", code, body)
	}

	logs := queryAuditLogs(t, ts.db, "LOGOUT")
	if len(logs) != 1 {
		t.Fatalf("expected 1 LOGOUT audit row, got %d", len(logs))
	}
	if logs[0].RequestID == "" {
		t.Fatalf("audit row has empty request_id")
	}
	assertNoSensitiveInAudit(t, logs[0].DetailJSON)
}

// TestCreateOperatorWritesAuditLog verifies that creating an Operator writes
// an audit row with request_id and no sensitive fields.
func TestCreateOperatorWritesAuditLog(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	token := loginAndGetToken(t, ts, "owner1", "Pass1234")

	code, body, _ := doRequestWithToken(t, "POST", ts.srv.URL+"/users", token,
		map[string]string{"username": "op1", "password": "Pass1234", "role": "OPERATOR"})
	if code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%v", code, body)
	}

	logs := queryAuditLogs(t, ts.db, "CREATE_OPERATOR")
	if len(logs) != 1 {
		t.Fatalf("expected 1 CREATE_OPERATOR audit row, got %d", len(logs))
	}
	if logs[0].RequestID == "" {
		t.Fatalf("audit row has empty request_id")
	}
	assertNoSensitiveInAudit(t, logs[0].DetailJSON)
}

// TestDisableUserWritesAuditLog verifies that disabling a user writes
// an audit row with request_id and no sensitive fields.
func TestDisableUserWritesAuditLog(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	createTestUser(t, ts.db, "op1", "Pass1234", "OPERATOR", "ACTIVE")
	token := loginAndGetToken(t, ts, "owner1", "Pass1234")

	code, body, _ := doRequestWithToken(t, "POST", ts.srv.URL+"/users/2/disable", token, nil)
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", code, body)
	}

	logs := queryAuditLogs(t, ts.db, "DISABLE_USER")
	if len(logs) != 1 {
		t.Fatalf("expected 1 DISABLE_USER audit row, got %d", len(logs))
	}
	if logs[0].RequestID == "" {
		t.Fatalf("audit row has empty request_id")
	}
	assertNoSensitiveInAudit(t, logs[0].DetailJSON)
}

// TestLoginAuditFailureRollsBackAll verifies that if the audit log INSERT
// (INSERT INTO operation_log) fails inside the login transaction, the fail
// count reset and refresh session insertion are also rolled back; no partial
// writes. The refresh_session INSERT must succeed first, proving the test
// reaches the audit INSERT, not just the session INSERT.
// Establishes fail_count=1 first, then asserts it remains 1 after rollback.
func TestLoginAuditFailureRollsBackAll(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// Step 1: Do a wrong-password login with the REAL db to set fail_count to 1.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler0 := appauth.NewHandler(db, "test-jwt-secret-for-r7-audit", logger)
	mux0 := httpserver.New()
	mux0 = appauth.MountRoutes(mux0, handler0, db)
	srv0 := httptest.NewServer(logging.NewMiddleware(logger)(mux0))

	req0, _ := http.NewRequest("POST", srv0.URL+"/auth/login",
		bytes.NewReader([]byte(`{"username":"owner1","password":"wrongpass1"}`)))
	req0.Header.Set("Content-Type", "application/json")
	resp0, err := http.DefaultClient.Do(req0)
	if err != nil {
		t.Fatalf("wrong-password login: %v", err)
	}
	resp0.Body.Close()
	srv0.Close()

	// Verify fail_count is exactly 1.
	var failCount int
	err = db.QueryRow(`SELECT login_fail_count FROM user_account WHERE username = 'owner1'`).Scan(&failCount)
	if err != nil {
		t.Fatalf("query fail count: %v", err)
	}
	if failCount != 1 {
		t.Fatalf("expected fail_count=1 after wrong password, got %d", failCount)
	}

	// Step 2: Use failingAuditDB which only fails INSERT INTO operation_log.
	// This lets UPDATE user_account and INSERT INTO refresh_session succeed
	// inside the tx, then fails on the audit INSERT, proving the test
	// reaches the audit step, not just the session INSERT.
	failingDB := &failingAuditDB{db: db}

	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r7-audit", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	// Attempt successful login; the audit INSERT should fail inside the tx.
	req, _ := http.NewRequest("POST", srv.URL+"/auth/login",
		bytes.NewReader([]byte(`{"username":"owner1","password":"Pass1234"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 when audit INSERT fails, got %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var body map[string]any
	_ = json.Unmarshal(data, &body)
	if body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got %v", body["code"])
	}

	// Assert fail_count is STILL 1 (the tx rolled back, so the reset did not persist).
	var finalFailCount int
	err = db.QueryRow(`SELECT login_fail_count FROM user_account WHERE username = 'owner1'`).Scan(&finalFailCount)
	if err != nil {
		t.Fatalf("query final fail count: %v", err)
	}
	if finalFailCount != 1 {
		t.Fatalf("expected fail_count=1 after rollback (not reset), got %d", finalFailCount)
	}

	// Assert 0 refresh sessions (the session INSERT succeeded inside the tx
	// but was rolled back when the audit INSERT failed).
	var sessionCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM refresh_session WHERE user_id = (SELECT id FROM user_account WHERE username = 'owner1')`).Scan(&sessionCount)
	if err != nil {
		t.Fatalf("query session count: %v", err)
	}
	if sessionCount != 0 {
		t.Fatalf("expected 0 refresh sessions after rollback, got %d", sessionCount)
	}

	// Assert 0 LOGIN audit rows.
	var auditCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM operation_log WHERE action = 'LOGIN'`).Scan(&auditCount)
	if err != nil {
		t.Fatalf("query audit count: %v", err)
	}
	if auditCount != 0 {
		t.Fatalf("expected 0 LOGIN audit rows after rollback, got %d", auditCount)
	}
}

// TestLogoutRevokeFailureDoesNotReturnSuccess verifies that if the logout
// transaction fails, the response is not 200 and the session is still active.
func TestLogoutRevokeFailureDoesNotReturnSuccess(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// First login with real DB to get a valid token and cookie.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := appauth.NewHandler(db, "test-jwt-secret-for-r5-logout", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))

	_, _, loginResp := doRequest(t, "POST", srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "Pass1234"}, nil)
	cookie := extractRefreshCookie(loginResp)
	token := loginAndGetTokenWith(t, srv, "owner1", "Pass1234")
	srv.Close()

	// Now set up a new server with failingSessionUpdateDB so the logout
	// revoke UPDATE fails.
	failingDB := &failingSessionUpdateDB{db: db, closedDB: db}
	handler2 := appauth.NewHandler(failingDB, "test-jwt-secret-for-r5-logout", logger)
	mux2 := httpserver.New()
	mux2 = appauth.MountRoutes(mux2, handler2, db)
	srv2 := httptest.NewServer(logging.NewMiddleware(logger)(mux2))
	defer srv2.Close()

	// Attempt logout; the revoke should fail.
	code, body, _ := doRequestWithTokenAndCookie(t, "POST", srv2.URL+"/auth/logout", token, cookie)
	if code == http.StatusOK {
		t.Fatalf("must not return 200 when logout revoke fails, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got body=%v", body)
	}

	// The session should still be active (revoked_at IS NULL).
	var revokedAt sql.NullString
	_ = db.QueryRow(`SELECT revoked_at FROM refresh_session WHERE token_hash = ?`,
		auth.HashRefreshToken(cookie.Value)).Scan(&revokedAt)
	if revokedAt.Valid {
		t.Fatalf("session must still be active (revoked_at IS NULL) after failed logout")
	}
}

// TestDisableRevokeFailureRollsBack verifies that if the session revoke
// fails inside the disable transaction, the account status is NOT set to
// DISABLED; everything rolls back.
func TestDisableRevokeFailureRollsBack(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	createTestUser(t, db, "op1", "Pass1234", "OPERATOR", "ACTIVE")

	// Login op1 to create an active session.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := appauth.NewHandler(db, "test-jwt-secret-for-r5-disable", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))

	_, _, opLoginResp := doRequest(t, "POST", srv.URL+"/auth/login",
		map[string]string{"username": "op1", "password": "Pass1234"}, nil)
	opCookie := extractRefreshCookie(opLoginResp)
	ownerToken := loginAndGetTokenWith(t, srv, "owner1", "Pass1234")
	srv.Close()

	// Now set up a new server with failingSessionUpdateDB so the disable
	// revoke sessions UPDATE fails.
	failingDB := &failingSessionUpdateDB{db: db, closedDB: db}
	handler2 := appauth.NewHandler(failingDB, "test-jwt-secret-for-r5-disable", logger)
	mux2 := httpserver.New()
	mux2 = appauth.MountRoutes(mux2, handler2, db)
	srv2 := httptest.NewServer(logging.NewMiddleware(logger)(mux2))
	defer srv2.Close()

	// Attempt disable; the session revoke should fail.
	code, body, _ := doRequestWithToken(t, "POST", srv2.URL+"/users/2/disable", ownerToken, nil)
	if code == http.StatusOK {
		t.Fatalf("must not return 200 when disable revoke fails, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got body=%v", body)
	}

	// The account status should still be ACTIVE (rolled back).
	var status string
	_ = db.QueryRow(`SELECT status FROM user_account WHERE id = 2`).Scan(&status)
	if status != "ACTIVE" {
		t.Fatalf("account must still be ACTIVE after rollback, got %s", status)
	}

	// The session should still be active (revoked_at IS NULL).
	if opCookie != nil {
		var revokedAt sql.NullString
		_ = db.QueryRow(`SELECT revoked_at FROM refresh_session WHERE token_hash = ?`,
			auth.HashRefreshToken(opCookie.Value)).Scan(&revokedAt)
		if revokedAt.Valid {
			t.Fatalf("session must still be active after rollback")
		}
	}

	// No DISABLE_USER audit should exist.
	var auditCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM operation_log WHERE action = 'DISABLE_USER'`).Scan(&auditCount)
	if auditCount != 0 {
		t.Fatalf("expected 0 DISABLE_USER audit rows after rollback, got %d", auditCount)
	}
}

// TestRefreshAndDisableConcurrentNoActiveSession verifies that when refresh
// is paused inside its transaction (after the ACTIVE check, before revoking
// the old session) and disable completes fully, refresh cannot succeed and
// the disabled account ends up with no active refresh sessions.
//
// Uses a channel barrier in a test-only Tx wrapper to create deterministic
// interleaving: refresh's tx blocks after the ACTIVE check SELECT and before
// the UPDATE refresh_session (revoke). Disable then commits fully (setting
// status=DISABLED and revoking all sessions). Refresh is released — its
// UPDATE ... WHERE revoked_at IS NULL finds rows=0 (disable already revoked),
// so it rolls back and returns 401.
func TestRefreshAndDisableConcurrentNoActiveSession(t *testing.T) {
	tmpDir := t.TempDir()
	dsn := "file:" + filepath.Join(tmpDir, "test.db")
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	// Allow 2 concurrent connections so refresh's tx and disable's tx
	// can both get a connection. WAL mode ensures reads don't block writes.
	db.SetMaxOpenConns(2)

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	createTestUser(t, db, "op1", "Pass1234", "OPERATOR", "ACTIVE")

	// Login op1 with real DB to get a refresh cookie.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler0 := appauth.NewHandler(db, "test-jwt-secret-for-r7-race", logger)
	mux0 := httpserver.New()
	mux0 = appauth.MountRoutes(mux0, handler0, db)
	srv0 := httptest.NewServer(logging.NewMiddleware(logger)(mux0))

	_, _, opLoginResp := doRequest(t, "POST", srv0.URL+"/auth/login",
		map[string]string{"username": "op1", "password": "Pass1234"}, nil)
	opCookie := extractRefreshCookie(opLoginResp)
	if opCookie == nil {
		t.Fatalf("no refresh cookie for op1")
	}
	ownerToken := loginAndGetTokenWith(t, srv0, "owner1", "Pass1234")
	srv0.Close()

	// Barrier channels: refresh signals it has completed the ACTIVE check
	// and is about to revoke; test signals refresh to proceed after disable.
	reached := make(chan struct{})
	release := make(chan struct{})
	done := make(chan struct{})

	barrierDB := &barrierRefreshDB{db: db, reached: reached, release: release}

	handler := appauth.NewHandler(barrierDB, "test-jwt-secret-for-r7-race", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	// Start refresh in a goroutine — it will block at the barrier
	// (after ACTIVE check, before revoke UPDATE).
	var refreshCode int
	var refreshBody map[string]any
	go func() {
		refreshCode, refreshBody, _ = doRequestWithCookie(t, "POST", srv.URL+"/auth/refresh", opCookie)
		close(done)
	}()

	// Wait for refresh to reach the barrier.
	select {
	case <-reached:
	case <-time.After(10 * time.Second):
		t.Fatalf("refresh did not reach barrier within 10s")
	}

	// Disable op1 (user id=2) while refresh is blocked.
	// Disable's tx goes through barrierDB.BeginTx -> barrierRefreshTx,
	// but barrierRefreshTx only blocks on UPDATE refresh_session ... WHERE id = ?
	// (refresh's revoke pattern), not WHERE user_id = ? (disable's revoke-all).
	disableCode, disableBody, _ := doRequestWithToken(t, "POST", srv.URL+"/users/2/disable", ownerToken, nil)
	if disableCode != http.StatusOK {
		t.Fatalf("disable must succeed, got %d body=%v", disableCode, disableBody)
	}

	// Release the barrier — refresh continues.
	close(release)

	// Wait for refresh to complete.
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("refresh did not complete within 10s after release")
	}

	// Refresh must NOT return 200 — disable already revoked the session.
	if refreshCode == http.StatusOK {
		t.Fatalf("refresh must not succeed after disable committed, got 200 body=%v", refreshBody)
	}

	// Final invariant: no active sessions for the disabled user.
	var activeCount int
	err = db.QueryRow(`SELECT COUNT(*) FROM refresh_session WHERE user_id = 2 AND revoked_at IS NULL`).Scan(&activeCount)
	if err != nil {
		t.Fatalf("query active sessions: %v", err)
	}
	if activeCount != 0 {
		t.Fatalf("expected 0 active sessions for disabled user, got %d", activeCount)
	}

	// Account must be DISABLED.
	var status string
	err = db.QueryRow(`SELECT status FROM user_account WHERE id = 2`).Scan(&status)
	if err != nil {
		t.Fatalf("query status: %v", err)
	}
	if status != "DISABLED" {
		t.Fatalf("expected DISABLED, got %s", status)
	}
}

// TestDBFailureReturns50002Not40101Or42201 verifies that database/transaction
// failures return HTTP 500 + 50002, not 40101 or 42201.
func TestDBFailureReturns50002Not40101Or42201(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// Use failingTxDB so BeginTx fails during login.
	failingDB := &failingTxDB{db: db}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r5-errcode", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	// Login should fail with 500 + 50002 (BeginTx fails).
	req, _ := http.NewRequest("POST", srv.URL+"/auth/login",
		bytes.NewReader([]byte(`{"username":"owner1","password":"Pass1234"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var body map[string]any
	_ = json.Unmarshal(data, &body)
	if body["code"] == float64(40101) {
		t.Fatalf("must not return 40101 for DB failure, got 40101")
	}
	if body["code"] == float64(42201) {
		t.Fatalf("must not return 42201 for DB failure, got 42201")
	}
	if body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got %v", body["code"])
	}
}

// TestFailedLoginNoSuccessAudit verifies that a failed login attempt does
// not produce a LOGIN audit row (only successful writes are audited).
func TestFailedLoginNoSuccessAudit(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// Wrong password should fail.
	_, _, _ = doRequest(t, "POST", ts.srv.URL+"/auth/login",
		map[string]string{"username": "owner1", "password": "wrongpass1"}, nil)

	logs := queryAuditLogs(t, ts.db, "LOGIN")
	if len(logs) != 0 {
		t.Fatalf("expected 0 LOGIN audit rows for failed login, got %d", len(logs))
	}
}

// TestUnauthenticatedRequestNoAudit verifies that unauthenticated requests
// to protected endpoints do not produce audit rows.
func TestUnauthenticatedRequestNoAudit(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// No auth header on /users.
	_, _, _ = doRequest(t, "GET", ts.srv.URL+"/users", nil, nil)

	// No audit rows should exist at all.
	var totalAudit int
	_ = ts.db.QueryRow(`SELECT COUNT(*) FROM operation_log`).Scan(&totalAudit)
	if totalAudit != 0 {
		t.Fatalf("expected 0 audit rows for unauthenticated request, got %d", totalAudit)
	}
}

// TestConflictRequestNoSuccessAudit verifies that a conflict (e.g. duplicate
// username on create) does not produce a CREATE_OPERATOR audit row.
func TestConflictRequestNoSuccessAudit(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	createTestUser(t, ts.db, "op1", "Pass1234", "OPERATOR", "ACTIVE")
	token := loginAndGetToken(t, ts, "owner1", "Pass1234")

	// Try to create op1 again; should conflict.
	_, body, _ := doRequestWithToken(t, "POST", ts.srv.URL+"/users", token,
		map[string]string{"username": "op1", "password": "Pass1234", "role": "OPERATOR"})
	if body == nil || body["code"] != float64(40901) {
		t.Fatalf("expected 40901 conflict, got body=%v", body)
	}

	logs := queryAuditLogs(t, ts.db, "CREATE_OPERATOR")
	if len(logs) != 0 {
		t.Fatalf("expected 0 CREATE_OPERATOR audit rows for conflict, got %d", len(logs))
	}
}

// ==================== r5: Test helpers ====================

// newMigratedDB opens a temp DB and runs all migrations.
func newMigratedDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dsn := "file:" + filepath.Join(tmpDir, "test.db")
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	return db
}

// doRequestWithCookie performs an HTTP request with a cookie but no token.
func doRequestWithCookie(t *testing.T, method, url string, cookie *http.Cookie) (int, map[string]any, *http.Response) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var parsed map[string]any
	_ = json.Unmarshal(data, &parsed)
	return resp.StatusCode, parsed, resp
}

// doRequestWithTokenAndCookie performs an HTTP request with both a Bearer
// token and a cookie.
func doRequestWithTokenAndCookie(t *testing.T, method, url, token string, cookie *http.Cookie) (int, map[string]any, *http.Response) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var parsed map[string]any
	_ = json.Unmarshal(data, &parsed)
	return resp.StatusCode, parsed, resp
}

// doRequestOnCustomServer is a helper for login on a custom server setup.
func doRequestOnCustomServer(t *testing.T, db *sql.DB, failingDB interface{}, username, password string,
	mountFn func(h *appauth.Handler, mux *http.ServeMux), method, path string) (int, map[string]any, *http.Response) {
	t.Helper()
	// This is a simplified helper, not used in the main test flow.
	return 0, nil, nil
}

// ==================== r6: P1-1 Login/disable race: deterministic interleaving ====================

// failingInsertAfterUpdateTx wraps *sql.Tx and lets the first N ExecContext
// calls pass through, then fails the (N+1)th. This creates a deterministic
// interleaving point: the UPDATE succeeds (fail count reset) but the INSERT
// (refresh session) or audit fails, proving the tx rolls back.
type failingInsertAfterUpdateTx struct {
	tx        *sql.Tx
	callCount int
	failAt    int
}

func (f *failingInsertAfterUpdateTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	f.callCount++
	if f.callCount == f.failAt {
		return nil, fmt.Errorf("deterministic failure at call %d", f.failAt)
	}
	return f.tx.ExecContext(ctx, query, args...)
}
func (f *failingInsertAfterUpdateTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.tx.QueryRowContext(ctx, query, args...)
}
func (f *failingInsertAfterUpdateTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.tx.QueryContext(ctx, query, args...)
}
func (f *failingInsertAfterUpdateTx) Commit() error {
	return f.tx.Commit()
}
func (f *failingInsertAfterUpdateTx) Rollback() error {
	return f.tx.Rollback()
}

// loginRaceTx wraps *sql.Tx. On the first ExecContext (the login tx's
// UPDATE user_account SET login_fail_count=0 ... WHERE status='ACTIVE'),
// it FIRST disables the target account via a SEPARATE *sql.DB connection,
// THEN runs the tx's UPDATE. This creates a deterministic race: the
// outside read saw ACTIVE, but by the time the tx UPDATE runs, the account
// is DISABLED.
//
// A separate *sql.DB is needed because database.Open sets SetMaxOpenConns(1),
// so the tx holds the only connection on the main DB. The sideDB is a
// second connection to the same SQLite file with its own pool.
type loginRaceTx struct {
	tx           *sql.Tx
	sideDB       *sql.DB
	disabledUser int64
	disabled     bool
}

func (f *loginRaceTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if !f.disabled && strings.Contains(query, "UPDATE user_account") {
		// Disable the account via a separate connection BEFORE the tx's
		// UPDATE runs. SQLite BEGIN DEFERRED doesn't hold a write lock
		// until the first write, so the sideDB's UPDATE can acquire and
		// release the write lock before the tx's UPDATE.
		_, _ = f.sideDB.ExecContext(ctx,
			`UPDATE user_account SET status = 'DISABLED' WHERE id = ?`,
			f.disabledUser,
		)
		f.disabled = true
	}
	return f.tx.ExecContext(ctx, query, args...)
}
func (f *loginRaceTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.tx.QueryRowContext(ctx, query, args...)
}
func (f *loginRaceTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.tx.QueryContext(ctx, query, args...)
}
func (f *loginRaceTx) Commit() error {
	return f.tx.Commit()
}
func (f *loginRaceTx) Rollback() error {
	return f.tx.Rollback()
}

// loginRaceDB wraps *sql.DB. Its BeginTx returns a loginRaceTx that
// disables the target account before the first tx UPDATE, using a
// separate sideDB connection to avoid pool deadlock.
type loginRaceDB struct {
	db           *sql.DB
	sideDB       *sql.DB
	disabledUser int64
}

func (f *loginRaceDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	tx, err := f.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &loginRaceTx{tx: tx, sideDB: f.sideDB, disabledUser: f.disabledUser}, nil
}
func (f *loginRaceDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.db.ExecContext(ctx, query, args...)
}
func (f *loginRaceDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.db.QueryRowContext(ctx, query, args...)
}
func (f *loginRaceDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.db.QueryContext(ctx, query, args...)
}

// TestLoginDisableRaceNoActiveSession verifies that when an account is
// disabled between the login's outside read and the login transaction's
// UPDATE, the login fails (no token/cookie), no refresh session is created,
// and no LOGIN audit is written. The UPDATE ... WHERE status='ACTIVE' sees
// rows-affected=0 and rolls back.
func TestLoginDisableRaceNoActiveSession(t *testing.T) {
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

	createTestUser(t, db, "victim", "Pass1234", "OPERATOR", "ACTIVE")
	var victimID int64
	_ = db.QueryRow(`SELECT id FROM user_account WHERE username = 'victim'`).Scan(&victimID)

	// Open a separate DB connection to the same SQLite file for the
	// disable operation. This avoids the SetMaxOpenConns(1) deadlock.
	sideDB, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open side db: %v", err)
	}
	defer sideDB.Close()

	// loginRaceDB will disable the account inside the tx's first UPDATE,
	// using the sideDB connection to avoid pool deadlock.
	raceDB := &loginRaceDB{db: db, sideDB: sideDB, disabledUser: victimID}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := appauth.NewHandler(raceDB, "test-jwt-secret-for-r6-race", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	// Attempt login: the outside read sees ACTIVE, but the tx UPDATE
	// with status='ACTIVE' will see rows-affected=0 because the account
	// was disabled inside BeginTx.
	req, _ := http.NewRequest("POST", srv.URL+"/auth/login",
		bytes.NewReader([]byte(`{"username":"victim","password":"Pass1234"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp.Body.Close()

	// Must NOT get 200; the login must fail.
	if resp.StatusCode == http.StatusOK {
		t.Fatalf("login must not succeed when account is disabled during tx, got 200")
	}

	// Must not return a token in the body.
	data, _ := io.ReadAll(resp.Body)
	var body map[string]any
	_ = json.Unmarshal(data, &body)
	if body != nil {
		if d, ok := body["data"].(map[string]any); ok {
			if _, hasToken := d["accessToken"]; hasToken {
				t.Fatalf("must not issue access token when account disabled during tx")
			}
		}
	}

	// No refresh session should exist.
	var sessionCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM refresh_session WHERE user_id = ?`, victimID).Scan(&sessionCount)
	if sessionCount != 0 {
		t.Fatalf("expected 0 refresh sessions after race, got %d", sessionCount)
	}

	// No LOGIN audit should exist.
	var auditCount int
	_ = db.QueryRow(`SELECT COUNT(*) FROM operation_log WHERE action = 'LOGIN'`).Scan(&auditCount)
	if auditCount != 0 {
		t.Fatalf("expected 0 LOGIN audit rows after race, got %d", auditCount)
	}

	// The account must be DISABLED.
	var status string
	_ = db.QueryRow(`SELECT status FROM user_account WHERE id = ?`, victimID).Scan(&status)
	if status != "DISABLED" {
		t.Fatalf("expected DISABLED, got %s", status)
	}
}

// ==================== r6: P1-2 Error code fault injection ====================

// failingQueryRowDB wraps *sql.DB and makes QueryRowContext return a Row
// from a closed DB (Scan will fail) for any query hitting user_account.
// This simulates a DB error during the AuthMiddleware or Me query.
type failingQueryRowDB struct {
	db       *sql.DB
	closedDB *sql.DB
}

func (f *failingQueryRowDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	tx, err := f.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
func (f *failingQueryRowDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.db.ExecContext(ctx, query, args...)
}
func (f *failingQueryRowDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if strings.Contains(query, "user_account") {
		return f.closedDB.QueryRowContext(ctx, "SELECT 1")
	}
	return f.db.QueryRowContext(ctx, query, args...)
}
func (f *failingQueryRowDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.db.QueryContext(ctx, query, args...)
}

// TestAuthMiddlewareDBErrorReturns50002 verifies that when the AuthMiddleware
// DB query fails (not sql.ErrNoRows), the response is 500 + 50002, not 40101.
func TestAuthMiddlewareDBErrorReturns50002(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// First login with real DB to get a valid token.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler0 := appauth.NewHandler(db, "test-jwt-secret-for-r6-mw", logger)
	mux0 := httpserver.New()
	mux0 = appauth.MountRoutes(mux0, handler0, db)
	srv0 := httptest.NewServer(logging.NewMiddleware(logger)(mux0))
	token := loginAndGetTokenWith(t, srv0, "owner1", "Pass1234")
	srv0.Close()

	// Now set up a server where AuthMiddleware uses a DB wrapper that fails
	// the user_account query. We can't wrap AuthMiddleware's DB directly
	// (it takes *sql.DB), so we close the real DB to cause a query error.
	// Actually, we need a different approach: use a separate closed DB for
	// AuthMiddleware. But MountRoutes takes *sql.DB for AuthMiddleware.
	//
	// Strategy: close the db, then make a request. The AuthMiddleware will
	// try to query the closed DB and get a non-ErrNoRows error.
	closedDB, _ := sql.Open("sqlite", "file:"+filepath.Join(tmpDir, "closed.db"))
	closedDB.Close()

	handler := appauth.NewHandler(db, "test-jwt-secret-for-r6-mw", logger)
	mux := httpserver.New()
	// Pass the closed DB to AuthMiddleware; any query will fail.
	mux = appauth.MountRoutes(mux, handler, closedDB)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	// GET /auth/me with valid token; AuthMiddleware query fails.
	req, _ := http.NewRequest("GET", srv.URL+"/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 for AuthMiddleware DB error, got %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var body map[string]any
	_ = json.Unmarshal(data, &body)
	if body["code"] == float64(40101) {
		t.Fatalf("must not return 40101 for DB error, got 40101")
	}
	if body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got %v", body["code"])
	}
}

// TestMeDBErrorReturns50002 verifies that when the Me handler's DB query
// fails (not sql.ErrNoRows), the response is 500 + 50002, not 40401.
func TestMeDBErrorReturns50002(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// First login with real DB to get a valid token.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler0 := appauth.NewHandler(db, "test-jwt-secret-for-r6-me", logger)
	mux0 := httpserver.New()
	mux0 = appauth.MountRoutes(mux0, handler0, db)
	srv0 := httptest.NewServer(logging.NewMiddleware(logger)(mux0))
	token := loginAndGetTokenWith(t, srv0, "owner1", "Pass1234")
	srv0.Close()

	// Now use failingQueryRowDB so the Me query fails.
	closedDB, _ := sql.Open("sqlite", "file:"+filepath.Join(tmpDir, "closed2.db"))
	closedDB.Close()
	failingDB := &failingQueryRowDB{db: db, closedDB: closedDB}

	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r6-me", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	// GET /auth/me with valid token; Me query fails (not ErrNoRows).
	req, _ := http.NewRequest("GET", srv.URL+"/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 for Me DB error, got %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var body map[string]any
	_ = json.Unmarshal(data, &body)
	if body["code"] == float64(40401) {
		t.Fatalf("must not return 40401 for DB error, got 40401")
	}
	if body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got %v", body["code"])
	}
}

// TestCreateUserNonUniqueDBErrorReturns50002 verifies that when CreateUser's
// INSERT fails for a non-unique-constraint reason, the response is 500 + 50002,
// not 40901.
func TestCreateUserNonUniqueDBErrorReturns50002(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// First login with real DB to get a valid token.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler0 := appauth.NewHandler(db, "test-jwt-secret-for-r6-create", logger)
	mux0 := httpserver.New()
	mux0 = appauth.MountRoutes(mux0, handler0, db)
	srv0 := httptest.NewServer(logging.NewMiddleware(logger)(mux0))
	token := loginAndGetTokenWith(t, srv0, "owner1", "Pass1234")
	srv0.Close()

	// Use failingInsertDB so the INSERT into user_account fails with a
	// non-unique-constraint error (our wrapper returns a generic error).
	closedDB, _ := sql.Open("sqlite", "file:"+filepath.Join(tmpDir, "closed3.db"))
	closedDB.Close()
	failingDB := &failingInsertDB{db: db, closedDB: closedDB}

	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r6-create", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	// POST /users — the INSERT will fail with a non-unique error.
	code, body, _ := doRequestWithToken(t, "POST", srv.URL+"/users", token,
		map[string]string{"username": "op1", "password": "Pass1234", "role": "OPERATOR"})

	if code == http.StatusConflict {
		t.Fatalf("must not return 40901 for non-unique DB error, got 409 body=%v", body)
	}
	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500 for non-unique DB error, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got body=%v", body)
	}
}

// ==================== r7: P1-1 failingAuditTx — only fails INSERT INTO operation_log ====================

// failingAuditTx wraps *sql.Tx and only fails INSERT INTO operation_log.
// All other queries (UPDATE user_account, INSERT INTO refresh_session) pass
// through. This proves the test reaches the audit INSERT step, not just the
// session INSERT.
type failingAuditTx struct {
	tx *sql.Tx
}

func (f *failingAuditTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if strings.Contains(query, "operation_log") {
		return nil, fmt.Errorf("simulated audit INSERT failure")
	}
	return f.tx.ExecContext(ctx, query, args...)
}
func (f *failingAuditTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.tx.QueryRowContext(ctx, query, args...)
}
func (f *failingAuditTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.tx.QueryContext(ctx, query, args...)
}
func (f *failingAuditTx) Commit() error {
	return f.tx.Commit()
}
func (f *failingAuditTx) Rollback() error {
	return f.tx.Rollback()
}

// failingAuditDB wraps *sql.DB and returns failingAuditTx from BeginTx.
type failingAuditDB struct {
	db *sql.DB
}

func (f *failingAuditDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	tx, err := f.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &failingAuditTx{tx: tx}, nil
}
func (f *failingAuditDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.db.ExecContext(ctx, query, args...)
}
func (f *failingAuditDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.db.QueryRowContext(ctx, query, args...)
}
func (f *failingAuditDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.db.QueryContext(ctx, query, args...)
}

// ==================== r7: P1-2 barrierRefreshTx — channel barrier for deterministic interleaving ====================

// barrierRefreshTx wraps *sql.Tx. On the first ExecContext containing
// "UPDATE refresh_session" AND "WHERE id = ?" (refresh's revoke pattern),
// it signals reached and blocks on release. This pauses refresh after the
// ACTIVE check (QueryRowContext) and before the revoke UPDATE.
//
// Disable's revoke-all query contains "WHERE user_id = ?" and does NOT
// trigger the barrier, so disable runs to completion.
type barrierRefreshTx struct {
	tx       *sql.Tx
	reached  chan struct{}
	release  chan struct{}
	signaled bool
}

func (f *barrierRefreshTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if !f.signaled && strings.Contains(query, "UPDATE refresh_session") && strings.Contains(query, "WHERE id = ?") {
		close(f.reached)
		select {
		case <-f.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		f.signaled = true
	}
	return f.tx.ExecContext(ctx, query, args...)
}
func (f *barrierRefreshTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.tx.QueryRowContext(ctx, query, args...)
}
func (f *barrierRefreshTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.tx.QueryContext(ctx, query, args...)
}
func (f *barrierRefreshTx) Commit() error {
	return f.tx.Commit()
}
func (f *barrierRefreshTx) Rollback() error {
	return f.tx.Rollback()
}

// barrierRefreshDB wraps *sql.DB and returns barrierRefreshTx from BeginTx.
type barrierRefreshDB struct {
	db      *sql.DB
	reached chan struct{}
	release chan struct{}
}

func (f *barrierRefreshDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	tx, err := f.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &barrierRefreshTx{tx: tx, reached: f.reached, release: f.release}, nil
}
func (f *barrierRefreshDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.db.ExecContext(ctx, query, args...)
}
func (f *barrierRefreshDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.db.QueryRowContext(ctx, query, args...)
}
func (f *barrierRefreshDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.db.QueryContext(ctx, query, args...)
}

// ==================== r7: P1-3.1 failingRowsAffected — RowsAffected() returns error ====================

// failingRowsAffectedResult wraps sql.Result and makes RowsAffected() fail.
type failingRowsAffectedResult struct {
	result sql.Result
}

func (r *failingRowsAffectedResult) LastInsertId() (int64, error) {
	return r.result.LastInsertId()
}
func (r *failingRowsAffectedResult) RowsAffected() (int64, error) {
	return 0, fmt.Errorf("simulated RowsAffected failure")
}

// failingRowsAffectedTx wraps *sql.Tx and returns failingRowsAffectedResult
// for the first ExecContext matching targetQuery.
type failingRowsAffectedTx struct {
	tx          *sql.Tx
	targetQuery string
	triggered   bool
}

func (f *failingRowsAffectedTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	res, err := f.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if !f.triggered && strings.Contains(query, f.targetQuery) {
		f.triggered = true
		return &failingRowsAffectedResult{result: res}, nil
	}
	return res, nil
}
func (f *failingRowsAffectedTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.tx.QueryRowContext(ctx, query, args...)
}
func (f *failingRowsAffectedTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.tx.QueryContext(ctx, query, args...)
}
func (f *failingRowsAffectedTx) Commit() error {
	return f.tx.Commit()
}
func (f *failingRowsAffectedTx) Rollback() error {
	return f.tx.Rollback()
}

// failingRowsAffectedDB wraps *sql.DB and returns failingRowsAffectedTx
// from BeginTx with the given target query.
type failingRowsAffectedDB struct {
	db          *sql.DB
	targetQuery string
}

func (f *failingRowsAffectedDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	tx, err := f.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &failingRowsAffectedTx{tx: tx, targetQuery: f.targetQuery}, nil
}
func (f *failingRowsAffectedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.db.ExecContext(ctx, query, args...)
}
func (f *failingRowsAffectedDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.db.QueryRowContext(ctx, query, args...)
}
func (f *failingRowsAffectedDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.db.QueryContext(ctx, query, args...)
}

// TestLoginRowsAffectedErrorReturns50002 verifies that when RowsAffected()
// fails on the login tx's UPDATE user_account, the response is 500 + 50002.
func TestLoginRowsAffectedErrorReturns50002(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// failingRowsAffectedDB targets "UPDATE user_account" (login's fail count reset).
	failingDB := &failingRowsAffectedDB{db: db, targetQuery: "UPDATE user_account"}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r7-rowsaffected", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	req, _ := http.NewRequest("POST", srv.URL+"/auth/login",
		bytes.NewReader([]byte(`{"username":"owner1","password":"Pass1234"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	data, _ := io.ReadAll(resp.Body)
	var body map[string]any
	_ = json.Unmarshal(data, &body)
	if body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got %v", body["code"])
	}
}

// TestRefreshRowsAffectedErrorReturns50002 verifies that when RowsAffected()
// fails on the refresh tx's UPDATE refresh_session (revoke), the response is
// 500 + 50002.
func TestRefreshRowsAffectedErrorReturns50002(t *testing.T) {
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

	createTestUser(t, db, "op1", "Pass1234", "OPERATOR", "ACTIVE")

	// Login with real DB to get a refresh cookie.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler0 := appauth.NewHandler(db, "test-jwt-secret-for-r7-rowsaffected", logger)
	mux0 := httpserver.New()
	mux0 = appauth.MountRoutes(mux0, handler0, db)
	srv0 := httptest.NewServer(logging.NewMiddleware(logger)(mux0))
	_, _, opLoginResp := doRequest(t, "POST", srv0.URL+"/auth/login",
		map[string]string{"username": "op1", "password": "Pass1234"}, nil)
	opCookie := extractRefreshCookie(opLoginResp)
	if opCookie == nil {
		t.Fatalf("no refresh cookie")
	}
	srv0.Close()

	// failingRowsAffectedDB targets "UPDATE refresh_session" (refresh's revoke).
	failingDB := &failingRowsAffectedDB{db: db, targetQuery: "UPDATE refresh_session"}

	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r7-rowsaffected", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	code, body, _ := doRequestWithCookie(t, "POST", srv.URL+"/auth/refresh", opCookie)
	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got body=%v", body)
	}
}

// TestDisableRowsAffectedErrorReturns50002 verifies that when RowsAffected()
// fails on the disable tx's UPDATE user_account (set DISABLED), the response
// is 500 + 50002.
func TestDisableRowsAffectedErrorReturns50002(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	createTestUser(t, db, "op1", "Pass1234", "OPERATOR", "ACTIVE")

	// Login with real DB to get owner token.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler0 := appauth.NewHandler(db, "test-jwt-secret-for-r7-rowsaffected", logger)
	mux0 := httpserver.New()
	mux0 = appauth.MountRoutes(mux0, handler0, db)
	srv0 := httptest.NewServer(logging.NewMiddleware(logger)(mux0))
	ownerToken := loginAndGetTokenWith(t, srv0, "owner1", "Pass1234")
	srv0.Close()

	// failingRowsAffectedDB targets "UPDATE user_account" (disable's set status).
	// But login also uses "UPDATE user_account" — we need to be more specific.
	// Disable's query is "UPDATE user_account SET status = 'DISABLED'".
	// Login's query is "UPDATE user_account SET login_fail_count = 0".
	// We target "SET status = 'DISABLED'" to only affect disable.
	failingDB := &failingRowsAffectedDB{db: db, targetQuery: "SET status = 'DISABLED'"}

	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r7-rowsaffected", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	code, body, _ := doRequestWithToken(t, "POST", srv.URL+"/users/2/disable", ownerToken, nil)
	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got body=%v", body)
	}
}

// ==================== r7: P1-3.2 failingLastInsertId — LastInsertId() returns error ====================

// failingLastInsertIdResult wraps sql.Result and makes LastInsertId() fail.
type failingLastInsertIdResult struct {
	result sql.Result
}

func (r *failingLastInsertIdResult) LastInsertId() (int64, error) {
	return 0, fmt.Errorf("simulated LastInsertId failure")
}
func (r *failingLastInsertIdResult) RowsAffected() (int64, error) {
	return r.result.RowsAffected()
}

// failingLastInsertIdTx wraps *sql.Tx and returns failingLastInsertIdResult
// for INSERT INTO user_account.
type failingLastInsertIdTx struct {
	tx        *sql.Tx
	triggered bool
}

func (f *failingLastInsertIdTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	res, err := f.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if !f.triggered && strings.Contains(query, "INSERT INTO user_account") {
		f.triggered = true
		return &failingLastInsertIdResult{result: res}, nil
	}
	return res, nil
}
func (f *failingLastInsertIdTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.tx.QueryRowContext(ctx, query, args...)
}
func (f *failingLastInsertIdTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.tx.QueryContext(ctx, query, args...)
}
func (f *failingLastInsertIdTx) Commit() error {
	return f.tx.Commit()
}
func (f *failingLastInsertIdTx) Rollback() error {
	return f.tx.Rollback()
}

// failingLastInsertIdDB wraps *sql.DB and returns failingLastInsertIdTx.
type failingLastInsertIdDB struct {
	db *sql.DB
}

func (f *failingLastInsertIdDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	tx, err := f.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &failingLastInsertIdTx{tx: tx}, nil
}
func (f *failingLastInsertIdDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.db.ExecContext(ctx, query, args...)
}
func (f *failingLastInsertIdDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.db.QueryRowContext(ctx, query, args...)
}
func (f *failingLastInsertIdDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.db.QueryContext(ctx, query, args...)
}

// TestCreateUserLastInsertIdErrorReturns50002 verifies that when
// LastInsertId() fails on the CreateUser tx's INSERT INTO user_account,
// the response is 500 + 50002 and the tx rolls back (no user created).
func TestCreateUserLastInsertIdErrorReturns50002(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// Login with real DB to get owner token.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler0 := appauth.NewHandler(db, "test-jwt-secret-for-r7-lastinsertid", logger)
	mux0 := httpserver.New()
	mux0 = appauth.MountRoutes(mux0, handler0, db)
	srv0 := httptest.NewServer(logging.NewMiddleware(logger)(mux0))
	token := loginAndGetTokenWith(t, srv0, "owner1", "Pass1234")
	srv0.Close()

	failingDB := &failingLastInsertIdDB{db: db}

	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r7-lastinsertid", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	code, body, _ := doRequestWithToken(t, "POST", srv.URL+"/users", token,
		map[string]string{"username": "op1", "password": "Pass1234", "role": "OPERATOR"})

	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got body=%v", body)
	}

	// Verify no user was created (tx rolled back).
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM user_account WHERE username = 'op1'`).Scan(&count)
	if err != nil {
		t.Fatalf("query count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 users with username op1 (rollback), got %d", count)
	}
}

// ==================== r7: P1-3.3 failingScanDB — ListUsers Scan/rows.Err error ====================

// failingScanDB wraps *sql.DB. For the ListUsers query (SELECT ... FROM
// user_account WHERE deleted_at), it returns rows with incompatible column
// types ('text' for id), causing a Scan error when scanning into int64.
type failingScanDB struct {
	db *sql.DB
}

func (f *failingScanDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (appauth.Tx, error) {
	tx, err := f.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
func (f *failingScanDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.db.ExecContext(ctx, query, args...)
}
func (f *failingScanDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.db.QueryRowContext(ctx, query, args...)
}
func (f *failingScanDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if strings.Contains(query, "FROM user_account") && strings.Contains(query, "deleted_at IS NULL") {
		// Return rows with incompatible types: 'text' for id (expected int64).
		return f.db.QueryContext(ctx, `SELECT 'text' AS id, 'x' AS username, 'x' AS role, 'x' AS display_name, 'x' AS status, 'x' AS created_at LIMIT 1`)
	}
	return f.db.QueryContext(ctx, query, args...)
}

// TestListUsersScanErrorReturns50002 verifies that when the ListUsers query
// produces a Scan error (incompatible column types), the response is
// 500 + 50002, not 200 with partial/empty results.
func TestListUsersScanErrorReturns50002(t *testing.T) {
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

	createTestUser(t, db, "owner1", "Pass1234", "OWNER", "ACTIVE")

	// Login with real DB to get owner token.
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	handler0 := appauth.NewHandler(db, "test-jwt-secret-for-r7-scan", logger)
	mux0 := httpserver.New()
	mux0 = appauth.MountRoutes(mux0, handler0, db)
	srv0 := httptest.NewServer(logging.NewMiddleware(logger)(mux0))
	token := loginAndGetTokenWith(t, srv0, "owner1", "Pass1234")
	srv0.Close()

	failingDB := &failingScanDB{db: db}

	handler := appauth.NewHandler(failingDB, "test-jwt-secret-for-r7-scan", logger)
	mux := httpserver.New()
	mux = appauth.MountRoutes(mux, handler, db)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	code, body, _ := doRequestWithToken(t, "GET", srv.URL+"/users", token, nil)
	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(50002) {
		t.Fatalf("expected code 50002, got body=%v", body)
	}
}

// ==================== r7: P1-3.4 Real duplicate username → 40901 ====================

// TestCreateUserDuplicateUsernameReturns40901 verifies that creating a user
// with an already-existing username returns 409 + 40901 (not 50002).
func TestCreateUserDuplicateUsernameReturns40901(t *testing.T) {
	ts := newTestServer(t)
	createTestUser(t, ts.db, "owner1", "Pass1234", "OWNER", "ACTIVE")
	createTestUser(t, ts.db, "existingop", "Pass1234", "OPERATOR", "ACTIVE")
	token := loginAndGetToken(t, ts, "owner1", "Pass1234")

	code, body, _ := doRequestWithToken(t, "POST", ts.srv.URL+"/users", token,
		map[string]string{"username": "existingop", "password": "Pass1234", "role": "OPERATOR"})

	if code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate username, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(40901) {
		t.Fatalf("expected code 40901, got body=%v", body)
	}
}
