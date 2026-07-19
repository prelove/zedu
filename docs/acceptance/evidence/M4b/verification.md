# M4b Resend notification outbox-lite verification

- `007_m4b_notification_outbox` has an up/down/up migration test and unique idempotency key.
- Lesson create/cancel queues student/parent email recipients in the same transaction; queue failure rolls back the lesson write.
- Resend delivery occurs only after a row is committed and claimed; fake sender tests cover SENT and FAILED outcomes.
- Owner/Operator can list, process and retry eligible failed rows; no API key is returned or logged.
- Real Resend smoke sending remains an explicit post-deploy/UAT operation and is not executed from automated tests.
