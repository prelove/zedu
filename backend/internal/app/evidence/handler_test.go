package evidence

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	platformauth "github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/platform/logging"
)

const testJWTSecret = "evidence-test-secret-must-be-at-least-32"

func TestAttachmentRoutesRequireAuth(t *testing.T) {
	db, srv, _, _ := newEvidenceTestServer(t, Config{})
	defer db.Close()
	defer srv.Close()

	status, body := evidenceJSONRequest(t, http.MethodPost, srv.URL+"/finance/payments/1/attachments", "", newMultipartBody(t, "file", "proof.png", validPNG()))
	if status != http.StatusUnauthorized || body["code"] != float64(httpserver.CodeUnauth) {
		t.Fatalf("unauth upload status=%d body=%v", status, body)
	}

	status, body = evidenceJSONRequest(t, http.MethodGet, srv.URL+"/finance/payments/1/attachments", "", nil)
	if status != http.StatusUnauthorized || body["code"] != float64(httpserver.CodeUnauth) {
		t.Fatalf("unauth list status=%d body=%v", status, body)
	}

	status, body = evidenceJSONRequest(t, http.MethodGet, srv.URL+"/finance/payments/1/attachments/1/content", "", nil)
	if status != http.StatusUnauthorized || body["code"] != float64(httpserver.CodeUnauth) {
		t.Fatalf("unauth download status=%d body=%v", status, body)
	}

	var audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM operation_log`).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if audits != 0 {
		t.Fatalf("unauth routes wrote %d audits", audits)
	}
}

func TestAttachmentUploadListAndDownload(t *testing.T) {
	db, srv, storageRoot, _ := newEvidenceTestServer(t, Config{})
	defer db.Close()
	defer srv.Close()

	paymentID := seedEvidencePayment(t, db)
	operatorID := evidenceUser(t, db, "operator", "OPERATOR")
	token := evidenceToken(t, operatorID, "OPERATOR")

	status, body := evidenceJSONRequest(t, http.MethodPost, srv.URL+"/finance/payments/"+itoa(paymentID)+"/attachments", token, newMultipartBody(t, "file", `C:\tmp\proof.txt`, validPNG()))
	if status != http.StatusCreated || body["code"] != float64(0) {
		t.Fatalf("upload status=%d body=%v", status, body)
	}

	data := body["data"].(map[string]any)
	attachmentID := int64(data["id"].(float64))
	if got := data["fileName"].(string); got != "proof.png" {
		t.Fatalf("fileName=%q want proof.png", got)
	}
	if got := data["fileType"].(string); got != "image/png" {
		t.Fatalf("fileType=%q", got)
	}

	status, body = evidenceJSONRequest(t, http.MethodGet, srv.URL+"/finance/payments/"+itoa(paymentID)+"/attachments", token, nil)
	if status != http.StatusOK || body["code"] != float64(0) {
		t.Fatalf("list status=%d body=%v", status, body)
	}
	list := body["data"].(map[string]any)
	items := list["items"].([]any)
	if len(items) != 1 {
		t.Fatalf("list items=%d", len(items))
	}

	rawStatus, rawHeader, rawBody := evidenceRawRequest(t, http.MethodGet, srv.URL+"/finance/payments/"+itoa(paymentID)+"/attachments/"+itoa(attachmentID)+"/content", token)
	if rawStatus != http.StatusOK {
		t.Fatalf("download status=%d body=%s", rawStatus, string(rawBody))
	}
	if got := rawHeader.Get("Content-Type"); got != "image/png" {
		t.Fatalf("download content-type=%q", got)
	}
	if got := rawHeader.Get("Content-Disposition"); !strings.Contains(got, "attachment") || !strings.Contains(got, "proof.png") {
		t.Fatalf("download content-disposition=%q", got)
	}
	if !bytes.Equal(rawBody, validPNG()) {
		t.Fatalf("download body mismatch")
	}

	var relPath string
	if err := db.QueryRow(`SELECT file_path FROM payment_attachment WHERE id = ?`, attachmentID).Scan(&relPath); err != nil {
		t.Fatal(err)
	}
	if filepath.IsAbs(relPath) || strings.Contains(relPath, "..") {
		t.Fatalf("unsafe stored file_path=%q", relPath)
	}
	if _, err := os.Stat(filepath.Join(storageRoot, "uploads", filepath.FromSlash(relPath))); err != nil {
		t.Fatalf("published file missing: %v", err)
	}

	assertAuditRow(t, db, "PAYMENT_ATTACHMENT_UPLOAD")
	assertAuditRow(t, db, "PAYMENT_ATTACHMENT_DOWNLOAD")
}

func TestAttachmentUploadRejectsUnsupportedTypeAndMaxThree(t *testing.T) {
	db, srv, storageRoot, _ := newEvidenceTestServer(t, Config{})
	defer db.Close()
	defer srv.Close()

	paymentID := seedEvidencePayment(t, db)
	operatorID := evidenceUser(t, db, "operator", "OPERATOR")
	token := evidenceToken(t, operatorID, "OPERATOR")

	status, body := evidenceJSONRequest(t, http.MethodPost, srv.URL+"/finance/payments/"+itoa(paymentID)+"/attachments", token, newMultipartBody(t, "file", "notes.txt", []byte("hello")))
	if status != http.StatusUnprocessableEntity || body["code"] != float64(httpserver.CodeInvalidState) {
		t.Fatalf("unsupported status=%d body=%v", status, body)
	}
	assertTempDirEmpty(t, filepath.Join(storageRoot, "uploads", ".tmp"))

	for i := 0; i < 3; i++ {
		status, body = evidenceJSONRequest(t, http.MethodPost, srv.URL+"/finance/payments/"+itoa(paymentID)+"/attachments", token, newMultipartBody(t, "file", "proof.png", validPNG()))
		if status != http.StatusCreated || body["code"] != float64(0) {
			t.Fatalf("upload %d status=%d body=%v", i+1, status, body)
		}
	}

	status, body = evidenceJSONRequest(t, http.MethodPost, srv.URL+"/finance/payments/"+itoa(paymentID)+"/attachments", token, newMultipartBody(t, "file", "proof.png", validPNG()))
	if status != http.StatusUnprocessableEntity || body["code"] != float64(httpserver.CodeInvalidState) {
		t.Fatalf("fourth upload status=%d body=%v", status, body)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payment_attachment WHERE payment_id = ?`, paymentID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("attachment count=%d", count)
	}
	assertTempDirEmpty(t, filepath.Join(storageRoot, "uploads", ".tmp"))
}

func TestAttachmentUploadRenameFailureCompensatesMetadataAndTemp(t *testing.T) {
	storage := NewStorage(t.TempDir())
	storage.rename = func(string, string) error { return errors.New("rename failed") }

	db, srv, _, _ := newEvidenceTestServer(t, Config{Storage: storage})
	defer db.Close()
	defer srv.Close()

	paymentID := seedEvidencePayment(t, db)
	operatorID := evidenceUser(t, db, "operator", "OPERATOR")
	token := evidenceToken(t, operatorID, "OPERATOR")

	status, body := evidenceJSONRequest(t, http.MethodPost, srv.URL+"/finance/payments/"+itoa(paymentID)+"/attachments", token, newMultipartBody(t, "file", "proof.png", validPNG()))
	if status != http.StatusInternalServerError || body["code"] != float64(httpserver.CodeDatabase) {
		t.Fatalf("rename failure status=%d body=%v", status, body)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payment_attachment WHERE payment_id = ?`, paymentID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("metadata remained after compensation: %d", count)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM operation_log WHERE action = 'PAYMENT_ATTACHMENT_UPLOAD'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("success audit remained after compensation: %d", count)
	}
	assertTempDirEmpty(t, storage.tmpDir)
}

func TestAttachmentDownloadRejectsWrongPaymentAndUnsafeMetadataPath(t *testing.T) {
	db, srv, storageRoot, _ := newEvidenceTestServer(t, Config{})
	defer db.Close()
	defer srv.Close()

	operatorID := evidenceUser(t, db, "operator", "OPERATOR")
	token := evidenceToken(t, operatorID, "OPERATOR")
	payment1 := seedEvidencePayment(t, db)
	payment2 := seedEvidencePayment(t, db)

	if err := os.MkdirAll(filepath.Join(storageRoot, "uploads"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(storageRoot, "outside-secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := db.Exec(`INSERT INTO payment_attachment (payment_id, file_name, file_path, file_type, file_size, uploaded_by) VALUES (?, 'proof.pdf', '../outside-secret.txt', 'application/pdf', 12, ?)`, payment1, operatorID)
	if err != nil {
		t.Fatal(err)
	}
	attachmentID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	status, body := evidenceJSONRequest(t, http.MethodGet, srv.URL+"/finance/payments/"+itoa(payment2)+"/attachments/"+itoa(attachmentID)+"/content", token, nil)
	if status != http.StatusNotFound || body["code"] != float64(httpserver.CodeNotFound) {
		t.Fatalf("wrong payment status=%d body=%v", status, body)
	}

	status, body = evidenceJSONRequest(t, http.MethodGet, srv.URL+"/finance/payments/"+itoa(payment1)+"/attachments/"+itoa(attachmentID)+"/content", token, nil)
	if status != http.StatusInternalServerError || body["code"] != float64(httpserver.CodeDatabase) {
		t.Fatalf("unsafe path status=%d body=%v", status, body)
	}
}

func newEvidenceTestServer(t *testing.T, cfg Config) (*sql.DB, *httptest.Server, string, *slog.Logger) {
	t.Helper()

	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "evidence.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	if cfg.DataRoot == "" && cfg.Storage == nil {
		cfg.DataRoot = t.TempDir()
	}
	mux := httpserver.New()
	MountRoutes(mux, NewHandler(db, logger, cfg), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	root := cfg.DataRoot
	if cfg.Storage != nil {
		root = cfg.Storage.dataRoot
	}
	return db, srv, root, logger
}

func seedEvidencePayment(t *testing.T, db *sql.DB) int64 {
	t.Helper()

	if _, err := db.Exec(`INSERT INTO course_domain (name, code, type) VALUES ('Language', ?, 'LANGUAGE')`, "LANG"+time.Now().Format("150405.000000000")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO course_track (domain_id, name, code) VALUES ((SELECT MAX(id) FROM course_domain), 'Track', ?)`, "TRK"+time.Now().Format("150405.000000000")); err != nil {
		t.Fatal(err)
	}
	result, err := db.Exec(`INSERT INTO student (name) VALUES (?)`, "Student "+time.Now().Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	studentID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	result, err = db.Exec(`INSERT INTO student_course_enrollment (student_id, domain_id, track_id) VALUES (?, (SELECT MAX(id) FROM course_domain), (SELECT MAX(id) FROM course_track))`, studentID)
	if err != nil {
		t.Fatal(err)
	}
	enrollmentID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	result, err = db.Exec(`INSERT INTO student_payment (payment_no, student_id, enrollment_id, original_amount, original_currency, fx_rate_to_base, amount_base, lessons_added, payment_method_code, paid_at, status) VALUES (?, ?, ?, '1000', 'JPY', '1', 1000, 1, 'CASH', '2026-07-19T00:00:00Z', 'CONFIRMED')`, "pay-"+time.Now().Format("20060102150405.000000000"), studentID, enrollmentID)
	if err != nil {
		t.Fatal(err)
	}
	paymentID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return paymentID
}

func evidenceUser(t *testing.T, db *sql.DB, username, role string) int64 {
	t.Helper()
	hash, err := platformauth.HashPassword("Pass1234")
	if err != nil {
		t.Fatal(err)
	}
	result, err := db.Exec(`INSERT INTO user_account (username, password_hash, role, display_name) VALUES (?, ?, ?, ?)`, username+"-"+time.Now().Format("150405.000000000"), hash, role, username)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func evidenceToken(t *testing.T, userID int64, role string) string {
	t.Helper()
	token, err := platformauth.SignAccessToken(testJWTSecret, userID, role, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

type multipartBody struct {
	contentType string
	body        []byte
}

func newMultipartBody(t *testing.T, field, name string, content []byte) multipartBody {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile(field, name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return multipartBody{contentType: writer.FormDataContentType(), body: buf.Bytes()}
}

func evidenceJSONRequest(t *testing.T, method, url, token string, body any) (int, map[string]any) {
	t.Helper()

	var reqBody io.Reader
	contentType := ""
	switch input := body.(type) {
	case nil:
	case multipartBody:
		reqBody = bytes.NewReader(input.body)
		contentType = input.contentType
	default:
		raw, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		reqBody = bytes.NewReader(raw)
		contentType = "application/json"
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	payload := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, payload
}

func evidenceRawRequest(t *testing.T, method, url, token string) (int, http.Header, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, resp.Header.Clone(), body
}

func assertTempDirEmpty(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("temp dir not empty: %d entries", len(entries))
	}
}

func assertAuditRow(t *testing.T, db *sql.DB, action string) {
	t.Helper()
	var detail, requestID string
	if err := db.QueryRow(`SELECT detail_json, request_id FROM operation_log WHERE action = ? ORDER BY id DESC LIMIT 1`, action).Scan(&detail, &requestID); err != nil {
		t.Fatalf("audit %s: %v", action, err)
	}
	if requestID == "" {
		t.Fatalf("audit %s missing request_id", action)
	}
	if strings.Contains(detail, ".tmp") || strings.Contains(detail, "uploads") || strings.Contains(detail, "\\") || strings.Contains(detail, "/payments/") {
		t.Fatalf("audit %s leaked path detail=%q", action, detail)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(detail), &payload); err != nil {
		t.Fatalf("audit %s invalid json: %v (%q)", action, err, detail)
	}
}

func validPNG() []byte {
	return []byte{
		0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R',
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89,
	}
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
