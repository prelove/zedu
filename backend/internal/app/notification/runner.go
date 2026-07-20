package notification

import (
	"context"
	"database/sql"
	"fmt"
	"html"
	"strings"
	"time"
)

// ReminderConfig controls the reminder scan window. The window is fixed at
// 30 minutes for MVP; no configuration UI is exposed.
type ReminderConfig struct {
	Window time.Duration
}

// ReminderRunner scans SCHEDULED lessons whose start time falls within the
// configured window and queues LESSON_REMINDER outbox rows. It does not send
// emails directly — the existing ClaimAndSend runner handles delivery.
// The runner is idempotent: re-scanning the same lesson produces no duplicate
// outbox rows because of the unique idempotency_key.
type ReminderRunner struct {
	db  *sql.DB
	cfg ReminderConfig
}

// NewReminderRunner creates a runner that scans for lessons starting within
// the configured window. If window is zero, it defaults to 30 minutes.
func NewReminderRunner(db *sql.DB, cfg ReminderConfig) *ReminderRunner {
	if cfg.Window <= 0 {
		cfg.Window = 30 * time.Minute
	}
	return &ReminderRunner{db: db, cfg: cfg}
}

// ScanReminders finds SCHEDULED lessons within the reminder window and queues
// LESSON_REMINDER outbox rows for each recipient. It does not modify lesson
// state. Idempotency is guaranteed by the unique idempotency_key.
func (r *ReminderRunner) ScanReminders(ctx context.Context) error {
	now := time.Now().UTC()
	windowEnd := now.Add(r.cfg.Window)
	rows, err := r.db.QueryContext(ctx, `
		SELECT l.id, l.lesson_no, l.student_id, l.scheduled_start_at, l.timezone
		FROM lesson l
		WHERE l.status = 'SCHEDULED'
		  AND l.scheduled_start_at >= ?
		  AND l.scheduled_start_at <= ?`, now, windowEnd)
	if err != nil {
		return fmt.Errorf("scan reminders: %w", err)
	}
	type candidate struct {
		lessonID  int64
		studentID int64
		lessonNo  string
		startUTC  time.Time
		timezone  string
	}
	candidates := make([]candidate, 0)
	for rows.Next() {
		var item candidate
		if err := rows.Scan(&item.lessonID, &item.lessonNo, &item.studentID, &item.startUTC, &item.timezone); err != nil {
			return fmt.Errorf("scan reminder row: %w", err)
		}
		candidates = append(candidates, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	// database.Open deliberately uses one SQLite connection. Closing the scan
	// cursor before recipient queries prevents a self-deadlock on that pool.
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close reminder scan: %w", err)
	}
	for _, item := range candidates {
		if err := r.queueReminderForLesson(ctx, item.lessonID, item.studentID, item.lessonNo, item.startUTC, item.timezone); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReminderRunner) queueReminderForLesson(ctx context.Context, lessonID, studentID int64, lessonNo string, startUTC time.Time, timezone string) error {
	emails, err := r.collectRecipientEmails(ctx, studentID)
	if err != nil {
		return err
	}
	if len(emails) == 0 {
		return nil
	}
	subject := "Zedu lesson reminder " + lessonNo
	body := fmt.Sprintf("<p>Reminder: lesson <strong>%s</strong> starts soon.</p><p>UTC: %s<br>Timezone: %s</p>",
		html.EscapeString(lessonNo),
		html.EscapeString(startUTC.Format(time.RFC3339)),
		html.EscapeString(timezone))
	for _, email := range emails {
		key := fmt.Sprintf("lesson:%d:LESSON_REMINDER:%s", lessonID, email)
		if _, err := r.db.ExecContext(ctx,
			`INSERT OR IGNORE INTO notification_outbox(lesson_id,event_type,recipient_email,subject,html_body,idempotency_key) VALUES (?,?,?,?,?,?)`,
			lessonID, "LESSON_REMINDER", email, subject, body, key); err != nil {
			return fmt.Errorf("queue reminder: %w", err)
		}
	}
	return nil
}

func (r *ReminderRunner) collectRecipientEmails(ctx context.Context, studentID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT email FROM student WHERE id=? AND email IS NOT NULL AND trim(email)<>''
		 UNION
		 SELECT email FROM parent WHERE student_id=? AND email IS NOT NULL AND trim(email)<>''`,
		studentID, studentID)
	if err != nil {
		return nil, fmt.Errorf("collect recipient emails: %w", err)
	}
	defer rows.Close()
	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		email = strings.ToLower(strings.TrimSpace(email))
		if email != "" {
			emails = append(emails, email)
		}
	}
	return emails, rows.Err()
}
