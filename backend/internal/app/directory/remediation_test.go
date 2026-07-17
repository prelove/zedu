package directory_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

// openMigrated opens a temp SQLite DB and runs all migrations up.
func openMigrated(dsn string) (*sql.DB, error) {
	db, err := database.Open(dsn)
	if err != nil {
		return nil, err
	}
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// ==================== P1-1: Nested lists must return pagination envelope ====================

func TestListParentsReturnsPaginationEnvelope(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, sb := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "S1"})
	studentID := int64(dataMap(t, sb)["id"].(float64))
	req(t, "POST", fmt.Sprintf("%s/students/%d/parents", ts.srv.URL, studentID), tok, map[string]any{"name": "P1", "relationship": "FATHER"})
	req(t, "POST", fmt.Sprintf("%s/students/%d/parents", ts.srv.URL, studentID), tok, map[string]any{"name": "P2", "relationship": "MOTHER"})

	status, body := req(t, "GET", fmt.Sprintf("%s/students/%d/parents?page=1&pageSize=1", ts.srv.URL, studentID), tok, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", status, body)
	}
	d := dataMap(t, body)
	// Must have items/page/pageSize/total, not a bare array.
	if d["items"] == nil {
		t.Fatalf("expected items field in pagination envelope, got data=%v", d)
	}
	if d["page"] == nil || int(d["page"].(float64)) != 1 {
		t.Fatalf("expected page=1, got %v", d["page"])
	}
	if d["pageSize"] == nil || int(d["pageSize"].(float64)) != 1 {
		t.Fatalf("expected pageSize=1, got %v", d["pageSize"])
	}
	if d["total"] == nil || int(d["total"].(float64)) != 2 {
		t.Fatalf("expected total=2, got %v", d["total"])
	}
	items, ok := d["items"].([]any)
	if !ok {
		t.Fatalf("expected items to be array, got %T", d["items"])
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item on page, got %d", len(items))
	}
}

func TestListParentsEmptyReturnsEmptyArrayNotNull(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, sb := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "S1"})
	studentID := int64(dataMap(t, sb)["id"].(float64))

	_, body := req(t, "GET", fmt.Sprintf("%s/students/%d/parents", ts.srv.URL, studentID), tok, nil)
	d := dataMap(t, body)
	items, ok := d["items"].([]any)
	if !ok {
		t.Fatalf("expected items to be array even when empty, got %T: %v", d["items"], d["items"])
	}
	if items == nil {
		t.Fatalf("expected items to be non-nil empty array, got nil")
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestListCapabilitiesReturnsPaginationEnvelope(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, tb := req(t, "POST", ts.srv.URL+"/teachers", tok, map[string]any{"name": "T1"})
	teacherID := int64(dataMap(t, tb)["id"].(float64))
	domainID, trackID, levelID := insertDomainTrackLevel(t, ts.db)
	req(t, "POST", fmt.Sprintf("%s/teachers/%d/capabilities", ts.srv.URL, teacherID), tok, map[string]any{"domainId": domainID, "trackId": trackID, "levelId": levelID})

	_, body := req(t, "GET", fmt.Sprintf("%s/teachers/%d/capabilities", ts.srv.URL, teacherID), tok, nil)
	d := dataMap(t, body)
	if d["items"] == nil {
		t.Fatalf("expected items field, got %v", d)
	}
	if d["total"] == nil || int(d["total"].(float64)) != 1 {
		t.Fatalf("expected total=1, got %v", d["total"])
	}
}

func TestListAvailabilityReturnsPaginationEnvelope(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, tb := req(t, "POST", ts.srv.URL+"/teachers", tok, map[string]any{"name": "T1"})
	teacherID := int64(dataMap(t, tb)["id"].(float64))
	req(t, "POST", fmt.Sprintf("%s/teachers/%d/availability", ts.srv.URL, teacherID), tok, map[string]any{"weekday": 1, "startTime": "10:00", "endTime": "11:00"})

	_, body := req(t, "GET", fmt.Sprintf("%s/teachers/%d/availability", ts.srv.URL, teacherID), tok, nil)
	d := dataMap(t, body)
	if d["items"] == nil {
		t.Fatalf("expected items field, got %v", d)
	}
	if d["total"] == nil || int(d["total"].(float64)) != 1 {
		t.Fatalf("expected total=1, got %v", d["total"])
	}
}

// ==================== P1-6: Rollback failure must map to 50002 (directory) ====================

// failingRollbackTx wraps a real *sql.Tx and fails Rollback.
type failingRollbackTx struct {
	repository.Tx
	rolledBack bool
}

func (f *failingRollbackTx) Rollback() error {
	f.rolledBack = true
	return fmt.Errorf("simulated rollback failure")
}

// failingRollbackDB wraps *sql.DB and returns failingRollbackTx from BeginTx.
// It also injects a validation error (via the business function) so that
// Rollback is called; the Rollback itself fails.
type failingRollbackDB struct {
	*sql.DB
}

func (f *failingRollbackDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (repository.Tx, error) {
	tx, err := f.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &failingRollbackTx{Tx: tx}, nil
}
func (f *failingRollbackDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.DB.ExecContext(ctx, query, args...)
}
func (f *failingRollbackDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.DB.QueryRowContext(ctx, query, args...)
}
func (f *failingRollbackDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.DB.QueryContext(ctx, query, args...)
}

func TestRollbackFailureMapsTo50002Directory(t *testing.T) {
	// Use a fault DB that fails Rollback. The business function will return a
	// conflict error (duplicate email), and the Rollback failure must surface as
	// 50002, not 40901.
	dsn := "file:" + t.TempDir() + "/rb_dir.db"
	db, err := openMigrated(dsn)
	if err != nil {
		t.Fatalf("open+migrate: %v", err)
	}
	defer db.Close()
	// Seed a student with email to cause conflict.
	_, err = db.Exec(`INSERT INTO student (name, email, status) VALUES ('A', 'dup@example.com', 'ACTIVE')`)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	faultDB := &failingRollbackDB{DB: db}
	ts := newTestServerWithFaultDB(t, faultDB)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")

	status, body := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "B", "email": "dup@example.com"})
	// Must be 500/50002 (rollback failure), not 409/40901 (the business conflict).
	if status != http.StatusInternalServerError || codeOf(t, body) != float64(httpserver.CodeDatabase) {
		t.Fatalf("expected 500/50002 on rollback failure, got %d %v", status, body)
	}
	// Must not leak SQLite text.
	msg, _ := body["message"].(string)
	if strings.Contains(strings.ToLower(msg), "sqlite") {
		t.Fatalf("response leaked sqlite text: %s", msg)
	}
}

// ==================== P2-1: UpdateAvailability must merge existing effective range ====================

func TestUpdateAvailabilityEffectiveFromAfterExistingTo(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, tb := req(t, "POST", ts.srv.URL+"/teachers", tok, map[string]any{"name": "T1"})
	teacherID := int64(dataMap(t, tb)["id"].(float64))
	// Create with effective range 2024-01-01 to 2024-06-30.
	_, body := req(t, "POST", fmt.Sprintf("%s/teachers/%d/availability", ts.srv.URL, teacherID), tok,
		map[string]any{"weekday": 1, "startTime": "10:00", "endTime": "11:00", "effectiveFrom": "2024-01-01", "effectiveTo": "2024-06-30"})
	availID := int64(dataMap(t, body)["id"].(float64))

	// PATCH only effective_from to 2024-07-01 (after existing effective_to).
	// Must merge with existing effective_to=2024-06-30 and reject 42201.
	status, body2 := req(t, "PATCH", fmt.Sprintf("%s/teachers/%d/availability/%d", ts.srv.URL, teacherID, availID), tok,
		map[string]any{"effectiveFrom": "2024-07-01"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body2) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for effective_from after existing effective_to, got %d %v", status, body2)
	}
	// Original record unchanged.
	_, gb := req(t, "GET", fmt.Sprintf("%s/teachers/%d/availability", ts.srv.URL, teacherID), tok, nil)
	items := dataMap(t, gb)["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("expected 1 availability slot, got %d", len(items))
	}
}

// ==================== P2-2: Empty body PATCH returns 42201 ====================

func TestPatchStudentEmptyBody42201(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, sb := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "S1"})
	studentID := int64(dataMap(t, sb)["id"].(float64))

	// Empty JSON object.
	status, body := req(t, "PATCH", fmt.Sprintf("%s/students/%d", ts.srv.URL, studentID), tok, map[string]any{})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for empty body PATCH, got %d %v", status, body)
	}
	// No audit written.
	if auditCountForAction(t, ts.db, "STUDENT_UPDATE") != 0 {
		t.Fatalf("expected 0 audit rows for empty PATCH")
	}
}

func TestPatchStudentRawEmptyBody42201(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, sb := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "S1"})
	studentID := int64(dataMap(t, sb)["id"].(float64))

	// Raw empty body (not even {}).
	httpReq, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/students/%d", ts.srv.URL, studentID), bytes.NewReader([]byte{}))
	httpReq.Header.Set("Authorization", "Bearer "+tok)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for raw empty body, got %d", resp.StatusCode)
	}
}
