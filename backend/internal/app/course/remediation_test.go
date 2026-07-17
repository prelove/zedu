package course_test

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

// ==================== P1-1: Nested lists must return pagination envelope ====================

func TestListEnrollmentsReturnsPaginationEnvelope(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	status, body := req(t, "GET", fmt.Sprintf("%s/students/%d/enrollments?page=1&pageSize=1", ts.srv.URL, studentID), tok, nil)
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", status, body)
	}
	d := dataMap(t, body)
	if d["items"] == nil {
		t.Fatalf("expected items field, got %v", d)
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
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestListEnrollmentsEmptyReturnsEmptyArrayNotNull(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, _, _, _ := seedStudentAndCourse(t, ts)

	_, body := req(t, "GET", fmt.Sprintf("%s/students/%d/enrollments", ts.srv.URL, studentID), tok, nil)
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

func TestListAssignmentsReturnsPaginationEnvelope(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	teacher1 := seedTeacher(t, ts, "T1")
	req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok, map[string]any{"teacherId": teacher1, "roleType": "MAIN"})

	_, body := req(t, "GET", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok, nil)
	d := dataMap(t, body)
	if d["items"] == nil {
		t.Fatalf("expected items field, got %v", d)
	}
	if d["total"] == nil || int(d["total"].(float64)) != 1 {
		t.Fatalf("expected total=1, got %v", d["total"])
	}
}

// ==================== P1-2: ENDED/PAUSED student cannot update enrollment ====================

func TestUpdateEnrollmentPausedStudent42201(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	// Pause the student via directory API.
	ownerID := createUser(t, ts.db, "owner_dir", "OWNER")
	dirTok := tokenFor(t, ownerID, "OWNER")
	req(t, "PATCH", fmt.Sprintf("%s/students/%d", ts.srv.URL, studentID), dirTok, map[string]any{"status": "PAUSED"})

	// Now try to PATCH the enrollment (e.g. change note) — must fail 42201.
	status, body := req(t, "PATCH", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok, map[string]any{"note": "updated"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for PATCH enrollment with paused student, got %d %v", status, body)
	}
	// No audit written.
	if auditCountForAction(t, ts.db, "ENROLLMENT_UPDATE") != 0 {
		t.Fatalf("expected 0 audit rows for failed enrollment update")
	}
}

// ==================== P1-3: Referenced track/level cannot re-parent ====================

func TestUpdateTrackReparentRejectedWhenCapabilityReferences(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	_ = studentID
	_ = levelID

	// Create a capability referencing this track.
	ownerID := createUser(t, ts.db, "owner_dir", "OWNER")
	dirTok := tokenFor(t, ownerID, "OWNER")
	_, tb := req(t, "POST", ts.srv.URL+"/teachers", dirTok, map[string]any{"name": "T1"})
	teacherID := int64(dataMap(t, tb)["id"].(float64))
	req(t, "POST", fmt.Sprintf("%s/teachers/%d/capabilities", ts.srv.URL, teacherID), dirTok,
		map[string]any{"domainId": domainID, "trackId": trackID, "levelId": levelID})

	// Create a second domain to reparent to.
	_, db2 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "数学", "code": "MATH", "type": "K12"})
	otherDomainID := int64(dataMap(t, db2)["id"].(float64))

	// Try to reparent track to other domain — must fail 42201.
	status, body := req(t, "PATCH", fmt.Sprintf("%s/tracks/%d", ts.srv.URL, trackID), tok, map[string]any{"domainId": otherDomainID})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for reparenting referenced track, got %d %v", status, body)
	}
	// No audit.
	if auditCountForAction(t, ts.db, "TRACK_UPDATE") != 0 {
		t.Fatalf("expected 0 audit rows for rejected track reparent")
	}
}

func TestUpdateLevelReparentRejectedWhenEnrollmentReferences(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	_ = enrID

	// Create a second track under same domain to reparent level to.
	_, tb2 := req(t, "POST", ts.srv.URL+"/tracks", tok, map[string]any{"domainId": domainID, "name": "EJU", "code": "EJU"})
	otherTrackID := int64(dataMap(t, tb2)["id"].(float64))

	// Try to reparent level to other track — must fail 42201 (enrollment references it).
	status, body := req(t, "PATCH", fmt.Sprintf("%s/levels/%d", ts.srv.URL, levelID), tok, map[string]any{"trackId": otherTrackID})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for reparenting referenced level, got %d %v", status, body)
	}
	if auditCountForAction(t, ts.db, "LEVEL_UPDATE") != 0 {
		t.Fatalf("expected 0 audit rows for rejected level reparent")
	}
}

func TestUpdateTrackReparentSucceedsWhenUnreferenced(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	_, _, trackID, _ := seedStudentAndCourse(t, ts)

	// Create a second domain.
	_, db2 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "数学", "code": "MATH", "type": "K12"})
	otherDomainID := int64(dataMap(t, db2)["id"].(float64))

	// Track is unreferenced — reparent should succeed.
	status, body := req(t, "PATCH", fmt.Sprintf("%s/tracks/%d", ts.srv.URL, trackID), tok, map[string]any{"domainId": otherDomainID})
	if status != http.StatusOK {
		t.Fatalf("expected 200 for unreferenced track reparent, got %d body=%v", status, body)
	}
	if int64(dataMap(t, body)["domainId"].(float64)) != otherDomainID {
		t.Fatalf("expected domainId=%d, got %v", otherDomainID, dataMap(t, body)["domainId"])
	}
}

// ==================== P1-4: Level change writes event, preserves enrollment current level ====================

func TestLevelChangeWritesEventAndPreservesEnrollmentLevel(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	// Create a second level under same track.
	_, lb2 := req(t, "POST", ts.srv.URL+"/levels", tok, map[string]any{"trackId": trackID, "name": "N4", "code": "N4"})
	newLevelID := int64(dataMap(t, lb2)["id"].(float64))

	// PATCH enrollment's currentLevelId to new level.
	status, body := req(t, "PATCH", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok,
		map[string]any{"currentLevelId": newLevelID})
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%v", status, body)
	}

	// Enrollment's current_level_id must NOT be the new level (preserved original).
	_, gb := req(t, "GET", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok, nil)
	enrData := dataMap(t, gb)
	currentLevel := enrData["currentLevelId"]
	if currentLevel != nil && int64(currentLevel.(float64)) == newLevelID {
		t.Fatalf("enrollment current_level_id was overwritten to new level; expected preserved original %d, got %v", levelID, currentLevel)
	}

	// A student_level_event row must exist.
	var eventCount int
	ts.db.QueryRow(`SELECT COUNT(*) FROM student_level_event WHERE enrollment_id = ? AND to_level_id = ?`, enrID, newLevelID).Scan(&eventCount)
	if eventCount != 1 {
		t.Fatalf("expected 1 level event for level change, got %d", eventCount)
	}

	// Verify event has from_level_id, to_level_id, event_type=MANUAL, operator_id.
	var fromLevel, toLevel, operatorID int64
	var eventType string
	ts.db.QueryRow(`SELECT from_level_id, to_level_id, event_type, operator_id FROM student_level_event WHERE enrollment_id = ? ORDER BY id DESC LIMIT 1`, enrID).Scan(&fromLevel, &toLevel, &eventType, &operatorID)
	if fromLevel != levelID {
		t.Fatalf("expected from_level_id=%d, got %d", levelID, fromLevel)
	}
	if toLevel != newLevelID {
		t.Fatalf("expected to_level_id=%d, got %d", newLevelID, toLevel)
	}
	if eventType != "MANUAL" {
		t.Fatalf("expected event_type=MANUAL, got %s", eventType)
	}
	if operatorID == 0 {
		t.Fatalf("expected non-zero operator_id")
	}
}

func TestCourseSelectionChangeAuditContainsBeforeAfter(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	// Create a second domain/track/level.
	_, db2 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "数学", "code": "MATH", "type": "K12"})
	newDomainID := int64(dataMap(t, db2)["id"].(float64))
	_, tb2 := req(t, "POST", ts.srv.URL+"/tracks", tok, map[string]any{"domainId": newDomainID, "name": "Algebra", "code": "ALG"})
	newTrackID := int64(dataMap(t, tb2)["id"].(float64))
	_, lb2 := req(t, "POST", ts.srv.URL+"/levels", tok, map[string]any{"trackId": newTrackID, "name": "A1", "code": "A1"})
	newLevelID := int64(dataMap(t, lb2)["id"].(float64))

	// PATCH domain, track, and currentLevelId — audit must include before/after.
	patchStatus, patchBody := req(t, "PATCH", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok,
		map[string]any{"domainId": newDomainID, "trackId": newTrackID, "currentLevelId": newLevelID})
	if patchStatus != http.StatusOK {
		t.Fatalf("PATCH enrollment failed: status=%d body=%v", patchStatus, patchBody)
	}

	var detailJSON string
	ts.db.QueryRow(`SELECT detail_json FROM operation_log WHERE action = 'ENROLLMENT_UPDATE' AND target_id = ? ORDER BY id DESC LIMIT 1`, enrID).Scan(&detailJSON)
	if !strings.Contains(detailJSON, "before") {
		t.Fatalf("audit detail missing 'before' snapshot: %s", detailJSON)
	}
	if !strings.Contains(detailJSON, "after") {
		t.Fatalf("audit detail missing 'after' snapshot: %s", detailJSON)
	}
}

func TestLevelEventFaultInjectionNoHalfWrite(t *testing.T) {
	ts := newTestServerWithFaultDB(t, func(db *sql.DB) repository.DB {
		return &failingAuditDBC{DB: db}
	})
	// Seed via SQL.
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
	_, err = ts.db.Exec(`INSERT INTO course_level (track_id, name, code) VALUES (?, 'N4', 'N4')`, trackID)
	newLevelID := lastInsertID(t, ts.db)
	_, err = ts.db.Exec(`INSERT INTO student_course_enrollment (student_id, domain_id, track_id, current_level_id, enrollment_type, status) VALUES (?, ?, ?, ?, 'ONE_TO_ONE', 'ACTIVE')`, studentID, domainID, trackID, levelID)
	enrID := lastInsertID(t, ts.db)

	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")

	// PATCH currentLevelId — audit will fail, so level event + audit must all roll back.
	status, body := req(t, "PATCH", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok,
		map[string]any{"currentLevelId": newLevelID})
	if status != http.StatusInternalServerError || codeOf(t, body) != float64(httpserver.CodeDatabase) {
		t.Fatalf("expected 500/50002 on audit failure, got %d %v", status, body)
	}
	// No level event should exist.
	var n int
	ts.db.QueryRow(`SELECT COUNT(*) FROM student_level_event WHERE enrollment_id = ?`, enrID).Scan(&n)
	if n != 0 {
		t.Fatalf("expected 0 level events after rollback, got %d", n)
	}
}

// ==================== P1-5: Invalid level IDs and roleType rejected at service layer ====================

func TestEnrollmentCreateZeroLevelID42201(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, _ := seedStudentAndCourse(t, ts)

	status, body := req(t, "POST", fmt.Sprintf("%s/students/%d/enrollments", ts.srv.URL, studentID), tok,
		map[string]any{"domainId": domainID, "trackId": trackID, "currentLevelId": 0, "enrollmentType": "ONE_TO_ONE"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for currentLevelId=0, got %d %v", status, body)
	}
}

func TestEnrollmentCreateNegativeLevelID42201(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, _ := seedStudentAndCourse(t, ts)

	status, body := req(t, "POST", fmt.Sprintf("%s/students/%d/enrollments", ts.srv.URL, studentID), tok,
		map[string]any{"domainId": domainID, "trackId": trackID, "targetLevelId": -5, "enrollmentType": "ONE_TO_ONE"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for targetLevelId=-5, got %d %v", status, body)
	}
}

func TestAssignmentCreateInvalidRoleType42201(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)
	teacher1 := seedTeacher(t, ts, "T1")

	status, body := req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
		map[string]any{"teacherId": teacher1, "roleType": "BOGUS"})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for invalid roleType, got %d %v", status, body)
	}
}

func TestAssignmentCreateValidRoleTypesAccepted(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	for _, rt := range []string{"MAIN", "SUBSTITUTE", "ASSISTANT"} {
		teacherID := seedTeacher(t, ts, "T_"+rt)
		status, body := req(t, "POST", fmt.Sprintf("%s/enrollments/%d/assignments", ts.srv.URL, enrID), tok,
			map[string]any{"teacherId": teacherID, "roleType": rt})
		if status != http.StatusCreated {
			t.Fatalf("expected 201 for roleType=%s, got %d body=%v", rt, status, body)
		}
		if dataMap(t, body)["roleType"] != rt {
			t.Fatalf("expected roleType=%s, got %v", rt, dataMap(t, body)["roleType"])
		}
	}
}

// ==================== P1-6: Rollback failure must map to 50002 (course) ====================

type failingRollbackTxC struct {
	repository.Tx
}

func (f *failingRollbackTxC) Rollback() error {
	return fmt.Errorf("simulated rollback failure")
}

type failingRollbackDBC struct {
	*sql.DB
}

func (f *failingRollbackDBC) BeginTx(ctx context.Context, opts *sql.TxOptions) (repository.Tx, error) {
	tx, err := f.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &failingRollbackTxC{Tx: tx}, nil
}
func (f *failingRollbackDBC) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return f.DB.ExecContext(ctx, query, args...)
}
func (f *failingRollbackDBC) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return f.DB.QueryRowContext(ctx, query, args...)
}
func (f *failingRollbackDBC) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return f.DB.QueryContext(ctx, query, args...)
}

func TestRollbackFailureMapsTo50002Course(t *testing.T) {
	ts := newTestServerWithFaultDB(t, func(db *sql.DB) repository.DB {
		return &failingRollbackDBC{DB: db}
	})
	// Seed a domain with code DUP to cause conflict.
	_, err := ts.db.Exec(`INSERT INTO course_domain (name, code, type) VALUES ('A', 'DUP', 'OTHER')`)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	tok := tokenFor(t, ownerID, "OWNER")

	// Create domain with duplicate code — conflict triggers rollback, which fails.
	status, body := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "B", "code": "DUP", "type": "OTHER"})
	if status != http.StatusInternalServerError || codeOf(t, body) != float64(httpserver.CodeDatabase) {
		t.Fatalf("expected 500/50002 on rollback failure, got %d %v", status, body)
	}
	msg, _ := body["message"].(string)
	if strings.Contains(strings.ToLower(msg), "sqlite") {
		t.Fatalf("response leaked sqlite text: %s", msg)
	}
}

// ==================== P2-2: Empty body PATCH returns 42201 ====================

func TestPatchDomainEmptyBody42201(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	_, db1 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "日语", "code": "JP", "type": "LANGUAGE"})
	domainID := int64(dataMap(t, db1)["id"].(float64))

	status, body := req(t, "PATCH", fmt.Sprintf("%s/course-domains/%d", ts.srv.URL, domainID), tok, map[string]any{})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for empty body PATCH domain, got %d %v", status, body)
	}
	if auditCountForAction(t, ts.db, "DOMAIN_UPDATE") != 0 {
		t.Fatalf("expected 0 audit rows for empty PATCH")
	}
}

func TestPatchEnrollmentEmptyBody42201(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	studentID, domainID, trackID, levelID := seedStudentAndCourse(t, ts)
	enrID := createEnrollment(t, ts, tok, studentID, domainID, trackID, levelID)

	status, body := req(t, "PATCH", fmt.Sprintf("%s/enrollments/%d", ts.srv.URL, enrID), tok, map[string]any{})
	if status != http.StatusUnprocessableEntity || codeOf(t, body) != float64(httpserver.CodeInvalidState) {
		t.Fatalf("expected 422/42201 for empty body PATCH enrollment, got %d %v", status, body)
	}
	if auditCountForAction(t, ts.db, "ENROLLMENT_UPDATE") != 0 {
		t.Fatalf("expected 0 audit rows for empty PATCH enrollment")
	}
}

func TestPatchDomainRawEmptyBody42201(t *testing.T) {
	ts := newTestServer(t)
	tok := ownerToken(t, ts)
	_, db1 := req(t, "POST", ts.srv.URL+"/course-domains", tok, map[string]any{"name": "日语", "code": "JP", "type": "LANGUAGE"})
	domainID := int64(dataMap(t, db1)["id"].(float64))

	httpReq, _ := http.NewRequest("PATCH", fmt.Sprintf("%s/course-domains/%d", ts.srv.URL, domainID), bytes.NewReader([]byte{}))
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
