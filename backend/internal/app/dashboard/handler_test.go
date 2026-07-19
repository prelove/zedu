package dashboard_test

import (
	"database/sql"
	"encoding/json"
	"github.com/prelove/zedu/backend/internal/app/dashboard"
	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

const secret = "dashboard-test-secret-must-be-32-chars"

func TestDashboardRequiresAuthAndIsReadOnly(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "db.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	mux := httpserver.New()
	dashboard.MountRoutes(mux, db, secret)
	server := httptest.NewServer(mux)
	defer server.Close()
	resp, err := http.Get(server.URL + "/dashboard")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("unauth=%d", resp.StatusCode)
	}
	resp.Body.Close()
	id := seed(t, db)
	token, err := auth.SignAccessToken(secret, id, "OWNER", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		Code int            `json:"code"`
		Data map[string]int `json:"data"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Data["pendingLessonConfirmations"] != 0 || body.Data["failedNotifications"] != 0 {
		t.Fatalf("body=%#v", body)
	}
}
func seed(t *testing.T, db *sql.DB) int64 {
	h, err := auth.HashPassword("Pass1234")
	if err != nil {
		t.Fatal(err)
	}
	r, err := db.Exec(`INSERT INTO user_account(username,password_hash,role,display_name) VALUES('dash-owner',?,'OWNER','dash')`, h)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := r.LastInsertId()
	return id
}
