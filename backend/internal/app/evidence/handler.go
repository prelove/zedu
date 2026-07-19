package evidence

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
)

type Handler struct {
	svc    *Service
	logger *slog.Logger
}

func NewHandler(db any, logger *slog.Logger, cfg Config) *Handler {
	storage := cfg.Storage
	if storage == nil {
		storage = NewStorage(cfg.DataRoot)
	}
	if err := storage.CleanupTemp(); err != nil && logger != nil {
		logger.Error("cleanup evidence temp files", slog.Any("error", err))
	}
	return &Handler{svc: NewService(repository.AsDB(db), storage), logger: logger}
}

func MountRoutes(mux *http.ServeMux, h *Handler, authDB *sql.DB, jwtSecret string) {
	authMW := httpserver.AuthMiddleware(jwtSecret, authDB)
	withFinanceRole := func(next http.HandlerFunc) http.Handler {
		return authMW(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := httpserver.UserFromContext(r.Context())
			if !ok {
				httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
				return
			}
			if !isFinanceRole(user.Role) {
				httpserver.WriteErrorFromContext(w, r, http.StatusForbidden, httpserver.CodeForbidden, "FORBIDDEN")
				return
			}
			next.ServeHTTP(w, r)
		}))
	}

	mux.Handle("POST /finance/payments/{id}/attachments", withFinanceRole(h.uploadAttachment))
	mux.Handle("GET /finance/payments/{id}/attachments", withFinanceRole(h.listAttachments))
	mux.Handle("GET /finance/payments/{paymentId}/attachments/{attachmentId}/content", withFinanceRole(h.downloadAttachment))
}

func (h *Handler) uploadAttachment(w http.ResponseWriter, r *http.Request) {
	paymentID, ok := parseID(r.PathValue("id"))
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}

	reader, err := r.MultipartReader()
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_STATE")
		return
	}

	var (
		partFound bool
		fileName  string
	)
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_STATE")
			return
		}
		if part.FileName() == "" {
			_ = part.Close()
			continue
		}
		if partFound || part.FormName() != "file" {
			_ = part.Close()
			httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_STATE")
			return
		}
		partFound = true
		fileName = part.FileName()
		user, _ := httpserver.UserFromContext(r.Context())
		attachment, err := h.svc.UploadAttachment(r.Context(), user, paymentID, fileName, part, httpserver.RequestIDFromContext(r.Context()))
		_ = part.Close()
		if err != nil {
			h.writeError(w, r, err)
			return
		}
		httpserver.WriteSuccess(w, http.StatusCreated, attachment)
		return
	}

	httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_STATE")
}

func (h *Handler) listAttachments(w http.ResponseWriter, r *http.Request) {
	paymentID, ok := parseID(r.PathValue("id"))
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	user, _ := httpserver.UserFromContext(r.Context())
	pq := httpserver.ParsePage(r)
	value, err := h.svc.ListAttachments(r.Context(), user, paymentID, pq.Page, pq.PageSize)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpserver.WriteSuccess(w, http.StatusOK, value)
}

func (h *Handler) downloadAttachment(w http.ResponseWriter, r *http.Request) {
	paymentID, ok := parseID(r.PathValue("paymentId"))
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	attachmentID, ok := parseID(r.PathValue("attachmentId"))
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}

	user, _ := httpserver.UserFromContext(r.Context())
	content, err := h.svc.OpenAttachment(r.Context(), user, paymentID, attachmentID, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	defer content.File.Close()

	info, err := content.File.Stat()
	if err != nil {
		h.writeError(w, r, repository.ErrDatabase)
		return
	}
	w.Header().Set("Content-Type", content.Attachment.FileType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", content.Attachment.FileName))
	http.ServeContent(w, r, content.Attachment.FileName, info.ModTime(), content.File)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, code, message := http.StatusInternalServerError, httpserver.CodeDatabase, "DATABASE_ERROR"
	switch {
	case errors.Is(err, ErrForbidden):
		status, code, message = http.StatusForbidden, httpserver.CodeForbidden, "FORBIDDEN"
	case errors.Is(err, ErrNotFound):
		status, code, message = http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND"
	case errors.Is(err, ErrInvalidState):
		status, code, message = http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_STATE"
	}
	if status >= 500 && h.logger != nil {
		h.logger.Error("evidence service error", slog.String("request_id", httpserver.RequestIDFromContext(r.Context())), slog.Any("error", err))
	}
	httpserver.WriteErrorFromContext(w, r, status, code, message)
}

func parseID(raw string) (int64, bool) {
	value, err := strconv.ParseInt(raw, 10, 64)
	return value, err == nil && value > 0
}
