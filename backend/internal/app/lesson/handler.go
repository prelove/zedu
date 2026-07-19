package lesson

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(db any, logger *slog.Logger) *Handler {
	return &Handler{svc: NewService(repository.AsDB(db)), logger: logger}
}

func MountRoutes(mux *http.ServeMux, h *Handler, db *sql.DB, secret string) {
	authenticated := httpserver.AuthMiddleware(secret, db)
	mux.Handle("GET /lessons", authenticated(http.HandlerFunc(h.list)))
	mux.Handle("POST /lessons", authenticated(http.HandlerFunc(h.create)))
	mux.Handle("GET /lessons/{id}", authenticated(http.HandlerFunc(h.get)))
	mux.Handle("PATCH /lessons/{id}", authenticated(http.HandlerFunc(h.update)))
	mux.Handle("POST /lessons/{id}/cancel", authenticated(http.HandlerFunc(h.cancel)))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input Write
	if !decodeJSON(w, r, &input) {
		return
	}
	user, _ := httpserver.UserFromContext(r.Context())
	id, err := h.svc.Create(r.Context(), user, input, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	lesson, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, lesson)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := lessonID(w, r)
	if !ok {
		return
	}
	var input ScheduleUpdate
	if !decodeJSON(w, r, &input) {
		return
	}
	user, _ := httpserver.UserFromContext(r.Context())
	if err := h.svc.Update(r.Context(), user, id, input, httpserver.RequestIDFromContext(r.Context())); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	lesson, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, lesson)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request) {
	id, ok := lessonID(w, r)
	if !ok {
		return
	}
	var input CancelWrite
	if !decodeJSON(w, r, &input) {
		return
	}
	user, _ := httpserver.UserFromContext(r.Context())
	if err := h.svc.Cancel(r.Context(), user, id, input.Reason, httpserver.RequestIDFromContext(r.Context())); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	lesson, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, lesson)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := lessonID(w, r)
	if !ok {
		return
	}
	lesson, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, lesson)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_FILTER")
		return
	}
	result, err := h.svc.List(r.Context(), filter)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, result)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusInternalServerError, httpserver.CodeInternal, "INTERNAL_ERROR"
	switch {
	case errors.Is(err, repository.ErrDatabase):
		status, code, message = http.StatusInternalServerError, httpserver.CodeDatabase, "DATABASE_ERROR"
	case errors.Is(err, ErrForbidden):
		status, code, message = http.StatusForbidden, httpserver.CodeForbidden, "FORBIDDEN"
	case errors.Is(err, ErrNotFound):
		status, code, message = http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND"
	case errors.Is(err, ErrInvalidState):
		status, code, message = http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_STATE"
	}
	h.logger.Error("lesson request failed", slog.String("request_id", httpserver.RequestIDFromContext(r.Context())), slog.Int("status", status))
	httpserver.WriteErrorFromContext(w, r, status, code, message)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return false
	}
	return true
}

func lessonID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return 0, false
	}
	return id, true
}

func parseListFilter(r *http.Request) (ListFilter, error) {
	query := r.URL.Query()
	parseID := func(key string) (int64, error) {
		if query.Get(key) == "" {
			return 0, nil
		}
		return strconv.ParseInt(query.Get(key), 10, 64)
	}
	studentID, err := parseID("studentId")
	if err != nil || studentID < 0 {
		return ListFilter{}, errors.New("invalid studentId")
	}
	teacherID, err := parseID("teacherId")
	if err != nil || teacherID < 0 {
		return ListFilter{}, errors.New("invalid teacherId")
	}
	page, err := parseQueryInt(query.Get("page"))
	if err != nil {
		return ListFilter{}, err
	}
	pageSize, err := parseQueryInt(query.Get("pageSize"))
	if err != nil {
		return ListFilter{}, err
	}
	return ListFilter{StudentID: studentID, TeacherID: teacherID, Status: strings.TrimSpace(query.Get("status")), From: query.Get("from"), To: query.Get("to"), Page: page, PageSize: pageSize}, nil
}

func parseQueryInt(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.Atoi(value)
}
