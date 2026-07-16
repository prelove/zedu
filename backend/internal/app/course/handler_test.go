package course_test

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
	"sync/atomic"
	"testing"
	"time"

	"github.com/prelove/zedu/backend/internal/app/course"
	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/platform/logging"
	"github.com/prelove/zedu/backend/internal/repository"
)

const testJWTSecret = "course-test-secret-must-be-32-chars"

type testServer struct {
	db  *sql.DB
	srv *httptest.Server
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "course.db")
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	course.MountRoutes(mux, course.NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	t.Cleanup(func() { srv.Close(); db.Close() })
	return &testServer{db: db, srv: srv}
}

func newTestServerWithFaultDB(t *testing.T, faultDB repository.DB) *testServer {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "course_fault.db")
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	h := course.NewHandler(faultDB, logger)
	mux := httpserver.New()
	course.MountRoutes(mux, h, db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	t.Cleanup(func() { srv.Close(); db.Close() })
	return &testServer{db: db, srv: srv}
}

func createUser(t *testing.T, db *sql.DB, username, role string) int64 {
	t.Helper()
	hash, err := auth.HashPassword("Pass1234")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	res, err := db.Exec(`INSERT INTO user_account (username, password_hash, role, display_name) VALUES (?, ?, ?, ?)`, username, hash, role, username)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func tokenFor(t *testing.T, userID int64, role string) string {
	t.Helper()
	tok, err := auth.SignAccessToken(testJWTSecret, userID, role, time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return tok
}

func req(t *testing.T, method, url, token string, body any) (int, map[string]any) {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	httpReq, err := http.NewRequest(method, url, r)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var parsed map[string]any
	_ = json.Unmarshal(data, &parsed)
	return resp.StatusCode, parsed
}

func codeOf(t *testing.T, body map[string]any) float64 {
	t.Helper()
	c, ok := body["code"]
	if !ok {
		t.Fatalf("response missing code: %v", body)
	}
	return c.(float64)
}

func dataMap(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	d, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("response missing data object: %v", body)
	}
	return d
}

func auditCountForAction(t *testing.T, db *sql.DB, action string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM operation_log WHERE action = ?`, action).Scan(&n); err != nil {
		t.Fatalf("count audit: %v", err)
	}
	return n
}

func auditHasRequestID(t *testing.T, db *sql.DB, action string) bool {
	t.Helper()
	var rid string
	err := db.QueryRow(`SELECT request_id FROM operation_log WHERE action = ? ORDER BY id DESC LIMIT 1`, action).Scan(&rid)
	if err != nil {
		t.Fatalf("read audit request_id: %v", err)
	}
	return rid != "" && rid != "unknown"
}

// ==================== Auth/RBAC ====================

func TestUnauthenticatedReturns40101(t *testing.T) {
	ts := newTestServer(t)
	status, body := req(t, "GET", ts.srv.URL+"/course-domains", "", nil)
	if status != http.StatusUnauthorized || codeOf(t, body) != float64(httpserver.CodeUnauth) {
		t.Fatalf("expected 401/40101, got %d %v", status, body)
	}
}

func TestOwnerAndOperatorCanAccessDictionary(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	opID := createUser(t, ts.db, "op", "OPERATOR")
	for _, tok := range []string{tokenFor(t, ownerID, "OWNER"), tokenFor(t, opID, "OPERATOR")} {
		status, body := req(t, "GET", ts.srv.URL+"/course-domains", tok, nil)
		if status != http.StatusOK || codeOf(t, body) != 0 {
			t.Fatalf("expected 200/0, got %d %v", status, body)
		}
	}
}

// ==================== Domain ====================

func TestCreateDomainAndDuplicateCode40901(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	status, body := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "日语", "code": "JP", "type": "LANGUAGE"})
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%v", status, body)
	}
	if dataMap(t, body)["code"] != "JP" {
		t.Fatalf("expected code JP, got %v", dataMap(t, body)["code"])
	}
	if auditCountForAction(t, ts.db, "DOMAIN_CREATE") != 1 || !auditHasRequestID(t, ts.db, "DOMAIN_CREATE") {
		t.Fatalf("expected audited DOMAIN_CREATE with request_id")
	}
	// Duplicate code -> 40901.
	status, body = req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "Japanese", "code": "JP", "type": "LANGUAGE"})
	if status != http.StatusConflict || codeOf(t, body) != float64(httpserver.CodeConflict) {
		t.Fatalf("expected 409/40901 for duplicate domain code, got %d %v", status, body)
	}
}

func TestDomainInvalidType42201(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	status, body := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "X", "code": "X", "type": "BOGUS"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for bad type, got %d %v", status, body)
	}
}

// ==================== Track ====================

func TestCreateTrackAndDuplicateCodeWithinDomain40901(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, db := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "日语", "code": "JP", "type": "LANGUAGE"})
	domainID := int64(dataMap(t, db)["id"].(float64))

	status, body := req(t, "POST", ts.srv.URL+"/tracks", tok, map[string]any{"domainId": domainID, "name": "JLPT", "code": "JLPT"})
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%v", status, body)
	}
	// Same code within same domain -> 40901.
	status, body = req(t, "POST", ts.srv.URL+"/tracks", tok, map[string]any{"domainId": domainID, "name": "JLPT2", "code": "JLPT"})
	if status != http.StatusConflict || codeOf(t, body) != float64(httpserver.CodeConflict) {
		t.Fatalf("expected 409/40901 for duplicate track code in domain, got %d %v", status, body)
	}
}

func TestCreateTrackMissingDomain40401(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	status, body := req(t, "POST", ts.srv.URL+"/tracks", tok, map[string]any{"domainId": 999999, "name": "T", "code": "T"})
	if status != http.StatusNotFound || codeOf(t, body) != float64(httpserver.CodeNotFound) {
		t.Fatalf("expected 404/40401 for missing domain, got %d %v", status, body)
	}
}

// ==================== Level ====================

func TestCreateLevelVerifiesHierarchy(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, db1 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "日语", "code": "JP", "type": "LANGUAGE"})
	domainID := int64(dataMap(t, db1)["id"].(float64))
	_, tb := req(t, "POST", ts.srv.URL+"/tracks", tok, map[string]any{"domainId": domainID, "name": "JLPT", "code": "JLPT"})
	trackID := int64(dataMap(t, tb)["id"].(float64))

	status, body := req(t, "POST", ts.srv.URL+"/levels", tok, map[string]any{"trackId": trackID, "name": "N5", "code": "N5"})
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%v", status, body)
	}
	// Missing track -> 40401.
	status, body = req(t, "POST", ts.srv.URL+"/levels", tok, map[string]any{"trackId": 999999, "name": "X", "code": "X"})
	if status != http.StatusNotFound || codeOf(t, body) != float64(httpserver.CodeNotFound) {
		t.Fatalf("expected 404/40401 for missing track, got %d %v", status, body)
	}
}

// ==================== Capability Tag ====================

func TestCreateTagAndDuplicateCodeWithinDomain40901(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, db1 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "日语", "code": "JP", "type": "LANGUAGE"})
	domainID := int64(dataMap(t, db1)["id"].(float64))

	status, body := req(t, "POST", ts.srv.URL+"/capability-tags", tok, map[string]any{"domainId": domainID, "name": "会话", "code": "SPEAK"})
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%v", status, body)
	}
	status, body = req(t, "POST", ts.srv.URL+"/capability-tags", tok, map[string]any{"domainId": domainID, "name": "会话2", "code": "SPEAK"})
	if status != http.StatusConflict || codeOf(t, body) != float64(httpserver.CodeConflict) {
		t.Fatalf("expected 409/40901 for duplicate tag code in domain, got %d %v", status, body)
	}
}

// ==================== Disable preserves references ====================

func TestDisableDomainPreservesReferences(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, db1 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "日语", "code": "JP", "type": "LANGUAGE"})
	domainID := int64(dataMap(t, db1)["id"].(float64))
	req(t, "POST", ts.srv.URL+"/tracks", tok, map[string]any{"domainId": domainID, "name": "JLPT", "code": "JLPT"})

	// Disable the domain (PATCH enabled=false). No DELETE route exists.
	status, body := req(t, "PATCH", fmt.Sprintf("%s/course-domains/%d", ts.srv.URL, domainID), tok, map[string]any{"enabled": false})
	if status != http.StatusOK {
		t.Fatalf("expected 200 on disable, got %d body=%v", status, body)
	}
	// Track still exists (relation preserved).
	var n int
	ts.db.QueryRow(`SELECT COUNT(*) FROM course_track WHERE domain_id = ?`, domainID).Scan(&n)
	if n != 1 {
		t.Fatalf("expected track preserved after domain disable, got %d", n)
	}
}

// ==================== No DELETE routes ====================

func TestNoDeleteRoutesForDictionary(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, db1 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "日语", "code": "JP", "type": "LANGUAGE"})
	id := int64(dataMap(t, db1)["id"].(float64))
	status, _ := req(t, "DELETE", fmt.Sprintf("%s/course-domains/%d", ts.srv.URL, id), tok, nil)
	if status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
		t.Fatalf("DELETE /course-domains/{id} should not be registered, got %d", status)
	}
}

// ==================== Audit / transaction fault injection ====================

type failingAuditTx struct {
	repository.Tx
}

func (f *failingAuditTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if strings.Contains(query, "operation_log") {
		return nil, fmt.Errorf("simulated audit insert failure")
	}
	return f.Tx.ExecContext(ctx, query, args...)
}

type failingAuditDB struct {
	*sql.DB
}

func (f *failingAuditDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (repository.Tx, error) {
	tx, err := f.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &failingAuditTx{Tx: tx}, nil
}
func (f *failingAuditDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.DB.ExecContext(ctx, query, args...)
}
func (f *failingAuditDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.DB.QueryRowContext(ctx, query, args...)
}
func (f *failingAuditDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.DB.QueryContext(ctx, query, args...)
}

func TestAuditFailureRollsBackDomainCreate(t *testing.T) {
	dsn := "file:" + filepath.Join(t.TempDir(), "course_fault_audit.db")
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	faultDB := &failingAuditDB{DB: db}
	ts := newTestServerWithFaultDB(t, faultDB)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")

	status, body := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "X", "code": "X", "type": "OTHER"})
	if status != http.StatusInternalServerError || codeOf(t, body) != float64(httpserver.CodeDatabase) {
		t.Fatalf("expected 500/50002 on audit failure, got %d %v", status, body)
	}
	var n int
	ts.db.QueryRow(`SELECT COUNT(*) FROM course_domain WHERE code='X'`).Scan(&n)
	if n != 0 {
		t.Fatalf("expected 0 domains after rollback, got %d", n)
	}
	if auditCountForAction(t, ts.db, "DOMAIN_CREATE") != 0 {
		t.Fatalf("expected 0 audit rows after rollback")
	}
}

// ==================== Concurrent duplicate code ====================

func TestConcurrentDuplicateDomainCodeExactlyOneWinner(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	const n = 8
	var wg sync.WaitGroup
	var success, conflict int64
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			status, _ := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": fmt.Sprintf("D%d", i), "code": "RACE", "type": "OTHER"})
			switch status {
			case http.StatusCreated:
				atomic.AddInt64(&success, 1)
			case http.StatusConflict:
				atomic.AddInt64(&conflict, 1)
			}
		}(i)
	}
	wg.Wait()
	if success != 1 {
		t.Fatalf("expected exactly 1 success, got %d (conflicts=%d)", success, conflict)
	}
	if conflict != n-1 {
		t.Fatalf("expected %d conflicts, got %d", n-1, conflict)
	}
}

// ==================== Negative scope ====================

func TestNoLessonPaymentRoutes(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	for _, p := range []string{"/lessons", "/payments", "/notifications", "/payouts"} {
		status, _ := req(t, "GET", ts.srv.URL+p, tok, nil)
		if status != http.StatusNotFound {
			t.Fatalf("route %s should not be registered, got %d", p, status)
		}
	}
}
