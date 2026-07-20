// Package notification provides the bounded M4b outbox and Resend delivery path.
package notification

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/prelove/zedu/backend/internal/repository"
	"html"
	"strings"
)

// QueueLesson queues student and parent email notifications in the caller's transaction.
func QueueLesson(ctx context.Context, tx repository.Tx, lessonID, studentID int64, lessonNo, event, startUTC, timezone string) error {
	rows, err := tx.QueryContext(ctx, `SELECT email FROM student WHERE id=? AND email IS NOT NULL AND trim(email)<>'' UNION SELECT email FROM parent WHERE student_id=? AND email IS NOT NULL AND trim(email)<>''`, studentID, studentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	subject := "Zedu lesson " + lessonNo
	body := fmt.Sprintf("<p>Lesson <strong>%s</strong> %s.</p><p>UTC: %s<br>Timezone: %s</p>", html.EscapeString(lessonNo), html.EscapeString(strings.ToLower(strings.TrimPrefix(event, "LESSON_"))), html.EscapeString(startUTC), html.EscapeString(timezone))
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return err
		}
		email = strings.ToLower(strings.TrimSpace(email))
		if email == "" {
			continue
		}
		key := fmt.Sprintf("lesson:%d:%s:%s", lessonID, event, email)
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO notification_outbox(lesson_id,event_type,recipient_email,subject,html_body,idempotency_key) VALUES (?,?,?,?,?,?)`, lessonID, event, email, subject, body, key); err != nil {
			return err
		}
	}
	return rows.Err()
}

type Sender interface {
	Send(context.Context, string, string, string) (string, error)
}
type Outbox struct {
	ID             int64  `json:"id"`
	LessonID       int64  `json:"lessonId"`
	EventType      string `json:"eventType"`
	RecipientEmail string `json:"recipientEmail"`
	Status         string `json:"status"`
	Attempts       int    `json:"attempts"`
	LastError      string `json:"lastError,omitempty"`
}

func ClaimAndSend(ctx context.Context, db repository.DB, sender Sender) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return repository.ErrDatabase
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	var id int64
	var to, subject, body string
	err = tx.QueryRowContext(ctx, `SELECT id,recipient_email,subject,html_body FROM notification_outbox WHERE status IN ('PENDING','FAILED') AND attempts<3 AND available_at<=CURRENT_TIMESTAMP ORDER BY id LIMIT 1`).Scan(&id, &to, &subject, &body)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return repository.ErrDatabase
	}
	result, err := tx.ExecContext(ctx, `UPDATE notification_outbox SET status='PROCESSING',attempts=attempts+1,locked_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP WHERE id=? AND status IN ('PENDING','FAILED')`, id)
	if err != nil {
		return repository.ErrDatabase
	}
	n, _ := result.RowsAffected()
	if n != 1 {
		return nil
	}
	if err = tx.Commit(); err != nil {
		return repository.ErrDatabase
	}
	committed = true
	msgID, sendErr := sender.Send(ctx, to, subject, body)
	if sendErr != nil {
		_, err = db.ExecContext(ctx, `UPDATE notification_outbox SET status='FAILED',last_error=?,available_at=DATETIME(CURRENT_TIMESTAMP, '+5 minutes'),updated_at=CURRENT_TIMESTAMP WHERE id=?`, sanitize(sendErr.Error()), id)
		if err != nil {
			return repository.ErrDatabase
		}
		return nil
	}
	if _, err = db.ExecContext(ctx, `UPDATE notification_outbox SET status='SENT',provider_message_id=?,sent_at=CURRENT_TIMESTAMP,updated_at=CURRENT_TIMESTAMP WHERE id=?`, msgID, id); err != nil {
		return repository.ErrDatabase
	}
	return nil
}
func sanitize(v string) string {
	return "delivery failed"
}
