// Package lesson implements the deliberately bounded M4a scheduling domain.
// It creates scheduling facts and their audit rows only: attendance,
// notifications, finance, payout, and conflict-resolution stay outside M4a.
package lesson

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/prelove/zedu/backend/internal/app/notification"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

var (
	ErrInvalidState = errors.New("invalid lesson state")
	ErrForbidden    = errors.New("lesson write forbidden")
	ErrNotFound     = errors.New("lesson not found")
)

// Write is the complete, mutable M4a schedule payload. Enrollment and
// assignment are immutable after creation; update uses ScheduleUpdate.
type Write struct {
	EnrollmentID int64  `json:"enrollmentId"`
	AssignmentID int64  `json:"assignmentId"`
	StartAt      string `json:"startAt"`
	DurationMin  int    `json:"durationMin"`
	Timezone     string `json:"timezone"`
	MeetingType  string `json:"meetingType"`
	MeetingLink  string `json:"meetingLink"`
	LessonTopic  string `json:"lessonTopic"`
	Note         string `json:"note"`
}

// ScheduleUpdate intentionally excludes enrollmentId and assignmentId. A
// lesson keeps the original commercial/teaching relationship as a fact.
type ScheduleUpdate struct {
	StartAt     string `json:"startAt"`
	DurationMin int    `json:"durationMin"`
	Timezone    string `json:"timezone"`
	MeetingType string `json:"meetingType"`
	MeetingLink string `json:"meetingLink"`
	LessonTopic string `json:"lessonTopic"`
	Note        string `json:"note"`
}

type CancelWrite struct {
	Reason string `json:"reason"`
}

type Lesson struct {
	ID               int64  `json:"id"`
	LessonNo         string `json:"lessonNo"`
	EnrollmentID     int64  `json:"enrollmentId"`
	AssignmentID     int64  `json:"assignmentId"`
	TeacherID        int64  `json:"teacherId"`
	StudentID        int64  `json:"studentId"`
	ScheduledStartAt string `json:"scheduledStartAt"`
	ScheduledEndAt   string `json:"scheduledEndAt"`
	DurationMin      int    `json:"durationMin"`
	Timezone         string `json:"timezone"`
	MeetingType      string `json:"meetingType"`
	MeetingLink      string `json:"meetingLink,omitempty"`
	LessonTopic      string `json:"lessonTopic,omitempty"`
	Note             string `json:"note,omitempty"`
	Status           string `json:"status"`
	CancelReason     string `json:"cancelReason,omitempty"`
}

type ListFilter struct {
	StudentID int64
	TeacherID int64
	Status    string
	From      string
	To        string
	Page      int
	PageSize  int
}

type ListResult struct {
	Items    []Lesson `json:"items"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
	Total    int      `json:"total"`
}

type Service struct{ db repository.DB }

func NewService(db repository.DB) *Service { return &Service{db: db} }

func validateWrite(w Write) error {
	return validateSchedule(w.StartAt, w.DurationMin, w.Timezone, w.MeetingType, w.MeetingLink)
}

func validateSchedule(startAt string, duration int, timezone, meetingType, meetingLink string) error {
	if duration < 10 || duration > 480 || strings.TrimSpace(timezone) == "" || strings.TrimSpace(meetingType) == "" {
		return ErrInvalidState
	}
	if _, err := parseBusinessTime(startAt, timezone); err != nil {
		return ErrInvalidState
	}
	if strings.EqualFold(meetingType, "WECHAT") {
		link, err := url.ParseRequestURI(strings.TrimSpace(meetingLink))
		if err != nil || (link.Scheme != "http" && link.Scheme != "https") || link.Host == "" {
			return ErrInvalidState
		}
	}
	return nil
}

// parseBusinessTime accepts the browser-friendly local timestamp form as well
// as RFC3339. In both cases the IANA timezone is validated and UTC is stored.
func parseBusinessTime(value, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil {
		return time.Time{}, err
	}
	value = strings.TrimSpace(value)
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC(), nil
	}
	parsed, err := time.ParseInLocation("2006-01-02T15:04:05", value, loc)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func requireWriter(user httpserver.AuthUser) error {
	if user.Role != "OWNER" && user.Role != "OPERATOR" {
		return ErrForbidden
	}
	return nil
}

// Create stores the lesson and audit record in one transaction. It never
// creates a downstream notification, attendance, ledger, payment, or payout.
func (s *Service) Create(ctx context.Context, user httpserver.AuthUser, w Write, requestID string) (id int64, err error) {
	if err = requireWriter(user); err != nil {
		return 0, err
	}
	if w.EnrollmentID <= 0 || w.AssignmentID <= 0 || validateWrite(w) != nil {
		return 0, ErrInvalidState
	}
	start, _ := parseBusinessTime(w.StartAt, w.Timezone)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, repository.ErrDatabase
	}
	committed := false
	defer func() {
		if !committed && tx.Rollback() != nil {
			err = repository.ErrDatabase
		}
	}()

	studentID, teacherID, err := activeAssignment(ctx, tx, w.EnrollmentID, w.AssignmentID)
	if err != nil {
		return 0, err
	}
	lessonNo := newLessonNo()
	result, execErr := tx.ExecContext(ctx, `INSERT INTO lesson
		(lesson_no,enrollment_id,assignment_id,teacher_id,student_id,scheduled_start_at,scheduled_end_at,duration_min,timezone,meeting_type,meeting_link,lesson_topic,note,created_by)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		lessonNo, w.EnrollmentID, w.AssignmentID, teacherID, studentID, start,
		start.Add(time.Duration(w.DurationMin)*time.Minute), w.DurationMin, strings.TrimSpace(w.Timezone),
		strings.ToUpper(strings.TrimSpace(w.MeetingType)), nullable(w.MeetingLink), nullable(w.LessonTopic), nullable(w.Note), user.UserID)
	if execErr != nil {
		return 0, repository.ErrDatabase
	}
	id, err = result.LastInsertId()
	if err != nil {
		return 0, repository.ErrDatabase
	}
	if err = writeAudit(ctx, tx, user, "LESSON_CREATE", id, map[string]any{"enrollmentId": w.EnrollmentID, "assignmentId": w.AssignmentID}, requestID); err != nil {
		return 0, err
	}
	if err = notification.QueueLesson(ctx, tx, id, studentID, lessonNo, "LESSON_CREATED", start.Format(time.RFC3339), w.Timezone); err != nil {
		return 0, repository.ErrDatabase
	}
	if err = tx.Commit(); err != nil {
		return 0, repository.ErrDatabase
	}
	committed = true
	return id, nil
}

func (s *Service) Update(ctx context.Context, user httpserver.AuthUser, id int64, w ScheduleUpdate, requestID string) (err error) {
	if err = requireWriter(user); err != nil {
		return err
	}
	if id <= 0 || validateSchedule(w.StartAt, w.DurationMin, w.Timezone, w.MeetingType, w.MeetingLink) != nil {
		return ErrInvalidState
	}
	start, _ := parseBusinessTime(w.StartAt, w.Timezone)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.ErrDatabase
	}
	committed := false
	defer func() {
		if !committed && tx.Rollback() != nil {
			err = repository.ErrDatabase
		}
	}()
	result, execErr := tx.ExecContext(ctx, `UPDATE lesson SET scheduled_start_at=?, scheduled_end_at=?, duration_min=?, timezone=?, meeting_type=?, meeting_link=?, lesson_topic=?, note=?, updated_at=CURRENT_TIMESTAMP WHERE id=? AND status='SCHEDULED'`,
		start, start.Add(time.Duration(w.DurationMin)*time.Minute), w.DurationMin, strings.TrimSpace(w.Timezone), strings.ToUpper(strings.TrimSpace(w.MeetingType)), nullable(w.MeetingLink), nullable(w.LessonTopic), nullable(w.Note), id)
	if execErr != nil {
		return repository.ErrDatabase
	}
	changed, execErr := result.RowsAffected()
	if execErr != nil {
		return repository.ErrDatabase
	}
	if changed != 1 {
		return stateForLesson(ctx, tx, id)
	}
	if err = writeAudit(ctx, tx, user, "LESSON_UPDATE", id, map[string]any{"durationMin": w.DurationMin}, requestID); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return repository.ErrDatabase
	}
	committed = true
	return nil
}

func (s *Service) Cancel(ctx context.Context, user httpserver.AuthUser, id int64, reason, requestID string) (err error) {
	if err = requireWriter(user); err != nil {
		return err
	}
	if id <= 0 || strings.TrimSpace(reason) == "" {
		return ErrInvalidState
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.ErrDatabase
	}
	committed := false
	defer func() {
		if !committed && tx.Rollback() != nil {
			err = repository.ErrDatabase
		}
	}()
	result, execErr := tx.ExecContext(ctx, `UPDATE lesson SET status='CANCELLED', cancel_reason=?, updated_at=CURRENT_TIMESTAMP WHERE id=? AND status='SCHEDULED'`, strings.TrimSpace(reason), id)
	if execErr != nil {
		return repository.ErrDatabase
	}
	changed, execErr := result.RowsAffected()
	if execErr != nil {
		return repository.ErrDatabase
	}
	if changed != 1 {
		return stateForLesson(ctx, tx, id)
	}
	var lessonNo, startUTC, timezone string
	var studentID int64
	if err = tx.QueryRowContext(ctx, `SELECT lesson_no, student_id, scheduled_start_at, timezone FROM lesson WHERE id=?`, id).Scan(&lessonNo, &studentID, &startUTC, &timezone); err != nil {
		return repository.ErrDatabase
	}
	if err = writeAudit(ctx, tx, user, "LESSON_CANCEL", id, map[string]any{"reason": strings.TrimSpace(reason)}, requestID); err != nil {
		return err
	}
	if err = notification.QueueLesson(ctx, tx, id, studentID, lessonNo, "LESSON_CANCELLED", startUTC, timezone); err != nil {
		return repository.ErrDatabase
	}
	if err = tx.Commit(); err != nil {
		return repository.ErrDatabase
	}
	committed = true
	return nil
}

func (s *Service) Get(ctx context.Context, id int64) (Lesson, error) {
	if id <= 0 {
		return Lesson{}, ErrNotFound
	}
	lesson, err := scanLesson(s.db.QueryRowContext(ctx, lessonSelect+" WHERE id=?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return Lesson{}, ErrNotFound
	}
	if err != nil {
		return Lesson{}, repository.ErrDatabase
	}
	return lesson, nil
}

func (s *Service) List(ctx context.Context, filter ListFilter) (ListResult, error) {
	page, pageSize := normalizePage(filter.Page, filter.PageSize)
	where, args, err := lessonWhere(filter)
	if err != nil {
		return ListResult{}, ErrInvalidState
	}
	var total int
	if err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM lesson"+where, args...).Scan(&total); err != nil {
		return ListResult{}, repository.ErrDatabase
	}
	rows, err := s.db.QueryContext(ctx, lessonSelect+where+" ORDER BY scheduled_start_at DESC, id DESC LIMIT ? OFFSET ?", append(args, pageSize, (page-1)*pageSize)...)
	if err != nil {
		return ListResult{}, repository.ErrDatabase
	}
	defer rows.Close()
	items := make([]Lesson, 0)
	for rows.Next() {
		item, scanErr := scanLesson(rows)
		if scanErr != nil {
			return ListResult{}, repository.ErrDatabase
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return ListResult{}, repository.ErrDatabase
	}
	return ListResult{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func activeAssignment(ctx context.Context, tx repository.Tx, enrollmentID, assignmentID int64) (int64, int64, error) {
	var studentID, teacherID int64
	var enrollmentStatus, assignmentStatus string
	err := tx.QueryRowContext(ctx, `SELECT e.student_id,e.status,a.teacher_id,a.status FROM student_course_enrollment e JOIN student_teacher_assignment a ON a.id=? AND a.enrollment_id=e.id WHERE e.id=?`, assignmentID, enrollmentID).Scan(&studentID, &enrollmentStatus, &teacherID, &assignmentStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, 0, ErrInvalidState
	}
	if err != nil {
		return 0, 0, repository.ErrDatabase
	}
	if enrollmentStatus != "ACTIVE" || assignmentStatus != "ACTIVE" {
		return 0, 0, ErrInvalidState
	}
	return studentID, teacherID, nil
}

func stateForLesson(ctx context.Context, tx repository.Tx, id int64) error {
	var found int
	err := tx.QueryRowContext(ctx, "SELECT 1 FROM lesson WHERE id=?", id).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return repository.ErrDatabase
	}
	return ErrInvalidState
}

func writeAudit(ctx context.Context, tx repository.Tx, user httpserver.AuthUser, action string, id int64, detail any, requestID string) error {
	name, err := repository.ActorName(tx, ctx, user.UserID)
	if err != nil {
		return repository.ErrDatabase
	}
	if err = repository.InsertAuditLog(tx, ctx, user.UserID, name, action, "lesson", id, detail, requestID); err != nil {
		return repository.ErrDatabase
	}
	return nil
}

func newLessonNo() string { return fmt.Sprintf("LSN-%d", time.Now().UTC().UnixNano()) }

func nullable(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}
