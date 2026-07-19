package notification_test

import (
	"context"
	"database/sql"
	"errors"
	"github.com/prelove/zedu/backend/internal/app/notification"
	"github.com/prelove/zedu/backend/internal/platform/database"
	"github.com/prelove/zedu/backend/internal/repository"
	"path/filepath"
	"testing"
)

type fakeSender struct {
	calls int
	fail  bool
}

func (s *fakeSender) Send(_ context.Context, _, _, _ string) (string, error) {
	s.calls++
	if s.fail {
		return "", errors.New("provider unavailable")
	}
	return "re_123", nil
}
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open("file:" + filepath.Join(t.TempDir(), "notification.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err = database.MigrateUp(db, filepath.Join("..", "..", "..", "migrations")); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
func TestClaimAndSendRecordsSuccessAndFailure(t *testing.T) {
	db := testDB(t)
	for _, c := range []struct {
		key    string
		sender *fakeSender
		want   string
	}{{"success", &fakeSender{}, "SENT"}, {"failure", &fakeSender{fail: true}, "FAILED"}} {
		if _, err := db.Exec(`INSERT INTO notification_outbox(lesson_id,event_type,recipient_email,subject,html_body,idempotency_key) VALUES (1,'LESSON_CREATED','x@example.test','s','b',?)`, c.key); err != nil {
			t.Fatal(err)
		}
		if err := notification.ClaimAndSend(context.Background(), repository.NewDB(db), c.sender); err != nil {
			t.Fatal(err)
		}
		var status string
		if err := db.QueryRow(`SELECT status FROM notification_outbox WHERE id=(SELECT max(id) FROM notification_outbox)`).Scan(&status); err != nil {
			t.Fatal(err)
		}
		if status != c.want || c.sender.calls != 1 {
			t.Fatalf("status=%s calls=%d", status, c.sender.calls)
		}
	}
}
func TestOutboxIdempotencyKeyIsUnique(t *testing.T) {
	db := testDB(t)
	_, err := db.Exec(`INSERT INTO notification_outbox(lesson_id,event_type,recipient_email,subject,html_body,idempotency_key) VALUES (1,'LESSON_CREATED','x@example.test','s','b','same')`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO notification_outbox(lesson_id,event_type,recipient_email,subject,html_body,idempotency_key) VALUES (1,'LESSON_CREATED','x@example.test','s','b','same')`); err == nil {
		t.Fatal("duplicate key must fail")
	}
}
