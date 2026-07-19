package lesson_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/prelove/zedu/backend/internal/app/attendance"
	"github.com/prelove/zedu/backend/internal/app/lesson"
	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/platform/logging"
)

const lessonTestSecret = "lesson-test-secret-must-be-32-chars"

type lessonServer struct {
	db  *sql.DB
	srv *httptest.Server
}

func newLessonServer(t *testing.T) *lessonServer {
	t.Helper()
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "lesson.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err = database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	lesson.MountRoutes(mux, lesson.NewHandler(db, logger), db, lessonTestSecret)
	attendance.MountRoutes(mux, attendance.NewHandler(db), db, lessonTestSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	t.Cleanup(func() { srv.Close(); _ = db.Close() })
	return &lessonServer{db: db, srv: srv}
}

func TestLessonConfirmationCreatesAtomicFacts(t *testing.T) {
	ts := newLessonServer(t)
	userID := seedLessonUser(t, ts.db, "confirm-owner", "OWNER")
	enrollmentID, assignmentID := seedActiveTeachingRelationship(t, ts.db)
	token := lessonToken(t, userID, "OWNER")
	status, data := lessonRequest(t, http.MethodPost, ts.srv.URL+"/lessons", token, map[string]any{"enrollmentId": enrollmentID, "assignmentId": assignmentID, "startAt": "2026-08-01T19:00:00", "durationMin": 60, "timezone": "Asia/Tokyo", "meetingType": "OFFLINE"})
	if status != http.StatusCreated {
		t.Fatalf("create = %d %#v", status, data)
	}
	id := int64(responseData(t, data)["id"].(float64))
	status, data = lessonRequest(t, http.MethodPost, ts.srv.URL+"/lessons/"+itoa(id)+"/confirm", token, map[string]any{"outcomeType": "ATTENDED", "lessonDeducted": "1", "chargeAmount": 0, "teacherPayAmount": 0, "actualDurationMin": 60})
	if status != http.StatusOK {
		t.Fatalf("confirm = %d %#v", status, data)
	}
	assertCount(t, ts.db, "SELECT COUNT(*) FROM attendance WHERE lesson_id=?", 1, id)
	assertCount(t, ts.db, "SELECT COUNT(*) FROM lesson_finance WHERE lesson_id=?", 1, id)
	assertCount(t, ts.db, "SELECT COUNT(*) FROM teacher_account_ledger WHERE lesson_id=?", 1, id)
	var lessonStatus string
	if err := ts.db.QueryRow("SELECT status FROM lesson WHERE id=?", id).Scan(&lessonStatus); err != nil || lessonStatus != "COMPLETED" {
		t.Fatalf("lesson status=%s err=%v", lessonStatus, err)
	}
	status, data = lessonRequest(t, http.MethodPost, ts.srv.URL+"/lessons/"+itoa(id)+"/confirm", token, map[string]any{"outcomeType": "ATTENDED", "lessonDeducted": "1", "chargeAmount": 0, "teacherPayAmount": 0})
	if status != http.StatusUnprocessableEntity || responseCode(data) != 42201 {
		t.Fatalf("duplicate = %d %#v", status, data)
	}
}

func TestConcurrentLessonConfirmationHasOneWinner(t *testing.T) {
	ts := newLessonServer(t)
	userID := seedLessonUser(t, ts.db, "concurrent-confirm-owner", "OWNER")
	enrollmentID, assignmentID := seedActiveTeachingRelationship(t, ts.db)
	token := lessonToken(t, userID, "OWNER")
	status, data := lessonRequest(t, http.MethodPost, ts.srv.URL+"/lessons", token, map[string]any{"enrollmentId": enrollmentID, "assignmentId": assignmentID, "startAt": "2026-08-01T19:00:00", "durationMin": 60, "timezone": "Asia/Tokyo", "meetingType": "OFFLINE"})
	if status != http.StatusCreated {
		t.Fatalf("create = %d %#v", status, data)
	}
	id := int64(responseData(t, data)["id"].(float64))
	var wg sync.WaitGroup
	results := make(chan int, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			code, _ := lessonRequest(t, http.MethodPost, ts.srv.URL+"/lessons/"+itoa(id)+"/confirm", token, map[string]any{"outcomeType": "ATTENDED", "lessonDeducted": "1", "chargeAmount": 0, "teacherPayAmount": 0})
			results <- code
		}()
	}
	wg.Wait()
	close(results)
	success := 0
	for code := range results {
		if code == http.StatusOK {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("successful confirmations = %d, want 1", success)
	}
	assertCount(t, ts.db, "SELECT COUNT(*) FROM attendance WHERE lesson_id=?", 1, id)
}

func TestLessonLifecycleUsesOnlyLessonAndAuditFacts(t *testing.T) {
	ts := newLessonServer(t)
	userID := seedLessonUser(t, ts.db, "owner", "OWNER")
	enrollmentID, assignmentID := seedActiveTeachingRelationship(t, ts.db)
	token := lessonToken(t, userID, "OWNER")

	status, data := lessonRequest(t, http.MethodPost, ts.srv.URL+"/lessons", token, map[string]any{
		"enrollmentId": enrollmentID, "assignmentId": assignmentID, "startAt": "2026-08-01T19:00:00", "durationMin": 60,
		"timezone": "Asia/Tokyo", "meetingType": "OFFLINE", "lessonTopic": "会话练习", "note": "UTF-8 😀",
	})
	if status != http.StatusCreated {
		t.Fatalf("create status = %d, response=%v", status, data)
	}
	created := responseData(t, data)
	lessonID := int64(created["id"].(float64))
	if created["status"] != "SCHEDULED" || created["timezone"] != "Asia/Tokyo" || created["scheduledStartAt"] != "2026-08-01T10:00:00Z" {
		t.Fatalf("unexpected created lesson: %#v", created)
	}
	assertCount(t, ts.db, "SELECT COUNT(*) FROM operation_log WHERE action='LESSON_CREATE' AND target_id=?", 1, lessonID)
	assertCount(t, ts.db, "SELECT COUNT(*) FROM payment_order", 0)

	status, data = lessonRequest(t, http.MethodPatch, ts.srv.URL+"/lessons/"+itoa(lessonID), token, map[string]any{
		"startAt": "2026-08-02T20:00:00", "durationMin": 90, "timezone": "Asia/Tokyo", "meetingType": "WECHAT", "meetingLink": "https://example.test/meeting", "lessonTopic": "更新", "note": "更新备注",
	})
	if status != http.StatusOK {
		t.Fatalf("update status = %d, response=%v", status, data)
	}
	updated := responseData(t, data)
	if updated["scheduledStartAt"] != "2026-08-02T11:00:00Z" || updated["durationMin"] != float64(90) {
		t.Fatalf("unexpected update: %#v", updated)
	}

	status, data = lessonRequest(t, http.MethodGet, ts.srv.URL+"/lessons?studentId=1&status=SCHEDULED", token, nil)
	if status != http.StatusOK {
		t.Fatalf("list status = %d, response=%v", status, data)
	}
	items := responseData(t, data)["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("list items = %#v", items)
	}

	status, data = lessonRequest(t, http.MethodPost, ts.srv.URL+"/lessons/"+itoa(lessonID)+"/cancel", token, map[string]any{"reason": "学生请假"})
	if status != http.StatusOK {
		t.Fatalf("cancel status = %d, response=%v", status, data)
	}
	if responseData(t, data)["status"] != "CANCELLED" {
		t.Fatalf("cancel response = %#v", data)
	}
	assertCount(t, ts.db, "SELECT COUNT(*) FROM operation_log WHERE action IN ('LESSON_CREATE','LESSON_UPDATE','LESSON_CANCEL') AND target_id=?", 3, lessonID)

	status, data = lessonRequest(t, http.MethodPatch, ts.srv.URL+"/lessons/"+itoa(lessonID), token, map[string]any{"startAt": "2026-08-03T20:00:00", "durationMin": 60, "timezone": "Asia/Tokyo", "meetingType": "OFFLINE"})
	if status != http.StatusUnprocessableEntity || responseCode(data) != 42201 {
		t.Fatalf("cancelled update must be 42201, got %d %#v", status, data)
	}
}

func TestLessonRejectsInvalidRelationshipAndAuthorization(t *testing.T) {
	ts := newLessonServer(t)
	ownerID := seedLessonUser(t, ts.db, "owner", "OWNER")
	viewerID := seedLessonUser(t, ts.db, "viewer", "VIEWER")
	enrollmentID, assignmentID := seedActiveTeachingRelationship(t, ts.db)
	body := map[string]any{"enrollmentId": enrollmentID, "assignmentId": assignmentID, "startAt": "2026-08-01T19:00:00", "durationMin": 60, "timezone": "Asia/Tokyo", "meetingType": "OFFLINE"}
	status, data := lessonRequest(t, http.MethodPost, ts.srv.URL+"/lessons", lessonToken(t, viewerID, "VIEWER"), body)
	if status != http.StatusForbidden || responseCode(data) != 40301 {
		t.Fatalf("viewer write = %d %#v", status, data)
	}
	if _, err := ts.db.Exec("UPDATE student_teacher_assignment SET status='ENDED' WHERE id=?", assignmentID); err != nil {
		t.Fatal(err)
	}
	status, data = lessonRequest(t, http.MethodPost, ts.srv.URL+"/lessons", lessonToken(t, ownerID, "OWNER"), body)
	if status != http.StatusUnprocessableEntity || responseCode(data) != 42201 {
		t.Fatalf("inactive assignment = %d %#v", status, data)
	}
	assertCount(t, ts.db, "SELECT COUNT(*) FROM lesson", 0)
	assertCount(t, ts.db, "SELECT COUNT(*) FROM operation_log WHERE action LIKE 'LESSON_%'", 0)
}

func TestLessonRequiresAuthenticationAndRejectsBadSchedulingInput(t *testing.T) {
	ts := newLessonServer(t)
	status, data := lessonRequest(t, http.MethodGet, ts.srv.URL+"/lessons", "", nil)
	if status != http.StatusUnauthorized || responseCode(data) != 40101 {
		t.Fatalf("unauthenticated = %d %#v", status, data)
	}
	userID := seedLessonUser(t, ts.db, "operator", "OPERATOR")
	enrollmentID, assignmentID := seedActiveTeachingRelationship(t, ts.db)
	status, data = lessonRequest(t, http.MethodPost, ts.srv.URL+"/lessons", lessonToken(t, userID, "OPERATOR"), map[string]any{"enrollmentId": enrollmentID, "assignmentId": assignmentID, "startAt": "bad", "durationMin": 9, "timezone": "No/Such_Zone", "meetingType": "WECHAT", "meetingLink": "bad"})
	if status != http.StatusUnprocessableEntity || responseCode(data) != 42201 {
		t.Fatalf("invalid schedule = %d %#v", status, data)
	}
	assertCount(t, ts.db, "SELECT COUNT(*) FROM lesson", 0)
}

func seedLessonUser(t *testing.T, db *sql.DB, username, role string) int64 {
	t.Helper()
	hash, err := auth.HashPassword("Pass1234")
	if err != nil {
		t.Fatal(err)
	}
	result, err := db.Exec("INSERT INTO user_account (username,password_hash,role,display_name) VALUES (?,?,?,?)", username, hash, role, username)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := result.LastInsertId()
	return id
}

func seedActiveTeachingRelationship(t *testing.T, db *sql.DB) (int64, int64) {
	t.Helper()
	execID := func(query string, args ...any) int64 {
		result, err := db.Exec(query, args...)
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
		id, _ := result.LastInsertId()
		return id
	}
	domainID := execID("INSERT INTO course_domain(name,code,type) VALUES ('Japanese','JP','LANGUAGE')")
	trackID := execID("INSERT INTO course_track(domain_id,name,code) VALUES (?,?,?)", domainID, "Beginner", "BEGINNER")
	levelID := execID("INSERT INTO course_level(track_id,name,code) VALUES (?,?,?)", trackID, "N5", "N5")
	studentID := execID("INSERT INTO student(name,timezone) VALUES (?,?)", "学生", "Asia/Tokyo")
	teacherID := execID("INSERT INTO teacher(name) VALUES (?)", "老师")
	enrollmentID := execID("INSERT INTO student_course_enrollment(student_id,domain_id,track_id,current_level_id,target_level_id,status) VALUES (?,?,?,?,?,'ACTIVE')", studentID, domainID, trackID, levelID, levelID)
	assignmentID := execID("INSERT INTO student_teacher_assignment(enrollment_id,student_id,teacher_id,role_type,status,start_date) VALUES (?,?,?,'MAIN','ACTIVE','2026-01-01')", enrollmentID, studentID, teacherID)
	return enrollmentID, assignmentID
}

func lessonToken(t *testing.T, userID int64, role string) string {
	t.Helper()
	token, err := auth.SignAccessToken(lessonTestSecret, userID, role, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func lessonRequest(t *testing.T, method, url, token string, body any) (int, map[string]any) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, data
}

func responseData(t *testing.T, response map[string]any) map[string]any {
	t.Helper()
	data, ok := response["data"].(map[string]any)
	if !ok {
		t.Fatalf("missing data: %#v", response)
	}
	return data
}
func responseCode(response map[string]any) int {
	value, _ := response["code"].(float64)
	return int(value)
}
func itoa(value int64) string { return strconv.FormatInt(value, 10) }
func assertCount(t *testing.T, db *sql.DB, query string, expected int, args ...any) {
	t.Helper()
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != expected {
		t.Fatalf("count query %q = %d, want %d", query, count, expected)
	}
}
