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
	mux = appauth.MountRoutes(mux, handler)
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

	// Refresh — should return new access token and new refresh cookie.
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
		t.Fatalf("refresh token must rotate — new cookie same as old")
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

	// Replay old cookie — must fail.
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

	// Operator tries to list users — Owner-only.
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
	mux = appauth.MountRoutes(mux, handler)
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
	mux = appauth.MountRoutes(mux, handler)
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
	mux = appauth.MountRoutes(mux, handler)
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
	mux = appauth.MountRoutes(mux, handler)
	wrapped := logging.NewMiddleware(logger)(mux)
	srv := httptest.NewServer(wrapped)
	defer srv.Close()

	// Login with wrong password. The SELECT succeeds (not an UPDATE),
	// password verification fails, then the atomic UPDATE ... RETURNING
	// fails because failingDB intercepts it.
	code, body, _ := doRequest(t, "POST", srv.URL+"/auth/login",
		map[string]string{"username": "victim", "password": "wrongpass1"}, nil)

	// Must NOT be 40102 (LOGIN_FAILED) — that would mask the DB error.
	if code == http.StatusUnauthorized && body != nil && body["code"] == float64(40102) {
		t.Fatalf("must not return 40102 when fail-count UPDATE fails; got code=40102 body=%v", body)
	}

	// Must be 500 with code 42201 (internal error).
	if code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when fail-count UPDATE fails, got %d body=%v", code, body)
	}
	if body == nil || body["code"] != float64(42201) {
		t.Fatalf("expected code 42201, got body=%v", body)
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

func (f *failingUpdateDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return f.db.BeginTx(ctx, opts)
}

func (f *failingUpdateDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.db.ExecContext(ctx, query, args...)
}

func (f *failingUpdateDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if strings.Contains(strings.ToUpper(query), "UPDATE") {
		// Return a Row from a closed DB — Scan will fail with "sql: database is closed".
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
	mux = appauth.MountRoutes(mux, handler)
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

	// Call refresh with the real cookie — the DB query will fail.
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
