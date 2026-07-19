package finance

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

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
func MountRoutes(mux *http.ServeMux, h *Handler, authDB *sql.DB, jwtSecret string) {
	authMW := httpserver.AuthMiddleware(jwtSecret, authDB)
	mux.Handle("GET /system/base-currency", authMW(http.HandlerFunc(h.getBaseCurrency)))
	mux.Handle("PUT /system/base-currency", authMW(httpserver.RequireRole("OWNER", http.HandlerFunc(h.putBaseCurrency))))
	mux.Handle("GET /system/payment-methods", authMW(http.HandlerFunc(h.listPaymentMethods)))
	mux.Handle("POST /system/payment-methods", authMW(httpserver.RequireRole("OWNER", http.HandlerFunc(h.createPaymentMethod))))
	mux.Handle("PATCH /system/payment-methods/{code}", authMW(httpserver.RequireRole("OWNER", http.HandlerFunc(h.updatePaymentMethod))))
	mux.Handle("POST /finance/payments", authMW(http.HandlerFunc(h.createPayment)))
	mux.Handle("GET /finance/payments", authMW(http.HandlerFunc(h.listPayments)))
	mux.Handle("GET /finance/payments/{id}", authMW(http.HandlerFunc(h.getPayment)))
	mux.Handle("GET /finance/ledger/student/{studentId}", authMW(http.HandlerFunc(h.listStudentLedger)))
	mux.Handle("POST /finance/payments/{id}/void", authMW(http.HandlerFunc(h.voidPayment)))
}
func (h *Handler) listPayments(w http.ResponseWriter, r *http.Request) {
	pq := httpserver.ParsePage(r)
	filter := PaymentFilter{PaymentNo: r.URL.Query().Get("paymentNo"), Status: r.URL.Query().Get("status")}
	if _, err := fmt.Sscan(r.URL.Query().Get("studentId"), &filter.StudentID); err != nil && r.URL.Query().Get("studentId") != "" {
		h.respond(w, r, nil, ErrInvalidState)
		return
	}
	if _, err := fmt.Sscan(r.URL.Query().Get("enrollmentId"), &filter.EnrollmentID); err != nil && r.URL.Query().Get("enrollmentId") != "" {
		h.respond(w, r, nil, ErrInvalidState)
		return
	}
	value, err := h.svc.ListPayments(r.Context(), filter, pq.Page, pq.PageSize)
	h.respond(w, r, value, err)
}
func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request) {
	var id int64
	if _, err := fmt.Sscan(r.PathValue("id"), &id); err != nil || id <= 0 {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	value, err := h.svc.GetPayment(r.Context(), id)
	h.respond(w, r, value, err)
}
func (h *Handler) listStudentLedger(w http.ResponseWriter, r *http.Request) {
	var id int64
	if _, err := fmt.Sscan(r.PathValue("studentId"), &id); err != nil || id <= 0 {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	pq := httpserver.ParsePage(r)
	value, err := h.svc.ListStudentLedger(r.Context(), id, pq.Page, pq.PageSize)
	h.respond(w, r, value, err)
}

func (h *Handler) voidPayment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	var id int64
	if _, err := fmt.Sscan(r.PathValue("id"), &id); err != nil || id <= 0 {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}
	u, _ := httpserver.UserFromContext(r.Context())
	value, err := h.svc.VoidPayment(r.Context(), u, id, input.Reason, httpserver.RequestIDFromContext(r.Context()))
	h.respond(w, r, value, err)
}

func (h *Handler) createPayment(w http.ResponseWriter, r *http.Request) {
	var input PaymentWrite
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	u, _ := httpserver.UserFromContext(r.Context())
	value, reused, err := h.svc.CreatePayment(r.Context(), u, input, httpserver.RequestIDFromContext(r.Context()))
	if err != nil {
		h.respond(w, r, nil, err)
		return
	}
	status := http.StatusCreated
	if reused {
		status = http.StatusOK
	}
	httpserver.WriteSuccess(w, status, value)
}

func (h *Handler) listPaymentMethods(w http.ResponseWriter, r *http.Request) {
	u, _ := httpserver.UserFromContext(r.Context())
	items, err := h.svc.ListPaymentMethods(r.Context(), u)
	h.respond(w, r, items, err)
}
func (h *Handler) createPaymentMethod(w http.ResponseWriter, r *http.Request) {
	var item PaymentMethod
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	u, _ := httpserver.UserFromContext(r.Context())
	value, err := h.svc.CreatePaymentMethod(r.Context(), u, item, httpserver.RequestIDFromContext(r.Context()))
	h.respond(w, r, value, err)
}
func (h *Handler) updatePaymentMethod(w http.ResponseWriter, r *http.Request) {
	var item PaymentMethod
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	u, _ := httpserver.UserFromContext(r.Context())
	value, err := h.svc.UpdatePaymentMethod(r.Context(), u, r.PathValue("code"), item, httpserver.RequestIDFromContext(r.Context()))
	h.respond(w, r, value, err)
}

func (h *Handler) getBaseCurrency(w http.ResponseWriter, r *http.Request) {
	u, _ := httpserver.UserFromContext(r.Context())
	value, err := h.svc.GetBaseCurrency(r.Context(), u)
	h.respond(w, r, value, err)
}
func (h *Handler) putBaseCurrency(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Currency string `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_BODY")
		return
	}
	u, _ := httpserver.UserFromContext(r.Context())
	value, err := h.svc.UpdateBaseCurrency(r.Context(), u, req.Currency, httpserver.RequestIDFromContext(r.Context()))
	h.respond(w, r, value, err)
}
func (h *Handler) respond(w http.ResponseWriter, r *http.Request, value any, err error) {
	if err == nil {
		httpserver.WriteSuccess(w, http.StatusOK, value)
		return
	}
	status, code, message := http.StatusInternalServerError, httpserver.CodeDatabase, "DATABASE_ERROR"
	if errors.Is(err, ErrForbidden) {
		status, code, message = http.StatusForbidden, httpserver.CodeForbidden, "FORBIDDEN"
	} else if errors.Is(err, ErrInvalidState) {
		status, code, message = http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_STATE"
	}
	if errors.Is(err, ErrNotFound) {
		status, code, message = http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND"
	}
	if errors.Is(err, ErrConflict) {
		status, code, message = http.StatusConflict, httpserver.CodeConflict, "CONFLICT"
	}
	if status >= 500 {
		h.logger.Error("finance service error", slog.String("request_id", httpserver.RequestIDFromContext(r.Context())), slog.Any("error", err))
	}
	httpserver.WriteError(w, status, code, message, httpserver.RequestIDFromContext(r.Context()))
}
