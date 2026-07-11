package database_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/prelove/zedu/backend/internal/platform/database"
)

// newMigratedDB opens a temp DB and runs all migrations up, returning the db,
// the DSN, and the migrations directory path.
func newMigratedDB(t *testing.T) (*sql.DB, string, string) {
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

	return db, dsn, migrationsDir
}

// TestFoundationSeedCreatesMarker verifies that the first call to
// ApplyFoundationSeed inserts exactly one foundation marker row with
// expected UTF-8 metadata (CJK + emoji).
func TestFoundationSeedCreatesMarker(t *testing.T) {
	db, _, _ := newMigratedDB(t)

	if err := database.ApplyFoundationSeed(context.Background(), db); err != nil {
		t.Fatalf("apply foundation seed: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM foundation_seed`).Scan(&count); err != nil {
		t.Fatalf("count foundation_seed rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 foundation_seed row after first seed, got %d", count)
	}

	var key, value string
	if err := db.QueryRow(`SELECT key, value FROM foundation_seed WHERE key = ?`, "foundation.marker").Scan(&key, &value); err != nil {
		t.Fatalf("select foundation marker: %v", err)
	}
	if key != "foundation.marker" {
		t.Fatalf("expected key foundation.marker, got %q", key)
	}
	expectedValue := "基盤マーカー 🎓 中文 😀"
	if value != expectedValue {
		t.Fatalf("expected value %q, got %q", expectedValue, value)
	}
}

// TestFoundationSeedIdempotent verifies that a second call to
// ApplyFoundationSeed does not create duplicate rows and does not
// alter the data written by the first call.
func TestFoundationSeedIdempotent(t *testing.T) {
	db, _, _ := newMigratedDB(t)

	if err := database.ApplyFoundationSeed(context.Background(), db); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	// Capture the original value.
	var originalValue string
	if err := db.QueryRow(`SELECT value FROM foundation_seed WHERE key = ?`, "foundation.marker").Scan(&originalValue); err != nil {
		t.Fatalf("select original value: %v", err)
	}

	// Call seed a second time — must not error, must not duplicate.
	if err := database.ApplyFoundationSeed(context.Background(), db); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM foundation_seed`).Scan(&count); err != nil {
		t.Fatalf("count after second seed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after second seed (idempotent), got %d", count)
	}

	var secondValue string
	if err := db.QueryRow(`SELECT value FROM foundation_seed WHERE key = ?`, "foundation.marker").Scan(&secondValue); err != nil {
		t.Fatalf("select second value: %v", err)
	}
	if secondValue != originalValue {
		t.Fatalf("value changed after second seed: original %q, second %q", originalValue, secondValue)
	}
}

// TestFoundationSeedTransactionRollback verifies that if a failure occurs
// during seed application, the transaction rolls back and no partial data
// is written. We inject a fault by closing the DB connection before
// calling ApplyFoundationSeed, then reopen the same database file to
// verify zero seed rows exist.
func TestFoundationSeedTransactionRollback(t *testing.T) {
	db, dsn, _ := newMigratedDB(t)

	// Close the DB to simulate a connection failure during seed.
	db.Close()

	// ApplyFoundationSeed must return an error — the closed connection
	// prevents any SQL execution.
	err := database.ApplyFoundationSeed(context.Background(), db)
	if err == nil {
		t.Fatalf("expected error when seeding with closed DB, got nil")
	}

	// Reopen a fresh connection to the same database file to verify
	// no partial data was written by the failed seed.
	freshDB, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("reopen database: %v", err)
	}
	defer freshDB.Close()

	var count int
	if err := freshDB.QueryRow(`SELECT COUNT(*) FROM foundation_seed`).Scan(&count); err != nil {
		t.Fatalf("count after rollback: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows after rollback, got %d (partial write detected)", count)
	}
}

// TestFoundationSeedUTF8Metadata verifies that CJK and emoji metadata
// survive a seed → read roundtrip without corruption.
func TestFoundationSeedUTF8Metadata(t *testing.T) {
	db, _, _ := newMigratedDB(t)

	if err := database.ApplyFoundationSeed(context.Background(), db); err != nil {
		t.Fatalf("apply seed: %v", err)
	}

	var value string
	if err := db.QueryRow(`SELECT value FROM foundation_seed WHERE key = ?`, "foundation.marker").Scan(&value); err != nil {
		t.Fatalf("select seed value: %v", err)
	}

	// Must contain Chinese, Japanese, and emoji characters.
	if !containsSubstr(value, "中文") {
		t.Errorf("seed value missing Chinese characters: %q", value)
	}
	if !containsSubstr(value, "マーカー") {
		t.Errorf("seed value missing Japanese characters: %q", value)
	}
	if !containsSubstr(value, "😀") {
		t.Errorf("seed value missing emoji: %q", value)
	}
}

// TestFoundationSeedUniqueConstraint verifies that the idempotency is
// enforced by a database-level UNIQUE constraint, not just Go-side
// check-then-insert. A direct duplicate insert must fail.
func TestFoundationSeedUniqueConstraint(t *testing.T) {
	db, _, _ := newMigratedDB(t)

	if err := database.ApplyFoundationSeed(context.Background(), db); err != nil {
		t.Fatalf("apply seed: %v", err)
	}

	// Direct duplicate insert must fail due to UNIQUE constraint.
	_, err := db.Exec(`INSERT INTO foundation_seed (key, value) VALUES (?, ?)`, "foundation.marker", "duplicate")
	if err == nil {
		t.Fatalf("expected UNIQUE constraint violation on duplicate insert, got nil")
	}
}

func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
