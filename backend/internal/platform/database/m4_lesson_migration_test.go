package database_test

import (
	"path/filepath"
	"testing"

	"github.com/prelove/zedu/backend/internal/platform/database"
)

func TestM4LessonMigrationUpDownUpCreatesScopedLessonTable(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "m4.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	migrationsDir := filepath.Join("..", "..", "..", "migrations")
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatal(err)
	}
	if err := database.MigrateDown(db, migrationsDir); err != nil {
		t.Fatal(err)
	}
	if err := database.MigrateUp(db, migrationsDir); err != nil {
		t.Fatal(err)
	}
	var sql string
	if err := db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='lesson'`).Scan(&sql); err != nil {
		t.Fatalf("lesson table: %v", err)
	}
	if sql == "" {
		t.Fatal("lesson DDL is empty")
	}
}
