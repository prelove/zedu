package course_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

// ==================== helpers shared with handler_test.go ====================

// seedStudentAndCourse creates a student + domain/track/level and returns the
// IDs needed for enrollment tests.
func seedStudentAndCourse(t *testing.T, ts *testServer) (studentID, domainID, trackID, levelID int64) {
	t.Helper()
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, sb := req(t, "POST", ts.srv.URL+"/students", tok, map[string]any{"name": "S1"})
	studentID = int64(dataMap(t, sb)["id"].(float64))
	_, db1 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "日语", "code": "JP", "type": "LANGUAGE"})
	domainID = int64(dataMap(t, db1)["id"].(float64))
	_, tb := req(t, "POST", ts.srv.URL+"/tracks", tok, map[string]any{"domainId": domainID, "name": "JLPT", "code": "JLPT"})
	trackID = int64(dataMap(t, tb)["id"].(float64))
	_, lb := req(t, "POST", ts.srv.URL+"/levels", tok, map[string]any{"trackId": trackID, "name": "N5", "code": "N5"})
	levelID = int64(dataMap(t, lb)["id"].(float64))
	// Create a teacher.
	_, tb2 := req(t, "POST", ts.srv.URL+"/teachers", tok, map[string]any{"name": "T1"})
	_ = int64(dataMap(t, tb2)["id"].(float64))
	return
}

func seedTeacher(t *testing.T, ts *testServer, name string) int64 {
	t.Helper()
	ownerID := createUser(t, ts.db, "owner_"+name, "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")
	_, tb := req(t, "POST", ts.srv.URL+"/teachers", tok, map[string]any{"name": name})
	return int64(dataMap(t, tb)["id"].(float64))
}

func ownerToken(t *testing.T, ts *testServer) string {
	t.Helper()
	ownerID := createUser(t, ts.db, "owner_enr", "OWNER")
	return tokenFor(t, ownerID, "OWNER")
}

func createEnrollment(t *testing.T, ts *testServer, tok string, studentID, domainID, trackID, levelID int64) int64 {
	t.Helper()
	_, body := req(t, "POST", fmt.Sprintf("%s/students/%d/enrollments", ts.srv.URL, studentID), tok,
		map[string]any{"domainId": domainID, "trackId": trackID, "currentLevelId": levelID, "enrollmentType": "ONE_TO_ONE"})
	return int64(dataMap(t, body)["id"].(float64))
}

// ==================== Enrollment ====================

func TestCreateEnrollmentVerifiesHierarchy(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)

	// Valid enrollment.
	status, body := req(t, "POST", fmt.Sprintf("%s/students/%d/enrollments", ts.srv.URL, studentID), tok,
		map[string]any{"domainId": domainID, "trackId": trackID, "currentLevelId": levelID, "enrollmentType": "ONE_TO_ONE"})
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%v", status, body)
	}
	if dataMap(t, body)["status"] != "ACTIVE" {
		t.Fatalf("expected default status ACTIVE, got %v", dataMap(t, body)["status"])
	}
	if auditCountForAction(t, ts.db, "ENROLLMENT_CREATE") != 1 {
		t.Fatalf("expected 1 audit row for ENROLLMENT_CREATE")
	}

	// Hierarchy mismatch: track doesn't belong to a different domain.
	_, db2 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "数学", "code": "MATH", "type": "K12"})
	otherDomainID := int64(dataMap(t, db2)["id"].(float64))
	status, body = req(t, "POST", fmt.Sprintf("%s/students/%d/enrollments", ts.srv.URL, studentID), tok,
		map[string]any{"domainId": otherDomainID, "trackId": trackID, "enrollmentType": "ONE_TO_ONE"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for hierarchy mismatch, got %d %v", status, body)
	}
}

func TestCreateEnrollmentMissingStudent42201(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	_, _, trackID, levelID := seedStudentAndCourse(t, ts)
	// Use a non-existent student.
	status, body := req(t, "POST", ts.srv.URL+"/students/999999/enrollments", tok,
		map[string]any{"domainId": 1, "trackId": trackID, "currentLevelId": levelID, "enrollmentType": "ONE_TO_ONE"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for missing student, got %d %v", status, body)
	}
}

func TestEnrollmentStateMachine(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	// ACTIVE -> PAUSED (valid).
	status, body := req(t, "PATCH", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok, map[string]any{"status": "PAUSED"})
	if status != http.StatusOK || dataMap(t, body)["status"] != "PAUSED" {
		t.Fatalf("expected PAUSED, got %d %v", status, body)
	}

	// PAUSED -> ACTIVE (valid).
	status, body = req(t, "PATCH", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok, map[string]any{"status": "ACTIVE"})
	if status != http.StatusOK || dataMap(t, body)["status"] != "ACTIVE" {
		t.Fatalf("expected ACTIVE, got %d %v", status, body)
	}

	// ACTIVE -> COMPLETED (valid terminal).
	status, body = req(t, "PATCH", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok, map[string]any{"status": "COMPLETED"})
	if status != http.StatusOK || dataMap(t, body)["status"] != "COMPLETED" {
		t.Fatalf("expected COMPLETED, got %d %v", status, body)
	}

	// COMPLETED -> ACTIVE (invalid: terminal state cannot be restored).
	status, body = req(t, "PATCH", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok, map[string]any{"status": "ACTIVE"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for terminal->active, got %d %v", status, body)
	}
}

func TestEnrollmentUpdateMissing40401(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	status, body := req(t, "PATCH", ts.srv.URL+"/enrollments/999999", tok, map[string]any{"status": "PAUSED"})
	if status != http.StatusNotFound || codeOf(t, body) != float64(httpserver.CodeNotFound) {
		t.Fatalf("expected 404/40401, got %d %v", status, body)
	}
}

func TestListEnrollmentsForStudent(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	status, body := req(t, "GET", fmt.Sprintf("%s/students/%d/enrollments", ts.srv.URL, studentID), tok, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	items := body["data"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected 2 enrollments, got %d", len(items))
	}
}

// ==================== Assignment ====================

func TestCreateAssignmentAndAtomicReplacement(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	teacher1 := seedTeacher(t, ts, "T1")
	teacher2 := seedTeacher(t, ts, "T2")

	// First assignment.
	status, body := req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
		map[string]any{"teacherId": teacher1, "roleType": "MAIN"})
	if status != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%v", status, body)
	}
	assign1 := int64(dataMap(t, body)["id"].(float64))
	if dataMap(t, body)["status"] != "ACTIVE" {
		t.Fatalf("expected ACTIVE, got %v", dataMap(t, body)["status"])
	}

	// Verify only one ACTIVE assignment.
	n := countActiveAssignments(t, ts.db, enrID)
	if n != 1 {
		t.Fatalf("expected 1 ACTIVE assignment, got %d", n)
	}

	// Replace teacher: create a new assignment -> old one ends, new one ACTIVE.
	status, body = req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
		map[string]any{"teacherId": teacher2, "roleType": "MAIN", "reason": "REPLACE"})
	if status != http.StatusCreated {
		t.Fatalf("expected 201 on replacement, got %d body=%v", status, body)
	}
	assign2 := int64(dataMap(t, body)["id"].(float64))

	// Still exactly one ACTIVE.
	n = countActiveAssignments(t, ts.db, enrID)
	if n != 1 {
		t.Fatalf("expected 1 ACTIVE after replacement, got %d", n)
	}

	// Old assignment is ENDED.
	_, oldBody := req(t, "GET", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok, nil)
	items := oldBody["data"].([]any)
	var oldAssign, newAssign map[string]any
	for _, item := range items {
		m := item.(map[string]any)
		if int64(m["id"].(float64)) == assign1 {
			oldAssign = m
		}
		if int64(m["id"].(float64)) == assign2 {
			newAssign = m
		}
	}
	if oldAssign == nil || oldAssign["status"] != "ENDED" {
		t.Fatalf("expected old assignment ENDED, got %v", oldAssign)
	}
	if newAssign == nil || newAssign["status"] != "ACTIVE" {
		t.Fatalf("expected new assignment ACTIVE, got %v", newAssign)
	}
	if oldAssign["endDate"] == nil || oldAssign["endDate"] == "" {
		t.Fatalf("expected old assignment endDate set, got %v", oldAssign["endDate"])
	}

	// Audit: one END + two CREATEs.
	if auditCountForAction(t, ts.db, "ASSIGNMENT_END") != 1 {
		t.Fatalf("expected 1 ASSIGNMENT_END audit")
	}
	if auditCountForAction(t, ts.db, "ASSIGNMENT_CREATE") != 2 {
		t.Fatalf("expected 2 ASSIGNMENT_CREATE audits")
	}
}

func TestCreateAssignmentNonActiveEnrollment42201(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	teacher1 := seedTeacher(t, ts, "T1")

	// Pause the enrollment.
	req(t, "PATCH", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok, map[string]any{"status": "PAUSED"})

	// Cannot assign to a PAUSED enrollment.
	status, body := req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
		map[string]any{"teacherId": teacher1, "roleType": "MAIN"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for assignment on non-active enrollment, got %d %v", status, body)
	}
}

func TestCreateAssignmentInactiveTeacher42201(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	// Create a teacher and pause it.
	ownerID := createUser(t, ts.db, "owner_t", "OWNER")
	ownerTok := tokenFor(t, ownerID, "OWNER")
	_, tb := req(t, "POST", ts.srv.URL+"/teachers", ownerTok, map[string]any{"name": "TP"})
	teacherID := int64(dataMap(t, tb)["id"].(float64))
	req(t, "PATCH", fmt.Sprintf("%s/teachers/%d", ts.srv.URL, teacherID), ownerTok, map[string]any{"status": "PAUSED"})

	status, body := req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
		map[string]any{"teacherId": teacherID, "roleType": "MAIN"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for inactive teacher, got %d %v", status, body)
	}
}

func TestEndAssignmentById(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	teacher1 := seedTeacher(t, ts, "T1")

	_, body := req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
		map[string]any{"teacherId": teacher1, "roleType": "MAIN"})
	assignID := int64(dataMap(t, body)["id"].(float64))

	// End the assignment.
	status, body := req(t, "POST", fmt.Sprintf("%s/assignments/%d/end", ts.srv.URL, assignID), tok,
		map[string]any{"reason": "STUDENT_LEAVE"})
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", status, body)
	}
	if dataMap(t, body)["status"] != "ENDED" {
		t.Fatalf("expected ENDED, got %v", dataMap(t, body)["status"])
	}
	if dataMap(t, body)["endDate"] == nil || dataMap(t, body)["endDate"] == "" {
		t.Fatalf("expected endDate set, got %v", dataMap(t, body)["endDate"])
	}

	// No ACTIVE assignment left.
	n := countActiveAssignments(t, ts.db, enrID)
	if n != 0 {
		t.Fatalf("expected 0 ACTIVE after end, got %d", n)
	}

	// Ending again -> 42201 (already ENDED).
	status, body = req(t, "POST", fmt.Sprintf("%s/assignments/%d/end", ts.srv.URL, assignID), tok,
		map[string]any{"reason": "AGAIN"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for double-end, got %d %v", status, body)
	}
}

func TestEndAssignmentMissing40401(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	status, body := req(t, "POST", ts.srv.URL+"/assignments/999999/end", tok, map[string]any{"reason": "X"})
	if status != http.StatusNotFound || codeOf(t, body) != float64(httpserver.CodeNotFound) {
		t.Fatalf("expected 404/40401, got %d %v", status, body)
	}
}

func TestListAssignmentsForEnrollment(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	teacher1 := seedTeacher(t, ts, "T1")
	teacher2 := seedTeacher(t, ts, "T2")

	// Create first, then replace with second -> 2 assignments total (1 ENDED, 1 ACTIVE).
	req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
		map[string]any{"teacherId": teacher1, "roleType": "MAIN"})
	req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
		map[string]any{"teacherId": teacher2, "roleType": "MAIN", "reason": "REPLACE"})

	status, body := req(t, "GET", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	items := body["data"].([]any)
	if len(items) != 2 {
		t.Fatalf("expected 2 assignments (history preserved), got %d", len(items))
	}
}

// ==================== Concurrency: atomic replacement ====================

func TestConcurrentAssignmentCreateExactlyOneActive(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	teacher1 := seedTeacher(t, ts, "T1")

	const n = 8
	var wg sync.WaitGroup
	var success, conflict int64
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			status, _ := req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
				map[string]any{"teacherId": teacher1, "roleType": "MAIN", "reason": "CONCURRENT"})
			switch status {
			case http.StatusCreated:
				atomic.AddInt64(&success, 1)
			case http.StatusConflict:
				atomic.AddInt64(&conflict, 1)
			}
		}()
	}
	wg.Wait()
	// All concurrent creates should succeed (each replaces the previous within
	// its own transaction), but exactly one ACTIVE assignment remains.
	if success == 0 {
		t.Fatalf("expected at least 1 success, got 0 (conflicts=%d)", conflict)
	}
	active := countActiveAssignments(t, ts.db, enrID)
	if active != 1 {
		t.Fatalf("expected exactly 1 ACTIVE assignment after concurrent creates, got %d", active)
	}
}

// ==================== Audit / transaction fault injection ====================

type failingAuditTxC struct {
	repository.Tx
}

func (f *failingAuditTxC) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if strings.Contains(query, "operation_log") {
		return nil, fmt.Errorf("simulated audit insert failure")
	}
	return f.Tx.ExecContext(ctx, query, args...)
}

type failingAuditDBC struct {
	*sql.DB
}

func (f *failingAuditDBC) BeginTx(ctx context.Context, opts *sql.TxOptions) (repository.Tx, error) {
	tx, err := f.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &failingAuditTxC{Tx: tx}, nil
}
func (f *failingAuditDBC) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.DB.ExecContext(ctx, query, args...)
}
func (f *failingAuditDBC) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.DB.QueryRowContext(ctx, query, args...)
}
func (f *failingAuditDBC) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.DB.QueryContext(ctx, query, args...)
}

func TestAuditFailureRollsBackEnrollmentCreate(t *testing.T) {
	ts := newTestServerWithFaultDB(t, func(db *sql.DB) repository.DB {
		return &failingAuditDBC{DB: db}
	})

	// Seed student + course directly via SQL.
	_, err := ts.db.Exec(`INSERT INTO student (name, status) VALUES ('S1', 'ACTIVE')`)
	if err != nil {
		t.Fatalf("seed student: %v", err)
	}
	studentID := lastInsertID(t, ts.db)
	_, err = ts.db.Exec(`INSERT INTO course_domain (name, code, type) VALUES ('JP', 'JP', 'LANGUAGE')`)
	domainID := lastInsertID(t, ts.db)
	_, err = ts.db.Exec(`INSERT INTO course_track (domain_id, name, code) VALUES (?, 'JLPT', 'JLPT')`, domainID)
	trackID := lastInsertID(t, ts.db)
	_, err = ts.db.Exec(`INSERT INTO course_level (track_id, name, code) VALUES (?, 'N5', 'N5')`, trackID)
	levelID := lastInsertID(t, ts.db)
	_ = levelID

	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")

	status, body := req(t, "POST", fmt.Sprintf("%s/students/%d/enrollments", ts.srv.URL, studentID), tok,
		map[string]any{"domainId": domainID, "trackId": trackID, "enrollmentType": "ONE_TO_ONE"})
	if status != http.StatusInternalServerError || codeOf(t, body) != float64(httpserver.CodeDatabase) {
		t.Fatalf("expected 500/50002 on audit failure, got %d %v", status, body)
	}
	// No enrollment should exist (business write rolled back with audit).
	var n int
	ts.db.QueryRow(`SELECT COUNT(*) FROM student_course_enrollment WHERE student_id = ?`, studentID).Scan(&n)
	if n != 0 {
		t.Fatalf("expected 0 enrollments after rollback, got %d", n)
	}
	if auditCountForAction(t, ts.db, "ENROLLMENT_CREATE") != 0 {
		t.Fatalf("expected 0 audit rows after rollback")
	}
}

func TestAuditFailureRollsBackAssignmentReplacement(t *testing.T) {
	ts := newTestServerWithFaultDB(t, func(db *sql.DB) repository.DB {
		return &failingAuditDBC{DB: db}
	})

	// Seed student, course, teacher, enrollment, and an initial ACTIVE assignment.
	_, err := ts.db.Exec(`INSERT INTO student (name, status) VALUES ('S1', 'ACTIVE')`)
	studentID := lastInsertID(t, ts.db)
	_, err = ts.db.Exec(`INSERT INTO course_domain (name, code, type) VALUES ('JP', 'JP', 'LANGUAGE')`)
	domainID := lastInsertID(t, ts.db)
	_, err = ts.db.Exec(`INSERT INTO course_track (domain_id, name, code) VALUES (?, 'JLPT', 'JLPT')`, domainID)
	trackID := lastInsertID(t, ts.db)
	_, err = ts.db.Exec(`INSERT INTO course_level (track_id, name, code) VALUES (?, 'N5', 'N5')`, trackID)
	_, err = ts.db.Exec(`INSERT INTO teacher (name, status) VALUES ('T1', 'ACTIVE')`)
	teacherID := lastInsertID(t, ts.db)
	_, err = ts.db.Exec(`INSERT INTO student_course_enrollment (student_id, domain_id, track_id, enrollment_type, status) VALUES (?, ?, ?, 'ONE_TO_ONE', 'ACTIVE')`, studentID, domainID, trackID)
	enrID := lastInsertID(t, ts.db)
	_, err = ts.db.Exec(`INSERT INTO student_teacher_assignment (enrollment_id, student_id, teacher_id, role_type, status, start_date) VALUES (?, ?, ?, 'MAIN', 'ACTIVE', '2024-01-01')`, enrID, studentID, teacherID)
	if err != nil {
		t.Fatalf("seed assignment: %v", err)
	}

	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")

	// Attempt replacement (new assignment) -> audit failure should roll back
	// both the end of the old assignment AND the new assignment insert.
	status, body := req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
		map[string]any{"teacherId": teacherID, "roleType": "MAIN", "reason": "REPLACE"})
	if status != http.StatusInternalServerError || codeOf(t, body) != float64(httpserver.CodeDatabase) {
		t.Fatalf("expected 500/50002 on audit failure, got %d %v", status, body)
	}
	// The original ACTIVE assignment must still be ACTIVE (end rolled back).
	active := countActiveAssignments(t, ts.db, enrID)
	if active != 1 {
		t.Fatalf("expected 1 ACTIVE assignment after rollback (original preserved), got %d", active)
	}
	// No new assignment row should exist.
	var total int
	ts.db.QueryRow(`SELECT COUNT(*) FROM student_teacher_assignment WHERE enrollment_id = ?`, enrID).Scan(&total)
	if total != 1 {
		t.Fatalf("expected 1 total assignment after rollback, got %d", total)
	}
}

// ==================== No financial/lesson/notification side effects ====================

func TestEnrollmentCreateNoFinancialOrLessonRecords(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	// No lesson, attendance, payment, notification, or email records should exist.
	for _, table := range []string{"lesson", "attendance", "payment", "notification", "email_log", "payout"} {
		if tableExists(t, ts.db, table) {
			var n int
			ts.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&n)
			if n != 0 {
				t.Fatalf("expected 0 rows in %s after enrollment, got %d", table, n)
			}
		}
	}
}

func TestAssignmentCreateNoFinancialOrLessonRecords(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	teacher1 := seedTeacher(t, ts, "T1")
	req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
		map[string]any{"teacherId": teacher1, "roleType": "MAIN"})

	for _, table := range []string{"lesson", "attendance", "payment", "notification", "email_log", "payout"} {
		if tableExists(t, ts.db, table) {
			var n int
			ts.db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&n)
			if n != 0 {
				t.Fatalf("expected 0 rows in %s after assignment, got %d", table, n)
			}
		}
	}
}

// ==================== No DELETE routes ====================

func TestNoDeleteRoutesForEnrollment(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	status, _ := req(t, "DELETE", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok, nil)
	if status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
		t.Fatalf("DELETE /enrollments/{id} should not be registered, got %d", status)
	}
}

// ==================== helpers ====================

func countActiveAssignments(t *testing.T, db *sql.DB, enrollmentID int64) int {
	t.Helper()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM student_teacher_assignment WHERE enrollment_id = ? AND status = 'ACTIVE'`, enrollmentID).Scan(&n); err != nil {
		t.Fatalf("count active assignments: %v", err)
	}
	return n
}

func lastInsertID(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	var id int64
	if err := db.QueryRow(`SELECT last_insert_rowid()`).Scan(&id); err != nil {
		t.Fatalf("last insert id: %v", err)
	}
	return id
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&n)
	if err != nil {
		t.Fatalf("check table %s: %v", name, err)
	}
	return n > 0
}

// Suppress unused import warnings for bytes/io/time when only some tests use them.
var _ = bytes.NewReader
var _ = io.Discard
var _ = time.Now
