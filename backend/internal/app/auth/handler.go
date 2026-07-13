package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prelove/zedu/backend/internal/platform/auth"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
)

// RefreshCookieName is the cookie name for the refresh token.
const RefreshCookieName = "zedu_refresh"

// refreshTokenDuration is the lifetime of a refresh session.
const refreshTokenDuration = 14 * 24 * time.Hour

// accessTokenDuration is the lifetime of an access token.
const accessTokenDuration = 60 * time.Minute

// maxLoginFailures is the number of consecutive failures before lockout.
const maxLoginFailures = 5

// lockoutDuration is how long an account is locked after too many failures.
const lockoutDuration = 15 * time.Minute

// Tx is the minimal transaction interface used by the auth Handler.
// *sql.Tx satisfies this interface; tests may provide a wrapper to inject
// failures on specific query types within a transaction.
type Tx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Commit() error
	Rollback() error
}

// dbAdapter wraps *sql.DB to satisfy the DB interface by converting
// *sql.Tx returned by BeginTx into the Tx interface.
type dbAdapter struct {
	*sql.DB
}

func (a *dbAdapter) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := a.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// DB is the minimal database interface used by the auth Handler.
// *sql.DB does NOT directly satisfy this interface (BeginTx returns *sql.Tx,
// not Tx); use NewHandler which wraps *sql.DB in a dbAdapter. Test wrappers
// that implement DB directly are also accepted.
type DB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Handler implements the auth HTTP handlers.
type Handler struct {
	db        DB
	jwtSecret string
	logger    *slog.Logger
}

// NewHandler creates a new auth handler. db may be *sql.DB (wrapped in a
// dbAdapter) or any type that directly implements the DB interface (used by
// tests for fault injection).
func NewHandler(db any, jwtSecret string, logger *slog.Logger) *Handler {
	switch d := db.(type) {
	case DB:
		return &Handler{db: d, jwtSecret: jwtSecret, logger: logger}
	case *sql.DB:
		return &Handler{db: &dbAdapter{d}, jwtSecret: jwtSecret, logger: logger}
	default:
		panic(fmt.Sprintf("NewHandler: unsupported db type %T", db))
	}
}

// MountRoutes mounts auth routes onto the given mux. It returns the mux
// so callers can chain further route mounting. authDB is the real *sql.DB
// used by AuthMiddleware to confirm account status on every protected request;
// it must not be nil and must not be a test wrapper.
func MountRoutes(mux *http.ServeMux, h *Handler, authDB *sql.DB) *http.ServeMux {
	// Public routes (no auth middleware).
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/refresh", h.Refresh)

	// Authenticated routes — each wrapped with AuthMiddleware.
	authMW := httpserver.AuthMiddleware(h.jwtSecret, authDB)

	mux.Handle("POST /auth/logout", authMW(http.HandlerFunc(h.Logout)))
	mux.Handle("GET /auth/me", authMW(http.HandlerFunc(h.Me)))
	mux.Handle("GET /users", authMW(http.HandlerFunc(h.handleUsers)))
	mux.Handle("POST /users", authMW(http.HandlerFunc(h.handleUsers)))
	mux.Handle("POST /users/{id}/disable", authMW(http.HandlerFunc(h.handleDisableUser)))

	return mux
}

// LoginRequest is the JSON body for POST /auth/login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the data payload for a successful login.
type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	Role        string `json:"role"`
}

// writeDBError writes a 500 + 50002 response for database/transaction failures.
func (h *Handler) writeDBError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	rid := httpserver.RequestIDFromContext(r.Context())
	h.logger.Error(msg,
		slog.String("request_id", rid),
		slog.Any("error", err),
	)
	httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeDatabase, "DATABASE_ERROR")
}

// writeInternalError writes a 500 + 50001 response for non-database internal failures.
func (h *Handler) writeInternalError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	rid := httpserver.RequestIDFromContext(r.Context())
	h.logger.Error(msg,
		slog.String("request_id", rid),
		slog.Any("error", err),
	)
	httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeInternal, "INTERNAL_ERROR")
}

// insertAuditLog inserts an operation_log row within the given transaction.
// detail must not contain password, hash, token, or Authorization values.
func insertAuditLog(tx Tx, ctx context.Context, actorID int64, actorName, action, targetType string, targetID int64, detailJSON, requestID string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO operation_log (operator_id, operator_name, action, target_type, target_id, detail_json, request_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		actorID, actorName, action, targetType, targetID, detailJSON, requestID,
	)
	return err
}

// Login handles POST /auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	rid := httpserver.RequestIDFromContext(r.Context())

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeLoginFailed, "LOGIN_FAILED")
		return
	}

	// Look up user by username.
	var (
		id           int64
		passwordHash string
		role         string
		status       string
		username     string
		lockedUntil  *time.Time
	)
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, username, password_hash, role, status, locked_until FROM user_account WHERE username = ?`,
		req.Username,
	).Scan(&id, &username, &passwordHash, &role, &status, &lockedUntil)

	// Same error for unknown user and wrong password — do not disclose account existence.
	authFailedCode := httpserver.CodeLoginFailed
	authFailedMsg := "LOGIN_FAILED"

	if err == sql.ErrNoRows {
		auth.VerifyPasswordDummy(req.Password)
		h.logger.Info("login failed: unknown user",
			slog.String("request_id", rid),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, authFailedCode, authFailedMsg)
		return
	}
	if err != nil {
		h.writeDBError(w, r, "login query error", err)
		return
	}

	// Check lockout.
	if lockedUntil != nil && time.Now().UTC().Before(*lockedUntil) {
		h.logger.Info("login rejected: account locked",
			slog.String("request_id", rid),
			slog.Int64("user_id", id),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeLocked, "ACCOUNT_LOCKED")
		return
	}

	// Check if account is active.
	if status != "ACTIVE" {
		h.logger.Info("login failed: account not active",
			slog.String("request_id", rid),
			slog.Int64("user_id", id),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, authFailedCode, authFailedMsg)
		return
	}

	// Verify password.
	if !auth.VerifyPassword(passwordHash, req.Password) {
		lockTime := time.Now().UTC().Add(lockoutDuration)
		var (
			newFailCount   int
			newLockedUntil *time.Time
		)
		err = h.db.QueryRowContext(r.Context(),
			`UPDATE user_account
			 SET login_fail_count = login_fail_count + 1,
			     locked_until = CASE WHEN login_fail_count + 1 >= ? THEN ? ELSE locked_until END
			 WHERE id = ?
			 RETURNING login_fail_count, locked_until`,
			maxLoginFailures, lockTime, id,
		).Scan(&newFailCount, &newLockedUntil)

		if err != nil {
			h.writeDBError(w, r, "login fail count update error", err)
			return
		}

		if newFailCount >= maxLoginFailures {
			h.logger.Info("login failed: account locked after max failures",
				slog.String("request_id", rid),
				slog.Int64("user_id", id),
				slog.Int("fail_count", newFailCount),
			)
			httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeLocked, "ACCOUNT_LOCKED")
			return
		}

		h.logger.Info("login failed: wrong password",
			slog.String("request_id", rid),
			slog.Int64("user_id", id),
			slog.Int("fail_count", newFailCount),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, authFailedCode, authFailedMsg)
		return
	}

	// === Successful login: reset fail count + insert refresh session + audit in one transaction ===
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		h.writeInternalError(w, r, "generate refresh token", err)
		return
	}
	refreshHash := auth.HashRefreshToken(refreshToken)
	expiresAt := time.Now().UTC().Add(refreshTokenDuration)
	now := time.Now().UTC()

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.writeDBError(w, r, "login begin tx", err)
		return
	}

	_, err = tx.ExecContext(r.Context(),
		`UPDATE user_account SET login_fail_count = 0, locked_until = NULL, last_login_at = ? WHERE id = ?`,
		now, id,
	)
	if err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "login reset fail count", err)
		return
	}

	_, err = tx.ExecContext(r.Context(),
		`INSERT INTO refresh_session (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		id, refreshHash, expiresAt,
	)
	if err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "login insert refresh session", err)
		return
	}

	detail := `{"action":"login"}`
	if err := insertAuditLog(tx, r.Context(), id, username, "LOGIN", "USER", id, detail, rid); err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "login audit log", err)
		return
	}

	if err := tx.Commit(); err != nil {
		h.writeDBError(w, r, "login commit", err)
		return
	}

	// Issue access token (after tx committed; JWT is stateless).
	accessToken, err := auth.SignAccessToken(h.jwtSecret, id, role, accessTokenDuration)
	if err != nil {
		h.writeInternalError(w, r, "sign access token", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    refreshToken,
		Path:     "/auth",
		MaxAge:   int(refreshTokenDuration.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	h.logger.Info("login success",
		slog.String("request_id", rid),
		slog.Int64("user_id", id),
		slog.String("role", role),
	)

	httpserver.WriteSuccess(w, http.StatusOK, LoginResponse{
		AccessToken: accessToken,
		Role:        role,
	})
}

// Refresh handles POST /auth/refresh. Reads refresh token from cookie,
// rotates it in a single transaction, and returns a new access token.
// The rotation is atomic: ACTIVE check, revoke old session, insert new session,
// and audit log are all in one BeginTx. Concurrent requests with the same
// token will see rows-affected=0 on the conditional UPDATE and be rejected.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	rid := httpserver.RequestIDFromContext(r.Context())

	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	refreshToken := cookie.Value
	refreshHash := auth.HashRefreshToken(refreshToken)

	// Look up the session (read-only, outside the rotation tx).
	var (
		sessionID int64
		userID    int64
		role      string
		username  string
		revokedAt *time.Time
		expiresAt time.Time
	)
	err = h.db.QueryRowContext(r.Context(),
		`SELECT rs.id, rs.user_id, ua.role, ua.username, rs.revoked_at, rs.expires_at
		 FROM refresh_session rs
		 JOIN user_account ua ON ua.id = rs.user_id
		 WHERE rs.token_hash = ?`,
		refreshHash,
	).Scan(&sessionID, &userID, &role, &username, &revokedAt, &expiresAt)

	if err == sql.ErrNoRows {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}
	if err != nil {
		h.writeDBError(w, r, "refresh query error", err)
		return
	}

	// Check if session is revoked.
	if revokedAt != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	// Check if session is expired.
	if time.Now().UTC().After(expiresAt) {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	// Generate new refresh token (crypto rand, safe outside tx).
	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		h.writeInternalError(w, r, "generate refresh token", err)
		return
	}
	newRefreshHash := auth.HashRefreshToken(newRefreshToken)
	newExpiresAt := time.Now().UTC().Add(refreshTokenDuration)
	now := time.Now().UTC()

	// Atomic rotation: single transaction.
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.writeDBError(w, r, "refresh begin tx", err)
		return
	}

	// Re-check ACTIVE status inside the transaction (don't trust outside read).
	var status string
	err = tx.QueryRowContext(r.Context(),
		`SELECT status FROM user_account WHERE id = ?`,
		userID,
	).Scan(&status)
	if err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "refresh active check", err)
		return
	}
	if status != "ACTIVE" {
		_ = tx.Rollback()
		// Revoke the session inside the same tx for consistency.
		_, _ = tx.ExecContext(r.Context(),
			`UPDATE refresh_session SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL`,
			now, sessionID,
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	// Conditionally revoke the old session. If another concurrent request
	// already revoked it, rows-affected will be 0 and we roll back.
	res, err := tx.ExecContext(r.Context(),
		`UPDATE refresh_session SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL`,
		now, sessionID,
	)
	if err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "refresh revoke old session", err)
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected != 1 {
		_ = tx.Rollback()
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	// Insert the new session in the same transaction.
	_, err = tx.ExecContext(r.Context(),
		`INSERT INTO refresh_session (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		userID, newRefreshHash, newExpiresAt,
	)
	if err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "refresh insert new session", err)
		return
	}

	// Audit log in the same transaction.
	detail := `{"action":"refresh"}`
	if err := insertAuditLog(tx, r.Context(), userID, username, "REFRESH", "SESSION", sessionID, detail, rid); err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "refresh audit log", err)
		return
	}

	if err := tx.Commit(); err != nil {
		h.writeDBError(w, r, "refresh commit", err)
		return
	}

	// Issue new access token (after tx committed; JWT is stateless).
	accessToken, err := auth.SignAccessToken(h.jwtSecret, userID, role, accessTokenDuration)
	if err != nil {
		h.writeInternalError(w, r, "sign access token", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    newRefreshToken,
		Path:     "/auth",
		MaxAge:   int(refreshTokenDuration.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	httpserver.WriteSuccess(w, http.StatusOK, LoginResponse{
		AccessToken: accessToken,
		Role:        role,
	})
}

// Logout handles POST /auth/logout. Revokes the refresh session from the cookie
// and writes an audit log entry in the same transaction. If the DB write fails,
// does not return success and does not clear the cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	rid := httpserver.RequestIDFromContext(r.Context())
	user, ok := httpserver.UserFromContext(r.Context())
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	cookie, err := r.Cookie(RefreshCookieName)
	if err != nil || cookie.Value == "" {
		// No cookie — still write audit and return success (idempotent logout).
		tx, err := h.db.BeginTx(r.Context(), nil)
		if err != nil {
			h.writeDBError(w, r, "logout begin tx", err)
			return
		}
		detail := `{"action":"logout","reason":"no_cookie"}`
		if err := insertAuditLog(tx, r.Context(), user.UserID, "", "LOGOUT", "USER", user.UserID, detail, rid); err != nil {
			_ = tx.Rollback()
			h.writeDBError(w, r, "logout audit log", err)
			return
		}
		if err := tx.Commit(); err != nil {
			h.writeDBError(w, r, "logout commit", err)
			return
		}
		// Clear cookie and return success.
		http.SetCookie(w, &http.Cookie{
			Name: RefreshCookieName, Value: "", Path: "/auth", MaxAge: -1,
			HttpOnly: true, Secure: true, SameSite: http.SameSiteStrictMode,
		})
		httpserver.WriteSuccess(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	refreshHash := auth.HashRefreshToken(cookie.Value)
	now := time.Now().UTC()

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.writeDBError(w, r, "logout begin tx", err)
		return
	}

	// Revoke the session.
	_, err = tx.ExecContext(r.Context(),
		`UPDATE refresh_session SET revoked_at = ? WHERE token_hash = ? AND revoked_at IS NULL`,
		now, refreshHash,
	)
	if err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "logout revoke session", err)
		return
	}
	// Even if rowsAffected == 0 (already revoked or not found), we still
	// write the audit and return success — logout is idempotent.

	// Audit log in the same transaction.
	detail := `{"action":"logout"}`
	if err := insertAuditLog(tx, r.Context(), user.UserID, "", "LOGOUT", "USER", user.UserID, detail, rid); err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "logout audit log", err)
		return
	}

	if err := tx.Commit(); err != nil {
		h.writeDBError(w, r, "logout commit", err)
		return
	}

	// Clear the cookie only after successful commit.
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookieName,
		Value:    "",
		Path:     "/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	httpserver.WriteSuccess(w, http.StatusOK, map[string]any{"ok": true})
}

// Me handles GET /auth/me. Returns the current authenticated user's info.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := httpserver.UserFromContext(r.Context())
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	var username, role, displayName string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT username, role, display_name FROM user_account WHERE id = ?`,
		user.UserID,
	).Scan(&username, &role, &displayName)
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "NOT_FOUND")
		return
	}

	httpserver.WriteSuccess(w, http.StatusOK, map[string]any{
		"id":          user.UserID,
		"username":    username,
		"role":        role,
		"displayName": displayName,
	})
}

// handleUsers routes GET (list) and POST (create) for /users.
func (h *Handler) handleUsers(w http.ResponseWriter, r *http.Request) {
	user, ok := httpserver.UserFromContext(r.Context())
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	// Only Owner can manage users.
	if user.Role != "OWNER" {
		httpserver.WriteErrorFromContext(w, r, http.StatusForbidden, httpserver.CodeForbidden, "FORBIDDEN")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.ListUsers(w, r)
	case http.MethodPost:
		h.CreateUser(w, r, user)
	default:
		httpserver.WriteErrorFromContext(w, r, http.StatusForbidden, httpserver.CodeForbidden, "METHOD_NOT_ALLOWED")
	}
}

// ListUsers handles GET /users (Owner-only).
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, username, role, display_name, status, created_at FROM user_account WHERE deleted_at IS NULL ORDER BY id`,
	)
	if err != nil {
		h.writeDBError(w, r, "list users query", err)
		return
	}
	defer rows.Close()

	type UserItem struct {
		ID          int64  `json:"id"`
		Username    string `json:"username"`
		Role        string `json:"role"`
		DisplayName string `json:"displayName"`
		Status      string `json:"status"`
	}
	var items []UserItem
	for rows.Next() {
		var u UserItem
		var createdAt string
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.DisplayName, &u.Status, &createdAt); err != nil {
			continue
		}
		items = append(items, u)
	}

	if items == nil {
		items = []UserItem{}
	}

	httpserver.WriteSuccess(w, http.StatusOK, map[string]any{
		"items":    items,
		"page":     1,
		"pageSize": 100,
		"total":    len(items),
	})
}

// CreateUserRequest is the JSON body for POST /users.
type CreateUserRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	DisplayName string `json:"displayName"`
}

// CreateUser handles POST /users (Owner-only). Only OPERATOR accounts can be
// created via this endpoint; OWNER accounts are created only via onboarding
// or DB seeding. Password must meet PRD rules (>=8 chars, letter+digit).
// The user creation and audit log are in the same transaction.
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request, actor httpserver.AuthUser) {
	rid := httpserver.RequestIDFromContext(r.Context())

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_REQUEST")
		return
	}

	if req.Username == "" {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "USERNAME_REQUIRED")
		return
	}

	if err := auth.ValidatePassword(req.Password); err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "INVALID_PASSWORD")
		return
	}

	// Only OPERATOR can be created via this endpoint.
	if req.Role != "OPERATOR" {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "ROLE_MUST_BE_OPERATOR")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.writeInternalError(w, r, "hash password", err)
		return
	}

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.writeDBError(w, r, "create user begin tx", err)
		return
	}

	result, err := tx.ExecContext(r.Context(),
		`INSERT INTO user_account (username, password_hash, role, display_name) VALUES (?, ?, ?, ?)`,
		req.Username, hash, req.Role, displayName,
	)
	if err != nil {
		_ = tx.Rollback()
		httpserver.WriteErrorFromContext(w, r, http.StatusConflict, httpserver.CodeConflict, "USERNAME_CONFLICT")
		return
	}

	id, _ := result.LastInsertId()

	// Audit log in the same transaction.
	detail := `{"action":"create_operator","username":"` + req.Username + `"}`
	if err := insertAuditLog(tx, r.Context(), actor.UserID, "", "CREATE_OPERATOR", "USER", id, detail, rid); err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "create user audit log", err)
		return
	}

	if err := tx.Commit(); err != nil {
		h.writeDBError(w, r, "create user commit", err)
		return
	}

	httpserver.WriteSuccess(w, http.StatusCreated, map[string]any{
		"id":       id,
		"username": req.Username,
		"role":     req.Role,
	})
}

// handleDisableUser handles POST /users/{id}/disable (Owner-only).
func (h *Handler) handleDisableUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := extractUserIDFromPath(r.URL.Path, "/users/")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "INVALID_USER_ID")
		return
	}
	h.DisableUser(w, r, userID)
}

// DisableUser handles POST /users/{id}/disable (Owner-only).
// Sets status to DISABLED, revokes all active refresh sessions, and writes
// an audit log entry — all in the same transaction. If any step fails, the
// entire operation rolls back, leaving no "disabled but sessions active" state.
func (h *Handler) DisableUser(w http.ResponseWriter, r *http.Request, userID int64) {
	rid := httpserver.RequestIDFromContext(r.Context())
	user, ok := httpserver.UserFromContext(r.Context())
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}
	if user.Role != "OWNER" {
		httpserver.WriteErrorFromContext(w, r, http.StatusForbidden, httpserver.CodeForbidden, "FORBIDDEN")
		return
	}

	// Prevent self-disable.
	if user.UserID == userID {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnprocessableEntity, httpserver.CodeInvalidState, "CANNOT_DISABLE_SELF")
		return
	}

	now := time.Now().UTC()

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.writeDBError(w, r, "disable begin tx", err)
		return
	}

	// Set status to DISABLED.
	result, err := tx.ExecContext(r.Context(),
		`UPDATE user_account SET status = 'DISABLED' WHERE id = ? AND deleted_at IS NULL`,
		userID,
	)
	if err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "disable set status", err)
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		_ = tx.Rollback()
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "USER_NOT_FOUND")
		return
	}

	// Revoke all active refresh sessions in the same transaction.
	_, err = tx.ExecContext(r.Context(),
		`UPDATE refresh_session SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`,
		now, userID,
	)
	if err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "disable revoke sessions", err)
		return
	}

	// Audit log in the same transaction.
	detail := `{"action":"disable"}`
	if err := insertAuditLog(tx, r.Context(), user.UserID, "", "DISABLE_USER", "USER", userID, detail, rid); err != nil {
		_ = tx.Rollback()
		h.writeDBError(w, r, "disable audit log", err)
		return
	}

	if err := tx.Commit(); err != nil {
		h.writeDBError(w, r, "disable commit", err)
		return
	}

	httpserver.WriteSuccess(w, http.StatusOK, map[string]any{"id": userID, "status": "DISABLED"})
}

// extractUserIDFromPath extracts a numeric ID from a path like /users/{id}/disable.
func extractUserIDFromPath(path, prefix string) (int64, bool) {
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	rest := path[len(prefix):]
	idx := strings.Index(rest, "/")
	var idStr string
	if idx >= 0 {
		idStr = rest[:idx]
	} else {
		idStr = rest
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, false
	}
	return id, true
}
