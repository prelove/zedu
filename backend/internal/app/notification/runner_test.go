package notification_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prelove/zedu/backend/internal/app/notification"
	"github.com/prelove/zedu/backend/internal/repository"
)

// TestReminderScanIsIdempotent verifies that scanning the same eligible lesson
// twice produces at most one LESSON_REMINDER outbox row per recipient.
func TestReminderScanIsIdempotent(t *testing.T) {
	db := testDB(t)
	studentID, lessonID := seedReminderLesson(t, db, time.Now().UTC().Add(15*time.Minute), "reminder-idempotent@example.test")
	runner := notification.NewReminderRunner(db, notification.ReminderConfig{Window: 30 * time.Minute})
	if err := runner.ScanReminders(context.Background()); err != nil {
		t.Fatal(err)
	}
	var firstCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM notification_outbox WHERE event_type='LESSON_REMINDER' AND lesson_id=?`, lessonID).Scan(&firstCount); err != nil {
		t.Fatal(err)
	}
	if firstCount != 1 {
		t.Fatalf("expected 1 reminder after first scan, got %d", firstCount)
	}
	if err := runner.ScanReminders(context.Background()); err != nil {
		t.Fatal(err)
	}
	var secondCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM notification_outbox WHERE event_type='LESSON_REMINDER' AND lesson_id=?`, lessonID).Scan(&secondCount); err != nil {
		t.Fatal(err)
	}
	if secondCount != 1 {
		t.Fatalf("duplicate reminder after re-scan: got %d", secondCount)
	}
	_ = studentID
}

// TestReminderScanWindowBoundary verifies that lessons outside the reminder
// window are not enqueued.
func TestReminderScanWindowBoundary(t *testing.T) {
	db := testDB(t)
	_, farLessonID := seedReminderLesson(t, db, time.Now().UTC().Add(3*time.Hour), "reminder-far@example.test")
	runner := notification.NewReminderRunner(db, notification.ReminderConfig{Window: 30 * time.Minute})
	if err := runner.ScanReminders(context.Background()); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM notification_outbox WHERE event_type='LESSON_REMINDER' AND lesson_id=?`, farLessonID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("lesson outside window should not be enqueued, got %d", count)
	}
}

// TestReminderScanSkipsNonScheduledLessons verifies that CANCELLED or COMPLETED
// lessons are not enqueued.
func TestReminderScanSkipsNonScheduledLessons(t *testing.T) {
	db := testDB(t)
	_, cancelledLessonID := seedReminderLesson(t, db, time.Now().UTC().Add(15*time.Minute), "reminder-cancelled@example.test")
	if _, err := db.Exec(`UPDATE lesson SET status='CANCELLED' WHERE id=?`, cancelledLessonID); err != nil {
		t.Fatal(err)
	}
	runner := notification.NewReminderRunner(db, notification.ReminderConfig{Window: 30 * time.Minute})
	if err := runner.ScanReminders(context.Background()); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM notification_outbox WHERE event_type='LESSON_REMINDER' AND lesson_id=?`, cancelledLessonID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("cancelled lesson should not be enqueued, got %d", count)
	}
}

// TestFailedNotificationRetriesAfterAvailableAt verifies that a FAILED row is
// retried only after available_at and that the lesson stays unchanged.
func TestFailedNotificationRetriesAfterAvailableAt(t *testing.T) {
	db := testDB(t)
	_, lessonID := seedReminderLesson(t, db, time.Now().UTC().Add(15*time.Minute), "retry@example.test")
	// Insert a FAILED row with future available_at and attempts=1.
	if _, err := db.Exec(`INSERT INTO notification_outbox(lesson_id,event_type,recipient_email,subject,html_body,idempotency_key,status,attempts,available_at) VALUES(?,'LESSON_CREATED','retry@example.test','s','b','retry-key','FAILED',1,?)`, lessonID, time.Now().UTC().Add(10*time.Minute)); err != nil {
		t.Fatal(err)
	}
	var lessonStatusBefore string
	if err := db.QueryRow(`SELECT status FROM lesson WHERE id=?`, lessonID).Scan(&lessonStatusBefore); err != nil {
		t.Fatal(err)
	}
	sender := &fakeSender{fail: true}
	if err := notification.ClaimAndSend(context.Background(), repository.NewDB(db), sender); err != nil {
		t.Fatal(err)
	}
	var status string
	var attempts int
	if err := db.QueryRow(`SELECT status,attempts FROM notification_outbox WHERE id=(SELECT max(id) FROM notification_outbox)`).Scan(&status, &attempts); err != nil {
		t.Fatal(err)
	}
	if status != "FAILED" || attempts != 1 || sender.calls != 0 {
		t.Fatalf("future failed row must not be retried: status=%s attempts=%d calls=%d", status, attempts, sender.calls)
	}
	if _, err := db.Exec(`UPDATE notification_outbox SET available_at=? WHERE id=(SELECT max(id) FROM notification_outbox)`, time.Now().UTC().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := notification.ClaimAndSend(context.Background(), repository.NewDB(db), sender); err != nil {
		t.Fatal(err)
	}
	var availableAt time.Time
	if err := db.QueryRow(`SELECT status,attempts,available_at FROM notification_outbox WHERE id=(SELECT max(id) FROM notification_outbox)`).Scan(&status, &attempts, &availableAt); err != nil {
		t.Fatal(err)
	}
	if status != "FAILED" || attempts != 2 || !availableAt.After(time.Now().UTC()) {
		t.Fatalf("failed retry must defer: status=%s attempts=%d availableAt=%s", status, attempts, availableAt)
	}
	var lessonStatusAfter string
	if err := db.QueryRow(`SELECT status FROM lesson WHERE id=?`, lessonID).Scan(&lessonStatusAfter); err != nil {
		t.Fatal(err)
	}
	if lessonStatusBefore != lessonStatusAfter {
		t.Fatalf("lesson status changed: before=%s after=%s", lessonStatusBefore, lessonStatusAfter)
	}
}

// TestNotificationRetryStopsAtThree verifies that after three attempts the
// runner does not attempt another send.
func TestNotificationRetryStopsAtThree(t *testing.T) {
	db := testDB(t)
	if _, err := db.Exec(`INSERT INTO notification_outbox(lesson_id,event_type,recipient_email,subject,html_body,idempotency_key,status,attempts,available_at) VALUES(1,'LESSON_CREATED','cap@example.test','s','b','cap-key','FAILED',3,?)`, time.Now().UTC().Add(-1*time.Minute)); err != nil {
		t.Fatal(err)
	}
	sender := &fakeSender{}
	if err := notification.ClaimAndSend(context.Background(), repository.NewDB(db), sender); err != nil {
		t.Fatal(err)
	}
	if sender.calls != 0 {
		t.Fatalf("runner should not retry after 3 attempts, calls=%d", sender.calls)
	}
}

// TestNotificationErrorIsSanitized verifies that the stored last_error does not
// contain secrets.
func TestNotificationErrorIsSanitized(t *testing.T) {
	db := testDB(t)
	if _, err := db.Exec(`INSERT INTO notification_outbox(lesson_id,event_type,recipient_email,subject,html_body,idempotency_key) VALUES(1,'LESSON_CREATED','san@example.test','s','b','san-key')`); err != nil {
		t.Fatal(err)
	}
	leaky := &fakeSender{fail: true, leakMsg: "Authorization: Bearer secret-token api_key=ABC123 password=hunter2"}
	if err := notification.ClaimAndSend(context.Background(), repository.NewDB(db), leaky); err != nil {
		t.Fatal(err)
	}
	var lastError string
	if err := db.QueryRow(`SELECT coalesce(last_error,'') FROM notification_outbox WHERE id=(SELECT max(id) FROM notification_outbox)`).Scan(&lastError); err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"Bearer", "secret-token", "api_key", "ABC123", "password", "hunter2"} {
		if contains(lastError, secret) {
			t.Fatalf("last_error leaks secret %q: %q", secret, lastError)
		}
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

func seedReminderLesson(t *testing.T, db *sql.DB, startUTC time.Time, studentEmail string) (int64, int64) {
	t.Helper()
	rStu, err := db.Exec(`INSERT INTO student(name,email,timezone,status) VALUES('reminder-student',?,'Asia/Tokyo','ACTIVE')`, studentEmail)
	if err != nil {
		t.Fatal(err)
	}
	studentID, _ := rStu.LastInsertId()
	rDom, err := db.Exec(`INSERT INTO course_domain(name,code,type,sort_order) VALUES('reminder-domain','RD','LANGUAGE',1)`)
	if err != nil {
		t.Fatal(err)
	}
	domainID, _ := rDom.LastInsertId()
	rTrk, err := db.Exec(`INSERT INTO course_track(domain_id,name,code,sort_order) VALUES(?,'reminder-track','RT',1)`, domainID)
	if err != nil {
		t.Fatal(err)
	}
	trackID, _ := rTrk.LastInsertId()
	rLvl, err := db.Exec(`INSERT INTO course_level(track_id,name,code,sort_order) VALUES(?,'reminder-level','RL',1)`, trackID)
	if err != nil {
		t.Fatal(err)
	}
	levelID, _ := rLvl.LastInsertId()
	rEnr, err := db.Exec(`INSERT INTO student_course_enrollment(student_id,domain_id,track_id,current_level_id,enrollment_type,status,charge_per_lesson_amount,lesson_balance,balance_amount) VALUES(?,?,?,?, 'ONE_TO_ONE','ACTIVE',3000,10,30000)`, studentID, domainID, trackID, levelID)
	if err != nil {
		t.Fatal(err)
	}
	enrollmentID, _ := rEnr.LastInsertId()
	rTch, err := db.Exec(`INSERT INTO teacher(name,default_rate_amount,status) VALUES('reminder-teacher',5000,'ACTIVE')`)
	if err != nil {
		t.Fatal(err)
	}
	teacherID, _ := rTch.LastInsertId()
	rAsg, err := db.Exec(`INSERT INTO student_teacher_assignment(enrollment_id,student_id,teacher_id,role_type,status,start_date) VALUES(?,?,?, 'MAIN','ACTIVE','2026-01-01')`, enrollmentID, studentID, teacherID)
	if err != nil {
		t.Fatal(err)
	}
	assignmentID, _ := rAsg.LastInsertId()
	end := startUTC.Add(60 * time.Minute)
	rLes, err := db.Exec(`INSERT INTO lesson(lesson_no,enrollment_id,assignment_id,teacher_id,student_id,scheduled_start_at,scheduled_end_at,duration_min,timezone,meeting_type,status) VALUES('LSN-REM',?,?,?,?,?,?,?,?, 'OFFLINE','SCHEDULED')`, enrollmentID, assignmentID, teacherID, studentID, startUTC, end, 60, "Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	lessonID, _ := rLes.LastInsertId()
	return studentID, lessonID
}

var _ = filepath.Join
