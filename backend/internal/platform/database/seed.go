package database

import (
	"context"
	"database/sql"
	"fmt"
)

// foundationMarkerKey is the stable key for the non-business foundation marker.
const foundationMarkerKey = "foundation.marker"

// foundationMarkerValue contains CJK and emoji characters to verify
// UTF-8 roundtrip in seed metadata. It has no business semantics.
const foundationMarkerValue = "基盤マーカー 🎓 中文 😀"

// FaultHook is a test-controllable hook called after the seed row is
// inserted within the transaction but before Commit. It receives the
// transaction so tests can verify in-transaction state. In production
// the hook is nil and the fast path is taken.
type FaultHook = func(tx *sql.Tx) error

var faultHook FaultHook

// SetFaultHook sets a test-controllable fault injection hook called
// between the seed INSERT and the transaction Commit. Passing nil
// clears the hook. This is not safe for concurrent use and is
// intended for testing only.
func SetFaultHook(h FaultHook) {
	faultHook = h
}

// ApplyFoundationSeed inserts the minimal, non-business foundation marker
// into the database. It is idempotent: repeated calls do not create
// duplicate rows and do not alter existing data.
//
// Idempotency is enforced by a database-level PRIMARY KEY constraint on
// the foundation_seed.key column, using INSERT OR IGNORE to silently
// skip already-present rows. The entire operation runs in a single
// transaction; on failure the transaction is rolled back and no
// partial data is written.
//
// This function MUST be called after MigrateUp has completed. It does
// not create schema — only data rows.
func ApplyFoundationSeed(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	// INSERT OR IGNORE relies on the PRIMARY KEY constraint for idempotency.
	// If the row already exists, the INSERT is silently skipped — no
	// duplicate, no error, no mutation of existing data.
	if _, err := tx.ExecContext(
		ctx,
		`INSERT OR IGNORE INTO foundation_seed (key, value) VALUES (?, ?)`,
		foundationMarkerKey,
		foundationMarkerValue,
	); err != nil {
		return fmt.Errorf("insert foundation marker: %w", err)
	}

	// Test fault injection point: after the seed row is written
	// within the transaction but before Commit. In production
	// faultHook is nil and this block is skipped entirely.
	if faultHook != nil {
		if err := faultHook(tx); err != nil {
			return fmt.Errorf("fault injected before commit: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed transaction: %w", err)
	}

	committed = true
	return nil
}
