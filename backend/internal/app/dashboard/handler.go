package dashboard

import (
	"database/sql"
	"github.com/prelove/zedu/backend/internal/platform/httpserver"
	"net/http"
)

func MountRoutes(mux *http.ServeMux, db *sql.DB, secret string) {
	mux.Handle("GET /dashboard", httpserver.AuthMiddleware(secret, db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pending, failed int
		if err := db.QueryRow(`SELECT COUNT(*) FROM lesson WHERE status='SCHEDULED'`).Scan(&pending); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		if err := db.QueryRow(`SELECT COUNT(*) FROM notification_outbox WHERE status='FAILED'`).Scan(&failed); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		httpserver.WriteSuccess(w, 200, map[string]int{"pendingLessonConfirmations": pending, "failedNotifications": failed})
	})))
}
