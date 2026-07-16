package directory

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

// validStudentStatuses and validTeacherStatuses mirror migration 003 CHECK
// constraints; the service rejects anything outside these sets before writing.
var validStudentStatuses = map[string]bool{"ACTIVE": true, "PAUSED": true, "ENDED": true}
var validTeacherStatuses = map[string]bool{"ACTIVE": true, "PAUSED": true, "ENDED": true}
var validCapabilityStatuses = map[string]bool{"ACTIVE": true, "PAUSED": true, "ENDED": true}
var validRelationships = map[string]bool{"FATHER": true, "MOTHER": true, "OTHER": true}

// Service is the people-directory application service. It owns authorization,
// validation, state transitions, transaction orchestration and audit. It does
// not depend on net/http.
type Service struct {
	db   repository.DB
	repo *Repository
}

// NewService creates a people-directory service backed by the given DB.
func NewService(db repository.DB) *Service {
	return &Service{db: db, repo: NewRepository()}
}

// actor captures the authenticated principal for audit purposes.
type actor struct {
	id   int64
	role string
}

func fromUser(u httpserver.AuthUser) actor {
	return actor{id: u.UserID, role: u.Role}
}

// authorize confirms the actor is Owner or Operator. Both roles may use the
// M2 business routes (Owner includes Operator).
func (a actor) authorize() error {
	if a.role != "OWNER" && a.role != "OPERATOR" {
		return ErrForbidden
	}
	return nil
}

// ErrForbidden is returned when the actor lacks permission. Callers map this to
// HTTP 403 / 40301.
var ErrForbidden = errors.New("forbidden")

// ---------- Student ----------

// CreateStudent creates a student and writes audit in one transaction.
func (s *Service) CreateStudent(ctx context.Context, u httpserver.AuthUser, w StudentWrite, requestID string) (Student, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Student{}, err
	}
	if err := validateStudentWrite(w, true); err != nil {
		return Student{}, err
	}
	var created Student
	err := s.inTx(ctx, func(tx repository.Tx) error {
		id, err := s.repo.InsertStudent(ctx, tx, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetStudent(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "STUDENT_CREATE", "student", id, map[string]any{"name": created.Name}, requestID)
	})
	if err != nil {
		return Student{}, err
	}
	return created, nil
}

// UpdateStudent patches a student and writes audit in one transaction.
func (s *Service) UpdateStudent(ctx context.Context, u httpserver.AuthUser, id int64, w StudentWrite, requestID string) (Student, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Student{}, err
	}
	if err := validateStudentWrite(w, false); err != nil {
		return Student{}, err
	}
	var updated Student
	err := s.inTx(ctx, func(tx repository.Tx) error {
		if _, err := s.repo.GetStudent(ctx, tx, id); err != nil {
			return err
		}
		if err := s.repo.UpdateStudent(ctx, tx, id, w); err != nil {
			return err
		}
		var err error
		updated, err = s.repo.GetStudent(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "STUDENT_UPDATE", "student", id, map[string]any{"name": updated.Name}, requestID)
	})
	if err != nil {
		return Student{}, err
	}
	return updated, nil
}

// GetStudent returns a student by id.
func (s *Service) GetStudent(ctx context.Context, u httpserver.AuthUser, id int64) (Student, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Student{}, err
	}
	return s.repo.GetStudent(ctx, s.db, id)
}

// ListStudents returns a page of students.
func (s *Service) ListStudents(ctx context.Context, u httpserver.AuthUser, status, search string, page, pageSize int) (httpserver.ListData, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return httpserver.ListData{}, err
	}
	if status != "" && !validStudentStatuses[status] {
		return httpserver.ListData{}, ErrInvalidState
	}
	items, total, err := s.repo.ListStudents(ctx, s.db, status, search, pageSize, (page-1)*pageSize)
	if err != nil {
		return httpserver.ListData{}, err
	}
	return httpserver.NewListData(items, total, httpserver.PageQuery{Page: page, PageSize: pageSize}), nil
}

func validateStudentWrite(w StudentWrite, isCreate bool) error {
	if isCreate {
		if w.Name == nil || strings.TrimSpace(*w.Name) == "" {
			return ErrInvalidState
		}
	}
	if w.Name != nil && strings.TrimSpace(*w.Name) == "" {
		return ErrInvalidState
	}
	if w.Email != nil {
		if e := strings.TrimSpace(*w.Email); e != "" && !isValidEmail(e) {
			return ErrInvalidState
		}
	}
	if w.Status != nil && !validStudentStatuses[*w.Status] {
		return ErrInvalidState
	}
	return nil
}

// ---------- Parent ----------

// CreateParent creates a parent under a student with audit in one transaction.
func (s *Service) CreateParent(ctx context.Context, u httpserver.AuthUser, studentID int64, w ParentWrite, requestID string) (Parent, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Parent{}, err
	}
	if err := validateParentWrite(w, true); err != nil {
		return Parent{}, err
	}
	var created Parent
	err := s.inTx(ctx, func(tx repository.Tx) error {
		if _, err := s.repo.GetStudent(ctx, tx, studentID); err != nil {
			return err
		}
		id, err := s.repo.InsertParent(ctx, tx, studentID, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetParent(ctx, tx, studentID, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "PARENT_CREATE", "parent", id, map[string]any{"studentId": studentID, "name": created.Name}, requestID)
	})
	if err != nil {
		return Parent{}, err
	}
	return created, nil
}

// UpdateParent patches a parent scoped to a student with audit.
func (s *Service) UpdateParent(ctx context.Context, u httpserver.AuthUser, studentID, parentID int64, w ParentWrite, requestID string) (Parent, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Parent{}, err
	}
	if err := validateParentWrite(w, false); err != nil {
		return Parent{}, err
	}
	var updated Parent
	err := s.inTx(ctx, func(tx repository.Tx) error {
		if _, err := s.repo.GetParent(ctx, tx, studentID, parentID); err != nil {
			return err
		}
		if err := s.repo.UpdateParent(ctx, tx, studentID, parentID, w); err != nil {
			return err
		}
		var err error
		updated, err = s.repo.GetParent(ctx, tx, studentID, parentID)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "PARENT_UPDATE", "parent", parentID, map[string]any{"studentId": studentID, "name": updated.Name}, requestID)
	})
	if err != nil {
		return Parent{}, err
	}
	return updated, nil
}

// ListParents returns parents for a student.
func (s *Service) ListParents(ctx context.Context, u httpserver.AuthUser, studentID int64) ([]Parent, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetStudent(ctx, s.db, studentID); err != nil {
		return nil, err
	}
	return s.repo.ListParents(ctx, s.db, studentID)
}

func validateParentWrite(w ParentWrite, isCreate bool) error {
	if isCreate {
		if w.Name == nil || strings.TrimSpace(*w.Name) == "" {
			return ErrInvalidState
		}
	}
	if w.Name != nil && strings.TrimSpace(*w.Name) == "" {
		return ErrInvalidState
	}
	if w.Relationship != nil && *w.Relationship != "" && !validRelationships[*w.Relationship] {
		return ErrInvalidState
	}
	return nil
}

// ---------- Teacher ----------

// CreateTeacher creates a teacher with audit in one transaction.
func (s *Service) CreateTeacher(ctx context.Context, u httpserver.AuthUser, w TeacherWrite, requestID string) (Teacher, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Teacher{}, err
	}
	if err := validateTeacherWrite(w, true); err != nil {
		return Teacher{}, err
	}
	var created Teacher
	err := s.inTx(ctx, func(tx repository.Tx) error {
		id, err := s.repo.InsertTeacher(ctx, tx, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetTeacher(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "TEACHER_CREATE", "teacher", id, map[string]any{"name": created.Name}, requestID)
	})
	if err != nil {
		return Teacher{}, err
	}
	return created, nil
}

// UpdateTeacher patches a teacher with audit.
func (s *Service) UpdateTeacher(ctx context.Context, u httpserver.AuthUser, id int64, w TeacherWrite, requestID string) (Teacher, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Teacher{}, err
	}
	if err := validateTeacherWrite(w, false); err != nil {
		return Teacher{}, err
	}
	var updated Teacher
	err := s.inTx(ctx, func(tx repository.Tx) error {
		if _, err := s.repo.GetTeacher(ctx, tx, id); err != nil {
			return err
		}
		if err := s.repo.UpdateTeacher(ctx, tx, id, w); err != nil {
			return err
		}
		var err error
		updated, err = s.repo.GetTeacher(ctx, tx, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "TEACHER_UPDATE", "teacher", id, map[string]any{"name": updated.Name}, requestID)
	})
	if err != nil {
		return Teacher{}, err
	}
	return updated, nil
}

// GetTeacher returns a teacher by id.
func (s *Service) GetTeacher(ctx context.Context, u httpserver.AuthUser, id int64) (Teacher, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Teacher{}, err
	}
	return s.repo.GetTeacher(ctx, s.db, id)
}

// ListTeachers returns a page of teachers.
func (s *Service) ListTeachers(ctx context.Context, u httpserver.AuthUser, status, search string, page, pageSize int) (httpserver.ListData, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return httpserver.ListData{}, err
	}
	if status != "" && !validTeacherStatuses[status] {
		return httpserver.ListData{}, ErrInvalidState
	}
	items, total, err := s.repo.ListTeachers(ctx, s.db, status, search, pageSize, (page-1)*pageSize)
	if err != nil {
		return httpserver.ListData{}, err
	}
	return httpserver.NewListData(items, total, httpserver.PageQuery{Page: page, PageSize: pageSize}), nil
}

func validateTeacherWrite(w TeacherWrite, isCreate bool) error {
	if isCreate {
		if w.Name == nil || strings.TrimSpace(*w.Name) == "" {
			return ErrInvalidState
		}
	}
	if w.Name != nil && strings.TrimSpace(*w.Name) == "" {
		return ErrInvalidState
	}
	if w.Status != nil && !validTeacherStatuses[*w.Status] {
		return ErrInvalidState
	}
	return nil
}

// ---------- Capability ----------

// CreateCapability creates a teacher capability after verifying the
// domain/track/level hierarchy, with audit in one transaction.
func (s *Service) CreateCapability(ctx context.Context, u httpserver.AuthUser, teacherID int64, w CapabilityWrite, requestID string) (Capability, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Capability{}, err
	}
	if err := validateCapabilityWrite(w, true); err != nil {
		return Capability{}, err
	}
	var created Capability
	err := s.inTx(ctx, func(tx repository.Tx) error {
		if _, err := s.repo.GetTeacher(ctx, tx, teacherID); err != nil {
			return err
		}
		if err := s.repo.VerifyHierarchy(ctx, tx, *w.DomainID, *w.TrackID, *w.LevelID); err != nil {
			return err
		}
		id, err := s.repo.InsertCapability(ctx, tx, teacherID, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetCapability(ctx, tx, teacherID, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "CAPABILITY_CREATE", "teacher_capability", id, map[string]any{"teacherId": teacherID, "trackId": created.TrackID, "levelId": created.LevelID}, requestID)
	})
	if err != nil {
		return Capability{}, err
	}
	return created, nil
}

// UpdateCapability patches a capability. Ending a capability sets effective_to
// and status=ENDED without deleting the row (history preserved).
func (s *Service) UpdateCapability(ctx context.Context, u httpserver.AuthUser, teacherID, capID int64, w CapabilityWrite, requestID string) (Capability, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Capability{}, err
	}
	if err := validateCapabilityWrite(w, false); err != nil {
		return Capability{}, err
	}
	var updated Capability
	err := s.inTx(ctx, func(tx repository.Tx) error {
		existing, err := s.repo.GetCapability(ctx, tx, teacherID, capID)
		if err != nil {
			return err
		}
		// If track/level/domain are being changed, re-verify hierarchy.
		domainID := existing.DomainID
		trackID := existing.TrackID
		levelID := existing.LevelID
		if w.DomainID != nil {
			domainID = *w.DomainID
		}
		if w.TrackID != nil {
			trackID = *w.TrackID
		}
		if w.LevelID != nil {
			levelID = *w.LevelID
		}
		if w.DomainID != nil || w.TrackID != nil || w.LevelID != nil {
			if err := s.repo.VerifyHierarchy(ctx, tx, domainID, trackID, levelID); err != nil {
				return err
			}
		}
		// Ending a capability: force effective_to when status moves to ENDED.
		if w.Status != nil && *w.Status == "ENDED" && (w.EffectiveTo == nil || *w.EffectiveTo == "") {
			today := time.Now().UTC().Format("2006-01-02")
			w.EffectiveTo = &today
		}
		if err := s.repo.UpdateCapability(ctx, tx, teacherID, capID, w); err != nil {
			return err
		}
		updated, err = s.repo.GetCapability(ctx, tx, teacherID, capID)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "CAPABILITY_UPDATE", "teacher_capability", capID, map[string]any{"teacherId": teacherID, "status": updated.Status}, requestID)
	})
	if err != nil {
		return Capability{}, err
	}
	return updated, nil
}

// ListCapabilities returns capabilities for a teacher.
func (s *Service) ListCapabilities(ctx context.Context, u httpserver.AuthUser, teacherID int64) ([]Capability, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetTeacher(ctx, s.db, teacherID); err != nil {
		return nil, err
	}
	return s.repo.ListCapabilities(ctx, s.db, teacherID)
}

func validateCapabilityWrite(w CapabilityWrite, isCreate bool) error {
	if isCreate {
		if w.DomainID == nil || w.TrackID == nil || w.LevelID == nil {
			return ErrInvalidState
		}
	}
	if w.Status != nil && !validCapabilityStatuses[*w.Status] {
		return ErrInvalidState
	}
	return nil
}

// ---------- Availability ----------

// CreateAvailability creates a teacher availability slot after validating time
// formats and ranges, with audit in one transaction.
func (s *Service) CreateAvailability(ctx context.Context, u httpserver.AuthUser, teacherID int64, w AvailabilityWrite, requestID string) (Availability, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Availability{}, err
	}
	if err := validateAvailabilityWrite(w, true); err != nil {
		return Availability{}, err
	}
	var created Availability
	err := s.inTx(ctx, func(tx repository.Tx) error {
		if _, err := s.repo.GetTeacher(ctx, tx, teacherID); err != nil {
			return err
		}
		id, err := s.repo.InsertAvailability(ctx, tx, teacherID, w)
		if err != nil {
			return err
		}
		created, err = s.repo.GetAvailability(ctx, tx, teacherID, id)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "AVAILABILITY_CREATE", "teacher_availability", id, map[string]any{"teacherId": teacherID, "weekday": created.Weekday}, requestID)
	})
	if err != nil {
		return Availability{}, err
	}
	return created, nil
}

// UpdateAvailability patches an availability slot with audit.
func (s *Service) UpdateAvailability(ctx context.Context, u httpserver.AuthUser, teacherID, availID int64, w AvailabilityWrite, requestID string) (Availability, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return Availability{}, err
	}
	if err := validateAvailabilityWrite(w, false); err != nil {
		return Availability{}, err
	}
	var updated Availability
	err := s.inTx(ctx, func(tx repository.Tx) error {
		existing, err := s.repo.GetAvailability(ctx, tx, teacherID, availID)
		if err != nil {
			return err
		}
		weekday := existing.Weekday
		start := existing.StartTime
		end := existing.EndTime
		if w.Weekday != nil {
			weekday = *w.Weekday
		}
		if w.StartTime != nil {
			start = *w.StartTime
		}
		if w.EndTime != nil {
			end = *w.EndTime
		}
		if err := validateTimeRange(weekday, start, end); err != nil {
			return err
		}
		if err := validateEffectiveRange(ptrString(w.EffectiveFrom), ptrString(w.EffectiveTo)); err != nil {
			return err
		}
		if err := s.repo.UpdateAvailability(ctx, tx, teacherID, availID, w); err != nil {
			return err
		}
		updated, err = s.repo.GetAvailability(ctx, tx, teacherID, availID)
		if err != nil {
			return err
		}
		return s.audit(ctx, tx, a.id, "AVAILABILITY_UPDATE", "teacher_availability", availID, map[string]any{"teacherId": teacherID, "weekday": updated.Weekday}, requestID)
	})
	if err != nil {
		return Availability{}, err
	}
	return updated, nil
}

// ListAvailability returns availability slots for a teacher.
func (s *Service) ListAvailability(ctx context.Context, u httpserver.AuthUser, teacherID int64) ([]Availability, error) {
	a := fromUser(u)
	if err := a.authorize(); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetTeacher(ctx, s.db, teacherID); err != nil {
		return nil, err
	}
	return s.repo.ListAvailability(ctx, s.db, teacherID)
}

func validateAvailabilityWrite(w AvailabilityWrite, isCreate bool) error {
	if isCreate {
		if w.Weekday == nil || w.StartTime == nil || w.EndTime == nil {
			return ErrInvalidState
		}
	}
	weekday := -1
	start, end := "", ""
	if w.Weekday != nil {
		weekday = *w.Weekday
	}
	if w.StartTime != nil {
		start = *w.StartTime
	}
	if w.EndTime != nil {
		end = *w.EndTime
	}
	if w.Weekday != nil || isCreate {
		if weekday < 0 || weekday > 6 {
			return ErrInvalidState
		}
	}
	if (w.StartTime != nil || w.EndTime != nil || isCreate) && start != "" && end != "" {
		if err := validateTimeRange(weekday, start, end); err != nil {
			return err
		}
	}
	if err := validateEffectiveRange(ptrString(w.EffectiveFrom), ptrString(w.EffectiveTo)); err != nil {
		return err
	}
	return nil
}

func validateTimeRange(weekday int, start, end string) error {
	if weekday < 0 || weekday > 6 {
		return ErrInvalidState
	}
	if !isValidTime(start) || !isValidTime(end) {
		return ErrInvalidState
	}
	if start >= end {
		return ErrInvalidState
	}
	return nil
}

func validateEffectiveRange(from, to string) error {
	if from != "" && !isValidDate(from) {
		return ErrInvalidState
	}
	if to != "" && !isValidDate(to) {
		return ErrInvalidState
	}
	if from != "" && to != "" && from > to {
		return ErrInvalidState
	}
	return nil
}

// isValidTime accepts HH:MM 24-hour format.
func isValidTime(s string) bool {
	if len(s) != 5 || s[2] != ':' {
		return false
	}
	hh, mm := s[0:2], s[3:5]
	for _, ch := range hh + mm {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	if hh < "00" || hh > "23" || mm < "00" || mm > "59" {
		return false
	}
	return true
}

// isValidDate accepts YYYY-MM-DD.
func isValidDate(s string) bool {
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return false
	}
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return false
	}
	return true
}

// isValidEmail performs a minimal structural email check.
func isValidEmail(s string) bool {
	at := strings.Index(s, "@")
	if at <= 0 || at == len(s)-1 {
		return false
	}
	return strings.Contains(s[at+1:], ".")
}

// ---------- shared helpers ----------

// inTx runs fn inside a single database transaction. The transaction is the
// only multi-table write boundary: every business write and the audit row share
// it, so any error rolls back both. A nil error from fn commits; otherwise the
// transaction is rolled back.
func (s *Service) inTx(ctx context.Context, fn func(repository.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// audit writes a non-sensitive operation_log row within the transaction. The
// actor name is loaded inside the same tx so a missing account is detected
// before commit.
func (s *Service) audit(ctx context.Context, tx repository.Tx, actorID int64, action, targetType string, targetID int64, detail map[string]any, requestID string) error {
	name, err := repository.ActorName(tx, ctx, actorID)
	if err != nil {
		return err
	}
	return repository.InsertAuditLog(tx, ctx, actorID, name, action, targetType, targetID, detail, requestID)
}
