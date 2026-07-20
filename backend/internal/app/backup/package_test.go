package backup_test

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/prelove/zedu/backend/internal/app/backup"
	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"net/http"
	"net/http/httptest"
)

const pkgSecret = "backup-pkg-test-secret-must-be-32-chars"

// TestBackupPackageContainsDBAttachmentsManifest verifies that a successful
// Owner backup contains the SQLite snapshot, an uploaded evidence file, a
// manifest with matching SHA-256 values, and a BACKUP_CREATE audit row.
func TestBackupPackageContainsDBAttachmentsManifest(t *testing.T) {
	db, dataRoot, backupDir := setupBackupEnv(t)
	// Seed an attachment file under uploads/payments/1/.
	attDir := filepath.Join(dataRoot, "uploads", "payments", "1")
	if err := os.MkdirAll(attDir, 0o755); err != nil {
		t.Fatal(err)
	}
	attPath := filepath.Join(attDir, "evidence.png")
	if err := os.WriteFile(attPath, []byte("PNG_DATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	ownerID := seedBackupUser(t, db, "OWNER")
	mux := httpserver.New()
	backup.MountRoutes(mux, db, pkgSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp := doBackupAuthed(t, srv.URL+"/system/backups", ownerID, "OWNER")
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	var out struct {
		Data struct {
			File string `json:"file"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	pkgPath := filepath.Join(backupDir, out.Data.File)
	info, err := os.Stat(pkgPath)
	if err != nil || !info.IsDir() {
		t.Fatalf("backup package not found: %v", err)
	}
	// Package should be a directory containing db, uploads, manifest.json.
	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	if !containsName(names, "manifest.json") || !containsName(names, "zedu.db") {
		t.Fatalf("package missing core files: %v", names)
	}
	// Verify manifest hashes match actual files.
	manifestBytes, err := os.ReadFile(filepath.Join(pkgPath, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Files []struct {
			Path string `json:"path"`
			Hash string `json:"sha256"`
			Size int64  `json:"size"`
		} `json:"files"`
	}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Files) == 0 {
		t.Fatal("manifest has no file entries")
	}
	for _, f := range manifest.Files {
		full := filepath.Join(pkgPath, filepath.FromSlash(f.Path))
		data, err := os.ReadFile(full)
		if err != nil {
			t.Fatalf("manifest file missing: %s", f.Path)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != f.Hash {
			t.Fatalf("hash mismatch for %s", f.Path)
		}
	}
	// Verify the attachment is included.
	foundAttachment := false
	for _, f := range manifest.Files {
		if strings.Contains(filepath.ToSlash(f.Path), "uploads/payments/1/evidence.png") {
			foundAttachment = true
			break
		}
	}
	if !foundAttachment {
		t.Fatal("attachment not included in backup package")
	}
	// Verify no secrets in manifest.
	manifestStr := string(manifestBytes)
	for _, secret := range []string{"password", "ZEDU_JWT_SECRET", "ZEDU_RESEND_API_KEY", "Bearer", "Authorization"} {
		if strings.Contains(strings.ToLower(manifestStr), strings.ToLower(secret)) {
			t.Fatalf("manifest leaks secret %q", secret)
		}
	}
	// Verify audit row.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM operation_log WHERE action='BACKUP_CREATE'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("audit count=%d err=%v", count, err)
	}
}

// TestBackupPackageOperatorForbidden verifies that a non-Owner gets 40301 and
// no package or audit row is created.
func TestBackupPackageOperatorForbidden(t *testing.T) {
	db, _, backupDir := setupBackupEnv(t)
	opID := seedBackupUser(t, db, "OPERATOR")
	mux := httpserver.New()
	backup.MountRoutes(mux, db, pkgSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp := doBackupAuthed(t, srv.URL+"/system/backups", opID, "OPERATOR")
	defer resp.Body.Close()
	if resp.StatusCode != 403 {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	var body struct {
		Code int `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.Code != 40301 {
		t.Fatalf("expected 40301, got %d", body.Code)
	}
	entries, err := os.ReadDir(backupDir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("operator should not create any package, found %d", len(entries))
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM operation_log WHERE action='BACKUP_CREATE'`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("audit should be 0, got %d", count)
	}
}

// TestBackupPackageFailureCleansStaging verifies that if audit write fails, no
// published package or success audit remains.
func TestBackupPackageFailureCleansStaging(t *testing.T) {
	db, _, backupDir := setupBackupEnv(t)
	ownerID := seedBackupUser(t, db, "OWNER")
	// Drop operation_log to force audit failure.
	if _, err := db.Exec(`DROP TABLE operation_log`); err != nil {
		t.Fatal(err)
	}
	mux := httpserver.New()
	backup.MountRoutes(mux, db, pkgSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp := doBackupAuthed(t, srv.URL+"/system/backups", ownerID, "OWNER")
	defer resp.Body.Close()
	if resp.StatusCode == 201 {
		t.Fatal("expected failure, got 201")
	}
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != ".tmp" {
			t.Fatalf("staging leaked into published dir: %s", e.Name())
		}
	}
}

// TestBackupPackageSameSecondNoConflict verifies that two backups in the same
// second do not collide.
func TestBackupPackageSameSecondNoConflict(t *testing.T) {
	db, _, backupDir := setupBackupEnv(t)
	ownerID := seedBackupUser(t, db, "OWNER")
	mux := httpserver.New()
	backup.MountRoutes(mux, db, pkgSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp1 := doBackupAuthed(t, srv.URL+"/system/backups", ownerID, "OWNER")
	resp1.Body.Close()
	resp2 := doBackupAuthed(t, srv.URL+"/system/backups", ownerID, "OWNER")
	resp2.Body.Close()
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		t.Fatal(err)
	}
	pkgCount := 0
	for _, e := range entries {
		if e.IsDir() && e.Name() != ".tmp" {
			pkgCount++
		}
	}
	if pkgCount < 2 {
		t.Fatalf("expected at least 2 packages, got %d", pkgCount)
	}
}

// TestBackupVerifyValidatesManifest verifies that the verify command succeeds
// on a valid package and fails on a tampered one without touching the active DB.
func TestBackupVerifyValidatesManifest(t *testing.T) {
	db, dataRoot, backupDir := setupBackupEnv(t)
	attDir := filepath.Join(dataRoot, "uploads", "payments", "1")
	if err := os.MkdirAll(attDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(attDir, "evidence.png"), []byte("PNG_DATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	ownerID := seedBackupUser(t, db, "OWNER")
	mux := httpserver.New()
	backup.MountRoutes(mux, db, pkgSecret)
	srv := httptest.NewServer(mux)
	defer srv.Close()
	resp := doBackupAuthed(t, srv.URL+"/system/backups", ownerID, "OWNER")
	defer resp.Body.Close()
	var out struct {
		Data struct {
			File string `json:"file"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	pkgPath := filepath.Join(backupDir, out.Data.File)
	target := filepath.Join(t.TempDir(), "restore-target")
	if err := backup.VerifyPackage(pkgPath, target); err != nil {
		t.Fatalf("verify valid package failed: %v", err)
	}
	// A package may not contain files omitted from its manifest.
	if err := os.WriteFile(filepath.Join(pkgPath, "unlisted.txt"), []byte("unexpected"), 0o644); err != nil {
		t.Fatal(err)
	}
	unlistedTarget := filepath.Join(t.TempDir(), "unlisted-target")
	if err := backup.VerifyPackage(pkgPath, unlistedTarget); err == nil {
		t.Fatal("expected package with unlisted file to fail verification")
	}
	if err := os.Remove(filepath.Join(pkgPath, "unlisted.txt")); err != nil {
		t.Fatal(err)
	}
	// Tamper with the db file.
	dbPath := filepath.Join(pkgPath, "zedu.db")
	if err := os.WriteFile(dbPath, []byte("TAMPERED"), 0o644); err != nil {
		t.Fatal(err)
	}
	tamperedTarget := filepath.Join(t.TempDir(), "tampered-target")
	if err := backup.VerifyPackage(pkgPath, tamperedTarget); err == nil {
		t.Fatal("expected tampered package to fail verification")
	}
	// Active DB should still be readable.
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM user_account`).Scan(&n); err != nil {
		t.Fatalf("active DB unreadable after tamper: %v", err)
	}
}

// --- helpers ---

func setupBackupEnv(t *testing.T) (*sql.DB, string, string) {
	t.Helper()
	dbDir := t.TempDir()
	db, err := database.Open("file:" + filepath.Join(dbDir, "source.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatal(err)
	}
	dataRoot := filepath.Join(t.TempDir(), "data")
	backupDir := filepath.Join(t.TempDir(), "backups")
	t.Setenv("ZEDU_DATA_ROOT", dataRoot)
	t.Setenv("ZEDU_BACKUP_DIR", backupDir)
	t.Cleanup(func() { db.Close() })
	return db, dataRoot, backupDir
}

func seedBackupUser(t *testing.T, db *sql.DB, role string) int64 {
	t.Helper()
	h, err := auth.HashPassword("Pass1234")
	if err != nil {
		t.Fatal(err)
	}
	r, err := db.Exec(`INSERT INTO user_account(username,password_hash,role,display_name) VALUES(?,?,?,?)`, role+"-pkg", h, role, role)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := r.LastInsertId()
	return id
}

func doBackupAuthed(t *testing.T, url string, userID int64, role string) *http.Response {
	t.Helper()
	tok, err := auth.SignAccessToken(pkgSecret, userID, role, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func containsName(names []string, target string) bool {
	for _, n := range names {
		if n == target {
			return true
		}
	}
	return false
}

var _ = io.EOF
