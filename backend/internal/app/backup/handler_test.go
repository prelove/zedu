package backup_test

import (
	"database/sql"
	"encoding/json"
	"github.com/prelove/zedu/backend/internal/app/backup"
	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const secret = "backup-test-secret-must-be-32-chars"

func TestOwnerBackupCreatesAuditedArtifact(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "source.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "backups")
	t.Setenv("ZEDU_BACKUP_DIR", dir)
	mux := httpserver.New()
	backup.MountRoutes(mux, db, secret)
	s := httptest.NewServer(mux)
	defer s.Close()
	owner := user(t, db, "OWNER")
	operator := user(t, db, "OPERATOR")
	tok := token(t, operator, "OPERATOR")
	req, _ := http.NewRequest(http.MethodPost, s.URL+"/system/backups", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 403 {
		t.Fatalf("operator=%d", resp.StatusCode)
	}
	resp.Body.Close()
	req, _ = http.NewRequest(http.MethodPost, s.URL+"/system/backups", nil)
	req.Header.Set("Authorization", "Bearer "+token(t, owner, "OWNER"))
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("owner=%d", resp.StatusCode)
	}
	var out struct {
		Data struct {
			File string `json:"file"`
		} `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(filepath.Join(dir, out.Data.File)); err != nil {
		t.Fatalf("backup artifact: %v", err)
	}
	var count int
	if err = db.QueryRow(`SELECT COUNT(*) FROM operation_log WHERE action='BACKUP_CREATE'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("audit=%d err=%v", count, err)
	}
}
func user(t *testing.T, db *sql.DB, role string) int64 {
	h, err := auth.HashPassword("Pass1234")
	if err != nil {
		t.Fatal(err)
	}
	r, err := db.Exec(`INSERT INTO user_account(username,password_hash,role,display_name) VALUES(?,?,?,?)`, role+"user", h, role, role)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := r.LastInsertId()
	return id
}
func token(t *testing.T, id int64, role string) string {
	v, err := auth.SignAccessToken(secret, id, role, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
