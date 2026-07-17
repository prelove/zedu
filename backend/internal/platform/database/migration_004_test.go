package database_test

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/prelove/zedu/backend/internal/platform/database"
)

// TestMigration004UpDownUp verifies that migration 004 (student_level_event)
// can be applied, rolled back, and re-applied cleanly on top of 001-003.
func TestMigration004UpDownUp(t *testing.T) {
	tmpDir := t.TempDir()
	dsn := "file:" + filepath.Join(tmpDir, "test_004.db")
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	migrationsDir := filepath.Join("..", "..", "..", "migrations")

	// Apply all migrations up (including 004).
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	// Verify student_level_event table exists.
	var name string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='student_level_event'`).Scan(&name)
	if err != nil {
		t.Fatalf("student_level_event table missing after up: %v", err)
	}

	// Verify columns.
	cols := tableColumns(t, db, "student_level_event")
	expected := []string{"id", "student_id", "enrollment_id", "from_level_id", "to_level_id", "event_type", "event_date", "evidence_note", "operator_id", "created_at"}
	for _, c := range expected {
		if _, ok := cols[c]; !ok {
			t.Fatalf("student_level_event missing column %s; got %v", c, cols)
		}
	}

	// Verify indexes exist.
	idxCount := 0
	err = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND tbl_name='student_level_event' AND name LIKE 'idx_%'`).Scan(&idxCount)
	if err != nil {
		t.Fatalf("count indexes: %v", err)
	}
	if idxCount < 2 {
		t.Fatalf("expected at least 2 indexes on student_level_event, got %d", idxCount)
	}

	// Roll back all migrations (down).
	if err := database.MigrateDown(db, migrationsDir); err != nil {
		t.Fatalf("migrate down: %v", err)
	}

	// Table should be gone.
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='student_level_event'`).Scan(&name)
	if err == nil {
		t.Fatalf("student_level_event still exists after down")
	}

	// Re-apply up.
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up again: %v", err)
	}

	// Table should exist again.
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='student_level_event'`).Scan(&name)
	if err != nil {
		t.Fatalf("student_level_event missing after re-up: %v", err)
	}
}

func tableColumns(t *testing.T, db *sql.DB, table string) map[string]bool {
	t.Helper()
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		t.Fatalf("pragma table_info %s: %v", table, err)
	}
	defer rows.Close()
	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		cols[name] = true
	}
	return cols
}
