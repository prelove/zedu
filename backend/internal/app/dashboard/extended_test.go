package dashboard_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/prelove/zedu/backend/internal/app/dashboard"
	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
)

const dashSecret = "dashboard-ext-test-secret-32-chars"

// TestDashboardExtendedCounts verifies that the dashboard returns all five
// read-only operational counts.
func TestDashboardExtendedCounts(t *testing.T) {
	db := openDashDB(t)
	ownerID := seedDashUser(t, db, "OWNER")
	mux := httpserver.New()
	dashboard.MountRoutes(mux, db, dashSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp := doDashAuthed(t, srv.URL+"/dashboard", ownerID, "OWNER")
	defer resp.Body.Close()
	var body struct {
		Code int `json:"code"`
		Data struct {
			TodayLessons               int `json:"todayLessons"`
			PendingLessonConfirmations int `json:"pendingLessonConfirmations"`
			RenewalNeededStudents      int `json:"renewalNeededStudents"`
			TeacherPayableAggregate    int `json:"teacherPayableAggregate"`
			FailedNotifications        int `json:"failedNotifications"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 {
		t.Fatalf("expected code 0, got %d", body.Code)
	}
	// All counts must be zero on empty DB.
	if body.Data.TodayLessons != 0 || body.Data.PendingLessonConfirmations != 0 || body.Data.RenewalNeededStudents != 0 || body.Data.TeacherPayableAggregate != 0 || body.Data.FailedNotifications != 0 {
		t.Fatalf("expected all zero, got %#v", body.Data)
	}
}

// TestDashboardReflectsConfirmedLessonPayable verifies that after confirming a
// lesson, the teacherPayableAggregate reflects the teacher pay amount.
func TestDashboardReflectsConfirmedLessonPayable(t *testing.T) {
	db := openDashDB(t)
	ownerID := seedDashUser(t, db, "OWNER")
	teacherID := seedDashTeacher(t, db)
	lessonID := seedDashScheduledLesson(t, db, teacherID)
	confirmDashLesson(t, db, lessonID, ownerID, 4000)
	mux := httpserver.New()
	dashboard.MountRoutes(mux, db, dashSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp := doDashAuthed(t, srv.URL+"/dashboard", ownerID, "OWNER")
	defer resp.Body.Close()
	var body struct {
		Code int `json:"code"`
		Data struct {
			TeacherPayableAggregate    int `json:"teacherPayableAggregate"`
			PendingLessonConfirmations int `json:"pendingLessonConfirmations"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Data.TeacherPayableAggregate != 4000 {
		t.Fatalf("expected payable 4000, got %d", body.Data.TeacherPayableAggregate)
	}
	if body.Data.PendingLessonConfirmations != 0 {
		t.Fatalf("expected 0 pending after confirm, got %d", body.Data.PendingLessonConfirmations)
	}
}

// TestDashboardReadOnlyNoSideEffects verifies that reading the dashboard does
// not write any ledger, lesson, outbox, or audit rows.
func TestDashboardReadOnlyNoSideEffects(t *testing.T) {
	db := openDashDB(t)
	ownerID := seedDashUser(t, db, "OWNER")
	mux := httpserver.New()
	dashboard.MountRoutes(mux, db, dashSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	var beforeOp, beforeLedger, beforeLesson, beforeOutbox int
	_ = db.QueryRow(`SELECT COUNT(*) FROM operation_log`).Scan(&beforeOp)
	_ = db.QueryRow(`SELECT COUNT(*) FROM teacher_account_ledger`).Scan(&beforeLedger)
	_ = db.QueryRow(`SELECT COUNT(*) FROM lesson`).Scan(&beforeLesson)
	_ = db.QueryRow(`SELECT COUNT(*) FROM notification_outbox`).Scan(&beforeOutbox)
	resp := doDashAuthed(t, srv.URL+"/dashboard", ownerID, "OWNER")
	resp.Body.Close()
	var afterOp, afterLedger, afterLesson, afterOutbox int
	_ = db.QueryRow(`SELECT COUNT(*) FROM operation_log`).Scan(&afterOp)
	_ = db.QueryRow(`SELECT COUNT(*) FROM teacher_account_ledger`).Scan(&afterLedger)
	_ = db.QueryRow(`SELECT COUNT(*) FROM lesson`).Scan(&afterLesson)
	_ = db.QueryRow(`SELECT COUNT(*) FROM notification_outbox`).Scan(&afterOutbox)
	if beforeOp != afterOp || beforeLedger != afterLedger || beforeLesson != afterLesson || beforeOutbox != afterOutbox {
		t.Fatalf("dashboard read had side effects: op %d→%d ledger %d→%d lesson %d→%d outbox %d→%d",
			beforeOp, afterOp, beforeLedger, afterLedger, beforeLesson, afterLesson, beforeOutbox, afterOutbox)
	}
}

// --- helpers ---

func openDashDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "dash.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func seedDashUser(t *testing.T, db *sql.DB, role string) int64 {
	t.Helper()
	h, err := auth.HashPassword("Pass1234")
	if err != nil {
		t.Fatal(err)
	}
	r, err := db.Exec(`INSERT INTO user_account(username,password_hash,role,display_name) VALUES(?,?,?,?)`, role+"-dash", h, role, role)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := r.LastInsertId()
	return id
}

func seedDashTeacher(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	r, err := db.Exec(`INSERT INTO teacher(name,default_rate_amount,status) VALUES('dash-teacher',5000,'ACTIVE')`)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := r.LastInsertId()
	return id
}

func seedDashScheduledLesson(t *testing.T, db *sql.DB, teacherID int64) int64 {
	t.Helper()
	rStu, err := db.Exec(`INSERT INTO student(name,timezone,status) VALUES('dash-student','Asia/Tokyo','ACTIVE')`)
	if err != nil {
		t.Fatal(err)
	}
	studentID, _ := rStu.LastInsertId()
	rDom, err := db.Exec(`INSERT INTO course_domain(name,code,type,sort_order) VALUES('dash-domain','DD','LANGUAGE',1)`)
	if err != nil {
		t.Fatal(err)
	}
	domainID, _ := rDom.LastInsertId()
	rTrk, err := db.Exec(`INSERT INTO course_track(domain_id,name,code,sort_order) VALUES(?,'dash-track','DT',1)`, domainID)
	if err != nil {
		t.Fatal(err)
	}
	trackID, _ := rTrk.LastInsertId()
	rLvl, err := db.Exec(`INSERT INTO course_level(track_id,name,code,sort_order) VALUES(?,'dash-level','DL',1)`, trackID)
	if err != nil {
		t.Fatal(err)
	}
	levelID, _ := rLvl.LastInsertId()
	rEnr, err := db.Exec(`INSERT INTO student_course_enrollment(student_id,domain_id,track_id,current_level_id,enrollment_type,status,charge_per_lesson_amount,lesson_balance,balance_amount) VALUES(?,?,?,?, 'ONE_TO_ONE','ACTIVE',4000,10,40000)`, studentID, domainID, trackID, levelID)
	if err != nil {
		t.Fatal(err)
	}
	enrollmentID, _ := rEnr.LastInsertId()
	rAsg, err := db.Exec(`INSERT INTO student_teacher_assignment(enrollment_id,student_id,teacher_id,role_type,status,start_date) VALUES(?,?,?, 'MAIN','ACTIVE','2026-01-01')`, enrollmentID, studentID, teacherID)
	if err != nil {
		t.Fatal(err)
	}
	assignmentID, _ := rAsg.LastInsertId()
	start := time.Now().UTC().Add(24 * time.Hour)
	end := start.Add(60 * time.Minute)
	rLes, err := db.Exec(`INSERT INTO lesson(lesson_no,enrollment_id,assignment_id,teacher_id,student_id,scheduled_start_at,scheduled_end_at,duration_min,timezone,meeting_type,status) VALUES('LSN-DASH',?,?,?,?,?,?,?,?, 'OFFLINE','SCHEDULED')`, enrollmentID, assignmentID, teacherID, studentID, start, end, 60, "Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	lessonID, _ := rLes.LastInsertId()
	return lessonID
}

func confirmDashLesson(t *testing.T, db *sql.DB, lessonID, operatorID int64, teacherPay int64) {
	t.Helper()
	var student, enroll, teacher int64
	var duration int
	if err := db.QueryRow(`SELECT student_id,enrollment_id,teacher_id,duration_min FROM lesson WHERE id=?`, lessonID).Scan(&student, &enroll, &teacher, &duration); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO attendance(lesson_id,outcome_type,actual_duration_min,lesson_deducted,charge_amount,teacher_pay_amount,confirmed_by) VALUES(?,'ATTENDED',?, '1', ?, ?, ?)`, lessonID, duration, teacherPay, teacherPay, operatorID); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO teacher_account_ledger(teacher_id,lesson_id,amount_delta,balance_after,operator_id) VALUES(?,?,?,COALESCE((SELECT balance_after FROM teacher_account_ledger WHERE teacher_id=? ORDER BY id DESC LIMIT 1),0)+?,?)`, teacher, lessonID, teacherPay, teacher, teacherPay, operatorID); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO lesson_finance(lesson_id,student_id,teacher_id,enrollment_id,charge_amount,teacher_pay_amount,gross_profit_amount) VALUES(?,?,?,?,?,?,0)`, lessonID, student, teacher, enroll, teacherPay, teacherPay); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if _, err := tx.Exec(`UPDATE lesson SET status='COMPLETED' WHERE id=?`, lessonID); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func doDashAuthed(t *testing.T, url string, userID int64, role string) *http.Response {
	t.Helper()
	tok, err := auth.SignAccessToken(dashSecret, userID, role, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
