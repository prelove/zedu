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

	// Five failed login attempts.
	for i := 0; i < 5; i++ {
		code, body, _ := doRequest(t, "POST", ts.srv.URL+"/auth/login",
			map[string]string{"username": "owner1", "password": "wrong"}, nil)
		if code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", i+1, code)
		}
		if body["code"] != float64(40102) {
			t.Fatalf("attempt %d: expected code 40102, got %v", i+1, body["code"])
		}
	}

	// Sixth attempt must be locked.
	code, body, _ := doRequest(t, "POST", ts.srv.URL+"/auth/login",
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
