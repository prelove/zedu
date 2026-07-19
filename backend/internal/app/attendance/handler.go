package attendance

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
	"net/http"
	"strconv"
	"strings"
)

var ErrInvalid = errors.New("invalid")

type ConfirmWrite struct {
	OutcomeType       string `json:"outcomeType"`
	LessonDeducted    string `json:"lessonDeducted"`
	ChargeAmount      int64  `json:"chargeAmount"`
	TeacherPayAmount  int64  `json:"teacherPayAmount"`
	ActualDurationMin int    `json:"actualDurationMin"`
	Note              string `json:"note"`
}
type Handler struct{ db repository.DB }

func NewHandler(db any) *Handler { return &Handler{repository.AsDB(db)} }
func MountRoutes(mux *http.ServeMux, h *Handler, db *sql.DB, secret string) {
	mux.Handle("POST /lessons/{id}/confirm", httpserver.AuthMiddleware(secret, db)(http.HandlerFunc(h.confirm)))
}
func (h *Handler) confirm(w http.ResponseWriter, r *http.Request) {
	u, _ := httpserver.UserFromContext(r.Context())
	if u.Role != "OWNER" && u.Role != "OPERATOR" {
		httpserver.WriteErrorFromContext(w, r, 403, httpserver.CodeForbidden, "FORBIDDEN")
		return
	}
	id, e := strconv.ParseInt(r.PathValue("id"), 10, 64)
	var in ConfirmWrite
	if e != nil || json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.OutcomeType) == "" || in.ChargeAmount < 0 || in.TeacherPayAmount < 0 || strings.TrimSpace(in.LessonDeducted) == "" {
		httpserver.WriteErrorFromContext(w, r, 422, httpserver.CodeInvalidState, "INVALID_STATE")
		return
	}
	if e = h.confirmTx(r.Context(), id, u, in, httpserver.RequestIDFromContext(r.Context())); e != nil {
		code, status := httpserver.CodeDatabase, 500
		if errors.Is(e, ErrInvalid) {
			code, status = httpserver.CodeInvalidState, 422
		}
		httpserver.WriteErrorFromContext(w, r, status, code, "INVALID_STATE")
		return
	}
	httpserver.WriteSuccess(w, 200, map[string]int64{"lessonId": id})
}
func (h *Handler) confirmTx(ctx context.Context, id int64, u httpserver.AuthUser, in ConfirmWrite, rid string) (err error) {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.ErrDatabase
	}
	ok := false
	defer func() {
		if !ok {
			tx.Rollback()
		}
	}()
	var student, enroll, teacher int64
	var duration int
	var charge int64
	if err = tx.QueryRowContext(ctx, `SELECT student_id,enrollment_id,teacher_id,duration_min FROM lesson WHERE id=? AND status='SCHEDULED'`, id).Scan(&student, &enroll, &teacher, &duration); err != nil {
		return ErrInvalid
	}
	if in.ActualDurationMin > duration*2 {
		return ErrInvalid
	}
	var sd, sc, st sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT suggested_lesson_deducted,suggested_charge_ratio,suggested_teacher_pay_ratio FROM attendance_outcome_type WHERE code=? AND enabled=1`, in.OutcomeType).Scan(&sd, &sc, &st); err != nil {
		return ErrInvalid
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO attendance(lesson_id,outcome_type,suggested_lesson_deducted,suggested_charge_ratio,suggested_teacher_pay_ratio,actual_duration_min,lesson_deducted,charge_amount,teacher_pay_amount,note,confirmed_by) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, id, in.OutcomeType, sd, sc, st, in.ActualDurationMin, in.LessonDeducted, in.ChargeAmount, in.TeacherPayAmount, in.Note, u.UserID); err != nil {
		return ErrInvalid
	}
	var bal int64
	var lessons string
	if err = tx.QueryRowContext(ctx, `SELECT balance_amount,lesson_balance FROM student_course_enrollment WHERE id=?`, enroll).Scan(&bal, &lessons); err != nil {
		return repository.ErrDatabase
	}
	if bal < in.ChargeAmount {
		return ErrInvalid
	}
	if _, err = tx.ExecContext(ctx, `UPDATE student_course_enrollment SET balance_amount=balance_amount-?,lesson_balance=lesson_balance-CAST(? AS REAL),updated_at=CURRENT_TIMESTAMP WHERE id=?`, in.ChargeAmount, in.LessonDeducted, enroll); err != nil {
		return repository.ErrDatabase
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO student_account_ledger(student_id,enrollment_id,biz_type,amount_delta,lesson_delta,balance_after,lesson_balance_after,operator_id,note) VALUES(?,?,'LESSON_CONFIRM',?,CAST(? AS REAL),?,CAST(? AS REAL),?,?)`, student, enroll, -in.ChargeAmount, "-"+in.LessonDeducted, bal-in.ChargeAmount, lessons, u.UserID, in.Note); err != nil {
		return repository.ErrDatabase
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO teacher_account_ledger(teacher_id,lesson_id,amount_delta,balance_after,operator_id,note) VALUES(?,?,?,COALESCE((SELECT balance_after FROM teacher_account_ledger WHERE teacher_id=? ORDER BY id DESC LIMIT 1),0)+?,?,?)`, teacher, id, in.TeacherPayAmount, teacher, in.TeacherPayAmount, u.UserID, in.Note); err != nil {
		return repository.ErrDatabase
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO lesson_finance(lesson_id,student_id,teacher_id,enrollment_id,charge_amount,teacher_pay_amount,gross_profit_amount) VALUES(?,?,?,?,?,?,?)`, id, student, teacher, enroll, in.ChargeAmount, in.TeacherPayAmount, in.ChargeAmount-in.TeacherPayAmount); err != nil {
		return repository.ErrDatabase
	}
	if _, err = tx.ExecContext(ctx, `UPDATE lesson SET status='COMPLETED',updated_at=CURRENT_TIMESTAMP WHERE id=? AND status='SCHEDULED'`, id); err != nil {
		return repository.ErrDatabase
	}
	name, _ := repository.ActorName(tx, ctx, u.UserID)
	if err = repository.InsertAuditLog(tx, ctx, u.UserID, name, "LESSON_CONFIRM", "lesson", id, map[string]any{"outcomeType": in.OutcomeType}, rid); err != nil {
		return repository.ErrDatabase
	}
	if err = tx.Commit(); err != nil {
		return repository.ErrDatabase
	}
	ok = true
	return nil
}
