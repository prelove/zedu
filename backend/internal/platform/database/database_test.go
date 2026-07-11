package database_test

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/prelove/zedu/backend/internal/platform/database"
)

func TestMigrationUpDownUp(t *testing.T) {
	db := newTempDB(t)
	defer db.Close()

	migrationsDir := filepath.Join("..", "..", "..", "migrations")

	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	if err := database.MigrateDown(db, migrationsDir); err != nil {
		t.Fatalf("migrate down: %v", err)
	}

	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatalf("migrate up again: %v", err)
	}

	// Verify schema_migrations table exists and tracks version.
	var version string
	err := db.QueryRow("SELECT version FROM schema_migrations LIMIT 1").Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		// ok, migrations may be empty
	} else if err != nil {
		t.Fatalf("verify schema_migrations: %v", err)
	}
}

func TestPragmaValues(t *testing.T) {
	db := newTempDB(t)
	defer db.Close()

	var fkOn int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkOn); err != nil {
		t.Fatalf("read foreign_keys: %v", err)
	}
	if fkOn != 1 {
		t.Fatalf("foreign_keys MUST be ON, got %d", fkOn)
	}

	var journal string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journal); err != nil {
		t.Fatalf("read journal_mode: %v", err)
	}
	if journal != "wal" {
		t.Fatalf("journal_mode MUST be WAL, got %q", journal)
	}

	var busy int
	if err := db.QueryRow("PRAGMA busy_timeout").Scan(&busy); err != nil {
		t.Fatalf("read busy_timeout: %v", err)
	}
	if busy < 5000 {
		t.Fatalf("busy_timeout MUST be >= 5000, got %d", busy)
	}
}

func TestPragmaForeignKeyEnforced(t *testing.T) {
	db := newTempDB(t)
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE parent (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE child (id INTEGER PRIMARY KEY, parent_id INTEGER REFERENCES parent(id))`); err != nil {
		t.Fatalf("create child: %v", err)
	}

	_, err := db.Exec(`INSERT INTO child (id, parent_id) VALUES (1, 999)`)
	if err == nil {
		t.Fatalf("expected foreign key violation, got nil")
	}
}

func TestUTF8Roundtrip(t *testing.T) {
	db := newTempDB(t)
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE utf8_test (id INTEGER PRIMARY KEY, value TEXT)`); err != nil {
		t.Fatalf("create utf8_test: %v", err)
	}

	value := "中文 日本語 😀 🎌 全角ＡＢＣ"
	if _, err := db.Exec(`INSERT INTO utf8_test (id, value) VALUES (1, ?)`, value); err != nil {
		t.Fatalf("insert utf8: %v", err)
	}

	var got string
	if err := db.QueryRow(`SELECT value FROM utf8_test WHERE id = 1`).Scan(&got); err != nil {
		t.Fatalf("select utf8: %v", err)
	}
	if got != value {
		t.Fatalf("utf8 roundtrip failed: expected %q, got %q", value, got)
	}
}

func newTempDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dsn := filepath.Join(tmpDir, "test.db")

	db, err := database.Open("file:" + dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
		os.Remove(dsn)
	})

	return db
}
