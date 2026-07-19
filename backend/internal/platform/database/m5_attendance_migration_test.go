package database_test

import (
	"github.com/prelove/zedu/backend/internal/platform/database"
	"path/filepath"
	"testing"
)

func TestM5AttendanceMigrationUpDownUp(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "m5.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dir := filepath.Join("..", "..", "..", "migrations")
	if err = database.MigrateUp(db, dir); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO attendance_outcome_type(code,name) VALUES('TEST','test')`); err != nil {
		t.Fatal(err)
	}
	if err = database.MigrateDown(db, dir); err != nil {
		t.Fatal(err)
	}
	if err = database.MigrateUp(db, dir); err != nil {
		t.Fatal(err)
	}
}
