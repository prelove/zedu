package finance

import (
	"context"
	"database/sql"
	"errors"
	"math/big"
	"regexp"
	"strings"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

var ErrInvalidState = errors.New("invalid state")
var ErrForbidden = errors.New("forbidden")
var ErrConflict = errors.New("conflict")

type Service struct {
	db   repository.DB
	repo Repository
}

func (s *Service) ListPaymentMethods(ctx context.Context, user httpserver.AuthUser) ([]PaymentMethod, error) {
	items, err := s.repo.ListPaymentMethods(ctx, s.db, user.Role == "OWNER")
	if err != nil {
		return nil, repository.ErrDatabase
	}
	return items, nil
}
func (s *Service) CreatePaymentMethod(ctx context.Context, user httpserver.AuthUser, item PaymentMethod, requestID string) (PaymentMethod, error) {
	if user.Role != "OWNER" {
		return PaymentMethod{}, ErrForbidden
	}
	item.Code = strings.ToUpper(strings.TrimSpace(item.Code))
	item.Name = strings.TrimSpace(item.Name)
	if !validPaymentMethod(item) {
		return PaymentMethod{}, ErrInvalidState
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PaymentMethod{}, repository.ErrDatabase
	}
	defer tx.Rollback()
	if err := s.repo.CreatePaymentMethod(ctx, tx, item); err != nil {
		if isUniqueViolation(err) {
			return PaymentMethod{}, ErrConflict
		}
		return PaymentMethod{}, repository.ErrDatabase
	}
	if err := s.repo.InsertAudit(ctx, tx, user.UserID, "PAYMENT_METHOD_CREATE", requestID); err != nil {
		return PaymentMethod{}, repository.ErrDatabase
	}
	if err := tx.Commit(); err != nil {
		return PaymentMethod{}, repository.ErrDatabase
	}
	return item, nil
}
func (s *Service) UpdatePaymentMethod(ctx context.Context, user httpserver.AuthUser, code string, item PaymentMethod, requestID string) (PaymentMethod, error) {
	if user.Role != "OWNER" {
		return PaymentMethod{}, ErrForbidden
	}
	item.Name = strings.TrimSpace(item.Name)
	item.Code = strings.ToUpper(strings.TrimSpace(code))
	if item.Code == "" || !validPaymentMethod(item) {
		return PaymentMethod{}, ErrInvalidState
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PaymentMethod{}, repository.ErrDatabase
	}
	defer tx.Rollback()
	found, err := s.repo.UpdatePaymentMethod(ctx, tx, item.Code, item)
	if err != nil {
		return PaymentMethod{}, repository.ErrDatabase
	}
	if !found {
		return PaymentMethod{}, ErrNotFound
	}
	if err := s.repo.InsertAudit(ctx, tx, user.UserID, "PAYMENT_METHOD_UPDATE", requestID); err != nil {
		return PaymentMethod{}, repository.ErrDatabase
	}
	if err := tx.Commit(); err != nil {
		return PaymentMethod{}, repository.ErrDatabase
	}
	return item, nil
}

var ErrNotFound = errors.New("not found")

func validPaymentMethod(item PaymentMethod) bool {
	if item.Name == "" || item.SortOrder < 0 {
		return false
	}
	for _, r := range item.Code {
		if !(r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
			return false
		}
	}
	return item.Code != ""
}

type PaymentWrite struct {
	PaymentNo         string `json:"paymentNo"`
	StudentID         int64  `json:"studentId"`
	EnrollmentID      int64  `json:"enrollmentId"`
	OriginalAmount    string `json:"originalAmount"`
	OriginalCurrency  string `json:"originalCurrency"`
	FXRateToBase      string `json:"fxRateToBase"`
	LessonsAdded      int64  `json:"lessonsAdded"`
	PaymentMethodCode string `json:"paymentMethodCode"`
	PaidAt            string `json:"paidAt"`
	Note              string `json:"note"`
}

func (s *Service) CreatePayment(ctx context.Context, user httpserver.AuthUser, w PaymentWrite, requestID string) (Payment, bool, error) {
	if !validPaymentWrite(w) {
		return Payment{}, false, ErrInvalidState
	}
	amount, err := baseAmount(w.OriginalAmount, w.FXRateToBase)
	if err != nil {
		return Payment{}, false, ErrInvalidState
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	defer tx.Rollback()
	baseCurrency, _, err := s.repo.BaseCurrency(ctx, tx)
	if err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	if w.OriginalCurrency != "JPY" && w.OriginalCurrency != "CNY" && w.OriginalCurrency != "USD" {
		return Payment{}, false, ErrInvalidState
	}
	if w.OriginalCurrency == baseCurrency && w.FXRateToBase != "1" {
		return Payment{}, false, ErrInvalidState
	}
	existing, existingErr := s.repo.PaymentByNo(ctx, tx, w.PaymentNo)
	if existingErr == nil {
		if samePaymentRequest(existing, w, amount) {
			return existing, true, nil
		}
		return Payment{}, false, ErrConflict
	}
	if !errors.Is(existingErr, sql.ErrNoRows) {
		return Payment{}, false, repository.ErrDatabase
	}
	active, err := s.repo.ActiveEnrollment(ctx, tx, w.StudentID, w.EnrollmentID)
	if err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	methodName, method, err := s.repo.EnabledPaymentMethod(ctx, tx, w.PaymentMethodCode)
	if err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	if !active || !method {
		return Payment{}, false, ErrInvalidState
	}
	p := Payment{PaymentNo: w.PaymentNo, StudentID: w.StudentID, EnrollmentID: w.EnrollmentID, AmountBase: amount, LessonsAdded: w.LessonsAdded, Status: "CONFIRMED", PaymentMethodCode: w.PaymentMethodCode, PaymentMethodName: methodName}
	id, inserted, err := s.repo.InsertPayment(ctx, tx, p, w.OriginalAmount, w.OriginalCurrency, w.FXRateToBase, w.PaymentMethodCode, methodName, w.PaidAt, w.Note, user.UserID)
	if err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	if !inserted {
		existing, err := s.repo.PaymentByNo(ctx, tx, w.PaymentNo)
		if err != nil {
			return Payment{}, false, repository.ErrDatabase
		}
		if samePaymentRequest(existing, w, amount) {
			return existing, true, nil
		}
		return Payment{}, false, ErrConflict
	}
	p.ID = id
	balance, lessons, err := s.repo.EnrollmentBalances(ctx, tx, w.EnrollmentID)
	if err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	if err := s.repo.InsertRechargeLedger(ctx, tx, p, balance+amount, lessons+float64(w.LessonsAdded), user.UserID, w.Note); err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	if err := s.repo.UpdateEnrollmentBalances(ctx, tx, w.EnrollmentID, balance+amount, lessons+float64(w.LessonsAdded)); err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	if err := s.repo.LockBaseCurrency(ctx, tx); err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	if err := s.repo.InsertAudit(ctx, tx, user.UserID, "PAYMENT_CREATE", requestID); err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	if err := tx.Commit(); err != nil {
		return Payment{}, false, repository.ErrDatabase
	}
	return p, false, nil
}

func samePaymentRequest(p Payment, w PaymentWrite, amount int64) bool {
	return p.StudentID == w.StudentID && p.EnrollmentID == w.EnrollmentID && p.AmountBase == amount && p.LessonsAdded == w.LessonsAdded && p.OriginalAmount == w.OriginalAmount && p.OriginalCurrency == w.OriginalCurrency && p.FXRateToBase == w.FXRateToBase && p.PaymentMethodCode == w.PaymentMethodCode && p.PaidAt == w.PaidAt && p.Note == w.Note
}
func validPaymentWrite(w PaymentWrite) bool {
	return uuidPattern.MatchString(w.PaymentNo) && w.StudentID > 0 && w.EnrollmentID > 0 && w.OriginalAmount != "" && w.OriginalCurrency != "" && w.FXRateToBase != "" && w.LessonsAdded > 0 && w.PaymentMethodCode != "" && w.PaidAt != ""
}

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func baseAmount(amount, rate string) (int64, error) {
	a, ok := new(big.Rat).SetString(amount)
	if !ok || a.Sign() <= 0 {
		return 0, ErrInvalidState
	}
	r, ok := new(big.Rat).SetString(rate)
	if !ok || r.Sign() <= 0 {
		return 0, ErrInvalidState
	}
	v := new(big.Rat).Mul(a, r)
	q := new(big.Int).Quo(v.Num(), v.Denom())
	rem := new(big.Int).Mod(v.Num(), v.Denom())
	if new(big.Int).Mul(rem, big.NewInt(2)).Cmp(v.Denom()) >= 0 {
		q.Add(q, big.NewInt(1))
	}
	if !q.IsInt64() {
		return 0, ErrInvalidState
	}
	return q.Int64(), nil
}

func (s *Service) VoidPayment(ctx context.Context, user httpserver.AuthUser, id int64, reason, requestID string) (Payment, error) {
	if id <= 0 || strings.TrimSpace(reason) == "" {
		return Payment{}, ErrInvalidState
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Payment{}, repository.ErrDatabase
	}
	defer tx.Rollback()
	p, err := s.repo.PaymentByID(ctx, tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Payment{}, ErrNotFound
	}
	if err != nil {
		return Payment{}, repository.ErrDatabase
	}
	if p.Status != "CONFIRMED" {
		return Payment{}, ErrInvalidState
	}
	balance, lessons, err := s.repo.EnrollmentBalances(ctx, tx, p.EnrollmentID)
	if err != nil {
		return Payment{}, repository.ErrDatabase
	}
	ok, err := s.repo.VoidPayment(ctx, tx, id, strings.TrimSpace(reason))
	if err != nil {
		return Payment{}, repository.ErrDatabase
	}
	if !ok {
		return Payment{}, ErrInvalidState
	}
	if err := s.repo.InsertVoidLedger(ctx, tx, p, balance-p.AmountBase, lessons-float64(p.LessonsAdded), user.UserID, reason); err != nil {
		return Payment{}, repository.ErrDatabase
	}
	if err := s.repo.UpdateEnrollmentBalances(ctx, tx, p.EnrollmentID, balance-p.AmountBase, lessons-float64(p.LessonsAdded)); err != nil {
		return Payment{}, repository.ErrDatabase
	}
	if err := s.repo.InsertAudit(ctx, tx, user.UserID, "PAYMENT_VOID", requestID); err != nil {
		return Payment{}, repository.ErrDatabase
	}
	if err := tx.Commit(); err != nil {
		return Payment{}, repository.ErrDatabase
	}
	p.Status = "VOIDED"
	return p, nil
}

type Page[T any] struct {
	Items    []T `json:"items"`
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

type PaymentFilter struct {
	PaymentNo    string
	StudentID    int64
	EnrollmentID int64
	Status       string
}

func (s *Service) ListPayments(ctx context.Context, filter PaymentFilter, page, pageSize int) (Page[Payment], error) {
	filter.PaymentNo = strings.TrimSpace(filter.PaymentNo)
	filter.Status = strings.ToUpper(strings.TrimSpace(filter.Status))
	if filter.StudentID < 0 || filter.EnrollmentID < 0 || (filter.Status != "" && filter.Status != "CONFIRMED" && filter.Status != "VOIDED") {
		return Page[Payment]{}, ErrInvalidState
	}
	items, total, err := s.repo.ListPayments(ctx, s.db, filter, pageSize, (page-1)*pageSize)
	if err != nil {
		return Page[Payment]{}, repository.ErrDatabase
	}
	return Page[Payment]{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}
func (s *Service) GetPayment(ctx context.Context, id int64) (Payment, error) {
	if id <= 0 {
		return Payment{}, ErrNotFound
	}
	p, err := s.repo.PaymentByID(ctx, s.db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return Payment{}, ErrNotFound
	}
	if err != nil {
		return Payment{}, repository.ErrDatabase
	}
	return p, nil
}
func (s *Service) ListStudentLedger(ctx context.Context, studentID int64, page, pageSize int) (Page[LedgerEntry], error) {
	if studentID <= 0 {
		return Page[LedgerEntry]{}, ErrNotFound
	}
	items, total, err := s.repo.ListStudentLedger(ctx, s.db, studentID, pageSize, (page-1)*pageSize)
	if err != nil {
		return Page[LedgerEntry]{}, repository.ErrDatabase
	}
	return Page[LedgerEntry]{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func NewService(db repository.DB) *Service { return &Service{db: db, repo: Repository{}} }

type BaseCurrency struct {
	Currency string `json:"currency"`
	Locked   bool   `json:"locked"`
}

func (s *Service) GetBaseCurrency(ctx context.Context, _ httpserver.AuthUser) (BaseCurrency, error) {
	currency, locked, err := s.repo.BaseCurrency(ctx, s.db)
	if err != nil {
		return BaseCurrency{}, repository.ErrDatabase
	}
	return BaseCurrency{Currency: currency, Locked: locked == "true"}, nil
}

func (s *Service) UpdateBaseCurrency(ctx context.Context, user httpserver.AuthUser, currency, requestID string) (BaseCurrency, error) {
	if user.Role != "OWNER" {
		return BaseCurrency{}, ErrForbidden
	}
	if currency != "JPY" && currency != "CNY" && currency != "USD" {
		return BaseCurrency{}, ErrInvalidState
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return BaseCurrency{}, repository.ErrDatabase
	}
	defer tx.Rollback()
	current, locked, err := s.repo.BaseCurrency(ctx, tx)
	if err != nil {
		return BaseCurrency{}, repository.ErrDatabase
	}
	facts, err := s.repo.HasM3FinancialFacts(ctx, tx)
	if err != nil {
		return BaseCurrency{}, repository.ErrDatabase
	}
	if locked == "true" || facts {
		return BaseCurrency{}, ErrInvalidState
	}
	if current != currency {
		if err := s.repo.UpdateBaseCurrency(ctx, tx, currency); err != nil {
			return BaseCurrency{}, repository.ErrDatabase
		}
	}
	if err := s.repo.InsertAudit(ctx, tx, user.UserID, "BASE_CURRENCY_UPDATE", requestID); err != nil {
		return BaseCurrency{}, repository.ErrDatabase
	}
	if err := tx.Commit(); err != nil {
		return BaseCurrency{}, repository.ErrDatabase
	}
	return BaseCurrency{Currency: currency, Locked: false}, nil
}
