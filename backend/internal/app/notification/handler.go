package notification

import (
	"database/sql"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
	"net/http"
	"strconv"
)

type Handler struct {
	db     repository.DB
	sender Sender
}

func NewHandler(db any, sender Sender) *Handler {
	return &Handler{db: repository.AsDB(db), sender: sender}
}
func MountRoutes(mux *http.ServeMux, h *Handler, db *sql.DB, secret string) {
	a := httpserver.AuthMiddleware(secret, db)
	mux.Handle("GET /notifications/outbox", a(http.HandlerFunc(h.list)))
	mux.Handle("POST /notifications/outbox/process", a(http.HandlerFunc(h.process)))
	mux.Handle("POST /notifications/outbox/{id}/retry", a(http.HandlerFunc(h.retry)))
}
func (h *Handler) authorized(w http.ResponseWriter, r *http.Request) bool {
	u, _ := httpserver.UserFromContext(r.Context())
	if u.Role == "OWNER" || u.Role == "OPERATOR" {
		return true
	}
	httpserver.WriteErrorFromContext(w, r, http.StatusForbidden, httpserver.CodeForbidden, "FORBIDDEN")
	return false
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	rows, err := h.db.QueryContext(r.Context(), `SELECT id,lesson_id,event_type,recipient_email,status,attempts,coalesce(last_error,'') FROM notification_outbox ORDER BY id DESC LIMIT 100`)
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
		return
	}
	defer rows.Close()
	items := []Outbox{}
	for rows.Next() {
		var x Outbox
		if err := rows.Scan(&x.ID, &x.LessonID, &x.EventType, &x.RecipientEmail, &x.Status, &x.Attempts, &x.LastError); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		items = append(items, x)
	}
	httpserver.WriteSuccess(w, 200, map[string]any{"items": items, "page": 1, "pageSize": 100, "total": len(items)})
}
func (h *Handler) process(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	if h.sender == nil {
		httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeInternal, "NOTIFICATION_NOT_CONFIGURED")
		return
	}
	if err := ClaimAndSend(r.Context(), h.db, h.sender); err != nil {
		httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
		return
	}
	httpserver.WriteSuccess(w, 200, map[string]string{"status": "processed"})
}
func (h *Handler) retry(w http.ResponseWriter, r *http.Request) {
	if !h.authorized(w, r) {
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		httpserver.WriteErrorFromContext(w, r, 404, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	res, err := h.db.ExecContext(r.Context(), `UPDATE notification_outbox SET status='PENDING',available_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP WHERE id=? AND status='FAILED' AND attempts<3`, id)
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
		return
	}
	n, _ := res.RowsAffected()
	if n != 1 {
		httpserver.WriteErrorFromContext(w, r, 422, httpserver.CodeInvalidState, "INVALID_STATE")
		return
	}
	httpserver.WriteSuccess(w, 200, map[string]int64{"id": id})
}
