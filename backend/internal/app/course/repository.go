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
