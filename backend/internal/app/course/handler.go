package course

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

// Handler implements the course-dictionary and enrollment/assignment HTTP
// handlers. It only decodes requests, reads identity context and encodes
// responses; all business logic lives in Service.
type Handler struct {
	svc    *Service
	logger *slog.Logger
}

// NewHandler creates a course handler. db may be *sql.DB (wrapped) or a
// repository.DB for fault-injection tests.
func NewHandler(db any, logger *slog.Logger) *Handler {
	return &Handler{svc: NewService(repository.AsDB(db)), logger: logger}
}

// MountRoutes mounts the course-dictionary routes (stage B). Enrollment and
// assignment routes are mounted by MountEnrollmentRoutes in stage C.
func MountRoutes(mux *http.ServeMux, h *Handler, authDB *sql.DB, jwtSecret string) {
	authMW := httpserver.AuthMiddleware(jwtSecret, authDB)

	mux.Handle("GET /course-domains", authMW(http.HandlerFunc(h.listDomains)))
	mux.Handle("POST /course-domains", authMW(http.HandlerFunc(h.createDomain)))
	mux.Handle("PATCH /course-domains/{id}", authMW(http.HandlerFunc(h.updateDomain)))

	mux.Handle("GET /tracks", authMW(http.HandlerFunc(h.listTracks)))
	mux.Handle("POST /tracks", authMW(http.HandlerFunc(h.createTrack)))
	mux.Handle("PATCH /tracks/{id}", authMW(http.HandlerFunc(h.updateTrack)))

	mux.Handle("GET /levels", authMW(http.HandlerFunc(h.listLevels)))
	mux.Handle("POST /levels", authMW(http.HandlerFunc(h.createLevel)))
	mux.Handle("PATCH /levels/{id}", authMW(http.HandlerFunc(h.updateLevel)))

	mux.Handle("GET /capability-tags", authMW(http.HandlerFunc(h.listTags)))
	mux.Handle("POST /capability-tags", authMW(http.HandlerFunc(h.createTag)))
	mux.Handle("PATCH /capability-tags/{id}", authMW(http.HandlerFunc(h.updateTag)))

	MountEnrollmentRoutes(mux, h, authDB, jwtSecret)
}

// MountEnrollmentRoutes mounts the enrollment and assignment routes (stage C).
func MountEnrollmentRoutes(mux *http.ServeMux, h *Handler, authDB *sql.DB, jwtSecret string) {
	authMW := httpserver.AuthMiddleware(jwtSecret, authDB)

	mux.Handle("GET /students/{id}/enrollments", authMW(http.HandlerFunc(h.listEnrollments)))
	mux.Handle("POST /students/{id}/enrollments", authMW(http.HandlerFunc(h.createEnrollment)))
	mux.Handle("GET /enrollments/{id}", authMW(http.HandlerFunc(h.getEnrollment)))
	mux.Handle("PATCH /enrollments/{id}", authMW(http.HandlerFunc(h.updateEnrollment)))

	mux.Handle("GET /enrollments/{id}/assignments", authMW(http.HandlerFunc(h.listAssignments)))
	mux.Handle("POST /enrollments/{id}/assignments", authMW(http.HandlerFunc(h.createAssignment)))
	mux.Handle("POST /assignments/{id}/end", authMW(http.HandlerFunc(h.endAssignment)))
}

// ---------- Domain ----------

func (h *Handler) listDomains(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	pq := httpserver.ParsePage(r)
	data, err := h.svc.ListDomains(r.Context(), user, r.URL.Query().Get("search"), pq.Page, pq.PageSize)
	h.respond(w, r, data, err)
}

func (h *Handler) createDomain(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	var w2 DomainWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateDomain(r.Context(), user, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) updateDomain(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 DomainWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	updated, err := h.svc.UpdateDomain(r.Context(), user, id, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, updated)
}

// ---------- Track ----------

func (h *Handler) listTracks(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	pq := httpserver.ParsePage(r)
	domainID, _ := strconv.ParseInt(r.URL.Query().Get("domainId"), 10, 64)
	data, err := h.svc.ListTracks(r.Context(), user, domainID, pq.Page, pq.PageSize)
	h.respond(w, r, data, err)
}

func (h *Handler) createTrack(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	var w2 TrackWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateTrack(r.Context(), user, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) updateTrack(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 TrackWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	updated, err := h.svc.UpdateTrack(r.Context(), user, id, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, updated)
}

// ---------- Level ----------

func (h *Handler) listLevels(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	pq := httpserver.ParsePage(r)
	trackID, _ := strconv.ParseInt(r.URL.Query().Get("trackId"), 10, 64)
	data, err := h.svc.ListLevels(r.Context(), user, trackID, pq.Page, pq.PageSize)
	h.respond(w, r, data, err)
}

func (h *Handler) createLevel(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	var w2 LevelWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateLevel(r.Context(), user, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) updateLevel(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 LevelWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	updated, err := h.svc.UpdateLevel(r.Context(), user, id, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, updated)
}

// ---------- Capability Tag ----------

func (h *Handler) listTags(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	pq := httpserver.ParsePage(r)
	domainID, _ := strconv.ParseInt(r.URL.Query().Get("domainId"), 10, 64)
	data, err := h.svc.ListTags(r.Context(), user, domainID, pq.Page, pq.PageSize)
	h.respond(w, r, data, err)
}

func (h *Handler) createTag(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	var w2 TagWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateTag(r.Context(), user, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) updateTag(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 TagWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	updated, err := h.svc.UpdateTag(r.Context(), user, id, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, updated)
}

// ---------- Enrollment ----------

func (h *Handler) listEnrollments(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	studentID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	items, err := h.svc.ListEnrollments(r.Context(), user, studentID)
	h.respond(w, r, items, err)
}

func (h *Handler) createEnrollment(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	studentID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 EnrollmentWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateEnrollment(r.Context(), user, studentID, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) getEnrollment(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	item, err := h.svc.GetEnrollment(r.Context(), user, id)
	h.respond(w, r, item, err)
}

func (h *Handler) updateEnrollment(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 EnrollmentWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	updated, err := h.svc.UpdateEnrollment(r.Context(), user, id, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, updated)
}

// ---------- Assignment ----------

func (h *Handler) listAssignments(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	enrollmentID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	items, err := h.svc.ListAssignments(r.Context(), user, enrollmentID)
	h.respond(w, r, items, err)
}

func (h *Handler) createAssignment(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	enrollmentID, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 AssignmentWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	created, err := h.svc.CreateAssignment(r.Context(), user, enrollmentID, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusCreated, created)
}

func (h *Handler) endAssignment(w http.ResponseWriter, r *http.Request) {
	user, _ := httpserver.UserFromContext(r.Context())
	id, ok := pathID(r, "id")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	var w2 EndAssignmentWrite
	if err := json.NewDecoder(r.Body).Decode(&w2); err != nil {
		// Body is optional; empty body is fine.
		w2 = EndAssignmentWrite{}
	}
	ended, err := h.svc.EndAssignment(r.Context(), user, id, w2, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, ended)
}

// ---------- helpers ----------

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

func (h *Handler) respond(w http.ResponseWriter, r *http.Request, data any, err error) {
	if err == nil {
		httpserver.WriteSuccess(w, http.StatusOK, data)
		return
	}
	status, code, msg := mapError(err)
	rid := httpserver.RequestIDFromContext(r.Context())
	if status >= 500 {
		h.logger.Error("course service error", slog.String("request_id", rid), slog.Any("error", err))
	}
	httpserver.WriteError(w, status, code, msg, rid)
}

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
