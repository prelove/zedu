// Package payable provides read-only teacher payable queries derived solely
// from the immutable teacher_account_ledger. It exposes no payout, settlement,
// adjustment, export, or write actions.
package payable

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
)

// TeacherPayableSummary is the aggregated unpaid amount per teacher, derived
// from the sum of amount_delta in teacher_account_ledger.
type TeacherPayableSummary struct {
	TeacherID    int64  `json:"teacherId"`
	TeacherName  string `json:"teacherName"`
	UnpaidAmount int64  `json:"unpaidAmount"`
	LessonCount  int    `json:"lessonCount"`
}

// TeacherPayableEntry is a single immutable ledger row describing one lesson's
// payable fact.
type TeacherPayableEntry struct {
	ID           int64  `json:"id"`
	LessonID     int64  `json:"lessonId"`
	LessonNo     string `json:"lessonNo"`
	AmountDelta  int64  `json:"amountDelta"`
	BalanceAfter int64  `json:"balanceAfter"`
	Note         string `json:"note,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

// MountRoutes mounts the read-only payable routes onto mux. Both OWNER and
// OPERATOR may read; no write routes exist in this capability.
func MountRoutes(mux *http.ServeMux, db *sql.DB, secret string) {
	auth := httpserver.AuthMiddleware(secret, db)
	mux.Handle("GET /teachers/payable", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		summary(w, r, db)
	})))
	mux.Handle("GET /teachers/{id}/payable", auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		detail(w, r, db)
	})))
}

func summary(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	page, pageSize := parsePage(r)
	offset := (page - 1) * pageSize
	rows, err := db.QueryContext(r.Context(), `
		SELECT t.id, t.name, COALESCE(SUM(l.amount_delta), 0), COUNT(l.id)
		FROM teacher t
		LEFT JOIN teacher_account_ledger l ON l.teacher_id = t.id
		WHERE t.deleted_at IS NULL
		GROUP BY t.id, t.name
		HAVING COALESCE(SUM(l.amount_delta), 0) > 0
		ORDER BY t.id
		LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
		return
	}
	defer rows.Close()
	items := make([]TeacherPayableSummary, 0)
	for rows.Next() {
		var s TeacherPayableSummary
		if err := rows.Scan(&s.TeacherID, &s.TeacherName, &s.UnpaidAmount, &s.LessonCount); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
		return
	}
	var total int
	if err := db.QueryRowContext(r.Context(), `
		SELECT COUNT(*) FROM (
			SELECT t.id FROM teacher t
			LEFT JOIN teacher_account_ledger l ON l.teacher_id = t.id
			WHERE t.deleted_at IS NULL
			GROUP BY t.id
			HAVING COALESCE(SUM(l.amount_delta), 0) > 0
		)`).Scan(&total); err != nil {
		httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
		return
	}
	httpserver.WriteSuccess(w, 200, map[string]any{"items": items, "page": page, "pageSize": pageSize, "total": total})
}

func detail(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpserver.WriteErrorFromContext(w, r, 404, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	page, pageSize := parsePage(r)
	offset := (page - 1) * pageSize
	rows, err := db.QueryContext(r.Context(), `
		SELECT l.id, l.lesson_id, COALESCE(s.lesson_no, ''), l.amount_delta, l.balance_after, COALESCE(l.note, ''), l.created_at
		FROM teacher_account_ledger l
		LEFT JOIN lesson s ON s.id = l.lesson_id
		WHERE l.teacher_id = ?
		ORDER BY l.id DESC
		LIMIT ? OFFSET ?`, id, pageSize, offset)
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
		return
	}
	defer rows.Close()
	items := make([]TeacherPayableEntry, 0)
	for rows.Next() {
		var e TeacherPayableEntry
		if err := rows.Scan(&e.ID, &e.LessonID, &e.LessonNo, &e.AmountDelta, &e.BalanceAfter, &e.Note, &e.CreatedAt); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
		return
	}
	var total int
	if err := db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM teacher_account_ledger WHERE teacher_id=?`, id).Scan(&total); err != nil {
		httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
		return
	}
	httpserver.WriteSuccess(w, 200, map[string]any{"items": items, "page": page, "pageSize": pageSize, "total": total})
}

func parsePage(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
