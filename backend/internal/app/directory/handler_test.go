package directory_test

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

	"github.com/prelove/zedu/backend/internal/app/directory"
	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/platform/logging"
	"github.com/prelove/zedu/backend/internal/repository"
)

const testJWTSecret = "directory-test-secret-must-be-32-chars"

// ---------- test harness ----------

type testServer struct {
	db  *sql.DB
	srv *httptest.Server
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "directory.db")
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	directory.MountRoutes(mux, directory.NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	t.Cleanup(func() { srv.Close(); db.Close() })
	return &testServer{db: db, srv: srv}
}

// newTestServerWithFaultDB mounts the directory handler with a custom
// repository.DB wrapper so tests can inject transaction failures.
func newTestServerWithFaultDB(t *testing.T, faultDB repository.DB) *testServer {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "directory_fault.db")
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	h := directory.NewHandler(faultDB, logger)
	mux := httpserver.New()
	// AuthMiddleware still uses the real *sql.DB for account-status checks.
	directory.MountRoutes(mux, h, db, testJWTSecret)
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

func auditCount(t *testing.T, db *sql.DB) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM operation_log`).Scan(&n); err != nil {
		t.Fatalf("count audit: %v", err)
	}
	return n
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

func auditDetailIsJSON(t *testing.T, db *sql.DB, action string) bool {
	t.Helper()
	var detail string
	err := db.QueryRow(`SELECT detail_json FROM operation_log WHERE action = ? ORDER BY id DESC LIMIT 1`, action).Scan(&detail)
	if err != nil {
		t.Fatalf("read audit detail: %v", err)
	}
	var v any
	return json.Unmarshal([]byte(detail), &v) == nil
}

func insertDomainTrackLevel(t *testing.T, db *sql.DB) (domainID, trackID, levelID int64) {
	t.Helper()
	res, err := db.Exec(`INSERT INTO course_domain (name, code, type) VALUES ('日语','JP','LANGUAGE')`)
	if err != nil {
		t.Fatalf("insert domain: %v", err)
	}
	domainID, _ = res.LastInsertId()
	res, err = db.Exec(`INSERT INTO course_track (domain_id, name, code) VALUES (?, 'JLPT','JLPT')`, domainID)
	if err != nil {
		t.Fatalf("insert track: %v", err)
	}
	trackID, _ = res.LastInsertId()
	res, err = db.Exec(`INSERT INTO course_level (track_id, name, code) VALUES (?, 'N5','N5')`, trackID)
	if err != nil {
		t.Fatalf("insert level: %v", err)
	}
	levelID, _ = res.LastInsertId()
	return
}

// ==================== Auth/RBAC ====================

func TestUnauthenticatedReturns40101(t *testing.T) {
	ts := newTestServer(t)
	status, body := req(t, "GET", ts.srv.URL+"/students", "", nil)
	if status != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", status)
	}
	if codeOf(t, body) != float64(httpserver.CodeUnauth) {
		t.Fatalf("expected 40101, got %v", body["code"])
	}
	if body["requestId"] == nil || body["requestId"] == "" {
		t.Fatalf("expected non-empty requestId: %v", body)
	}
}

func TestOwnerAndOperatorCanListStudents(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	opID := createUser(t, ts.db, "op", "OPERATOR")
	for _, tok := range []string{tokenFor(t, ownerID, "OWNER"), tokenFor(t, opID, "OPERATOR")} {
		status, body := req(t, "GET", ts.srv.URL+"/students", tok, nil)
		if status != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%v", status, body)
		}
		if codeOf(t, body) != 0 {
			t.Fatalf("expected code 0, got %v", body)
		}
	}
}

// ==================== Student ====================

func TestStudentWithoutEmailSaved(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	status, body := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "山田太郎"})
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%v", status, body)
	}
	if codeOf(t, body) != 0 {
		t.Fatalf("expected code 0, got %v", body)
	}
	d := dataMap(t, body)
	if d["name"] != "山田太郎" {
		t.Fatalf("expected name 山田太郎, got %v", d["name"])
	}
	if v, ok := d["email"]; ok && v != "" {
		t.Fatalf("expected empty email, got %v", d["email"])
	}
	if auditCountForAction(t, ts.db, "STUDENT_CREATE") != 1 {
		t.Fatalf("expected one audit row")
	}
	if !auditHasRequestID(t, ts.db, "STUDENT_CREATE") {
		t.Fatalf("audit row missing request_id")
	}
	if !auditDetailIsJSON(t, ts.db, "STUDENT_CREATE") {
		t.Fatalf("audit detail is not valid JSON")
	}
}

func TestDuplicateEmailCreateRejected40901(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "A", "email": "dup@example.com"})
	status, body := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "B", "email": "dup@example.com"})
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d", status)
	}
	if codeOf(t, body) != float64(httpserver.CodeConflict) {
		t.Fatalf("expected 40901, got %v", body["code"])
	}
	// No "bypass" action; only one student with that email exists.
	var n int
	ts.db.QueryRow(`SELECT COUNT(*) FROM student WHERE email='dup@example.com'`).Scan(&n)
	if n != 1 {
		t.Fatalf("expected exactly 1 student with dup email, got %d", n)
	}
}

func TestDuplicateEmailUpdateRejectedAndOriginalPreserved(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, body := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "A", "email": "a@example.com"})
	idA := int64(dataMap(t, body)["id"].(float64))
	req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "B", "email": "b@example.com"})
	status, body2 := req(t, "PATCH", fmt.Sprintf("%s/students/%d", ts.srv.URL, idA), tok, map[string]any{"email": "b@example.com"})
	if status != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%v", status, body2)
	}
	if codeOf(t, body2) != float64(httpserver.CodeConflict) {
		t.Fatalf("expected 40901, got %v", body2["code"])
	}
	// Original record preserved.
	_, body3 := req(t, "GET", fmt.Sprintf("%s/students/%d", ts.srv.URL, idA), tok, nil)
	if dataMap(t, body3)["email"] != "a@example.com" {
		t.Fatalf("original email not preserved: %v", body3)
	}
}

func TestConcurrentDuplicateEmailExactlyOneWinner(t *testing.T) {
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
			status, _ := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": fmt.Sprintf("S%d", i), "email": "race@example.com"})
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
	// Audit count equals the single winner; no partial audit.
	if auditCountForAction(t, ts.db, "STUDENT_CREATE") != 1 {
		t.Fatalf("expected exactly 1 audit row after race, got %d", auditCountForAction(t, ts.db, "STUDENT_CREATE"))
	}
}

func TestStudentUpdateMissingReturns40401(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	status, body := req(t, "PATCH", ts.srv.URL+"/students/999999", tok, map[string]any{"name": "X"})
	if status != http.StatusNotFound || codeOf(t, body) != float64(httpserver.CodeNotFound) {
		t.Fatalf("expected 404/40401, got %d %v", status, body)
	}
}

func TestStudentListPaginates(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	for i := 0; i < 5; i++ {
		req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": fmt.Sprintf("S%d", i)})
	}
	status, body := req(t, "GET", ts.srv.URL+"/students?page=1&pageSize=2", tok, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	d := dataMap(t, body)
	if int(d["total"].(float64)) != 5 {
		t.Fatalf("expected total 5, got %v", d["total"])
	}
	if int(d["pageSize"].(float64)) != 2 {
		t.Fatalf("expected pageSize 2, got %v", d["pageSize"])
	}
	items := d["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

// ==================== Parent ====================

func TestCreateParentAndCrossStudent40401(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, sb := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "S1"})
	studentA := int64(dataMap(t, sb)["id"].(float64))
	_, sb2 := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "S2"})
	studentB := int64(dataMap(t, sb2)["id"].(float64))

	// Create parent under studentA.
	status, body := req(t, "POST", fmt.Sprintf("%s/students/%d/parents", ts.srv.URL, studentA), tok, map[string]any{"name": "P1", "relationship": "FATHER"})
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%v", status, body)
	}
	parentID := int64(dataMap(t, body)["id"].(float64))

	// Cross-student PATCH must return 40401, not disclose the record.
	status, body = req(t, "PATCH", fmt.Sprintf("%s/students/%d/parents/%d", ts.srv.URL, studentB, parentID), tok, map[string]any{"name": "HACK"})
	if status != http.StatusNotFound || codeOf(t, body) != float64(httpserver.CodeNotFound) {
		t.Fatalf("expected 404/40401 for cross-student parent, got %d %v", status, body)
	}
	// No audit written for the failed attempt.
	if auditCountForAction(t, ts.db, "PARENT_UPDATE") != 0 {
		t.Fatalf("failed parent update must not produce audit")
	}
	// Original parent name unchanged.
	_, gb := req(t, "GET", fmt.Sprintf("%s/students/%d/parents", ts.srv.URL, studentA), tok, nil)
	items := dataMap(t, gb)["items"].([]any)
	if items[0].(map[string]any)["name"] != "P1" {
		t.Fatalf("parent name was changed by cross-student attempt: %v", items[0])
	}
}

func TestMultipleParentsForStudent(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, sb := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "S1"})
	studentID := int64(dataMap(t, sb)["id"].(float64))
	req(t, "POST", fmt.Sprintf("%s/students/%d/parents", ts.srv.URL, studentID), tok, map[string]any{"name": "Dad", "relationship": "FATHER"})
	req(t, "POST", fmt.Sprintf("%s/students/%d/parents", ts.srv.URL, studentID), tok, map[string]any{"name": "Mom", "relationship": "MOTHER"})
	_, body := req(t, "GET", fmt.Sprintf("%s/students/%d/parents", ts.srv.URL, studentID), tok, nil)
	items := dataMap(t, body)["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected 2 parents, got %d", len(items))
	}
}

// ==================== Teacher ====================

func TestCreateTeacherAndCapabilityUnique40901(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, tb := req(t, "POST", ts.srv.URL+"/teachers", tok, map[string]any{"name": "T1"})
	teacherID := int64(dataMap(t, tb)["id"].(float64))
	domainID, trackID, levelID := insertDomainTrackLevel(t, ts.db)

	payload := map[string]any{"domainId": domainID, "trackId": trackID, "levelId": levelID}
	status, body := req(t, "POST", fmt.Sprintf("%s/teachers/%d/capabilities", ts.srv.URL, teacherID), tok, payload)
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%v", status, body)
	}
	// Duplicate (teacher_id, track_id, level_id) -> 40901.
	status, body = req(t, "POST", fmt.Sprintf("%s/teachers/%d/capabilities", ts.srv.URL, teacherID), tok, payload)
	if status != http.StatusConflict || codeOf(t, body) != float64(httpserver.CodeConflict) {
		t.Fatalf("expected 409/40901 for duplicate capability, got %d %v", status, body)
	}
}

func TestEndCapabilityPreservesHistory(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, tb := req(t, "POST", ts.srv.URL+"/teachers", tok, map[string]any{"name": "T1"})
	teacherID := int64(dataMap(t, tb)["id"].(float64))
	domainID, trackID, levelID := insertDomainTrackLevel(t, ts.db)
	_, body := req(t, "POST", fmt.Sprintf("%s/teachers/%d/capabilities", ts.srv.URL, teacherID), tok, map[string]any{"domainId": domainID, "trackId": trackID, "levelId": levelID})
	capID := int64(dataMap(t, body)["id"].(float64))

	status, body := req(t, "PATCH", fmt.Sprintf("%s/teachers/%d/capabilities/%d", ts.srv.URL, teacherID, capID), tok, map[string]any{"status": "ENDED"})
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", status, body)
	}
	d := dataMap(t, body)
	if d["status"] != "ENDED" {
		t.Fatalf("expected ENDED, got %v", d["status"])
	}
	if d["effectiveTo"] == nil || d["effectiveTo"] == "" {
		t.Fatalf("expected effectiveTo set, got %v", d["effectiveTo"])
	}
	// Row still exists (history preserved, not deleted).
	var n int
	ts.db.QueryRow(`SELECT COUNT(*) FROM teacher_capability WHERE id = ?`, capID).Scan(&n)
	if n != 1 {
		t.Fatalf("expected capability row preserved, got count %d", n)
	}
}

func TestCapabilityHierarchyMismatch42201(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, tb := req(t, "POST", ts.srv.URL+"/teachers", tok, map[string]any{"name": "T1"})
	teacherID := int64(dataMap(t, tb)["id"].(float64))
	// Two unrelated domains/tracks/levels.
	d1, t1, l1 := insertDomainTrackLevel(t, ts.db)
	// second domain
	res, _ := ts.db.Exec(`INSERT INTO course_domain (name, code, type) VALUES ('数学','MATH','K12')`)
	d2, _ := res.LastInsertId()
	res, _ = ts.db.Exec(`INSERT INTO course_track (domain_id, name, code) VALUES (?, 'Algebra','ALG')`, d2)
	t2, _ := res.LastInsertId()
	res, _ = ts.db.Exec(`INSERT INTO course_level (track_id, name, code) VALUES (?, 'L1','L1')`, t2)
	l2, _ := res.LastInsertId()
	_ = d1
	_ = t1
	_ = l1
	// track t2 belongs to d2, but we claim domain d1 -> mismatch.
	status, body := req(t, "POST", fmt.Sprintf("%s/teachers/%d/capabilities", ts.srv.URL, teacherID), tok, map[string]any{"domainId": d1, "trackId": t2, "levelId": l2})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for hierarchy mismatch, got %d %v", status, body)
	}
}

// ==================== Availability ====================

func TestAvailabilityInvalidTimeRejected(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, tb := req(t, "POST", ts.srv.URL+"/teachers", tok, map[string]any{"name": "T1"})
	teacherID := int64(dataMap(t, tb)["id"].(float64))

	// end before start.
	status, body := req(t, "POST", fmt.Sprintf("%s/teachers/%d/availability", ts.srv.URL, teacherID), tok, map[string]any{"weekday": 1, "startTime": "20:00", "endTime": "10:00"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422 for end<start, got %d %v", status, body)
	}
	// bad weekday.
	status, body = req(t, "POST", fmt.Sprintf("%s/teachers/%d/availability", ts.srv.URL, teacherID), tok, map[string]any{"weekday": 9, "startTime": "10:00", "endTime": "11:00"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422 for bad weekday, got %d %v", status, body)
	}
	// bad time format.
	status, body = req(t, "POST", fmt.Sprintf("%s/teachers/%d/availability", ts.srv.URL, teacherID), tok, map[string]any{"weekday": 1, "startTime": "10", "endTime": "11:00"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422 for bad time, got %d %v", status, body)
	}
	// valid.
	status, body = req(t, "POST", fmt.Sprintf("%s/teachers/%d/availability", ts.srv.URL, teacherID), tok, map[string]any{"weekday": 1, "startTime": "10:00", "endTime": "11:00"})
	if status != http.StatusCreated {
		t.Fatalf("expected 201 for valid availability, got %d body=%v", status, body)
	}
}

// ==================== Audit / transaction fault injection ====================

// failingAuditTx wraps a real *sql.Tx and fails the operation_log INSERT.
type failingAuditTx struct {
	repository.Tx
	failedInsert atomic.Bool
}

func (f *failingAuditTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if strings.Contains(query, "operation_log") {
		f.failedInsert.Store(true)
		return nil, fmt.Errorf("simulated audit insert failure")
	}
	return f.Tx.ExecContext(ctx, query, args...)
}

// failingAuditDB wraps *sql.DB and returns failingAuditTx from BeginTx.
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

func TestAuditFailureRollsBackBusinessWrite(t *testing.T) {
	// Build a fault DB over a separate migrated database so the wrapper has a
	// real schema to transact against.
	dsn := "file:" + filepath.Join(t.TempDir(), "fault_audit.db")
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

	status, body := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "Rollback"})
	if status != http.StatusInternalServerError || codeOf(t, body) != float64(httpserver.CodeDatabase) {
		t.Fatalf("expected 500/50002 on audit failure, got %d %v", status, body)
	}
	// No student row should exist (business write rolled back with audit).
	var n int
	ts.db.QueryRow(`SELECT COUNT(*) FROM student WHERE name='Rollback'`).Scan(&n)
	if n != 0 {
		t.Fatalf("expected 0 students after rollback, got %d", n)
	}
	// No successful audit row.
	if auditCount(t, ts.db) != 0 {
		t.Fatalf("expected 0 audit rows after rollback, got %d", auditCount(t, ts.db))
	}
}

// ==================== Negative scope ====================

func TestNoLessonPaymentNotificationRoutes(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	for _, p := range []string{"/lessons", "/attendance", "/payments", "/notifications", "/payouts", "/backup"} {
		status, _ := req(t, "GET", ts.srv.URL+p, tok, nil)
		// Go 1.22 ServeMux returns 404 for unregistered routes.
		if status != http.StatusNotFound {
			t.Fatalf("route %s should not be registered, got %d", p, status)
		}
	}
}

func TestNoDeleteRoutesForStudents(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, sb := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "S1"})
	id := int64(dataMap(t, sb)["id"].(float64))
	status, _ := req(t, "DELETE", fmt.Sprintf("%s/students/%d", ts.srv.URL, id), tok, nil)
	// 405 (method not allowed) confirms the path exists for GET/PATCH but DELETE
	// is not registered; 404 would mean the path is entirely unknown. Both prove
	// no DELETE route is exposed.
	if status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
		t.Fatalf("DELETE /students/{id} should not be registered, got %d", status)
	}
}
