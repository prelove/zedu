## ADDED Requirements

### Requirement: Lesson notifications MUST be durably queued without coupling external delivery to lesson writes
The system SHALL write a unique PENDING notification outbox row in the same transaction as each created or cancelled lesson for every eligible recipient. It MUST NOT call Resend before the business transaction commits.

#### Scenario: Create lesson queues recipient notifications
- **WHEN** a scheduled lesson is created and its student or parent has a non-empty email
- **THEN** the lesson, audit record, and unique PENDING outbox records commit together

#### Scenario: Recipient has no email
- **WHEN** the lesson recipient has no usable email
- **THEN** no outbox row is created and lesson creation still succeeds

### Requirement: Outbox delivery MUST be idempotent, traceable and bounded
The system SHALL send a claimed outbox record through Resend only after it is committed, record SENT or FAILED status without exposing secrets, and limit automatic/manual processing to three attempts.

#### Scenario: Successful Resend delivery
- **WHEN** a claimed PENDING outbox row is accepted by Resend
- **THEN** it becomes SENT with the provider message id

#### Scenario: Failed delivery
- **WHEN** Resend rejects or cannot accept a claimed row
- **THEN** the row becomes FAILED with a sanitized error and the lesson remains unchanged

#### Scenario: Duplicate process request
- **WHEN** two processors attempt to claim the same PENDING row
- **THEN** only one sender call is made for that row

### Requirement: Authorized operators MUST be able to inspect and replay notification failures
The system SHALL allow authenticated OWNER and OPERATOR users to list notification outbox rows and reset a FAILED row below the attempt limit to PENDING. It MUST not expose API keys or raw provider credentials.

#### Scenario: Retry a failed notification
- **WHEN** an authorized operator retries a FAILED outbox row with fewer than three attempts
- **THEN** it returns to PENDING for later processing

#### Scenario: Reject a terminal retry
- **WHEN** an authorized operator retries a SENT row or a row that reached three attempts
- **THEN** the system returns a stable validation error
