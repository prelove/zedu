package dashboard

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/prelove/zedu/backend/internal/platform/httpserver"
)

// Dashboard is the read-only operational summary returned by GET /dashboard.
// All fields are derived from immutable facts; the handler performs no writes.
type Dashboard struct {
	TodayLessons               int   `json:"todayLessons"`
	PendingLessonConfirmations int   `json:"pendingLessonConfirmations"`
	RenewalNeededStudents      int   `json:"renewalNeededStudents"`
	TeacherPayableAggregate    int64 `json:"teacherPayableAggregate"`
	FailedNotifications        int   `json:"failedNotifications"`
}

func MountRoutes(mux *http.ServeMux, db *sql.DB, secret string) {
	mux.Handle("GET /dashboard", httpserver.AuthMiddleware(secret, db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var d Dashboard
		// Product dates are based on the system default of Asia/Tokyo while
		// timestamps remain stored as UTC.
		now := time.Now().UTC()
		startOfDay, endOfDay := tokyoDayRange(now)
		if err := db.QueryRow(`SELECT COUNT(*) FROM lesson WHERE scheduled_start_at >= ? AND scheduled_start_at < ?`, startOfDay, endOfDay).Scan(&d.TodayLessons); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		// Pending confirmations: lessons still SCHEDULED whose start time has passed.
		if err := db.QueryRow(`SELECT COUNT(*) FROM lesson WHERE status='SCHEDULED' AND scheduled_start_at <= ?`, now).Scan(&d.PendingLessonConfirmations); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		// Renewal needed: ACTIVE students with lesson_balance <= 2.
		if err := db.QueryRow(`SELECT COUNT(*) FROM student_course_enrollment WHERE status='ACTIVE' AND lesson_balance <= 2`).Scan(&d.RenewalNeededStudents); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		// Teacher payable aggregate: sum of amount_delta in teacher_account_ledger.
		var payable sql.NullInt64
		if err := db.QueryRow(`SELECT COALESCE(SUM(amount_delta), 0) FROM teacher_account_ledger`).Scan(&payable); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		d.TeacherPayableAggregate = payable.Int64
		// Failed notifications.
		if err := db.QueryRow(`SELECT COUNT(*) FROM notification_outbox WHERE status='FAILED'`).Scan(&d.FailedNotifications); err != nil {
			httpserver.WriteErrorFromContext(w, r, 500, httpserver.CodeDatabase, "DATABASE_ERROR")
			return
		}
		httpserver.WriteSuccess(w, 200, d)
	})))
}

func tokyoDayRange(now time.Time) (time.Time, time.Time) {
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		// Asia/Tokyo is provided by Go's standard time zone database. Keep a
		// deterministic fixed-offset fallback for constrained environments.
		tokyo = time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	local := now.In(tokyo)
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, tokyo)
	return start.UTC(), start.AddDate(0, 0, 1).UTC()
}
