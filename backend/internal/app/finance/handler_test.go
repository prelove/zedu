package finance

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

const testJWTSecret = "finance-test-secret-must-be-at-least-32-chars"

func TestBaseCurrencyOwnerCanReadAndOperatorCannotChange(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "finance.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	MountRoutes(mux, NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()

	ownerID := financeUser(t, db, "owner", "OWNER")
	operatorID := financeUser(t, db, "operator", "OPERATOR")
	ownerToken := financeToken(t, ownerID, "OWNER")
	operatorToken := financeToken(t, operatorID, "OPERATOR")

	status, body := financeRequest(t, http.MethodGet, srv.URL+"/system/base-currency", ownerToken, nil)
	if status != http.StatusOK || body["code"] != float64(0) {
		t.Fatalf("owner read status=%d body=%v", status, body)
	}
	status, body = financeRequest(t, http.MethodPut, srv.URL+"/system/base-currency", operatorToken, map[string]string{"currency": "CNY"})
	if status != http.StatusForbidden || body["code"] != float64(httpserver.CodeForbidden) {
		t.Fatalf("operator update status=%d body=%v", status, body)
	}
}

func TestPaymentMethodRoleFilteringAndOwnerWrite(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "methods.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	MountRoutes(mux, NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()
	ownerID := financeUser(t, db, "owner", "OWNER")
	operatorID := financeUser(t, db, "operator", "OPERATOR")
	owner := financeToken(t, ownerID, "OWNER")
	operator := financeToken(t, operatorID, "OPERATOR")
	status, body := financeRequest(t, http.MethodPatch, srv.URL+"/system/payment-methods/WECHAT", owner, map[string]any{"name": "微信", "sortOrder": 10, "enabled": false})
	if status != http.StatusOK || body["code"] != float64(0) {
		t.Fatalf("owner update: %d %v", status, body)
	}
	status, body = financeRequest(t, http.MethodGet, srv.URL+"/system/payment-methods", operator, nil)
	if status != http.StatusOK || body["code"] != float64(0) {
		t.Fatalf("operator list: %d %v", status, body)
	}
	items := body["data"].([]any)
	for _, raw := range items {
		if raw.(map[string]any)["code"] == "WECHAT" {
			t.Fatal("operator must not receive disabled payment method")
		}
	}
	status, body = financeRequest(t, http.MethodPost, srv.URL+"/system/payment-methods", operator, map[string]any{"code": "CARD", "name": "Card", "sortOrder": 60, "enabled": true})
	if status != http.StatusForbidden || body["code"] != float64(httpserver.CodeForbidden) {
		t.Fatalf("operator create: %d %v", status, body)
	}
}

func TestPaymentMethodDuplicateCodeReturnsConflict(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "duplicate.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	MountRoutes(mux, NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()
	ownerID := financeUser(t, db, "owner", "OWNER")
	owner := financeToken(t, ownerID, "OWNER")
	status, body := financeRequest(t, http.MethodPost, srv.URL+"/system/payment-methods", owner, map[string]any{"code": "WECHAT", "name": "duplicate", "sortOrder": 1, "enabled": true})
	if status != http.StatusConflict || body["code"] != float64(httpserver.CodeConflict) {
		t.Fatalf("duplicate create: %d %v", status, body)
	}
}

func TestLockedBaseCurrencyRejectsOwnerChangeWithoutAudit(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "locked.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE system_settings SET config_value='true' WHERE config_key='base_currency_locked'`); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	MountRoutes(mux, NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()
	ownerID := financeUser(t, db, "owner", "OWNER")
	owner := financeToken(t, ownerID, "OWNER")
	status, body := financeRequest(t, http.MethodPut, srv.URL+"/system/base-currency", owner, map[string]string{"currency": "CNY"})
	if status != http.StatusUnprocessableEntity || body["code"] != float64(httpserver.CodeInvalidState) {
		t.Fatalf("locked update: %d %v", status, body)
	}
	var currency string
	if err := db.QueryRow(`SELECT config_value FROM system_settings WHERE config_key='base_currency'`).Scan(&currency); err != nil {
		t.Fatal(err)
	}
	if currency != "JPY" {
		t.Fatalf("currency changed to %q", currency)
	}
	var audits int
	if err := db.QueryRow(`SELECT COUNT(*) FROM operation_log`).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if audits != 0 {
		t.Fatalf("rejected update wrote %d audits", audits)
	}
}

func TestBaseCurrencyAuditFailureRollsBack(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "audit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_currency_audit BEFORE INSERT ON operation_log BEGIN SELECT RAISE(FAIL, 'audit failed'); END`); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	MountRoutes(mux, NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()
	ownerID := financeUser(t, db, "owner", "OWNER")
	owner := financeToken(t, ownerID, "OWNER")
	status, body := financeRequest(t, http.MethodPut, srv.URL+"/system/base-currency", owner, map[string]string{"currency": "CNY"})
	if status != http.StatusInternalServerError || body["code"] != float64(httpserver.CodeDatabase) {
		t.Fatalf("audit failure: %d %v", status, body)
	}
	var currency string
	if err := db.QueryRow(`SELECT config_value FROM system_settings WHERE config_key='base_currency'`).Scan(&currency); err != nil {
		t.Fatal(err)
	}
	if currency != "JPY" {
		t.Fatalf("audit failure committed currency %q", currency)
	}
}

func TestPaymentMethodAuditFailureRollsBack(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "method-audit.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_method_audit BEFORE INSERT ON operation_log BEGIN SELECT RAISE(FAIL, 'audit failed'); END`); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	MountRoutes(mux, NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()
	ownerID := financeUser(t, db, "owner", "OWNER")
	owner := financeToken(t, ownerID, "OWNER")
	status, body := financeRequest(t, http.MethodPost, srv.URL+"/system/payment-methods", owner, map[string]any{"code": "CARD", "name": "Card", "sortOrder": 60, "enabled": true})
	if status != http.StatusInternalServerError || body["code"] != float64(httpserver.CodeDatabase) {
		t.Fatalf("audit failure: %d %v", status, body)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM payment_method WHERE code='CARD'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("audit failure committed %d payment methods", count)
	}
}

func TestCreatePaymentWritesAtomicRechargeFacts(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "payment.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO course_domain (name,code,type) VALUES ('语言','LANG','LANGUAGE')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO course_track (domain_id,name,code) VALUES (1,'方向','TRACK')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO student (name) VALUES ('学生')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO student_course_enrollment (student_id,domain_id,track_id) VALUES (1,1,1)`); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	MountRoutes(mux, NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()
	operatorID := financeUser(t, db, "operator", "OPERATOR")
	token := financeToken(t, operatorID, "OPERATOR")
	payload := map[string]any{"paymentNo": "11111111-1111-4111-8111-111111111111", "studentId": 1, "enrollmentId": 1, "originalAmount": "500.00", "originalCurrency": "CNY", "fxRateToBase": "21.8", "lessonsAdded": 10, "paymentMethodCode": "CASH", "paidAt": "2026-07-19T00:00:00Z"}
	invalidPaymentNo := map[string]any{}
	for key, value := range payload {
		invalidPaymentNo[key] = value
	}
	invalidPaymentNo["paymentNo"] = "not-a-uuid"
	status, body := financeRequest(t, http.MethodPost, srv.URL+"/finance/payments", token, invalidPaymentNo)
	if status != http.StatusUnprocessableEntity || body["code"] != float64(httpserver.CodeInvalidState) {
		t.Fatalf("invalid paymentNo: %d %v", status, body)
	}
	invalidRate := map[string]any{}
	for key, value := range payload {
		invalidRate[key] = value
	}
	invalidRate["originalCurrency"] = "JPY"
	invalidRate["fxRateToBase"] = "2"
	status, body = financeRequest(t, http.MethodPost, srv.URL+"/finance/payments", token, invalidRate)
	if status != http.StatusUnprocessableEntity || body["code"] != float64(httpserver.CodeInvalidState) {
		t.Fatalf("same currency invalid rate: %d %v", status, body)
	}
	unsupportedCurrency := map[string]any{}
	for key, value := range payload {
		unsupportedCurrency[key] = value
	}
	unsupportedCurrency["originalCurrency"] = "EUR"
	status, body = financeRequest(t, http.MethodPost, srv.URL+"/finance/payments", token, unsupportedCurrency)
	if status != http.StatusUnprocessableEntity || body["code"] != float64(httpserver.CodeInvalidState) {
		t.Fatalf("unsupported currency: %d %v", status, body)
	}
	status, body = financeRequest(t, http.MethodPost, srv.URL+"/finance/payments", token, payload)
	if status != http.StatusCreated || body["code"] != float64(0) {
		t.Fatalf("create payment: %d %v", status, body)
	}
	var payments, ledger, balance int
	if err := db.QueryRow(`SELECT COUNT(*) FROM student_payment`).Scan(&payments); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM student_account_ledger`).Scan(&ledger); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT balance_amount FROM student_course_enrollment WHERE id=1`).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if payments != 1 || ledger != 1 || balance != 10900 {
		t.Fatalf("facts payments=%d ledger=%d balance=%d", payments, ledger, balance)
	}
	status, body = financeRequest(t, http.MethodGet, srv.URL+"/finance/payments/1", token, nil)
	if status != http.StatusOK || body["code"] != float64(0) {
		t.Fatalf("payment detail: %d %v", status, body)
	}
	detail := body["data"].(map[string]any)
	if detail["originalAmount"] != "500.00" || detail["originalCurrency"] != "CNY" || detail["fxRateToBase"] != "21.8" {
		t.Fatalf("payment detail snapshot: %v", detail)
	}
	if detail["paymentMethodName"] == "" {
		t.Fatalf("payment method display snapshot missing: %v", detail)
	}
	status, body = financeRequest(t, http.MethodGet, srv.URL+"/finance/payments?paymentNo=11111111-1111-4111-8111-111111111111&status=CONFIRMED", token, nil)
	if status != http.StatusOK || body["code"] != float64(0) || body["data"].(map[string]any)["total"] != float64(1) {
		t.Fatalf("payment filter: %d %v", status, body)
	}
	status, body = financeRequest(t, http.MethodGet, srv.URL+"/finance/ledger/student/1", token, nil)
	if status != http.StatusOK || body["code"] != float64(0) || body["data"].(map[string]any)["total"] != float64(1) {
		t.Fatalf("student ledger: %d %v", status, body)
	}
	status, body = financeRequest(t, http.MethodPost, srv.URL+"/finance/payments", token, payload)
	if status != http.StatusOK || body["code"] != float64(0) {
		t.Fatalf("idempotent replay: %d %v", status, body)
	}
	payload["lessonsAdded"] = 9
	status, body = financeRequest(t, http.MethodPost, srv.URL+"/finance/payments", token, payload)
	if status != http.StatusConflict || body["code"] != float64(httpserver.CodeConflict) {
		t.Fatalf("conflicting replay: %d %v", status, body)
	}
}

func TestVoidPaymentWritesReversalOnce(t *testing.T) {
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "void.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO course_domain (name,code,type) VALUES ('语言','LANG','LANGUAGE'); INSERT INTO course_track (domain_id,name,code) VALUES (1,'方向','TRACK'); INSERT INTO student (name) VALUES ('学生'); INSERT INTO student_course_enrollment (student_id,domain_id,track_id) VALUES (1,1,1)`); err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	mux := httpserver.New()
	MountRoutes(mux, NewHandler(db, logger), db, testJWTSecret)
	srv := httptest.NewServer(logging.NewMiddleware(logger)(mux))
	defer srv.Close()
	operatorID := financeUser(t, db, "operator", "OPERATOR")
	token := financeToken(t, operatorID, "OPERATOR")
	payload := map[string]any{"paymentNo": "22222222-2222-4222-8222-222222222222", "studentId": 1, "enrollmentId": 1, "originalAmount": "1000", "originalCurrency": "JPY", "fxRateToBase": "1", "lessonsAdded": 2, "paymentMethodCode": "CASH", "paidAt": "2026-07-19T00:00:00Z"}
	status, _ := financeRequest(t, http.MethodPost, srv.URL+"/finance/payments", token, payload)
	if status != http.StatusCreated {
		t.Fatalf("create=%d", status)
	}
	status, body := financeRequest(t, http.MethodPost, srv.URL+"/finance/payments/1/void", token, map[string]string{"reason": "录入错误"})
	if status != http.StatusOK || body["code"] != float64(0) {
		t.Fatalf("void: %d %v", status, body)
	}
	var statusText string
	var balance int
	if err := db.QueryRow(`SELECT status FROM student_payment WHERE id=1`).Scan(&statusText); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT balance_amount FROM student_course_enrollment WHERE id=1`).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if statusText != "VOIDED" || balance != 0 {
		t.Fatalf("status=%s balance=%d", statusText, balance)
	}
	status, body = financeRequest(t, http.MethodPost, srv.URL+"/finance/payments/1/void", token, map[string]string{"reason": "again"})
	if status != http.StatusUnprocessableEntity || body["code"] != float64(httpserver.CodeInvalidState) {
		t.Fatalf("repeat void: %d %v", status, body)
	}
}

func financeUser(t *testing.T, db *sql.DB, username, role string) int64 {
	t.Helper()
	hash, err := platformauth.HashPassword("Pass1234")
	if err != nil {
		t.Fatal(err)
	}
	result, err := db.Exec(`INSERT INTO user_account (username, password_hash, role, display_name) VALUES (?, ?, ?, ?)`, username, hash, role, username)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func financeToken(t *testing.T, userID int64, role string) string {
	t.Helper()
	token, err := platformauth.SignAccessToken(testJWTSecret, userID, role, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func financeRequest(t *testing.T, method, url, token string, body any) (int, map[string]any) {
	t.Helper()
	var input io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		input = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, url, input)
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	out := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, out
}
