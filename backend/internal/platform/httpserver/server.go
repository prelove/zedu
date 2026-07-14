package httpserver

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/prelove/zedu/backend/internal/platform/auth"
)

// ctxKey is the context key type for auth values.
type ctxKey int

const (
	ctxKeyUser ctxKey = iota
)

// AuthUser holds the authenticated user's identity extracted from the JWT
// and confirmed against the database on every request.
type AuthUser struct {
	UserID int64
	Role   string
}

// UserFromContext returns the authenticated user from the context, if any.
func UserFromContext(ctx context.Context) (AuthUser, bool) {
	u, ok := ctx.Value(ctxKeyUser).(AuthUser)
	return u, ok
}

// New returns a minimal HTTP server with a health check endpoint.
// Auth routes are mounted by auth.MountRoutes.
func New() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}

// AuthMiddleware returns a middleware that validates the Bearer access token,
// then confirms the user is still ACTIVE in the database and loads the
// authoritative role from DB (not from the JWT). If the token is missing,
// invalid, or the account is no longer ACTIVE, it returns 401 with code 40101.
func AuthMiddleware(jwtSecret string, db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				WriteErrorFromContext(w, r, http.StatusUnauthorized, CodeUnauth, "AUTH_REQUIRED")
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := auth.VerifyAccessToken(jwtSecret, tokenStr)
			if err != nil {
				WriteErrorFromContext(w, r, http.StatusUnauthorized, CodeUnauth, "AUTH_REQUIRED")
				return
			}

			// Confirm the account is still ACTIVE and load the authoritative role.
			var role string
			err = db.QueryRowContext(r.Context(),
				`SELECT role FROM user_account WHERE id = ? AND status = 'ACTIVE' AND deleted_at IS NULL`,
				claims.UserID,
			).Scan(&role)
			if err != nil {
				if err == sql.ErrNoRows {
					// Account no longer ACTIVE or deleted → unauthenticated.
					WriteErrorFromContext(w, r, http.StatusUnauthorized, CodeUnauth, "AUTH_REQUIRED")
					return
				}
				// Database error (not "not found") → 500 + 50002.
				WriteErrorFromContext(w, r, http.StatusInternalServerError, CodeDatabase, "DATABASE_ERROR")
				return
			}

			user := AuthUser{UserID: claims.UserID, Role: role}
			ctx := context.WithValue(r.Context(), ctxKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole wraps a handler and rejects requests where the authenticated user
// does not have the required role. Owner includes Operator permissions.
func RequireRole(requiredRole string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			WriteErrorFromContext(w, r, http.StatusUnauthorized, CodeUnauth, "AUTH_REQUIRED")
			return
		}

		// Owner includes all Operator permissions.
		if user.Role == "OWNER" {
			next.ServeHTTP(w, r)
			return
		}

		if user.Role != requiredRole {
			WriteErrorFromContext(w, r, http.StatusForbidden, CodeForbidden, "FORBIDDEN")
			return
		}
		next.ServeHTTP(w, r)
	})
}
