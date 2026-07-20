## ADDED Requirements

### Requirement: Scheduled lessons MUST produce idempotent pre-lesson reminders
The system SHALL enqueue a `LESSON_REMINDER` outbox record for each eligible recipient of a SCHEDULED lesson in the configured fixed reminder window. The idempotency key MUST prevent duplicate reminders for the same lesson, event, and recipient.

#### Scenario: Reminder scan finds an eligible lesson
- **WHEN** the controlled notification runner scans a SCHEDULED lesson in the fixed reminder window with an email recipient
- **THEN** it SHALL create at most one pending reminder row for that lesson and recipient.

#### Scenario: Reminder scan repeats
- **WHEN** the runner scans the same eligible lesson more than once
- **THEN** it SHALL not create a duplicate reminder row.

### Requirement: Failed notification delivery MUST retry in a bounded and traceable manner
The controlled notification runner SHALL retry FAILED rows only after their `available_at` time and while attempts are below three. A send failure MUST leave the lesson and its business facts unchanged and MUST record a sanitized failure state.

#### Scenario: Retryable failure
- **WHEN** a retryable notification send fails before the third attempt
- **THEN** the outbox row SHALL remain FAILED with a future availability time and the lesson SHALL remain unchanged.

#### Scenario: Retry limit reached
- **WHEN** a notification has reached three attempts
- **THEN** the automatic runner SHALL not attempt another send, while the inspection UI continues to display its failure.

### Requirement: Notification operations MUST keep credentials and recipients safe
The system MUST NOT return or log Resend API keys, Authorization values, refresh tokens, or password material while processing notifications.

#### Scenario: Sender configuration failure
- **WHEN** sender configuration is absent or delivery fails
- **THEN** the response and stored error SHALL use a stable sanitized message without secret values.
