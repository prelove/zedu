package database_test

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/prelove/zedu/backend/internal/platform/database"
)

// newMigratedDBForM2 opens a temp DB, runs all migrations up, and returns the db.
func newMigratedDBForM2(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dsn := "file:" + filepath.Join(tmpDir, "test.db")

	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestM2MigrationUpDownUp verifies that migration 003 can be applied,
// rolled back, and re-applied cleanly.
func TestM2MigrationUpDownUp(t *testing.T) {
	tmpDir := t.TempDir()
	dsn := "file:" + filepath.Join(tmpDir, "test.db")

	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	migrationsDir := filepath.Join("..", "..", "..", "migrations")

	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	// Verify M2 tables exist.
	for _, table := range []string{
		"user_account", "refresh_session", "system_settings", "operation_log",
		"course_domain", "course_track", "course_level", "skill_tag",
		"student", "parent", "teacher", "teacher_availability", "teacher_capability",
		"student_course_enrollment", "student_teacher_assignment",
	} {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing after up: %v", table, err)
		}
	}

	if err := database.MigrateDown(db, migrationsDir); err != nil {
		t.Fatalf("migrate down: %v", err)
	}

	// Verify M2 tables are gone.
	for _, table := range []string{
		"user_account", "refresh_session", "system_settings", "operation_log",
		"student", "teacher", "teacher_capability",
		"student_course_enrollment", "student_teacher_assignment",
	} {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err == nil {
			t.Fatalf("table %s still exists after down", table)
		}
	}

	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up again: %v", err)
	}

	// Verify M2 tables exist again.
	for _, table := range []string{
		"user_account", "refresh_session", "system_settings", "operation_log",
		"student", "teacher", "teacher_capability",
	} {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s missing after re-up: %v", table, err)
		}
	}
}

// TestM2ForeignKeyEnforced verifies that FK constraints work on M2 tables.
func TestM2ForeignKeyEnforced(t *testing.T) {
	db := newMigratedDBForM2(t)

	// parent.student_id references student(id) — inserting with non-existent student must fail.
	_, err := db.Exec(`INSERT INTO parent (student_id, name) VALUES (999, 'Ghost Parent')`)
	if err == nil {
		t.Fatalf("expected FK violation on parent.student_id, got nil")
	}

	// refresh_session.user_id references user_account(id).
	_, err = db.Exec(`INSERT INTO refresh_session (user_id, token_hash, expires_at) VALUES (999, 'hash', '2026-01-01')`)
	if err == nil {
		t.Fatalf("expected FK violation on refresh_session.user_id, got nil")
	}
}

// TestM2UTF8Roundtrip verifies CJK and emoji survive insert/select on M2 tables.
func TestM2UTF8Roundtrip(t *testing.T) {
	db := newMigratedDBForM2(t)

	// Insert a user with CJK display name.
	cjkName := "管理者 中文 😀 🎌"
	_, err := db.Exec(`INSERT INTO user_account (username, password_hash, role, display_name) VALUES ('owner', 'hash', 'OWNER', ?)`, cjkName)
	if err != nil {
		t.Fatalf("insert user with CJK: %v", err)
	}

	var got string
	if err := db.QueryRow(`SELECT display_name FROM user_account WHERE username='owner'`).Scan(&got); err != nil {
		t.Fatalf("select user: %v", err)
	}
	if got != cjkName {
		t.Fatalf("CJK roundtrip failed: expected %q, got %q", cjkName, got)
	}

	// Insert a student with CJK name and emoji note.
	studentName := "王同学 日本語 😀"
	_, err = db.Exec(`INSERT INTO student (name, name_local, note) VALUES (?, '王', ?)`, studentName, "备注 emoji 🎓")
	if err != nil {
		t.Fatalf("insert student with CJK: %v", err)
	}

	var sName, sNote string
	if err := db.QueryRow(`SELECT name, note FROM student WHERE name=?`, studentName).Scan(&sName, &sNote); err != nil {
		t.Fatalf("select student: %v", err)
	}
	if sName != studentName {
		t.Fatalf("student name roundtrip: expected %q, got %q", studentName, sName)
	}
	if sNote != "备注 emoji 🎓" {
		t.Fatalf("student note roundtrip: expected %q, got %q", "备注 emoji 🎓", sNote)
	}
}

// TestM2StudentEmailNullMultipleRows verifies that multiple students with NULL email are allowed.
func TestM2StudentEmailNullMultipleRows(t *testing.T) {
	db := newMigratedDBForM2(t)

	for i := 0; i < 3; i++ {
		_, err := db.Exec(`INSERT INTO student (name) VALUES (?)`, fmt.Sprintf("Student-%d", i))
		if err != nil {
			t.Fatalf("insert student %d with NULL email: %v", i, err)
		}
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM student WHERE email IS NULL`).Scan(&count); err != nil {
		t.Fatalf("count NULL email students: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 students with NULL email, got %d", count)
	}
}

// TestM2StudentEmailNonNullUnique verifies that non-NULL emails are globally unique.
func TestM2StudentEmailNonNullUnique(t *testing.T) {
	db := newMigratedDBForM2(t)

	_, err := db.Exec(`INSERT INTO student (name, email) VALUES ('Alice', 'alice@example.com')`)
	if err != nil {
		t.Fatalf("insert first student: %v", err)
	}

	// Second student with same email must fail.
	_, err = db.Exec(`INSERT INTO student (name, email) VALUES ('Bob', 'alice@example.com')`)
	if err == nil {
		t.Fatalf("expected unique violation on duplicate email, got nil")
	}
}

// TestM2StudentEmailConcurrentDuplicate verifies that concurrent inserts of the same
// non-NULL email result in exactly one success and others fail with constraint violation.
func TestM2StudentEmailConcurrentDuplicate(t *testing.T) {
	// Use a separate DB with higher connection limit for concurrency.
	tmpDir := t.TempDir()
	dsn := "file:" + filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	if _, err := db.Exec(`PRAGMA journal_mode = WAL`); err != nil {
		t.Fatalf("set WAL: %v", err)
	}
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
		t.Fatalf("set busy timeout: %v", err)
	}

	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	const goroutines = 5
	const email = "concurrent@example.com"

	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	failures := 0

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := db.Exec(`INSERT INTO student (name, email) VALUES (?, ?)`,
				fmt.Sprintf("Concurrent-%d", idx), email)
			mu.Lock()
			if err != nil {
				failures++
			} else {
				successes++
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	if successes != 1 {
		t.Fatalf("expected exactly 1 concurrent insert success, got %d (failures=%d)", successes, failures)
	}
	if failures != goroutines-1 {
		t.Fatalf("expected %d concurrent failures, got %d", goroutines-1, failures)
	}
}

// TestM2TeacherCapabilityUniqueTriple verifies the (teacher_id, track_id, level_id) unique constraint.
func TestM2TeacherCapabilityUniqueTriple(t *testing.T) {
	db := newMigratedDBForM2(t)

	// Setup prerequisite data.
	_, err := db.Exec(`INSERT INTO course_domain (name, code, type) VALUES ('Japanese', 'JP', 'LANGUAGE')`)
	if err != nil {
		t.Fatalf("insert domain: %v", err)
	}
	_, err = db.Exec(`INSERT INTO course_track (domain_id, name, code) VALUES (1, 'JLPT', 'JLPT')`)
	if err != nil {
		t.Fatalf("insert track: %v", err)
	}
	_, err = db.Exec(`INSERT INTO course_level (track_id, name, code) VALUES (1, 'N1', 'N1')`)
	if err != nil {
		t.Fatalf("insert level: %v", err)
	}
	_, err = db.Exec(`INSERT INTO teacher (name) VALUES ('Sensei')`)
	if err != nil {
		t.Fatalf("insert teacher: %v", err)
	}

	// First capability insert succeeds.
	_, err = db.Exec(`INSERT INTO teacher_capability (teacher_id, domain_id, track_id, level_id) VALUES (1, 1, 1, 1)`)
	if err != nil {
		t.Fatalf("insert first capability: %v", err)
	}

	// Duplicate triple must fail.
	_, err = db.Exec(`INSERT INTO teacher_capability (teacher_id, domain_id, track_id, level_id) VALUES (1, 1, 1, 1)`)
	if err == nil {
		t.Fatalf("expected unique violation on duplicate (teacher_id, track_id, level_id), got nil")
	}
}

// TestM2UserAccountUsernameUnique verifies username uniqueness.
func TestM2UserAccountUsernameUnique(t *testing.T) {
	db := newMigratedDBForM2(t)

	_, err := db.Exec(`INSERT INTO user_account (username, password_hash, role, display_name) VALUES ('admin', 'hash1', 'OWNER', 'Admin')`)
	if err != nil {
		t.Fatalf("insert first user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO user_account (username, password_hash, role, display_name) VALUES ('admin', 'hash2', 'OPERATOR', 'Admin2')`)
	if err == nil {
		t.Fatalf("expected unique violation on duplicate username, got nil")
	}
}
