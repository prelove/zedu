package onboarding

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	platformauth "github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/platform/logging"
)

const testJWTSecret = "onboarding-test-secret-must-be-at-least-32-chars"

type testServer struct {
	db  *sql.DB
	srv *httptest.Server
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "onboarding.db")
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	MountRoutes(mux, NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	t.Cleanup(func() {
		srv.Close()
		db.Close()
	})
	return &testServer{db: db, srv: srv}
}

func createUser(t *testing.T, db *sql.DB, username, role string) int64 {
	t.Helper()
	hash, err := platformauth.HashPassword("Pass1234")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	result, err := db.Exec(`INSERT INTO user_account (username, password_hash, role, display_name) VALUES (?, ?, ?, ?)`, username, hash, role, username)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("read user id: %v", err)
	}
	return id
}

func tokenFor(t *testing.T, userID int64, role string) string {
	t.Helper()
	token, err := platformauth.SignAccessToken(testJWTSecret, userID, role, time.Hour)
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}
	return token
}

func request(t *testing.T, method, url, token string, body any) (int, map[string]any) {
	t.Helper()
	var input io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("encode request: %v", err)
		}
		input = bytes.NewReader(encoded)
	}
	req, err := http.NewRequest(method, url, input)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request: %v", err)
	}
	defer resp.Body.Close()
	decoded := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp.StatusCode, decoded
}

func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return count
}

func TestOwnerInitializesJapaneseTemplateIdempotently(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	token := tokenFor(t, ownerID, "OWNER")

	status, body := request(t, http.MethodPost, ts.srv.URL+"/onboarding/initialize", token, map[string]string{"template": "japanese"})
	if status != http.StatusOK || body["code"] != float64(0) {
		t.Fatalf("initialize status=%d body=%v", status, body)
	}
	if got := countRows(t, ts.db, "course_domain"); got != 1 {
		t.Fatalf("domains = %d, want 1", got)
	}
	if got := countRows(t, ts.db, "course_track"); got != 4 {
		t.Fatalf("tracks = %d, want 4", got)
	}
	if got := countRows(t, ts.db, "course_level"); got != 9 {
		t.Fatalf("levels = %d, want 9", got)
	}
	if got := countRows(t, ts.db, "skill_tag"); got != 9 {
		t.Fatalf("skill tags = %d, want 9", got)
	}

	status, body = request(t, http.MethodPost, ts.srv.URL+"/onboarding/initialize", token, map[string]string{"template": "japanese"})
	if status != http.StatusOK || body["code"] != float64(0) {
		t.Fatalf("idempotent initialize status=%d body=%v", status, body)
	}
	if got := countRows(t, ts.db, "operation_log"); got != 1 {
		t.Fatalf("successful initialization audit rows = %d, want 1", got)
	}
	var marker string
	if err := ts.db.QueryRow(`SELECT config_value FROM system_settings WHERE config_key = 'onboarding.template'`).Scan(&marker); err != nil {
		t.Fatalf("read initialization marker: %v", err)
	}
	if marker != "japanese" {
		t.Fatalf("marker = %q, want japanese", marker)
	}
}

func TestOwnerInitializesBlankAndK12Templates(t *testing.T) {
	for _, testCase := range []struct {
		name, template string
		wantDomains    int
	}{
		{name: "blank", template: "blank", wantDomains: 0},
		{name: "k12", template: "k12", wantDomains: 4},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ts := newTestServer(t)
			ownerID := createUser(t, ts.db, "owner", "OWNER")
			status, body := request(t, http.MethodPost, ts.srv.URL+"/onboarding/initialize", tokenFor(t, ownerID, "OWNER"), map[string]string{"template": testCase.template})
			if status != http.StatusOK || body["code"] != float64(0) {
				t.Fatalf("initialize status=%d body=%v", status, body)
			}
			if got := countRows(t, ts.db, "course_domain"); got != testCase.wantDomains {
				t.Fatalf("domains = %d, want %d", got, testCase.wantDomains)
			}
			if testCase.template == "k12" && (countRows(t, ts.db, "course_track") == 0 || countRows(t, ts.db, "course_level") == 0 || countRows(t, ts.db, "skill_tag") == 0) {
				t.Fatal("K12 hierarchy is incomplete")
			}
		})
	}
}

func TestInitializeRejectsUnknownTemplateWithoutSideEffects(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	settingsBefore := countRows(t, ts.db, "system_settings")
	status, body := request(t, http.MethodPost, ts.srv.URL+"/onboarding/initialize", tokenFor(t, ownerID, "OWNER"), map[string]string{"template": "surprise"})
	if status != http.StatusUnprocessableEntity || body["code"] != float64(httpserver.CodeInvalidState) {
		t.Fatalf("invalid template status=%d body=%v", status, body)
	}
	if countRows(t, ts.db, "course_domain") != 0 || countRows(t, ts.db, "system_settings") != settingsBefore || countRows(t, ts.db, "operation_log") != 0 {
		t.Fatalf("invalid template left side effects")
	}
}

func TestOperatorCannotInitializeAndUnauthenticatedIsRejected(t *testing.T) {
	ts := newTestServer(t)
	operatorID := createUser(t, ts.db, "operator", "OPERATOR")

	status, body := request(t, http.MethodPost, ts.srv.URL+"/onboarding/initialize", tokenFor(t, operatorID, "OPERATOR"), map[string]string{"template": "blank"})
	if status != http.StatusForbidden || body["code"] != float64(httpserver.CodeForbidden) {
		t.Fatalf("operator status=%d body=%v", status, body)
	}
	status, body = request(t, http.MethodPost, ts.srv.URL+"/onboarding/initialize", "", map[string]string{"template": "blank"})
	if status != http.StatusUnauthorized || body["code"] != float64(httpserver.CodeUnauth) {
		t.Fatalf("unauthenticated status=%d body=%v", status, body)
	}
	if got := countRows(t, ts.db, "operation_log"); got != 0 {
		t.Fatalf("audit rows = %d, want 0", got)
	}
}

func TestResetRejectsWhenProtectedBusinessDataExists(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	token := tokenFor(t, ownerID, "OWNER")
	status, _ := request(t, http.MethodPost, ts.srv.URL+"/onboarding/initialize", token, map[string]string{"template": "blank"})
	if status != http.StatusOK {
		t.Fatalf("initialize blank status=%d", status)
	}
	if _, err := ts.db.Exec(`INSERT INTO student (name) VALUES ('学生😀')`); err != nil {
		t.Fatalf("insert protected record: %v", err)
	}

	status, body := request(t, http.MethodPost, ts.srv.URL+"/onboarding/reset", token, map[string]string{"template": "japanese"})
	if status != http.StatusUnprocessableEntity || body["code"] != float64(httpserver.CodeInvalidState) {
		t.Fatalf("reset status=%d body=%v", status, body)
	}
	if got := countRows(t, ts.db, "course_domain"); got != 0 {
		t.Fatalf("domains after rejected reset = %d, want 0", got)
	}
	if got := countRows(t, ts.db, "operation_log"); got != 1 {
		t.Fatalf("audit rows after rejected reset = %d, want 1", got)
	}
}

func TestResetReplacesTemplateAndWritesAuditableFact(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	token := tokenFor(t, ownerID, "OWNER")
	status, _ := request(t, http.MethodPost, ts.srv.URL+"/onboarding/initialize", token, map[string]string{"template": "japanese"})
	if status != http.StatusOK {
		t.Fatalf("initialize status=%d", status)
	}

	status, body := request(t, http.MethodPost, ts.srv.URL+"/onboarding/reset", token, map[string]string{"template": "k12"})
	if status != http.StatusOK || body["code"] != float64(0) {
		t.Fatalf("reset status=%d body=%v", status, body)
	}
	if got := countRows(t, ts.db, "course_domain"); got != 4 {
		t.Fatalf("K12 domains = %d, want 4", got)
	}
	var action, targetType, detail, requestID, marker string
	var targetID int64
	if err := ts.db.QueryRow(`SELECT action, target_type, target_id, detail_json, request_id FROM operation_log ORDER BY id DESC LIMIT 1`).Scan(&action, &targetType, &targetID, &detail, &requestID); err != nil {
		t.Fatalf("read reset audit: %v", err)
	}
	if action != "ONBOARDING_RESET" || targetType != "system" || targetID != 1 || requestID == "" || requestID == "unknown" {
		t.Fatalf("audit action=%q target_type=%q target_id=%d request_id=%q", action, targetType, targetID, requestID)
	}
	var parsed map[string]string
	if err := json.Unmarshal([]byte(detail), &parsed); err != nil || parsed["template"] != "k12" {
		t.Fatalf("audit detail=%q parsed=%v err=%v", detail, parsed, err)
	}
	if _, found := parsed["password"]; found {
		t.Fatalf("audit detail leaked password field: %v", parsed)
	}
	if err := ts.db.QueryRow(`SELECT config_value FROM system_settings WHERE config_key = 'onboarding.template'`).Scan(&marker); err != nil {
		t.Fatalf("read marker: %v", err)
	}
	if marker != "k12" {
		t.Fatalf("marker=%q, want k12", marker)
	}
}

func TestInitializeRollsBackWhenAuditWriteFails(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	settingsBefore := countRows(t, ts.db, "system_settings")
	if _, err := ts.db.Exec(`CREATE TRIGGER fail_onboarding_audit BEFORE INSERT ON operation_log
		WHEN NEW.action = 'ONBOARDING_INITIALIZE'
		BEGIN SELECT RAISE(ABORT, 'injected audit failure'); END`); err != nil {
		t.Fatalf("create fault trigger: %v", err)
	}

	status, body := request(t, http.MethodPost, ts.srv.URL+"/onboarding/initialize", tokenFor(t, ownerID, "OWNER"), map[string]string{"template": "japanese"})
	if status != http.StatusInternalServerError || body["code"] != float64(httpserver.CodeDatabase) || body["message"] != "DATABASE_ERROR" {
		t.Fatalf("fault status=%d body=%v", status, body)
	}
	for _, table := range []string{"course_domain", "course_track", "course_level", "skill_tag", "operation_log"} {
		if got := countRows(t, ts.db, table); got != 0 {
			t.Fatalf("%s rows after rollback = %d, want 0", table, got)
		}
	}
	if got := countRows(t, ts.db, "system_settings"); got != settingsBefore {
		t.Fatalf("system_settings rows after rollback = %d, want baseline %d", got, settingsBefore)
	}
}

func TestResetRollsBackWhenAuditWriteFails(t *testing.T) {
	ts := newTestServer(t)
	ownerID := createUser(t, ts.db, "owner", "OWNER")
	token := tokenFor(t, ownerID, "OWNER")
	status, _ := request(t, http.MethodPost, ts.srv.URL+"/onboarding/initialize", token, map[string]string{"template": "japanese"})
	if status != http.StatusOK {
		t.Fatalf("initialize status=%d", status)
	}
	if _, err := ts.db.Exec(`CREATE TRIGGER fail_onboarding_reset_audit BEFORE INSERT ON operation_log
		WHEN NEW.action = 'ONBOARDING_RESET'
		BEGIN SELECT RAISE(ABORT, 'injected reset audit failure'); END`); err != nil {
		t.Fatalf("create fault trigger: %v", err)
	}

	status, body := request(t, http.MethodPost, ts.srv.URL+"/onboarding/reset", token, map[string]string{"template": "k12"})
	if status != http.StatusInternalServerError || body["code"] != float64(httpserver.CodeDatabase) || body["message"] != "DATABASE_ERROR" {
		t.Fatalf("fault status=%d body=%v", status, body)
	}
	if got := countRows(t, ts.db, "course_domain"); got != 1 {
		t.Fatalf("domains after rollback = %d, want japanese domain", got)
	}
	if got := countRows(t, ts.db, "course_track"); got != 4 {
		t.Fatalf("tracks after rollback = %d, want 4", got)
	}
	if got := countRows(t, ts.db, "course_level"); got != 9 {
		t.Fatalf("levels after rollback = %d, want 9", got)
	}
	if got := countRows(t, ts.db, "operation_log"); got != 1 {
		t.Fatalf("audit rows after rollback = %d, want initialization only", got)
	}
	var marker string
	if err := ts.db.QueryRow(`SELECT config_value FROM system_settings WHERE config_key = 'onboarding.template'`).Scan(&marker); err != nil {
		t.Fatalf("read retained marker: %v", err)
	}
	if marker != "japanese" {
		t.Fatalf("marker after rollback = %q, want japanese", marker)
	}
}
