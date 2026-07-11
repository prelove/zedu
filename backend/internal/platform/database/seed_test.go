package database_test

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
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

// TestFoundationSeedTransactionRollback verifies the full transaction
// rollback sequence using a test-controllable fault hook:
//
//  1. BeginTx succeeds (the hook is called, which means we passed BeginTx).
//  2. The first foundation_seed record is written within the transaction
//     (the hook queries the tx and finds 1 row).
//  3. The hook returns an error, injecting a fault before Commit.
//  4. ApplyFoundationSeed returns that error.
//  5. A fresh connection to the same DB file confirms 0 rows persisted.
//
// This replaces the previous approach of closing the DB before calling
// ApplyFoundationSeed, which only proved BeginTx failed — not that a
// real transaction with writes rolls back on commit-time faults.
func TestFoundationSeedTransactionRollback(t *testing.T) {
	db, dsn, _ := newMigratedDB(t)

	var hookCalled bool
	var rowsInTransaction int

	database.SetFaultHook(func(tx *sql.Tx) error {
		hookCalled = true
		// Query within the transaction to prove the seed row was
		// written before the fault is injected.
		if err := tx.QueryRow(`SELECT COUNT(*) FROM foundation_seed`).Scan(&rowsInTransaction); err != nil {
			return fmt.Errorf("query in transaction: %w", err)
		}
		return fmt.Errorf("injected fault before commit")
	})
	defer database.SetFaultHook(nil)

	err := database.ApplyFoundationSeed(context.Background(), db)
	if err == nil {
		t.Fatalf("expected error from injected fault, got nil")
	}

	// Prove BeginTx succeeded and the hook was called (which means
	// we got past BeginTx and the INSERT).
	if !hookCalled {
		t.Fatalf("fault hook was not called — BeginTx or INSERT may have failed before the hook")
	}

	// Prove the first foundation_seed record was written within the
	// transaction before the fault was injected.
	if rowsInTransaction != 1 {
		t.Fatalf("expected 1 row visible inside transaction before commit, got %d", rowsInTransaction)
	}

	// Reopen a fresh connection to the same database file to verify
	// the transaction was rolled back — no partial data persisted.
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
	if !strings.Contains(value, "中文") {
		t.Errorf("seed value missing Chinese characters: %q", value)
	}
	if !strings.Contains(value, "マーカー") {
		t.Errorf("seed value missing Japanese characters: %q", value)
	}
	if !strings.Contains(value, "😀") {
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
