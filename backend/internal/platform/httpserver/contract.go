package httpserver

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/prelove/zedu/backend/internal/platform/logging"
)

// ErrorCode is a stable business error code as defined in the M2 contract.
type ErrorCode int

const (
	CodeSuccess      ErrorCode = 0
	CodeUnauth       ErrorCode = 40101
	CodeLoginFailed  ErrorCode = 40102
	CodeLocked       ErrorCode = 40103
	CodeForbidden    ErrorCode = 40301
	CodeNotFound     ErrorCode = 40401
	CodeConflict     ErrorCode = 40901
	CodeInvalidState ErrorCode = 42201
	CodeInternal     ErrorCode = 50001
	CodeDatabase     ErrorCode = 50002
)

// successEnvelope is the unified success response outer structure.
type successEnvelope struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

// errorEnvelope is the unified error response outer structure.
type errorEnvelope struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	RequestID string    `json:"requestId"`
}

// WriteSuccess writes a unified success JSON response.
func WriteSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(successEnvelope{Code: 0, Data: data})
}

// WriteError writes a unified error JSON response.
func WriteError(w http.ResponseWriter, status int, code ErrorCode, message, requestID string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Code: code, Message: message, RequestID: requestID})
}

// WriteErrorFromContext writes an error response, extracting the request ID from context.
func WriteErrorFromContext(w http.ResponseWriter, r *http.Request, status int, code ErrorCode, message string) {
	rid, _ := logging.RequestID(r.Context())
	if rid == "" {
		rid = "unknown"
	}
	WriteError(w, status, code, message, rid)
}

// RequestIDFromContext extracts the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	rid, _ := logging.RequestID(ctx)
	return rid
}
