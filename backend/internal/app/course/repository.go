// Package course implements the course-dictionary, enrollment and assignment
// application domain for M2 stage B/C.
//
// Layering (per M2 design): HTTP handler -> application service -> repository.
// The repository only executes parameterized SQL; the service owns
// authorization, validation, hierarchy checks, transaction orchestration and
// audit; the handler only decodes/encodes.
package course

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/prelove/zedu/backend/internal/repository"
)

// ---------- Domain types ----------

// CourseDomain is the read model for a course domain.
type CourseDomain struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Type      string    `json:"type"`
	SortOrder int       `json:"sortOrder"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Track is the read model for a course track.
type Track struct {
	ID        int64     `json:"id"`
	DomainID  int64     `json:"domainId"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	SortOrder int       `json:"sortOrder"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Level is the read model for a course level.
type Level struct {
	ID                     int64     `json:"id"`
	TrackID                int64     `json:"trackId"`
	Name                   string    `json:"name"`
	Code                   string    `json:"code"`
	SortOrder              int       `json:"sortOrder"`
	MinAge                 *int      `json:"minAge,omitempty"`
	MaxAge                 *int      `json:"maxAge,omitempty"`
	MinLessonHours         *float64  `json:"minLessonHours,omitempty"`
	RecommendedLessonHours *float64  `json:"recommendedLessonHours,omitempty"`
	Enabled                bool      `json:"enabled"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}

// CapabilityTag is the read model for a skill tag.
type CapabilityTag struct {
	ID        int64     `json:"id"`
	DomainID  int64     `json:"domainId"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	SortOrder int       `json:"sortOrder"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
}

// ---------- Write payloads ----------

// DomainWrite captures create/update fields for a course domain.
type DomainWrite struct {
	Name      *string `json:"name,omitempty"`
	Code      *string `json:"code,omitempty"`
	Type      *string `json:"type,omitempty"`
	SortOrder *int    `json:"sortOrder,omitempty"`
	Enabled   *bool   `json:"enabled,omitempty"`
}

// TrackWrite captures create/update fields for a track.
type TrackWrite struct {
	DomainID  *int64  `json:"domainId,omitempty"`
	Name      *string `json:"name,omitempty"`
	Code      *string `json:"code,omitempty"`
	SortOrder *int    `json:"sortOrder,omitempty"`
	Enabled   *bool   `json:"enabled,omitempty"`
}

// LevelWrite captures create/update fields for a level.
type LevelWrite struct {
	TrackID                *int64   `json:"trackId,omitempty"`
	Name                   *string  `json:"name,omitempty"`
	Code                   *string  `json:"code,omitempty"`
	SortOrder              *int     `json:"sortOrder,omitempty"`
	MinAge                 *int     `json:"minAge,omitempty"`
	MaxAge                 *int     `json:"maxAge,omitempty"`
	MinLessonHours         *float64 `json:"minLessonHours,omitempty"`
	RecommendedLessonHours *float64 `json:"recommendedLessonHours,omitempty"`
	Enabled                *bool    `json:"enabled,omitempty"`
}

// TagWrite captures create/update fields for a capability tag.
type TagWrite struct {
	DomainID  *int64  `json:"domainId,omitempty"`
	Name      *string `json:"name,omitempty"`
	Code      *string `json:"code,omitempty"`
	SortOrder *int    `json:"sortOrder,omitempty"`
	Enabled   *bool   `json:"enabled,omitempty"`
}

// ---------- Sentinel errors ----------

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalidState = errors.New("invalid state")
	ErrForbidden    = errors.New("forbidden")
)

var validDomainTypes = map[string]bool{
	"LANGUAGE": true, "K12": true, "SPORT": true, "ART": true,
	"ACADEMIC": true, "CERTIFICATE": true, "OTHER": true,
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// ---------- Repository ----------

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

// ---------- Domain ----------

func (r *Repository) ListDomains(ctx context.Context, exec repository.Executor, search string, limit, offset int) ([]CourseDomain, int, error) {
	where := ""
	var args []any
	if search != "" {
		where = "WHERE name LIKE ? OR code LIKE ?"
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM course_domain "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := exec.QueryContext(ctx, "SELECT id, name, code, type, sort_order, enabled, created_at, updated_at FROM course_domain "+where+" ORDER BY sort_order ASC, id ASC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []CourseDomain
	for rows.Next() {
		var d CourseDomain
		var enabled int
		if err := rows.Scan(&d.ID, &d.Name, &d.Code, &d.Type, &d.SortOrder, &enabled, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		d.Enabled = enabled != 0
		out = append(out, d)
	}
	return out, total, rows.Err()
}

func (r *Repository) GetDomain(ctx context.Context, exec repository.Executor, id int64) (CourseDomain, error) {
	var d CourseDomain
	var enabled int
	err := exec.QueryRowContext(ctx, `SELECT id, name, code, type, sort_order, enabled, created_at, updated_at FROM course_domain WHERE id = ?`, id).
		Scan(&d.ID, &d.Name, &d.Code, &d.Type, &d.SortOrder, &enabled, &d.CreatedAt, &d.UpdatedAt)
	if err == sql.ErrNoRows {
		return CourseDomain{}, ErrNotFound
	}
	if err != nil {
		return CourseDomain{}, err
	}
	d.Enabled = enabled != 0
	return d, nil
}

func (r *Repository) InsertDomain(ctx context.Context, exec repository.Executor, w DomainWrite) (int64, error) {
	res, err := exec.ExecContext(ctx,
		`INSERT INTO course_domain (name, code, type, sort_order, enabled) VALUES (?, ?, ?, ?, ?)`,
		ptrString(w.Name), ptrString(w.Code), defaultStr(ptrString(w.Type), "OTHER"), ptrIntOr(w.SortOrder, 0), boolToInt(w.Enabled, true),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrConflict
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	return id, err
}

func (r *Repository) UpdateDomain(ctx context.Context, exec repository.Executor, id int64, w DomainWrite) error {
	sets, args := buildDomainSets(w)
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	res, err := exec.ExecContext(ctx, "UPDATE course_domain SET "+strings.Join(sets, ", ")+", updated_at = CURRENT_TIMESTAMP WHERE id = ?", args...)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func buildDomainSets(w DomainWrite) ([]string, []any) {
	var sets []string
	var args []any
	if w.Name != nil {
		sets, args = append(sets, "name = ?"), append(args, *w.Name)
	}
	if w.Code != nil {
		sets, args = append(sets, "code = ?"), append(args, *w.Code)
	}
	if w.Type != nil {
		sets, args = append(sets, "type = ?"), append(args, *w.Type)
	}
	if w.SortOrder != nil {
		sets, args = append(sets, "sort_order = ?"), append(args, *w.SortOrder)
	}
	if w.Enabled != nil {
		sets, args = append(sets, "enabled = ?"), append(args, boolToInt(w.Enabled, false))
	}
	return sets, args
}

// ---------- Track ----------

func (r *Repository) ListTracks(ctx context.Context, exec repository.Executor, domainID int64, limit, offset int) ([]Track, int, error) {
	where := "WHERE 1=1"
	var args []any
	if domainID > 0 {
		where = "WHERE domain_id = ?"
		args = append(args, domainID)
	}
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM course_track "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := exec.QueryContext(ctx, "SELECT id, domain_id, name, code, sort_order, enabled, created_at, updated_at FROM course_track "+where+" ORDER BY sort_order ASC, id ASC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []Track
	for rows.Next() {
		var t Track
		var enabled int
		if err := rows.Scan(&t.ID, &t.DomainID, &t.Name, &t.Code, &t.SortOrder, &enabled, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		t.Enabled = enabled != 0
		out = append(out, t)
	}
	return out, total, rows.Err()
}

func (r *Repository) GetTrack(ctx context.Context, exec repository.Executor, id int64) (Track, error) {
	var t Track
	var enabled int
	err := exec.QueryRowContext(ctx, `SELECT id, domain_id, name, code, sort_order, enabled, created_at, updated_at FROM course_track WHERE id = ?`, id).
		Scan(&t.ID, &t.DomainID, &t.Name, &t.Code, &t.SortOrder, &enabled, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return Track{}, ErrNotFound
	}
	if err != nil {
		return Track{}, err
	}
	t.Enabled = enabled != 0
	return t, nil
}

func (r *Repository) InsertTrack(ctx context.Context, exec repository.Executor, w TrackWrite) (int64, error) {
	res, err := exec.ExecContext(ctx,
		`INSERT INTO course_track (domain_id, name, code, sort_order, enabled) VALUES (?, ?, ?, ?, ?)`,
		ptrInt64(w.DomainID), ptrString(w.Name), ptrString(w.Code), ptrIntOr(w.SortOrder, 0), boolToInt(w.Enabled, true),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrConflict
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	return id, err
}

func (r *Repository) UpdateTrack(ctx context.Context, exec repository.Executor, id int64, w TrackWrite) error {
	sets, args := buildTrackSets(w)
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	res, err := exec.ExecContext(ctx, "UPDATE course_track SET "+strings.Join(sets, ", ")+", updated_at = CURRENT_TIMESTAMP WHERE id = ?", args...)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func buildTrackSets(w TrackWrite) ([]string, []any) {
	var sets []string
	var args []any
	if w.DomainID != nil {
		sets, args = append(sets, "domain_id = ?"), append(args, *w.DomainID)
	}
	if w.Name != nil {
		sets, args = append(sets, "name = ?"), append(args, *w.Name)
	}
	if w.Code != nil {
		sets, args = append(sets, "code = ?"), append(args, *w.Code)
	}
	if w.SortOrder != nil {
		sets, args = append(sets, "sort_order = ?"), append(args, *w.SortOrder)
	}
	if w.Enabled != nil {
		sets, args = append(sets, "enabled = ?"), append(args, boolToInt(w.Enabled, false))
	}
	return sets, args
}

// ---------- Level ----------

func (r *Repository) ListLevels(ctx context.Context, exec repository.Executor, trackID int64, limit, offset int) ([]Level, int, error) {
	where := "WHERE 1=1"
	var args []any
	if trackID > 0 {
		where = "WHERE track_id = ?"
		args = append(args, trackID)
	}
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM course_level "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := exec.QueryContext(ctx, "SELECT id, track_id, name, code, sort_order, min_age, max_age, min_lesson_hours, recommended_lesson_hours, enabled, created_at, updated_at FROM course_level "+where+" ORDER BY sort_order ASC, id ASC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []Level
	for rows.Next() {
		var l Level
		var enabled int
		if err := rows.Scan(&l.ID, &l.TrackID, &l.Name, &l.Code, &l.SortOrder, &l.MinAge, &l.MaxAge, &l.MinLessonHours, &l.RecommendedLessonHours, &enabled, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, 0, err
		}
		l.Enabled = enabled != 0
		out = append(out, l)
	}
	return out, total, rows.Err()
}

func (r *Repository) GetLevel(ctx context.Context, exec repository.Executor, id int64) (Level, error) {
	var l Level
	var enabled int
	err := exec.QueryRowContext(ctx, `SELECT id, track_id, name, code, sort_order, min_age, max_age, min_lesson_hours, recommended_lesson_hours, enabled, created_at, updated_at FROM course_level WHERE id = ?`, id).
		Scan(&l.ID, &l.TrackID, &l.Name, &l.Code, &l.SortOrder, &l.MinAge, &l.MaxAge, &l.MinLessonHours, &l.RecommendedLessonHours, &enabled, &l.CreatedAt, &l.UpdatedAt)
	if err == sql.ErrNoRows {
		return Level{}, ErrNotFound
	}
	if err != nil {
		return Level{}, err
	}
	l.Enabled = enabled != 0
	return l, nil
}

func (r *Repository) InsertLevel(ctx context.Context, exec repository.Executor, w LevelWrite) (int64, error) {
	res, err := exec.ExecContext(ctx,
		`INSERT INTO course_level (track_id, name, code, sort_order, min_age, max_age, min_lesson_hours, recommended_lesson_hours, enabled) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ptrInt64(w.TrackID), ptrString(w.Name), ptrString(w.Code), ptrIntOr(w.SortOrder, 0),
		nullableIntPtr(w.MinAge), nullableIntPtr(w.MaxAge), nullableFloatPtr(w.MinLessonHours), nullableFloatPtr(w.RecommendedLessonHours),
		boolToInt(w.Enabled, true),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrConflict
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	return id, err
}

func (r *Repository) UpdateLevel(ctx context.Context, exec repository.Executor, id int64, w LevelWrite) error {
	sets, args := buildLevelSets(w)
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	res, err := exec.ExecContext(ctx, "UPDATE course_level SET "+strings.Join(sets, ", ")+", updated_at = CURRENT_TIMESTAMP WHERE id = ?", args...)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func buildLevelSets(w LevelWrite) ([]string, []any) {
	var sets []string
	var args []any
	if w.TrackID != nil {
		sets, args = append(sets, "track_id = ?"), append(args, *w.TrackID)
	}
	if w.Name != nil {
		sets, args = append(sets, "name = ?"), append(args, *w.Name)
	}
	if w.Code != nil {
		sets, args = append(sets, "code = ?"), append(args, *w.Code)
	}
	if w.SortOrder != nil {
		sets, args = append(sets, "sort_order = ?"), append(args, *w.SortOrder)
	}
	if w.MinAge != nil {
		sets, args = append(sets, "min_age = ?"), append(args, nullableIntPtr(w.MinAge))
	}
	if w.MaxAge != nil {
		sets, args = append(sets, "max_age = ?"), append(args, nullableIntPtr(w.MaxAge))
	}
	if w.MinLessonHours != nil {
		sets, args = append(sets, "min_lesson_hours = ?"), append(args, nullableFloatPtr(w.MinLessonHours))
	}
	if w.RecommendedLessonHours != nil {
		sets, args = append(sets, "recommended_lesson_hours = ?"), append(args, nullableFloatPtr(w.RecommendedLessonHours))
	}
	if w.Enabled != nil {
		sets, args = append(sets, "enabled = ?"), append(args, boolToInt(w.Enabled, false))
	}
	return sets, args
}

// ---------- Capability Tag ----------

func (r *Repository) ListTags(ctx context.Context, exec repository.Executor, domainID int64, limit, offset int) ([]CapabilityTag, int, error) {
	where := "WHERE 1=1"
	var args []any
	if domainID > 0 {
		where = "WHERE domain_id = ?"
		args = append(args, domainID)
	}
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM skill_tag "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, limit, offset)
	rows, err := exec.QueryContext(ctx, "SELECT id, domain_id, name, code, sort_order, enabled, created_at FROM skill_tag "+where+" ORDER BY sort_order ASC, id ASC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []CapabilityTag
	for rows.Next() {
		var t CapabilityTag
		var enabled int
		if err := rows.Scan(&t.ID, &t.DomainID, &t.Name, &t.Code, &t.SortOrder, &enabled, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		t.Enabled = enabled != 0
		out = append(out, t)
	}
	return out, total, rows.Err()
}

func (r *Repository) GetTag(ctx context.Context, exec repository.Executor, id int64) (CapabilityTag, error) {
	var t CapabilityTag
	var enabled int
	err := exec.QueryRowContext(ctx, `SELECT id, domain_id, name, code, sort_order, enabled, created_at FROM skill_tag WHERE id = ?`, id).
		Scan(&t.ID, &t.DomainID, &t.Name, &t.Code, &t.SortOrder, &enabled, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return CapabilityTag{}, ErrNotFound
	}
	if err != nil {
		return CapabilityTag{}, err
	}
	t.Enabled = enabled != 0
	return t, nil
}

func (r *Repository) InsertTag(ctx context.Context, exec repository.Executor, w TagWrite) (int64, error) {
	res, err := exec.ExecContext(ctx,
		`INSERT INTO skill_tag (domain_id, name, code, sort_order, enabled) VALUES (?, ?, ?, ?, ?)`,
		ptrInt64(w.DomainID), ptrString(w.Name), ptrString(w.Code), ptrIntOr(w.SortOrder, 0), boolToInt(w.Enabled, true),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrConflict
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	return id, err
}

func (r *Repository) UpdateTag(ctx context.Context, exec repository.Executor, id int64, w TagWrite) error {
	sets, args := buildTagSets(w)
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	res, err := exec.ExecContext(ctx, "UPDATE skill_tag SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrConflict
		}
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func buildTagSets(w TagWrite) ([]string, []any) {
	var sets []string
	var args []any
	if w.DomainID != nil {
		sets, args = append(sets, "domain_id = ?"), append(args, *w.DomainID)
	}
	if w.Name != nil {
		sets, args = append(sets, "name = ?"), append(args, *w.Name)
	}
	if w.Code != nil {
		sets, args = append(sets, "code = ?"), append(args, *w.Code)
	}
	if w.SortOrder != nil {
		sets, args = append(sets, "sort_order = ?"), append(args, *w.SortOrder)
	}
	if w.Enabled != nil {
		sets, args = append(sets, "enabled = ?"), append(args, boolToInt(w.Enabled, false))
	}
	return sets, args
}

// ---------- helpers ----------

func ptrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func ptrIntOr(i *int, def int) int {
	if i == nil {
		return def
	}
	return *i
}

func defaultStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func nullableIntPtr(i *int) any {
	if i == nil {
		return nil
	}
	return *i
}

func nullableFloatPtr(f *float64) any {
	if f == nil {
		return nil
	}
	return *f
}

// boolToInt returns 1 for true, 0 for false. When the pointer is nil it returns
// the provided default (1 or 0) so create defaults to enabled.
func boolToInt(b *bool, def bool) int {
	if b == nil {
		if def {
			return 1
		}
		return 0
	}
	if *b {
		return 1
	}
	return 0
}

// ========== Enrollment & Assignment (stage C) ==========

// Enrollment is the read model for a student course enrollment. Financial
// fields are intentionally omitted from the M2 API surface; they remain at
// their schema defaults (0) and are not exposed or mutated here.
type Enrollment struct {
	ID             int64      `json:"id"`
	StudentID      int64      `json:"studentId"`
	DomainID       int64      `json:"domainId"`
	TrackID        int64      `json:"trackId"`
	CurrentLevelID *int64     `json:"currentLevelId,omitempty"`
	TargetLevelID  *int64     `json:"targetLevelId,omitempty"`
	EnrollmentType string     `json:"enrollmentType"`
	Status         string     `json:"status"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	EndedAt        *time.Time `json:"endedAt,omitempty"`
	Note           string     `json:"note,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// Assignment is the read model for a student-teacher assignment.
type Assignment struct {
	ID           int64      `json:"id"`
	EnrollmentID int64      `json:"enrollmentId"`
	StudentID    int64      `json:"studentId"`
	TeacherID    int64      `json:"teacherId"`
	RoleType     string     `json:"roleType"`
	RateAmount   *int64     `json:"rateAmount,omitempty"`
	Status       string     `json:"status"`
	StartDate    time.Time  `json:"startDate"`
	EndDate      *time.Time `json:"endDate,omitempty"`
	Reason       string     `json:"reason,omitempty"`
	Note         string     `json:"note,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// EnrollmentWrite captures create/update fields for an enrollment. Financial
// fields are deliberately absent: M2 does not create or mutate lesson balance,
// charge or balance amounts.
type EnrollmentWrite struct {
	DomainID       *int64  `json:"domainId,omitempty"`
	TrackID        *int64  `json:"trackId,omitempty"`
	CurrentLevelID *int64  `json:"currentLevelId,omitempty"`
	TargetLevelID  *int64  `json:"targetLevelId,omitempty"`
	EnrollmentType *string `json:"enrollmentType,omitempty"`
	Status         *string `json:"status,omitempty"`
	StartedAt      *string `json:"startedAt,omitempty"`
	Note           *string `json:"note,omitempty"`
}

// AssignmentWrite captures create fields for an assignment.
type AssignmentWrite struct {
	TeacherID *int64  `json:"teacherId,omitempty"`
	RoleType  *string `json:"roleType,omitempty"`
	Reason    *string `json:"reason,omitempty"`
	Note      *string `json:"note,omitempty"`
}

// EndAssignmentWrite captures the body for POST /assignments/{id}/end.
type EndAssignmentWrite struct {
	Reason *string `json:"reason,omitempty"`
}

// ---------- Enrollment ----------

// ListEnrollments returns enrollments for a student.
func (r *Repository) ListEnrollments(ctx context.Context, exec repository.Executor, studentID int64) ([]Enrollment, error) {
	rows, err := exec.QueryContext(ctx,
		`SELECT id, student_id, domain_id, track_id, current_level_id, target_level_id, enrollment_type, status, started_at, ended_at, COALESCE(note,''), created_at, updated_at
		 FROM student_course_enrollment WHERE student_id = ? AND deleted_at IS NULL ORDER BY id DESC`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Enrollment
	for rows.Next() {
		var e Enrollment
		var note string
		if err := rows.Scan(&e.ID, &e.StudentID, &e.DomainID, &e.TrackID, &e.CurrentLevelID, &e.TargetLevelID, &e.EnrollmentType, &e.Status, &e.StartedAt, &e.EndedAt, &note, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		e.Note = note
		out = append(out, e)
	}
	return out, rows.Err()
}

// GetEnrollment returns a non-deleted enrollment by id.
func (r *Repository) GetEnrollment(ctx context.Context, exec repository.Executor, id int64) (Enrollment, error) {
	var e Enrollment
	var note string
	err := exec.QueryRowContext(ctx,
		`SELECT id, student_id, domain_id, track_id, current_level_id, target_level_id, enrollment_type, status, started_at, ended_at, COALESCE(note,''), created_at, updated_at
		 FROM student_course_enrollment WHERE id = ? AND deleted_at IS NULL`, id,
	).Scan(&e.ID, &e.StudentID, &e.DomainID, &e.TrackID, &e.CurrentLevelID, &e.TargetLevelID, &e.EnrollmentType, &e.Status, &e.StartedAt, &e.EndedAt, &note, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		return Enrollment{}, ErrNotFound
	}
	if err != nil {
		return Enrollment{}, err
	}
	e.Note = note
	return e, nil
}

// InsertEnrollment inserts a new enrollment. Financial fields use schema
// defaults (0); M2 does not set charge/balance/lesson_balance.
func (r *Repository) InsertEnrollment(ctx context.Context, exec repository.Executor, studentID int64, w EnrollmentWrite) (int64, error) {
	res, err := exec.ExecContext(ctx,
		`INSERT INTO student_course_enrollment (student_id, domain_id, track_id, current_level_id, target_level_id, enrollment_type, status, started_at, note)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		studentID, ptrInt64(w.DomainID), ptrInt64(w.TrackID), nullableInt64Ptr(w.CurrentLevelID), nullableInt64Ptr(w.TargetLevelID),
		defaultStr(ptrString(w.EnrollmentType), "ONE_TO_ONE"), defaultStr(ptrString(w.Status), "ACTIVE"), nullableDate(ptrString(w.StartedAt)), nullableString(ptrString(w.Note)),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return id, err
}

// UpdateEnrollment patches a non-deleted enrollment.
func (r *Repository) UpdateEnrollment(ctx context.Context, exec repository.Executor, id int64, w EnrollmentWrite) error {
	sets, args := buildEnrollmentSets(w)
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	res, err := exec.ExecContext(ctx,
		"UPDATE student_course_enrollment SET "+strings.Join(sets, ", ")+", updated_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL", args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func buildEnrollmentSets(w EnrollmentWrite) ([]string, []any) {
	var sets []string
	var args []any
	if w.DomainID != nil {
		sets, args = append(sets, "domain_id = ?"), append(args, *w.DomainID)
	}
	if w.TrackID != nil {
		sets, args = append(sets, "track_id = ?"), append(args, *w.TrackID)
	}
	if w.CurrentLevelID != nil {
		sets, args = append(sets, "current_level_id = ?"), append(args, nullableInt64Ptr(w.CurrentLevelID))
	}
	if w.TargetLevelID != nil {
		sets, args = append(sets, "target_level_id = ?"), append(args, nullableInt64Ptr(w.TargetLevelID))
	}
	if w.EnrollmentType != nil {
		sets, args = append(sets, "enrollment_type = ?"), append(args, *w.EnrollmentType)
	}
	if w.Status != nil {
		sets, args = append(sets, "status = ?"), append(args, *w.Status)
	}
	if w.StartedAt != nil {
		sets, args = append(sets, "started_at = ?"), append(args, nullableDate(*w.StartedAt))
	}
	if w.Note != nil {
		sets, args = append(sets, "note = ?"), append(args, nullableString(*w.Note))
	}
	return sets, args
}

// ---------- Assignment ----------

// ListAssignments returns assignments for an enrollment.
func (r *Repository) ListAssignments(ctx context.Context, exec repository.Executor, enrollmentID int64) ([]Assignment, error) {
	rows, err := exec.QueryContext(ctx,
		`SELECT id, enrollment_id, student_id, teacher_id, role_type, rate_amount, status, start_date, end_date, COALESCE(reason,''), COALESCE(note,''), created_at, updated_at
		 FROM student_teacher_assignment WHERE enrollment_id = ? ORDER BY id ASC`, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Assignment
	for rows.Next() {
		var a Assignment
		var reason, note string
		if err := rows.Scan(&a.ID, &a.EnrollmentID, &a.StudentID, &a.TeacherID, &a.RoleType, &a.RateAmount, &a.Status, &a.StartDate, &a.EndDate, &reason, &note, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.Reason = reason
		a.Note = note
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetAssignment returns an assignment by id.
func (r *Repository) GetAssignment(ctx context.Context, exec repository.Executor, id int64) (Assignment, error) {
	var a Assignment
	var reason, note string
	err := exec.QueryRowContext(ctx,
		`SELECT id, enrollment_id, student_id, teacher_id, role_type, rate_amount, status, start_date, end_date, COALESCE(reason,''), COALESCE(note,''), created_at, updated_at
		 FROM student_teacher_assignment WHERE id = ?`, id,
	).Scan(&a.ID, &a.EnrollmentID, &a.StudentID, &a.TeacherID, &a.RoleType, &a.RateAmount, &a.Status, &a.StartDate, &a.EndDate, &reason, &note, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return Assignment{}, ErrNotFound
	}
	if err != nil {
		return Assignment{}, err
	}
	a.Reason = reason
	a.Note = note
	return a, nil
}

// EndActiveAssignment ends the current ACTIVE assignment for an enrollment in
// the current transaction. Returns the ended assignment id and whether a row
// was ended. The WHERE status='ACTIVE' guard plus the partial unique index make
// this safe under concurrency: only one ACTIVE can exist, and ending it removes
// it from the index so the subsequent insert cannot conflict.
func (r *Repository) EndActiveAssignment(ctx context.Context, exec repository.Executor, enrollmentID int64, reason string) (int64, bool, error) {
	today := time.Now().UTC().Format("2006-01-02")
	res, err := exec.ExecContext(ctx,
		`UPDATE student_teacher_assignment SET status = 'ENDED', end_date = ?, reason = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE enrollment_id = ? AND status = 'ACTIVE'`,
		today, reason, enrollmentID)
	if err != nil {
		return 0, false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, false, err
	}
	if n == 0 {
		return 0, false, nil
	}
	// Return the id of the row we just ended.
	var id int64
	err = exec.QueryRowContext(ctx, `SELECT id FROM student_teacher_assignment WHERE enrollment_id = ? AND status = 'ENDED' ORDER BY updated_at DESC, id DESC LIMIT 1`, enrollmentID).Scan(&id)
	if err != nil {
		return 0, true, err
	}
	return id, true, nil
}

// InsertAssignment inserts a new ACTIVE assignment. The partial unique index
// idx_assign_enrollment_active guarantees at most one ACTIVE per enrollment; a
// concurrent insert that violates it returns a UNIQUE error mapped to
// ErrConflict by the caller.
func (r *Repository) InsertAssignment(ctx context.Context, exec repository.Executor, enrollmentID, studentID, teacherID int64, w AssignmentWrite) (int64, error) {
	today := time.Now().UTC().Format("2006-01-02")
	res, err := exec.ExecContext(ctx,
		`INSERT INTO student_teacher_assignment (enrollment_id, student_id, teacher_id, role_type, status, start_date, reason, note)
		 VALUES (?, ?, ?, ?, 'ACTIVE', ?, ?, ?)`,
		enrollmentID, studentID, teacherID, defaultStr(ptrString(w.RoleType), "MAIN"), today,
		nullableString(ptrString(w.Reason)), nullableString(ptrString(w.Note)),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrConflict
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	return id, err
}

// EndAssignmentByID ends a specific assignment only if it is currently ACTIVE.
// Returns (ended bool, err). ended=false means the assignment was not ACTIVE
// (already ended or paused), so the caller returns 42201; a concurrent end
// surfaces as ended=false rather than a half-written record.
func (r *Repository) EndAssignmentByID(ctx context.Context, exec repository.Executor, id int64, reason string) (bool, error) {
	today := time.Now().UTC().Format("2006-01-02")
	res, err := exec.ExecContext(ctx,
		`UPDATE student_teacher_assignment SET status = 'ENDED', end_date = ?, reason = ?, updated_at = CURRENT_TIMESTAMP
		 WHERE id = ? AND status = 'ACTIVE'`,
		today, reason, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// CountActiveAssignments returns the number of ACTIVE assignments for an
// enrollment. Used by tests/invariants; the partial unique index keeps this at
// 0 or 1.
func (r *Repository) CountActiveAssignments(ctx context.Context, exec repository.Executor, enrollmentID int64) (int, error) {
	var n int
	err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM student_teacher_assignment WHERE enrollment_id = ? AND status = 'ACTIVE'`, enrollmentID).Scan(&n)
	return n, err
}

// ---------- helpers ----------

func nullableInt64Ptr(i *int64) any {
	if i == nil {
		return nil
	}
	return *i
}

// nullableDate returns nil for an empty string, else the string for a nullable
// TEXT/DATE column. SQLite stores dates as TEXT; we keep the raw value.
func nullableDate(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// nullableString returns nil for an empty string, else the string for a nullable
// TEXT column.
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// GetStudentActive returns a non-deleted student only if its status is ACTIVE.
// Used by enrollment creation: ENDED or missing students yield ErrInvalidState
// (per the frozen contract, missing/ended student -> 42201) rather than
// ErrNotFound, so callers do not disclose existence separately.
func (r *Repository) GetStudentActive(ctx context.Context, exec repository.Executor, id int64) (struct{}, error) {
	var status string
	err := exec.QueryRowContext(ctx, `SELECT status FROM student WHERE id = ? AND deleted_at IS NULL`, id).Scan(&status)
	if err == sql.ErrNoRows {
		return struct{}{}, ErrInvalidState
	}
	if err != nil {
		return struct{}{}, err
	}
	if status != "ACTIVE" {
		return struct{}{}, ErrInvalidState
	}
	return struct{}{}, nil
}

// GetTeacherActive returns a non-deleted teacher only if its status is ACTIVE.
// Used by assignment creation: a non-active teacher yields ErrInvalidState.
func (r *Repository) GetTeacherActive(ctx context.Context, exec repository.Executor, id int64) (struct{}, error) {
	var status string
	err := exec.QueryRowContext(ctx, `SELECT status FROM teacher WHERE id = ? AND deleted_at IS NULL`, id).Scan(&status)
	if err == sql.ErrNoRows {
		return struct{}{}, ErrInvalidState
	}
	if err != nil {
		return struct{}{}, err
	}
	if status != "ACTIVE" {
		return struct{}{}, ErrInvalidState
	}
	return struct{}{}, nil
}
