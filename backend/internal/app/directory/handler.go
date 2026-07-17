package directory

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

// Handler implements the people-directory HTTP handlers. It only decodes
// requests, reads identity context and encodes responses; all business logic
// lives in Service.
type Handler struct {
	svc    *Service
	logger *slog.Logger
}

// NewHandler creates a people-directory handler. db may be *sql.DB (wrapped) or
// a repository.DB for fault-injection tests.
func NewHandler(db any, logger *slog.Logger) *Handler {
	return &Handler{svc: NewService(repository.AsDB(db)), logger: logger}
}

// MountRoutes mounts the people-directory routes onto mux, each guarded by
// AuthMiddleware. authDB is the real *sql.DB used by the middleware.
func MountRoutes(mux *http.ServeMux, h *Handler, authDB *sql.DB, jwtSecret string) {
	authMW := httpserver.AuthMiddleware(jwtSecret, authDB)

	mux.Handle("GET /students", authMW(http.HandlerFunc(h.listStudents)))
	mux.Handle("POST /students", authMW(http.HandlerFunc(h.createStudent)))
	mux.Handle("GET /students/{id}", authMW(http.HandlerFunc(h.getStudent)))
	mux.Handle("PATCH /students/{id}", authMW(http.HandlerFunc(h.updateStudent)))

	mux.Handle("GET /students/{id}/parents", authMW(http.HandlerFunc(h.listParents)))
	mux.Handle("POST /students/{id}/parents", authMW(http.HandlerFunc(h.createParent)))
	mux.Handle("PATCH /students/{id}/parents/{parentId}", authMW(http.HandlerFunc(h.updateParent)))

	mux.Handle("GET /teachers", authMW(http.HandlerFunc(h.listTeachers)))
	mux.Handle("POST /teachers", authMW(http.HandlerFunc(h.createTeacher)))
	mux.Handle("GET /teachers/{id}", authMW(http.HandlerFunc(h.getTeacher)))
	mux.Handle("PATCH /teachers/{id}", authMW(http.HandlerFunc(h.updateTeacher)))

	mux.Handle("GET /teachers/{id}/capabilities", authMW(http.HandlerFunc(h.listCapabilities)))
	mux.Handle("POST /teachers/{id}/capabilities", authMW(http.HandlerFunc(h.createCapability)))
	mux.Handle("PATCH /teachers/{id}/capabilities/{capId}", authMW(http.HandlerFunc(h.updateCapability)))

	mux.Handle("GET /teachers/{id}/availability", authMW(http.HandlerFunc(h.listAvailability)))
	mux.Handle("POST /teachers/{id}/availability", authMW(http.HandlerFunc(h.createAvailability)))
	mux.Handle("PATCH /teachers/{id}/availability/{availId}", authMW(http.HandlerFunc(h.updateAvailability)))
}

// ---------- Student ----------

func (h *Handler) listStudents(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	pq := httpserver.ParsePage(r)
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	data, err := h.svc.ListStudents(r.Context(), user, status, search, pq.Page, pq.PageSize)
	h.respond(w, r, data, err)
}

func (h *Handler) createStudent(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	var w2 StudentWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateStudent(r.Context(), user, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) getStudent(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	s, err := h.svc.GetStudent(r.Context(), user, id)
	h.respond(w, r, s, err)
}

func (h *Handler) updateStudent(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 StudentWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	updated, err := h.svc.UpdateStudent(r.Context(), user, id, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, updated)
}

// ---------- Parent ----------

func (h *Handler) listParents(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	studentID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	pq := httpserver.ParsePage(r)
	data, err := h.svc.ListParents(r.Context(), user, studentID, pq.Page, pq.PageSize)
	h.respond(w, r, data, err)
}

func (h *Handler) createParent(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	studentID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 ParentWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateParent(r.Context(), user, studentID, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) updateParent(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	studentID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	parentID, ok := pathID(r, "parentId")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 ParentWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	updated, err := h.svc.UpdateParent(r.Context(), user, studentID, parentID, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, updated)
}

// ---------- Teacher ----------

func (h *Handler) listTeachers(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	pq := httpserver.ParsePage(r)
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	data, err := h.svc.ListTeachers(r.Context(), user, status, search, pq.Page, pq.PageSize)
	h.respond(w, r, data, err)
}

func (h *Handler) createTeacher(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	var w2 TeacherWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateTeacher(r.Context(), user, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) getTeacher(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	t, err := h.svc.GetTeacher(r.Context(), user, id)
	h.respond(w, r, t, err)
}

func (h *Handler) updateTeacher(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 TeacherWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	updated, err := h.svc.UpdateTeacher(r.Context(), user, id, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, updated)
}

// ---------- Capability ----------

func (h *Handler) listCapabilities(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	teacherID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	pq := httpserver.ParsePage(r)
	data, err := h.svc.ListCapabilities(r.Context(), user, teacherID, pq.Page, pq.PageSize)
	h.respond(w, r, data, err)
}

func (h *Handler) createCapability(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	teacherID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 CapabilityWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateCapability(r.Context(), user, teacherID, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) updateCapability(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	teacherID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	capID, ok := pathID(r, "capId")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 CapabilityWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	updated, err := h.svc.UpdateCapability(r.Context(), user, teacherID, capID, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, updated)
}

// ---------- Availability ----------

func (h *Handler) listAvailability(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	teacherID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	pq := httpserver.ParsePage(r)
	data, err := h.svc.ListAvailability(r.Context(), user, teacherID, pq.Page, pq.PageSize)
	h.respond(w, r, data, err)
}

func (h *Handler) createAvailability(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	teacherID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 AvailabilityWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateAvailability(r.Context(), user, teacherID, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) updateAvailability(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	teacherID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	availID, ok := pathID(r, "availId")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 AvailabilityWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	updated, err := h.svc.UpdateAvailability(r.Context(), user, teacherID, availID, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, updated)
}

// ---------- helpers ----------

// pathID extracts an integer path parameter. Go 1.22 ServeMux exposes path
// values via r.PathValue; missing/invalid values return ok=false.
func pathID(r *http.Request, key string) (int64, bool) {
	v := r.PathValue(key)
	if v == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(v, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

// respond maps a service error to the frozen HTTP status/code and writes the
// success payload otherwise.
func (h *Handler) respond(w http.ResponseWriter, r *http.Request, data any, err error) {
	if err == nil {
		httpserver.WriteSuccess(w, http.StatusOK, data)
		return
	}
	status, code, msg := mapError(err)
	rid := httpserver.RequestIDFromContext(r.Context())
	if status >= 500 {
		h.logger.Error("directory service error", slog.String("request_id", rid), slog.Any("error", err))
	}
	httpserver.WriteError(w, status, code, msg, rid)
}

// mapError translates a service/repository sentinel error into the frozen
// HTTP status, business code and stable error key. Database errors never leak
// the underlying SQLite text.
func mapError(err error) (int, httpserver.ErrorCode, string) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND"
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, httpserver.CodeConflict, "CONFLICT"
	case errors.Is(err, ErrInvalidState):
		return http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_STATE"
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, httpserver.CodeForbidden, "FORBIDDEN"
	default:
		return http.StatusInternalServerError, httpserver.CodeDatabase, "DATABASE_ERROR"
	}
}
