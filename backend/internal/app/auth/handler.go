package auth

import (
	"database/sql"
	"encoding/json"
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

// Handler implements the auth HTTP handlers.
type Handler struct {
	db        *sql.DB
	jwtSecret string
	logger    *slog.Logger
}

// NewHandler creates a new auth handler.
func NewHandler(db *sql.DB, jwtSecret string, logger *slog.Logger) *Handler {
	return &Handler{db: db, jwtSecret: jwtSecret, logger: logger}
}

// MountRoutes mounts auth routes onto the given mux. It returns the mux
// so callers can chain further route mounting.
func MountRoutes(mux *http.ServeMux, h *Handler) *http.ServeMux {
	// Public routes (no auth middleware).
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/refresh", h.Refresh)

	// Authenticated routes — each wrapped with AuthMiddleware.
	authMW := httpserver.AuthMiddleware(h.jwtSecret, h.db)

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
		lockedUntil  *time.Time
	)
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, password_hash, role, status, locked_until FROM user_account WHERE username = ?`,
		req.Username,
	).Scan(&id, &passwordHash, &role, &status, &lockedUntil)

	// Same error for unknown user and wrong password — do not disclose account existence.
	authFailedCode := httpserver.CodeLoginFailed
	authFailedMsg := "LOGIN_FAILED"

	if err == sql.ErrNoRows {
		// Dummy bcrypt compare to reduce timing-based user enumeration (P2.3).
		auth.VerifyPasswordDummy(req.Password)
		// Do not log the username — it is a user-supplied sensitive identifier.
		h.logger.Info("login failed: unknown user",
			slog.String("request_id", rid),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, authFailedCode, authFailedMsg)
		return
	}
	if err != nil {
		h.logger.Error("login query error",
			slog.String("request_id", rid),
			slog.Any("error", err),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, authFailedCode, authFailedMsg)
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
		// Same error as wrong password — don't disclose that account is disabled.
		h.logger.Info("login failed: account not active",
			slog.String("request_id", rid),
			slog.Int64("user_id", id),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, authFailedCode, authFailedMsg)
		return
	}

	// Verify password.
	if !auth.VerifyPassword(passwordHash, req.Password) {
		// Atomic increment: login_fail_count = login_fail_count + 1.
		// If the new count reaches maxLoginFailures, set locked_until in the
		// same statement. This avoids the read-modify-write race where
		// concurrent requests all read the same stale count and overwrite
		// each other, bypassing the 5th-failure lockout.
		lockTime := time.Now().UTC().Add(lockoutDuration)
		_, _ = h.db.ExecContext(r.Context(),
			`UPDATE user_account
			 SET login_fail_count = login_fail_count + 1,
			     locked_until = CASE WHEN login_fail_count + 1 >= ? THEN ? ELSE locked_until END
			 WHERE id = ?`,
			maxLoginFailures, lockTime, id,
		)

		// Read back the new count to determine the response code.
		var newFailCount int
		_ = h.db.QueryRowContext(r.Context(),
			`SELECT login_fail_count FROM user_account WHERE id = ?`,
			id,
		).Scan(&newFailCount)

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

	// Reset fail count on successful login.
	_, _ = h.db.ExecContext(r.Context(),
		`UPDATE user_account SET login_fail_count = 0, locked_until = NULL, last_login_at = ? WHERE id = ?`,
		time.Now().UTC(), id,
	)

	// Issue access token.
	accessToken, err := auth.SignAccessToken(h.jwtSecret, id, role, accessTokenDuration)
	if err != nil {
		h.logger.Error("sign access token",
			slog.String("request_id", rid),
			slog.Any("error", err),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, authFailedCode, authFailedMsg)
		return
	}

	// Issue refresh token.
	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		h.logger.Error("generate refresh token",
			slog.String("request_id", rid),
			slog.Any("error", err),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, authFailedCode, authFailedMsg)
		return
	}

	refreshHash := auth.HashRefreshToken(refreshToken)
	expiresAt := time.Now().UTC().Add(refreshTokenDuration)
	_, err = h.db.ExecContext(r.Context(),
		`INSERT INTO refresh_session (user_id, token_hash, expires_at) VALUES (?, ?, ?)`,
		id, refreshHash, expiresAt,
	)
	if err != nil {
		h.logger.Error("insert refresh session",
			slog.String("request_id", rid),
			slog.Any("error", err),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, authFailedCode, authFailedMsg)
		return
	}

	// Set refresh cookie.
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
// The rotation is atomic: revoke old session (conditional on revoked_at IS NULL)
// and insert new session in one BeginTx. Concurrent requests with the same
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
		status    string
		revokedAt *time.Time
		expiresAt time.Time
	)
	err = h.db.QueryRowContext(r.Context(),
		`SELECT rs.id, rs.user_id, ua.role, ua.status, rs.revoked_at, rs.expires_at
		 FROM refresh_session rs
		 JOIN user_account ua ON ua.id = rs.user_id
		 WHERE rs.token_hash = ?`,
		refreshHash,
	).Scan(&sessionID, &userID, &role, &status, &revokedAt, &expiresAt)

	if err == sql.ErrNoRows {
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}
	if err != nil {
		h.logger.Error("refresh query error",
			slog.String("request_id", rid),
			slog.Any("error", err),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
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

	// Check if account is still active.
	if status != "ACTIVE" {
		// Revoke the session (best-effort, no tx needed for read-only reject).
		_, _ = h.db.ExecContext(r.Context(),
			`UPDATE refresh_session SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL`,
			time.Now().UTC(), sessionID,
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusUnauthorized, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	// Generate new refresh token (crypto rand, safe outside tx).
	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}
	newRefreshHash := auth.HashRefreshToken(newRefreshToken)
	newExpiresAt := time.Now().UTC().Add(refreshTokenDuration)

	// Atomic rotation: single transaction.
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.Error("begin tx for refresh rotation",
			slog.String("request_id", rid),
			slog.Any("error", err),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	// Conditionally revoke the old session. If another concurrent request
	// already revoked it, rows-affected will be 0 and we roll back.
	res, err := tx.ExecContext(r.Context(),
		`UPDATE refresh_session SET revoked_at = ? WHERE id = ? AND revoked_at IS NULL`,
		time.Now().UTC(), sessionID,
	)
	if err != nil {
		_ = tx.Rollback()
		h.logger.Error("revoke old session in tx",
			slog.String("request_id", rid),
			slog.Any("error", err),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected != 1 {
		// Another request already rotated this token.
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
		h.logger.Error("insert new session in tx",
			slog.String("request_id", rid),
			slog.Any("error", err),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	if err := tx.Commit(); err != nil {
		h.logger.Error("commit refresh rotation",
			slog.String("request_id", rid),
			slog.Any("error", err),
		)
		httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeUnauth, "AUTH_REQUIRED")
		return
	}

	// Issue new access token (after tx committed; JWT is stateless).
	accessToken, err := auth.SignAccessToken(h.jwtSecret, userID, role, accessTokenDuration)
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeUnauth, "AUTH_REQUIRED")
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

// Logout handles POST /auth/logout. Revokes the refresh session from the cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(RefreshCookieName)
	if err == nil && cookie.Value != "" {
		refreshHash := auth.HashRefreshToken(cookie.Value)
		_, _ = h.db.ExecContext(r.Context(),
			`UPDATE refresh_session SET revoked_at = ? WHERE token_hash = ? AND revoked_at IS NULL`,
			time.Now().UTC(), refreshHash,
		)
	}

	// Clear the cookie.
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
		h.CreateUser(w, r)
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
		httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeInvalidState, "INTERNAL_ERROR")
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
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
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
		httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeInvalidState, "INTERNAL_ERROR")
		return
	}

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}

	result, err := h.db.ExecContext(r.Context(),
		`INSERT INTO user_account (username, password_hash, role, display_name) VALUES (?, ?, ?, ?)`,
		req.Username, hash, req.Role, displayName,
	)
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusConflict, httpserver.CodeConflict, "USERNAME_CONFLICT")
		return
	}

	id, _ := result.LastInsertId()

	httpserver.WriteSuccess(w, http.StatusCreated, map[string]any{
		"id":       id,
		"username": req.Username,
		"role":     req.Role,
	})
}

// handleDisableUser handles POST /users/{id}/disable (Owner-only).
// Extracts the user ID from the URL path.
func (h *Handler) handleDisableUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := extractUserIDFromPath(r.URL.Path, "/users/")
	if !ok {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "INVALID_USER_ID")
		return
	}
	h.DisableUser(w, r, userID)
}

// DisableUser handles POST /users/{id}/disable (Owner-only).
// Revokes all refresh sessions for the user and sets status to DISABLED.
func (h *Handler) DisableUser(w http.ResponseWriter, r *http.Request, userID int64) {
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

	// Set status to DISABLED.
	result, err := h.db.ExecContext(r.Context(),
		`UPDATE user_account SET status = 'DISABLED' WHERE id = ? AND deleted_at IS NULL`,
		userID,
	)
	if err != nil {
		httpserver.WriteErrorFromContext(w, r, http.StatusInternalServerError, httpserver.CodeInvalidState, "INTERNAL_ERROR")
		return
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		httpserver.WriteErrorFromContext(w, r, http.StatusNotFound, httpserver.CodeNotFound, "USER_NOT_FOUND")
		return
	}

	// Revoke all refresh sessions.
	_, _ = h.db.ExecContext(r.Context(),
		`UPDATE refresh_session SET revoked_at = ? WHERE user_id = ? AND revoked_at IS NULL`,
		time.Now().UTC(), userID,
	)

	httpserver.WriteSuccess(w, http.StatusOK, map[string]any{"id": userID, "status": "DISABLED"})
}

// extractUserIDFromPath extracts a numeric ID from a path like /users/{id}/disable.
func extractUserIDFromPath(path, prefix string) (int64, bool) {
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	rest := path[len(prefix):]
	// Find the next slash.
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
