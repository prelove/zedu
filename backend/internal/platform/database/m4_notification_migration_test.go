package database_test

import (
	"github.com/prelove/zedu/backend/internal/platform/database"
	"path/filepath"
	"testing"
)

func TestM4NotificationMigrationUpDownUp(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "m4b.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dir := filepath.Join("..", "..", "..", "migrations")
	if err = database.MigrateUp(db, dir); err != nil {
		t.Fatal(err)
	}
	if err = database.MigrateDown(db, dir); err != nil {
		t.Fatal(err)
	}
	if err = database.MigrateUp(db, dir); err != nil {
		t.Fatal(err)
	}
	var name string
	if err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='notification_outbox'`).Scan(&name); err != nil || name != "notification_outbox" {
		t.Fatalf("outbox missing: %v", err)
	}
}
