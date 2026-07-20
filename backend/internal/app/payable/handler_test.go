package payable_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/prelove/zedu/backend/internal/app/payable"
	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
)

const payableSecret = "payable-test-secret-must-be-32-chars"

func TestPayableSummaryRequiresAuth(t *testing.T) {
	db := openDB(t)
	mux := httpserver.New()
	payable.MountRoutes(mux, db, payableSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/teachers/payable")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
	var body struct {
		Code int `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.Code != 40101 {
		t.Fatalf("expected code 40101, got %d", body.Code)
	}
}

func TestPayableSummaryEmptyReturnsZero(t *testing.T) {
	db := openDB(t)
	ownerID := seedUser(t, db, "OWNER")
	mux := httpserver.New()
	payable.MountRoutes(mux, db, payableSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp := doAuthed(t, srv.URL+"/teachers/payable", ownerID, "OWNER")
	defer resp.Body.Close()
	var body struct {
		Code int `json:"code"`
		Data struct {
			Items []payable.TeacherPayableSummary `json:"items"`
			Total int                             `json:"total"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data.Total != 0 || len(body.Data.Items) != 0 {
		t.Fatalf("expected empty, got %#v", body)
	}
}

func TestPayableSummaryReflectsConfirmedLesson(t *testing.T) {
	db := openDB(t)
	ownerID := seedUser(t, db, "OWNER")
	teacherID := seedTeacher(t, db)
	lessonID := seedScheduledLesson(t, db, teacherID)
	confirmLesson(t, db, lessonID, ownerID, 3000)
	mux := httpserver.New()
	payable.MountRoutes(mux, db, payableSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp := doAuthed(t, srv.URL+"/teachers/payable", ownerID, "OWNER")
	defer resp.Body.Close()
	var body struct {
		Code int `json:"code"`
		Data struct {
			Items []payable.TeacherPayableSummary `json:"items"`
			Total int                             `json:"total"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data.Total != 1 || len(body.Data.Items) != 1 {
		t.Fatalf("expected one teacher, got %#v", body)
	}
	if body.Data.Items[0].TeacherID != teacherID || body.Data.Items[0].UnpaidAmount != 3000 {
		t.Fatalf("expected teacher %d amount 3000, got %#v", teacherID, body.Data.Items[0])
	}
}

func TestPayableDetailReflectsConfirmedLesson(t *testing.T) {
	db := openDB(t)
	ownerID := seedUser(t, db, "OWNER")
	teacherID := seedTeacher(t, db)
	lessonID := seedScheduledLesson(t, db, teacherID)
	confirmLesson(t, db, lessonID, ownerID, 2500)
	mux := httpserver.New()
	payable.MountRoutes(mux, db, payableSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp := doAuthed(t, srv.URL+path.Join("/teachers", strconv.FormatInt(teacherID, 10), "payable"), ownerID, "OWNER")
	defer resp.Body.Close()
	var body struct {
		Code int `json:"code"`
		Data struct {
			Items []payable.TeacherPayableEntry `json:"items"`
			Total int                           `json:"total"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data.Total != 1 || len(body.Data.Items) != 1 {
		t.Fatalf("expected one entry, got %#v", body)
	}
	if body.Data.Items[0].LessonID != lessonID || body.Data.Items[0].AmountDelta != 2500 {
		t.Fatalf("expected lesson %d amount 2500, got %#v", lessonID, body.Data.Items[0])
	}
}

func TestPayableReadHasNoSideEffects(t *testing.T) {
	db := openDB(t)
	ownerID := seedUser(t, db, "OWNER")
	teacherID := seedTeacher(t, db)
	lessonID := seedScheduledLesson(t, db, teacherID)
	confirmLesson(t, db, lessonID, ownerID, 3000)
	mux := httpserver.New()
	payable.MountRoutes(mux, db, payableSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	var beforeCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM teacher_account_ledger`).Scan(&beforeCount); err != nil {
		t.Fatal(err)
	}
	resp := doAuthed(t, srv.URL+"/teachers/payable", ownerID, "OWNER")
	resp.Body.Close()
	resp2 := doAuthed(t, srv.URL+path.Join("/teachers", strconv.FormatInt(teacherID, 10), "payable"), ownerID, "OWNER")
	resp2.Body.Close()
	var afterCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM teacher_account_ledger`).Scan(&afterCount); err != nil {
		t.Fatal(err)
	}
	if beforeCount != afterCount {
		t.Fatalf("payable read changed ledger rows: before=%d after=%d", beforeCount, afterCount)
	}
}

// --- helpers ---

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "payable.db"))
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

func seedUser(t *testing.T, db *sql.DB, role string) int64 {
	t.Helper()
	h, err := auth.HashPassword("Pass1234")
	if err != nil {
		t.Fatal(err)
	}
	r, err := db.Exec(`INSERT INTO user_account(username,password_hash,role,display_name) VALUES(?,?,?,?)`, role+"-payable", h, role, role)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := r.LastInsertId()
	return id
}

func seedTeacher(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	r, err := db.Exec(`INSERT INTO teacher(name,default_rate_amount,status) VALUES('payable-teacher',5000,'ACTIVE')`)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := r.LastInsertId()
	return id
}

func seedScheduledLesson(t *testing.T, db *sql.DB, teacherID int64) int64 {
	t.Helper()
	rStu, err := db.Exec(`INSERT INTO student(name,timezone,status) VALUES('payable-student','Asia/Tokyo','ACTIVE')`)
	if err != nil {
		t.Fatal(err)
	}
	studentID, _ := rStu.LastInsertId()
	rDom, err := db.Exec(`INSERT INTO course_domain(name,code,type,sort_order) VALUES('payable-domain','PD','LANGUAGE',1)`)
	if err != nil {
		t.Fatal(err)
	}
	domainID, _ := rDom.LastInsertId()
	rTrk, err := db.Exec(`INSERT INTO course_track(domain_id,name,code,sort_order) VALUES(?,'payable-track','PT',1)`, domainID)
	if err != nil {
		t.Fatal(err)
	}
	trackID, _ := rTrk.LastInsertId()
	rLvl, err := db.Exec(`INSERT INTO course_level(track_id,name,code,sort_order) VALUES(?,'payable-level','PL',1)`, trackID)
	if err != nil {
		t.Fatal(err)
	}
	levelID, _ := rLvl.LastInsertId()
	rEnr, err := db.Exec(`INSERT INTO student_course_enrollment(student_id,domain_id,track_id,current_level_id,enrollment_type,status,charge_per_lesson_amount,lesson_balance,balance_amount) VALUES(?,?,?,?, 'ONE_TO_ONE','ACTIVE',3000,10,30000)`, studentID, domainID, trackID, levelID)
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
	rLes, err := db.Exec(`INSERT INTO lesson(lesson_no,enrollment_id,assignment_id,teacher_id,student_id,scheduled_start_at,scheduled_end_at,duration_min,timezone,meeting_type,status) VALUES('LSN-PAY',?,?,?,?,?,?,?,?, 'OFFLINE','SCHEDULED')`, enrollmentID, assignmentID, teacherID, studentID, start, end, 60, "Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	lessonID, _ := rLes.LastInsertId()
	return lessonID
}

func confirmLesson(t *testing.T, db *sql.DB, lessonID, operatorID int64, teacherPay int64) {
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

func doAuthed(t *testing.T, url string, userID int64, role string) *http.Response {
	t.Helper()
	tok, err := auth.SignAccessToken(payableSecret, userID, role, time.Hour)
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
