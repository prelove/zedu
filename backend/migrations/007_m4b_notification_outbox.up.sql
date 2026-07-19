CREATE TABLE notification_outbox (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  lesson_id INTEGER NOT NULL REFERENCES lesson(id),
  event_type TEXT NOT NULL CHECK(event_type IN ('LESSON_CREATED','LESSON_CANCELLED')),
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
CREATE INDEX idx_notification_outbox_status ON notification_outbox(status, available_at);
CREATE INDEX idx_notification_outbox_lesson ON notification_outbox(lesson_id);
