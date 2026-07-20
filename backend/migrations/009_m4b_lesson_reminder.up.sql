-- 009: Add LESSON_REMINDER event type to notification_outbox.
-- SQLite cannot ALTER a CHECK constraint in place, so we recreate the table
-- with the expanded constraint and copy all existing data.

CREATE TABLE notification_outbox_new (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  lesson_id INTEGER NOT NULL REFERENCES lesson(id),
  event_type TEXT NOT NULL CHECK(event_type IN ('LESSON_CREATED','LESSON_CANCELLED','LESSON_REMINDER')),
  recipient_email TEXT NOT NULL,
  locale TEXT NOT NULL DEFAULT 'ja-JP',
  subject TEXT NOT NULL,
  html_body TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'PENDING' CHECK(status IN ('PENDING','PROCESSING','SENT','FAILED')),
  attempts INTEGER NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 3),
  idempotency_key TEXT NOT NULL UNIQUE,
  provider_message_id TEXT,
  last_error TEXT,
  available_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  locked_at DATETIME,
  sent_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO notification_outbox_new(id,lesson_id,event_type,recipient_email,locale,subject,html_body,status,attempts,idempotency_key,provider_message_id,last_error,available_at,locked_at,sent_at,created_at,updated_at)
SELECT id,lesson_id,event_type,recipient_email,locale,subject,html_body,status,attempts,idempotency_key,provider_message_id,last_error,available_at,locked_at,sent_at,created_at,updated_at FROM notification_outbox;

DROP TABLE notification_outbox;
ALTER TABLE notification_outbox_new RENAME TO notification_outbox;

CREATE INDEX idx_notification_outbox_status ON notification_outbox(status, available_at);
CREATE INDEX idx_notification_outbox_lesson ON notification_outbox(lesson_id);
