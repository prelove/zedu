package finance

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/prelove/zedu/backend/internal/repository"
)

type Repository struct{}

func (Repository) BaseCurrency(ctx context.Context, exec repository.Executor) (string, string, error) {
	var currency, locked string
	err := exec.QueryRowContext(ctx, `SELECT config_value FROM system_settings WHERE config_key='base_currency'`).Scan(&currency)
	if err != nil {
		return "", "", err
	}
	err = exec.QueryRowContext(ctx, `SELECT config_value FROM system_settings WHERE config_key='base_currency_locked'`).Scan(&locked)
	return currency, locked, err
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(strings.ToUpper(err.Error()), "UNIQUE")
}

type PaymentMethod struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
	Enabled   bool   `json:"enabled"`
}

type Payment struct {
	ID                int64  `json:"id"`
	PaymentNo         string `json:"paymentNo"`
	StudentID         int64  `json:"studentId"`
	EnrollmentID      int64  `json:"enrollmentId"`
	AmountBase        int64  `json:"amountBase"`
	LessonsAdded      int64  `json:"lessonsAdded"`
	Status            string `json:"status"`
	OriginalAmount    string `json:"originalAmount"`
	OriginalCurrency  string `json:"originalCurrency"`
	FXRateToBase      string `json:"fxRateToBase"`
	PaymentMethodCode string `json:"paymentMethodCode"`
	PaymentMethodName string `json:"paymentMethodName"`
	PaidAt            string `json:"paidAt"`
	Note              string `json:"note"`
}
type LedgerEntry struct {
	ID                 int64  `json:"id"`
	EnrollmentID       int64  `json:"enrollmentId"`
	BizType            string `json:"bizType"`
	AmountDelta        int64  `json:"amountDelta"`
	LessonDelta        int64  `json:"lessonDelta"`
	BalanceAfter       int64  `json:"balanceAfter"`
	LessonBalanceAfter int64  `json:"lessonBalanceAfter"`
	RelatedPaymentID   *int64 `json:"relatedPaymentId,omitempty"`
	Note               string `json:"note"`
	CreatedAt          string `json:"createdAt"`
}

func (Repository) InsertPayment(ctx context.Context, exec repository.Executor, p Payment, originalAmount, originalCurrency, fxRate, method, methodName, paidAt, note string, operatorID int64) (int64, bool, error) {
	result, err := exec.ExecContext(ctx, `INSERT OR IGNORE INTO student_payment (payment_no,student_id,enrollment_id,original_amount,original_currency,fx_rate_to_base,amount_base,lessons_added,payment_method_code,payment_method_name,paid_at,operator_id,note) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, p.PaymentNo, p.StudentID, p.EnrollmentID, originalAmount, originalCurrency, fxRate, p.AmountBase, p.LessonsAdded, method, methodName, paidAt, operatorID, note)
	if err != nil {
		return 0, false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, false, err
	}
	if rows == 0 {
		return 0, false, nil
	}
	id, err := result.LastInsertId()
	return id, true, err
}

func (Repository) PaymentByNo(ctx context.Context, exec repository.Executor, paymentNo string) (Payment, error) {
	var p Payment
	err := exec.QueryRowContext(ctx, `SELECT id,payment_no,student_id,enrollment_id,amount_base,lessons_added,status,original_amount,original_currency,fx_rate_to_base,payment_method_code,payment_method_name,paid_at,COALESCE(note,'') FROM student_payment WHERE payment_no=?`, paymentNo).Scan(&p.ID, &p.PaymentNo, &p.StudentID, &p.EnrollmentID, &p.AmountBase, &p.LessonsAdded, &p.Status, &p.OriginalAmount, &p.OriginalCurrency, &p.FXRateToBase, &p.PaymentMethodCode, &p.PaymentMethodName, &p.PaidAt, &p.Note)
	return p, err
}

func (Repository) PaymentByID(ctx context.Context, exec repository.Executor, id int64) (Payment, error) {
	var p Payment
	err := exec.QueryRowContext(ctx, `SELECT id,payment_no,student_id,enrollment_id,amount_base,lessons_added,status,original_amount,original_currency,fx_rate_to_base,payment_method_code,payment_method_name,paid_at,COALESCE(note,'') FROM student_payment WHERE id=?`, id).Scan(&p.ID, &p.PaymentNo, &p.StudentID, &p.EnrollmentID, &p.AmountBase, &p.LessonsAdded, &p.Status, &p.OriginalAmount, &p.OriginalCurrency, &p.FXRateToBase, &p.PaymentMethodCode, &p.PaymentMethodName, &p.PaidAt, &p.Note)
	return p, err
}
func (Repository) ListPayments(ctx context.Context, exec repository.Executor, filter PaymentFilter, limit, offset int) ([]Payment, int, error) {
	where, args := paymentFilterSQL(filter)
	var total int
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM student_payment`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(args, limit, offset)
	rows, err := exec.QueryContext(ctx, `SELECT id,payment_no,student_id,enrollment_id,amount_base,lessons_added,status,payment_method_code,payment_method_name FROM student_payment`+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Payment{}
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.PaymentNo, &p.StudentID, &p.EnrollmentID, &p.AmountBase, &p.LessonsAdded, &p.Status, &p.PaymentMethodCode, &p.PaymentMethodName); err != nil {
			return nil, 0, err
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}

func paymentFilterSQL(filter PaymentFilter) (string, []any) {
	where := " WHERE 1=1"
	args := []any{}
	if filter.PaymentNo != "" {
		where += " AND payment_no=?"
		args = append(args, filter.PaymentNo)
	}
	if filter.StudentID > 0 {
		where += " AND student_id=?"
		args = append(args, filter.StudentID)
	}
	if filter.EnrollmentID > 0 {
		where += " AND enrollment_id=?"
		args = append(args, filter.EnrollmentID)
	}
	if filter.Status != "" {
		where += " AND status=?"
		args = append(args, filter.Status)
	}
	return where, args
}
func (Repository) ListStudentLedger(ctx context.Context, exec repository.Executor, studentID int64, limit, offset int) ([]LedgerEntry, int, error) {
	var total int
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM student_account_ledger WHERE student_id=?`, studentID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := exec.QueryContext(ctx, `SELECT id,enrollment_id,biz_type,amount_delta,lesson_delta,balance_after,lesson_balance_after,related_payment_id,note,created_at FROM student_account_ledger WHERE student_id=? ORDER BY id DESC LIMIT ? OFFSET ?`, studentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []LedgerEntry{}
	for rows.Next() {
		var item LedgerEntry
		var related sql.NullInt64
		if err := rows.Scan(&item.ID, &item.EnrollmentID, &item.BizType, &item.AmountDelta, &item.LessonDelta, &item.BalanceAfter, &item.LessonBalanceAfter, &related, &item.Note, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		if related.Valid {
			item.RelatedPaymentID = &related.Int64
		}
		out = append(out, item)
	}
	return out, total, rows.Err()
}
func (Repository) VoidPayment(ctx context.Context, exec repository.Executor, id int64, reason string) (bool, error) {
	result, err := exec.ExecContext(ctx, `UPDATE student_payment SET status='VOIDED',voided_at=CURRENT_TIMESTAMP,void_reason=? WHERE id=? AND status='CONFIRMED'`, reason, id)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}
func (Repository) InsertVoidLedger(ctx context.Context, exec repository.Executor, p Payment, amountAfter int64, lessonAfter int64, operatorID int64, note string) error {
	_, err := exec.ExecContext(ctx, `INSERT INTO student_account_ledger (student_id,enrollment_id,biz_type,amount_delta,lesson_delta,balance_after,lesson_balance_after,related_payment_id,operator_id,note) VALUES (?,?,'VOID',?,?,?,?,?,?,?)`, p.StudentID, p.EnrollmentID, -p.AmountBase, -p.LessonsAdded, amountAfter, lessonAfter, p.ID, operatorID, note)
	return err
}

func (Repository) ActiveEnrollment(ctx context.Context, exec repository.Executor, studentID, enrollmentID int64) (bool, error) {
	var n int
	err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM student_course_enrollment e JOIN student s ON s.id=e.student_id WHERE e.id=? AND e.student_id=? AND e.status='ACTIVE' AND e.deleted_at IS NULL AND s.status='ACTIVE' AND s.deleted_at IS NULL`, enrollmentID, studentID).Scan(&n)
	return n == 1, err
}
func (Repository) EnabledPaymentMethod(ctx context.Context, exec repository.Executor, code string) (string, bool, error) {
	var name string
	err := exec.QueryRowContext(ctx, `SELECT name FROM payment_method WHERE code=? AND enabled=1`, code).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	return name, err == nil, err
}
func (Repository) EnrollmentBalances(ctx context.Context, exec repository.Executor, id int64) (int64, int64, error) {
	var amount int64
	var lessons int64
	err := exec.QueryRowContext(ctx, `SELECT balance_amount,lesson_balance FROM student_course_enrollment WHERE id=?`, id).Scan(&amount, &lessons)
	return amount, lessons, err
}
func (Repository) UpdateEnrollmentBalances(ctx context.Context, exec repository.Executor, id int64, amount int64, lessons int64) error {
	_, err := exec.ExecContext(ctx, `UPDATE student_course_enrollment SET balance_amount=?,lesson_balance=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`, amount, lessons, id)
	return err
}
func (Repository) InsertRechargeLedger(ctx context.Context, exec repository.Executor, p Payment, amountAfter int64, lessonAfter int64, operatorID int64, note string) error {
	_, err := exec.ExecContext(ctx, `INSERT INTO student_account_ledger (student_id,enrollment_id,biz_type,amount_delta,lesson_delta,balance_after,lesson_balance_after,related_payment_id,operator_id,note) VALUES (?,?,'RECHARGE',?,?,?,?,?,?,?)`, p.StudentID, p.EnrollmentID, p.AmountBase, p.LessonsAdded, amountAfter, lessonAfter, p.ID, operatorID, note)
	return err
}
func (Repository) LockBaseCurrency(ctx context.Context, exec repository.Executor) error {
	_, err := exec.ExecContext(ctx, `UPDATE system_settings SET config_value='true',updated_at=CURRENT_TIMESTAMP WHERE config_key='base_currency_locked'`)
	return err
}

func (Repository) ListPaymentMethods(ctx context.Context, exec repository.Executor, includeDisabled bool) ([]PaymentMethod, error) {
	query := `SELECT code,name,sort_order,enabled FROM payment_method`
	if !includeDisabled {
		query += ` WHERE enabled=1`
	}
	query += ` ORDER BY sort_order,code`
	rows, err := exec.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PaymentMethod{}
	for rows.Next() {
		var item PaymentMethod
		var enabled int
		if err := rows.Scan(&item.Code, &item.Name, &item.SortOrder, &enabled); err != nil {
			return nil, err
		}
		item.Enabled = enabled == 1
		out = append(out, item)
	}
	return out, rows.Err()
}
func (Repository) CreatePaymentMethod(ctx context.Context, exec repository.Executor, item PaymentMethod) error {
	_, err := exec.ExecContext(ctx, `INSERT INTO payment_method (code,name,sort_order,enabled) VALUES (?,?,?,?)`, item.Code, item.Name, item.SortOrder, boolInt(item.Enabled))
	return err
}
func (Repository) UpdatePaymentMethod(ctx context.Context, exec repository.Executor, code string, item PaymentMethod) (bool, error) {
	result, err := exec.ExecContext(ctx, `UPDATE payment_method SET name=?,sort_order=?,enabled=? WHERE code=?`, item.Name, item.SortOrder, boolInt(item.Enabled), code)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n > 0, err
}
func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func (Repository) HasM3FinancialFacts(ctx context.Context, exec repository.Executor) (bool, error) {
	var n int
	err := exec.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM student_payment) + (SELECT COUNT(*) FROM student_account_ledger)`).Scan(&n)
	return n > 0, err
}

func (Repository) UpdateBaseCurrency(ctx context.Context, exec repository.Executor, currency string) error {
	_, err := exec.ExecContext(ctx, `UPDATE system_settings SET config_value=?, updated_at=CURRENT_TIMESTAMP WHERE config_key='base_currency'`, currency)
	return err
}

func (Repository) InsertAudit(ctx context.Context, exec repository.Executor, userID int64, action, requestID string) error {
	var name string
	if err := exec.QueryRowContext(ctx, `SELECT username FROM user_account WHERE id=?`, userID).Scan(&name); err != nil {
		return err
	}
	_, err := exec.ExecContext(ctx, `INSERT INTO operation_log (operator_id, operator_name, action, target_type, target_id, detail_json, request_id) VALUES (?, ?, ?, 'system', 1, '{}', ?)`, userID, name, action, requestID)
	return err
}
