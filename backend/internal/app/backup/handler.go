package backup

import (
	"database/sql"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"github.com/prelove/zedu/backend/internal/repository"
	"net/http"
	"os"
	"path/filepath"
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
		name, err := CreatePackage(db, dir, os.Getenv("ZEDU_DATA_ROOT"))
		if err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		tx, err := db.BeginTx(r.Context(), nil)
		if err != nil {
			_ = os.RemoveAll(filepath.Join(dir, name))
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		actorName, _ := repository.ActorName(tx, r.Context(), u.UserID)
		if err = repository.InsertAuditLog(tx, r.Context(), u.UserID, actorName, "BACKUP_CREATE", "backup", 0, map[string]any{"package": name}, httpserver.RequestIDFromContext(r.Context())); err != nil {
			_ = tx.Rollback()
			_ = os.RemoveAll(filepath.Join(dir, name))
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		if err = tx.Commit(); err != nil {
			_ = os.RemoveAll(filepath.Join(dir, name))
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		httpserver.WriteSuccess(w, 201, map[string]string{"file": name})
	})))
}
