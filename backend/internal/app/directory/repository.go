// Package directory implements the people-directory application domain:
// students, parents, teachers, teacher capabilities and availability.
//
// Layering (per M2 design): HTTP handler → application service → repository.
// The repository only executes parameterized SQL and maps results; the service
// owns authorization, validation, state transitions, transaction orchestration
// and audit; the handler only decodes requests, reads identity context and
// encodes responses.
package directory

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/prelove/zedu/backend/internal/repository"
)

// ---------- Domain types ----------

// Student is the read model for a student record.
type Student struct {
	ID            int64      `json:"id"`
	Name          string     `json:"name"`
	NameLocal     string     `json:"nameLocal,omitempty"`
	Email         string     `json:"email,omitempty"`
	Phone         string     `json:"phone,omitempty"`
	Nationality   string     `json:"nationality,omitempty"`
	Timezone      string     `json:"timezone"`
	Status        string     `json:"status"`
	SourceChannel string     `json:"sourceChannel,omitempty"`
	Note          string     `json:"note,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	DeletedAt     *time.Time `json:"deletedAt,omitempty"`
}

// Parent is the read model for a parent contact scoped to a student.
type Parent struct {
	ID           int64     `json:"id"`
	StudentID    int64     `json:"studentId"`
	Name         string    `json:"name"`
	Email        string    `json:"email,omitempty"`
	Phone        string    `json:"phone,omitempty"`
	Relationship string    `json:"relationship,omitempty"`
	IsPrimary    bool      `json:"isPrimary"`
	Note         string    `json:"note,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Teacher is the read model for a teacher record.
type Teacher struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	NameLocal   string     `json:"nameLocal,omitempty"`
	Email       string     `json:"email,omitempty"`
	Phone       string     `json:"phone,omitempty"`
	Bio         string     `json:"bio,omitempty"`
	DefaultRate int64      `json:"defaultRate"`
	Status      string     `json:"status"`
	Note        string     `json:"note,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

// Capability is the read model for a teacher capability.
type Capability struct {
	ID            int64      `json:"id"`
	TeacherID     int64      `json:"teacherId"`
	DomainID      int64      `json:"domainId"`
	TrackID       int64      `json:"trackId"`
	LevelID       int64      `json:"levelId"`
	SkillTagCodes string     `json:"skillTagCodes,omitempty"`
	Status        string     `json:"status"`
	Verified      bool       `json:"verified"`
	EffectiveFrom *time.Time `json:"effectiveFrom,omitempty"`
	EffectiveTo   *time.Time `json:"effectiveTo,omitempty"`
	Note          string     `json:"note,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// Availability is the read model for a teacher availability slot.
type Availability struct {
	ID            int64      `json:"id"`
	TeacherID     int64      `json:"teacherId"`
	Weekday       int        `json:"weekday"`
	StartTime     string     `json:"startTime"`
	EndTime       string     `json:"endTime"`
	EffectiveFrom *time.Time `json:"effectiveFrom,omitempty"`
	EffectiveTo   *time.Time `json:"effectiveTo,omitempty"`
	Note          string     `json:"note,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// ---------- Write payloads ----------

// StudentWrite captures create/update fields for a student.
type StudentWrite struct {
	Name          *string `json:"name,omitempty"`
	NameLocal     *string `json:"nameLocal,omitempty"`
	Email         *string `json:"email,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Nationality   *string `json:"nationality,omitempty"`
	Timezone      *string `json:"timezone,omitempty"`
	Status        *string `json:"status,omitempty"`
	SourceChannel *string `json:"sourceChannel,omitempty"`
	Note          *string `json:"note,omitempty"`
}

// ParentWrite captures create/update fields for a parent.
type ParentWrite struct {
	Name         *string `json:"name,omitempty"`
	Email        *string `json:"email,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	Relationship *string `json:"relationship,omitempty"`
	IsPrimary    *bool   `json:"isPrimary,omitempty"`
	Note         *string `json:"note,omitempty"`
}

// TeacherWrite captures create/update fields for a teacher.
type TeacherWrite struct {
	Name        *string `json:"name,omitempty"`
	NameLocal   *string `json:"nameLocal,omitempty"`
	Email       *string `json:"email,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Bio         *string `json:"bio,omitempty"`
	DefaultRate *int64  `json:"defaultRate,omitempty"`
	Status      *string `json:"status,omitempty"`
	Note        *string `json:"note,omitempty"`
}

// CapabilityWrite captures create/update fields for a teacher capability.
type CapabilityWrite struct {
	DomainID      *int64  `json:"domainId,omitempty"`
	TrackID       *int64  `json:"trackId,omitempty"`
	LevelID       *int64  `json:"levelId,omitempty"`
	SkillTagCodes *string `json:"skillTagCodes,omitempty"`
	Status        *string `json:"status,omitempty"`
	Verified      *bool   `json:"verified,omitempty"`
	EffectiveFrom *string `json:"effectiveFrom,omitempty"`
	EffectiveTo   *string `json:"effectiveTo,omitempty"`
	Note          *string `json:"note,omitempty"`
}

// AvailabilityWrite captures create/update fields for a teacher availability.
type AvailabilityWrite struct {
	Weekday       *int    `json:"weekday,omitempty"`
	StartTime     *string `json:"startTime,omitempty"`
	EndTime       *string `json:"endTime,omitempty"`
	EffectiveFrom *string `json:"effectiveFrom,omitempty"`
	EffectiveTo   *string `json:"effectiveTo,omitempty"`
	Note          *string `json:"note,omitempty"`
}

// ---------- Sentinel errors ----------

// ErrNotFound is returned when a resource does not exist or is not owned by the
// referenced parent resource. Callers map this to HTTP 404 / 40401.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a unique constraint is violated. Callers map
// this to HTTP 409 / 40901.
var ErrConflict = errors.New("conflict")

// ErrInvalidState is returned when a state transition or relationship is not
// allowed. Callers map this to HTTP 422 / 42201.
var ErrInvalidState = errors.New("invalid state")

// isUniqueViolation reports whether err is a SQLite UNIQUE constraint failure.
// modernc.org/sqlite reports constraint failures with "UNIQUE" in the message.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// ---------- Repository ----------

// Repository provides parameterized SQL access to people-directory tables.
// Every method accepts repository.Executor so it can run inside a service
// transaction or against the bare DB for reads.
type Repository struct{}

// NewRepository returns a people-directory repository.
func NewRepository() *Repository { return &Repository{} }

// ListStudents returns a page of students matching the optional status filter
// and name search term (case-insensitive prefix/contains on name).
func (r *Repository) ListStudents(ctx context.Context, exec repository.Executor, status, search string, limit, offset int) ([]Student, int, error) {
	where := "WHERE deleted_at IS NULL"
	var args []any
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}
	if search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+search+"%")
	}
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM student "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := "SELECT id, name, COALESCE(name_local,''), COALESCE(email,''), COALESCE(phone,''), COALESCE(nationality,''), timezone, status, COALESCE(source_channel,''), COALESCE(note,''), created_at, updated_at, deleted_at FROM student " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := exec.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []Student
	for rows.Next() {
		var s Student
		var nameLocal, email, phone, nationality, sourceChannel, note string
		if err := rows.Scan(&s.ID, &s.Name, &nameLocal, &email, &phone, &nationality, &s.Timezone, &s.Status, &sourceChannel, &note, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt); err != nil {
			return nil, 0, err
		}
		s.NameLocal = nameLocal
		s.Email = email
		s.Phone = phone
		s.Nationality = nationality
		s.SourceChannel = sourceChannel
		s.Note = note
		out = append(out, s)
	}
	return out, total, rows.Err()
}

// GetStudent returns a single non-deleted student by id.
func (r *Repository) GetStudent(ctx context.Context, exec repository.Executor, id int64) (Student, error) {
	var s Student
	var nameLocal, email, phone, nationality, sourceChannel, note string
	err := exec.QueryRowContext(ctx,
		`SELECT id, name, COALESCE(name_local,''), COALESCE(email,''), COALESCE(phone,''), COALESCE(nationality,''), timezone, status, COALESCE(source_channel,''), COALESCE(note,''), created_at, updated_at, deleted_at
		 FROM student WHERE id = ? AND deleted_at IS NULL`, id,
	).Scan(&s.ID, &s.Name, &nameLocal, &email, &phone, &nationality, &s.Timezone, &s.Status, &sourceChannel, &note, &s.CreatedAt, &s.UpdatedAt, &s.DeletedAt)
	if err == sql.ErrNoRows {
		return Student{}, ErrNotFound
	}
	if err != nil {
		return Student{}, err
	}
	s.NameLocal = nameLocal
	s.Email = email
	s.Phone = phone
	s.Nationality = nationality
	s.SourceChannel = sourceChannel
	s.Note = note
	return s, nil
}

// InsertStudent inserts a new student and returns its id.
func (r *Repository) InsertStudent(ctx context.Context, exec repository.Executor, w StudentWrite) (int64, error) {
	name := ptrString(w.Name)
	email := nullableString(ptrString(w.Email))
	res, err := exec.ExecContext(ctx,
		`INSERT INTO student (name, name_local, email, phone, nationality, timezone, status, source_channel, note)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		name, nullableString(ptrString(w.NameLocal)), email, nullableString(ptrString(w.Phone)),
		nullableString(ptrString(w.Nationality)), defaultStr(ptrString(w.Timezone), "Asia/Tokyo"),
		defaultStr(ptrString(w.Status), "ACTIVE"), nullableString(ptrString(w.SourceChannel)), nullableString(ptrString(w.Note)),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrConflict
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateStudent patches a non-deleted student. Returns ErrNotFound if no row
// matched, ErrConflict on email uniqueness violation.
func (r *Repository) UpdateStudent(ctx context.Context, exec repository.Executor, id int64, w StudentWrite) error {
	sets, args := buildStudentSets(w)
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	res, err := exec.ExecContext(ctx,
		"UPDATE student SET "+strings.Join(sets, ", ")+", updated_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL", args...)
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

func buildStudentSets(w StudentWrite) ([]string, []any) {
	var sets []string
	var args []any
	if w.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *w.Name)
	}
	if w.NameLocal != nil {
		sets = append(sets, "name_local = ?")
		args = append(args, nullableString(*w.NameLocal))
	}
	if w.Email != nil {
		sets = append(sets, "email = ?")
		args = append(args, nullableString(*w.Email))
	}
	if w.Phone != nil {
		sets = append(sets, "phone = ?")
		args = append(args, nullableString(*w.Phone))
	}
	if w.Nationality != nil {
		sets = append(sets, "nationality = ?")
		args = append(args, nullableString(*w.Nationality))
	}
	if w.Timezone != nil {
		sets = append(sets, "timezone = ?")
		args = append(args, *w.Timezone)
	}
	if w.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *w.Status)
	}
	if w.SourceChannel != nil {
		sets = append(sets, "source_channel = ?")
		args = append(args, nullableString(*w.SourceChannel))
	}
	if w.Note != nil {
		sets = append(sets, "note = ?")
		args = append(args, nullableString(*w.Note))
	}
	return sets, args
}

// ---------- Parent ----------

// ListParents returns parents for a student.
func (r *Repository) ListParents(ctx context.Context, exec repository.Executor, studentID int64, limit, offset int) ([]Parent, int, error) {
	var total int
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM parent WHERE student_id = ?`, studentID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := exec.QueryContext(ctx,
		`SELECT id, student_id, name, COALESCE(email,''), COALESCE(phone,''), COALESCE(relationship,''), is_primary, COALESCE(note,''), created_at, updated_at
		 FROM parent WHERE student_id = ? ORDER BY is_primary DESC, id ASC LIMIT ? OFFSET ?`, studentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Parent{}
	for rows.Next() {
		var p Parent
		var email, phone, relationship, note string
		var isPrimary int
		if err := rows.Scan(&p.ID, &p.StudentID, &p.Name, &email, &phone, &relationship, &isPrimary, &note, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		p.Email = email
		p.Phone = phone
		p.Relationship = relationship
		p.IsPrimary = isPrimary != 0
		p.Note = note
		out = append(out, p)
	}
	return out, total, rows.Err()
}

// GetParent returns a parent only if it belongs to the given student; otherwise
// ErrNotFound (so cross-student access does not disclose the record).
func (r *Repository) GetParent(ctx context.Context, exec repository.Executor, studentID, parentID int64) (Parent, error) {
	var p Parent
	var email, phone, relationship, note string
	var isPrimary int
	err := exec.QueryRowContext(ctx,
		`SELECT id, student_id, name, COALESCE(email,''), COALESCE(phone,''), COALESCE(relationship,''), is_primary, COALESCE(note,''), created_at, updated_at
		 FROM parent WHERE id = ? AND student_id = ?`, parentID, studentID,
	).Scan(&p.ID, &p.StudentID, &p.Name, &email, &phone, &relationship, &isPrimary, &note, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return Parent{}, ErrNotFound
	}
	if err != nil {
		return Parent{}, err
	}
	p.Email = email
	p.Phone = phone
	p.Relationship = relationship
	p.IsPrimary = isPrimary != 0
	p.Note = note
	return p, nil
}

// InsertParent inserts a parent under a student.
func (r *Repository) InsertParent(ctx context.Context, exec repository.Executor, studentID int64, w ParentWrite) (int64, error) {
	res, err := exec.ExecContext(ctx,
		`INSERT INTO parent (student_id, name, email, phone, relationship, is_primary, note)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		studentID, ptrString(w.Name), nullableString(ptrString(w.Email)), nullableString(ptrString(w.Phone)),
		nullableString(ptrString(w.Relationship)), boolToInt(w.IsPrimary), nullableString(ptrString(w.Note)),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateParent patches a parent scoped to a student.
func (r *Repository) UpdateParent(ctx context.Context, exec repository.Executor, studentID, parentID int64, w ParentWrite) error {
	sets, args := buildParentSets(w)
	if len(sets) == 0 {
		// Still verify ownership so a no-op PATCH on a cross-student parent
		// returns 404 rather than 200.
		_, err := r.GetParent(ctx, exec, studentID, parentID)
		return err
	}
	args = append(args, parentID, studentID)
	res, err := exec.ExecContext(ctx,
		"UPDATE parent SET "+strings.Join(sets, ", ")+", updated_at = CURRENT_TIMESTAMP WHERE id = ? AND student_id = ?", args...)
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

func buildParentSets(w ParentWrite) ([]string, []any) {
	var sets []string
	var args []any
	if w.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *w.Name)
	}
	if w.Email != nil {
		sets = append(sets, "email = ?")
		args = append(args, nullableString(*w.Email))
	}
	if w.Phone != nil {
		sets = append(sets, "phone = ?")
		args = append(args, nullableString(*w.Phone))
	}
	if w.Relationship != nil {
		sets = append(sets, "relationship = ?")
		args = append(args, nullableString(*w.Relationship))
	}
	if w.IsPrimary != nil {
		sets = append(sets, "is_primary = ?")
		args = append(args, boolToInt(w.IsPrimary))
	}
	if w.Note != nil {
		sets = append(sets, "note = ?")
		args = append(args, nullableString(*w.Note))
	}
	return sets, args
}

// ---------- Teacher ----------

// ListTeachers returns a page of teachers.
func (r *Repository) ListTeachers(ctx context.Context, exec repository.Executor, status, search string, limit, offset int) ([]Teacher, int, error) {
	where := "WHERE deleted_at IS NULL"
	var args []any
	if status != "" {
		where += " AND status = ?"
		args = append(args, status)
	}
	if search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+search+"%")
	}
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM teacher "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := "SELECT id, name, COALESCE(name_local,''), COALESCE(email,''), COALESCE(phone,''), COALESCE(bio,''), default_rate_amount, status, COALESCE(note,''), created_at, updated_at, deleted_at FROM teacher " + where + " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := exec.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []Teacher
	for rows.Next() {
		var t Teacher
		var nameLocal, email, phone, bio, note string
		if err := rows.Scan(&t.ID, &t.Name, &nameLocal, &email, &phone, &bio, &t.DefaultRate, &t.Status, &note, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt); err != nil {
			return nil, 0, err
		}
		t.NameLocal = nameLocal
		t.Email = email
		t.Phone = phone
		t.Bio = bio
		t.Note = note
		out = append(out, t)
	}
	return out, total, rows.Err()
}

// GetTeacher returns a single non-deleted teacher by id.
func (r *Repository) GetTeacher(ctx context.Context, exec repository.Executor, id int64) (Teacher, error) {
	var t Teacher
	var nameLocal, email, phone, bio, note string
	err := exec.QueryRowContext(ctx,
		`SELECT id, name, COALESCE(name_local,''), COALESCE(email,''), COALESCE(phone,''), COALESCE(bio,''), default_rate_amount, status, COALESCE(note,''), created_at, updated_at, deleted_at
		 FROM teacher WHERE id = ? AND deleted_at IS NULL`, id,
	).Scan(&t.ID, &t.Name, &nameLocal, &email, &phone, &bio, &t.DefaultRate, &t.Status, &note, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
	if err == sql.ErrNoRows {
		return Teacher{}, ErrNotFound
	}
	if err != nil {
		return Teacher{}, err
	}
	t.NameLocal = nameLocal
	t.Email = email
	t.Phone = phone
	t.Bio = bio
	t.Note = note
	return t, nil
}

// InsertTeacher inserts a new teacher.
func (r *Repository) InsertTeacher(ctx context.Context, exec repository.Executor, w TeacherWrite) (int64, error) {
	res, err := exec.ExecContext(ctx,
		`INSERT INTO teacher (name, name_local, email, phone, bio, default_rate_amount, status, note)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		ptrString(w.Name), nullableString(ptrString(w.NameLocal)), nullableString(ptrString(w.Email)),
		nullableString(ptrString(w.Phone)), nullableString(ptrString(w.Bio)), defaultInt64(ptrInt64(w.DefaultRate), 0),
		defaultStr(ptrString(w.Status), "ACTIVE"), nullableString(ptrString(w.Note)),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateTeacher patches a non-deleted teacher.
func (r *Repository) UpdateTeacher(ctx context.Context, exec repository.Executor, id int64, w TeacherWrite) error {
	sets, args := buildTeacherSets(w)
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	res, err := exec.ExecContext(ctx,
		"UPDATE teacher SET "+strings.Join(sets, ", ")+", updated_at = CURRENT_TIMESTAMP WHERE id = ? AND deleted_at IS NULL", args...)
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

func buildTeacherSets(w TeacherWrite) ([]string, []any) {
	var sets []string
	var args []any
	if w.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *w.Name)
	}
	if w.NameLocal != nil {
		sets = append(sets, "name_local = ?")
		args = append(args, nullableString(*w.NameLocal))
	}
	if w.Email != nil {
		sets = append(sets, "email = ?")
		args = append(args, nullableString(*w.Email))
	}
	if w.Phone != nil {
		sets = append(sets, "phone = ?")
		args = append(args, nullableString(*w.Phone))
	}
	if w.Bio != nil {
		sets = append(sets, "bio = ?")
		args = append(args, nullableString(*w.Bio))
	}
	if w.DefaultRate != nil {
		sets = append(sets, "default_rate_amount = ?")
		args = append(args, *w.DefaultRate)
	}
	if w.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *w.Status)
	}
	if w.Note != nil {
		sets = append(sets, "note = ?")
		args = append(args, nullableString(*w.Note))
	}
	return sets, args
}

// ---------- Capability ----------

// ListCapabilities returns capabilities for a teacher.
func (r *Repository) ListCapabilities(ctx context.Context, exec repository.Executor, teacherID int64, limit, offset int) ([]Capability, int, error) {
	var total int
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM teacher_capability WHERE teacher_id = ?`, teacherID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := exec.QueryContext(ctx,
		`SELECT id, teacher_id, domain_id, track_id, level_id, COALESCE(skill_tag_codes,''), status, verified, effective_from, effective_to, COALESCE(note,''), created_at, updated_at
		 FROM teacher_capability WHERE teacher_id = ? ORDER BY id ASC LIMIT ? OFFSET ?`, teacherID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Capability{}
	for rows.Next() {
		var c Capability
		var tags, note string
		var verified int
		if err := rows.Scan(&c.ID, &c.TeacherID, &c.DomainID, &c.TrackID, &c.LevelID, &tags, &c.Status, &verified, &c.EffectiveFrom, &c.EffectiveTo, &note, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		c.SkillTagCodes = tags
		c.Verified = verified != 0
		c.Note = note
		out = append(out, c)
	}
	return out, total, rows.Err()
}

// GetCapability returns a capability by id scoped to a teacher.
func (r *Repository) GetCapability(ctx context.Context, exec repository.Executor, teacherID, capID int64) (Capability, error) {
	var c Capability
	var tags, note string
	var verified int
	err := exec.QueryRowContext(ctx,
		`SELECT id, teacher_id, domain_id, track_id, level_id, COALESCE(skill_tag_codes,''), status, verified, effective_from, effective_to, COALESCE(note,''), created_at, updated_at
		 FROM teacher_capability WHERE id = ? AND teacher_id = ?`, capID, teacherID,
	).Scan(&c.ID, &c.TeacherID, &c.DomainID, &c.TrackID, &c.LevelID, &tags, &c.Status, &verified, &c.EffectiveFrom, &c.EffectiveTo, &note, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return Capability{}, ErrNotFound
	}
	if err != nil {
		return Capability{}, err
	}
	c.SkillTagCodes = tags
	c.Verified = verified != 0
	c.Note = note
	return c, nil
}

// InsertCapability inserts a new teacher capability.
func (r *Repository) InsertCapability(ctx context.Context, exec repository.Executor, teacherID int64, w CapabilityWrite) (int64, error) {
	res, err := exec.ExecContext(ctx,
		`INSERT INTO teacher_capability (teacher_id, domain_id, track_id, level_id, skill_tag_codes, status, verified, effective_from, effective_to, note)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		teacherID, ptrInt64(w.DomainID), ptrInt64(w.TrackID), ptrInt64(w.LevelID), nullableString(ptrString(w.SkillTagCodes)),
		defaultStr(ptrString(w.Status), "ACTIVE"), boolToInt(w.Verified), nullableDate(ptrString(w.EffectiveFrom)), nullableDate(ptrString(w.EffectiveTo)), nullableString(ptrString(w.Note)),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrConflict
		}
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateCapability patches a capability scoped to a teacher. On the unique
// (teacher_id, track_id, level_id) violation it returns ErrConflict.
func (r *Repository) UpdateCapability(ctx context.Context, exec repository.Executor, teacherID, capID int64, w CapabilityWrite) error {
	sets, args := buildCapabilitySets(w)
	if len(sets) == 0 {
		_, err := r.GetCapability(ctx, exec, teacherID, capID)
		return err
	}
	args = append(args, capID, teacherID)
	res, err := exec.ExecContext(ctx,
		"UPDATE teacher_capability SET "+strings.Join(sets, ", ")+", updated_at = CURRENT_TIMESTAMP WHERE id = ? AND teacher_id = ?", args...)
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

func buildCapabilitySets(w CapabilityWrite) ([]string, []any) {
	var sets []string
	var args []any
	if w.DomainID != nil {
		sets = append(sets, "domain_id = ?")
		args = append(args, *w.DomainID)
	}
	if w.TrackID != nil {
		sets = append(sets, "track_id = ?")
		args = append(args, *w.TrackID)
	}
	if w.LevelID != nil {
		sets = append(sets, "level_id = ?")
		args = append(args, *w.LevelID)
	}
	if w.SkillTagCodes != nil {
		sets = append(sets, "skill_tag_codes = ?")
		args = append(args, nullableString(*w.SkillTagCodes))
	}
	if w.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *w.Status)
	}
	if w.Verified != nil {
		sets = append(sets, "verified = ?")
		args = append(args, boolToInt(w.Verified))
	}
	if w.EffectiveFrom != nil {
		sets = append(sets, "effective_from = ?")
		args = append(args, nullableDate(*w.EffectiveFrom))
	}
	if w.EffectiveTo != nil {
		sets = append(sets, "effective_to = ?")
		args = append(args, nullableDate(*w.EffectiveTo))
	}
	if w.Note != nil {
		sets = append(sets, "note = ?")
		args = append(args, nullableString(*w.Note))
	}
	return sets, args
}

// ---------- Availability ----------

// ListAvailability returns availability slots for a teacher.
func (r *Repository) ListAvailability(ctx context.Context, exec repository.Executor, teacherID int64, limit, offset int) ([]Availability, int, error) {
	var total int
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM teacher_availability WHERE teacher_id = ?`, teacherID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := exec.QueryContext(ctx,
		`SELECT id, teacher_id, weekday, start_time, end_time, effective_from, effective_to, COALESCE(note,''), created_at, updated_at
		 FROM teacher_availability WHERE teacher_id = ? ORDER BY weekday ASC, start_time ASC LIMIT ? OFFSET ?`, teacherID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Availability{}
	for rows.Next() {
		var a Availability
		var note string
		if err := rows.Scan(&a.ID, &a.TeacherID, &a.Weekday, &a.StartTime, &a.EndTime, &a.EffectiveFrom, &a.EffectiveTo, &note, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		a.Note = note
		out = append(out, a)
	}
	return out, total, rows.Err()
}

// GetAvailability returns an availability slot by id scoped to a teacher.
func (r *Repository) GetAvailability(ctx context.Context, exec repository.Executor, teacherID, availID int64) (Availability, error) {
	var a Availability
	var note string
	err := exec.QueryRowContext(ctx,
		`SELECT id, teacher_id, weekday, start_time, end_time, effective_from, effective_to, COALESCE(note,''), created_at, updated_at
		 FROM teacher_availability WHERE id = ? AND teacher_id = ?`, availID, teacherID,
	).Scan(&a.ID, &a.TeacherID, &a.Weekday, &a.StartTime, &a.EndTime, &a.EffectiveFrom, &a.EffectiveTo, &note, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return Availability{}, ErrNotFound
	}
	if err != nil {
		return Availability{}, err
	}
	a.Note = note
	return a, nil
}

// InsertAvailability inserts a new availability slot.
func (r *Repository) InsertAvailability(ctx context.Context, exec repository.Executor, teacherID int64, w AvailabilityWrite) (int64, error) {
	res, err := exec.ExecContext(ctx,
		`INSERT INTO teacher_availability (teacher_id, weekday, start_time, end_time, effective_from, effective_to, note)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		teacherID, ptrInt(w.Weekday), ptrString(w.StartTime), ptrString(w.EndTime),
		nullableDate(ptrString(w.EffectiveFrom)), nullableDate(ptrString(w.EffectiveTo)), nullableString(ptrString(w.Note)),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateAvailability patches an availability slot scoped to a teacher.
func (r *Repository) UpdateAvailability(ctx context.Context, exec repository.Executor, teacherID, availID int64, w AvailabilityWrite) error {
	sets, args := buildAvailabilitySets(w)
	if len(sets) == 0 {
		_, err := r.GetAvailability(ctx, exec, teacherID, availID)
		return err
	}
	args = append(args, availID, teacherID)
	res, err := exec.ExecContext(ctx,
		"UPDATE teacher_availability SET "+strings.Join(sets, ", ")+", updated_at = CURRENT_TIMESTAMP WHERE id = ? AND teacher_id = ?", args...)
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

func buildAvailabilitySets(w AvailabilityWrite) ([]string, []any) {
	var sets []string
	var args []any
	if w.Weekday != nil {
		sets = append(sets, "weekday = ?")
		args = append(args, *w.Weekday)
	}
	if w.StartTime != nil {
		sets = append(sets, "start_time = ?")
		args = append(args, *w.StartTime)
	}
	if w.EndTime != nil {
		sets = append(sets, "end_time = ?")
		args = append(args, *w.EndTime)
	}
	if w.EffectiveFrom != nil {
		sets = append(sets, "effective_from = ?")
		args = append(args, nullableDate(*w.EffectiveFrom))
	}
	if w.EffectiveTo != nil {
		sets = append(sets, "effective_to = ?")
		args = append(args, nullableDate(*w.EffectiveTo))
	}
	if w.Note != nil {
		sets = append(sets, "note = ?")
		args = append(args, nullableString(*w.Note))
	}
	return sets, args
}

// ---------- Hierarchy validation helpers ----------

// HierarchyRow holds the ids that must be consistent for a capability.
type HierarchyRow struct {
	DomainID int64
	TrackID  int64
	LevelID  int64
}

// VerifyHierarchy confirms that track belongs to domain and level belongs to
// track. Returns ErrInvalidState when the relationship is broken, ErrNotFound
// when any referenced row is missing.
func (r *Repository) VerifyHierarchy(ctx context.Context, exec repository.Executor, domainID, trackID, levelID int64) error {
	var trackDomainID int64
	err := exec.QueryRowContext(ctx, `SELECT domain_id FROM course_track WHERE id = ?`, trackID).Scan(&trackDomainID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if trackDomainID != domainID {
		return ErrInvalidState
	}
	var levelTrackID int64
	err = exec.QueryRowContext(ctx, `SELECT track_id FROM course_level WHERE id = ?`, levelID).Scan(&levelTrackID)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if levelTrackID != trackID {
		return ErrInvalidState
	}
	return nil
}

// ---------- helpers ----------

func ptrString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func ptrInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullableDate(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func defaultStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func defaultInt64(v, def int64) int64 {
	if v == 0 {
		return def
	}
	return v
}

func boolToInt(b *bool) int {
	if b == nil || !*b {
		return 0
	}
	return 1
}
