package backup

import (
	"database/sql"
	"fmt"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func MountRoutes(mux *http.ServeMux, db *sql.DB, secret string) {
	mux.Handle("POST /system/backups", httpserver.AuthMiddleware(secret, db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _ := httpserver.UserFromContext(r.Context())
		if u.Role != "OWNER" {
			httpserver.WriteErrorFromContext(w, r, 403, httpserver.CodeForbidden, "FORBIDDEN")
			return
		}
		dir := os.Getenv("ZEDU_BACKUP_DIR")
		if dir == "" {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeInternal, "BACKUP_NOT_CONFIGURED")
			return
		}
		if err := os.MkdirAll(dir, 0700); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeInternal, "BACKUP_FAILED")
			return
		}
		path := filepath.Join(dir, fmt.Sprintf("zedu-%s.db", time.Now().UTC().Format("20060102T150405Z")))
		if _, err := db.Exec("VACUUM INTO ?", path); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		name, _ := repository.ActorName(tx, r.Context(), u.UserID)
		if err = repository.InsertAuditLog(tx, r.Context(), u.UserID, name, "BACKUP_CREATE", "backup", 0, map[string]any{"file": filepath.Base(path)}, httpserver.RequestIDFromContext(r.Context())); err != nil || tx.Commit() != nil {
			tx.Rollback()
			_ = os.Remove(path)
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		httpserver.WriteSuccess(w, 201, map[string]string{"file": filepath.Base(path)})
	})))
}
