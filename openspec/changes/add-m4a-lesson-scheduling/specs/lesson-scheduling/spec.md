## ADDED Requirements

### Requirement: Authorized operator MUST be able to create a scheduled lesson from an active enrollment assignment
The system SHALL allow `OWNER` and `OPERATOR` to create a lesson only when the target enrollment is not in a terminal state and the target assignment is currently ACTIVE. A newly created lesson MUST default to `SCHEDULED`, generate a unique `lesson_no`, and freeze the referenced student and teacher facts needed for later attendance and notification flows.

#### Scenario: Create scheduled lesson successfully
- **WHEN** an `OWNER` or `OPERATOR` submits a create-lesson request for an enrollment with status `ACTIVE` and an ACTIVE assignment, with valid scheduled time, duration, meeting type, and optional topic/note
- **THEN** the system creates a lesson with status `SCHEDULED`, stores a unique `lesson_no`, and returns the created lesson detail

#### Scenario: Reject lesson creation for terminal enrollment
- **WHEN** an `OWNER` or `OPERATOR` submits a create-lesson request for an enrollment with status `COMPLETED` or `CANCELLED`
- **THEN** the system rejects the request with a stable business validation error and does not create a lesson

#### Scenario: Reject lesson creation for inactive assignment
- **WHEN** an `OWNER` or `OPERATOR` submits a create-lesson request for an assignment that is ended or not in `ACTIVE` status
- **THEN** the system rejects the request with a stable business validation error and does not create a lesson

#### Scenario: Reject lesson write by unauthorized role
- **WHEN** a caller without `OWNER` or `OPERATOR` role attempts to create a lesson
- **THEN** the system denies access and does not create a lesson

### Requirement: Lesson scheduling input MUST satisfy the frozen validation contract
The system SHALL validate lesson scheduling input consistently on create and update. `duration_min` MUST be within the approved range, meeting-link rules MUST match meeting type, and invalid payloads MUST be rejected before persistence.

#### Scenario: Reject duration outside approved range
- **WHEN** a create or update request provides `duration_min` outside the approved M4a range
- **THEN** the system rejects the request with a validation error and does not persist the change

#### Scenario: Reject invalid meeting link for WeChat lesson
- **WHEN** a create or update request sets `meeting_type` to `WECHAT` and provides an invalid `meeting_link`
- **THEN** the system rejects the request with a validation error and does not persist the change

#### Scenario: Accept lesson without meeting link for offline meeting
- **WHEN** a create or update request sets `meeting_type` to an offline mode that does not require a link and all other fields are valid
- **THEN** the system accepts the request without requiring `meeting_link`

### Requirement: Lesson time MUST be normalized to UTC while preserving submitted timezone
The system SHALL interpret the submitted business-local time using the provided `timezone`, store `scheduled_start_at` and `scheduled_end_at` in UTC, and preserve the original timezone for later display and reminder calculations.

#### Scenario: Persist Asia/Tokyo lesson time in UTC
- **WHEN** an authorized operator creates a lesson for local time `19:00` with timezone `Asia/Tokyo`
- **THEN** the stored `scheduled_start_at` and `scheduled_end_at` are converted to the correct UTC timestamps and the lesson retains timezone `Asia/Tokyo`

#### Scenario: Return normalized lesson time in detail response
- **WHEN** a caller fetches lesson detail after creation
- **THEN** the response includes the stored UTC timestamps together with the persisted timezone field needed to reconstruct the business-local schedule

### Requirement: Only scheduled lessons MUST remain editable or cancellable
The system SHALL allow update and cancel operations only while a lesson remains in `SCHEDULED` status. `COMPLETED` and `CANCELLED` lessons MUST be treated as immutable business facts for M4a.

#### Scenario: Update scheduled lesson successfully
- **WHEN** an authorized operator updates a lesson currently in `SCHEDULED` status with a valid future schedule or note change
- **THEN** the system persists the new lesson data and keeps the lesson within the approved status model

#### Scenario: Reject update for completed lesson
- **WHEN** an authorized operator attempts to update a lesson already marked `COMPLETED`
- **THEN** the system rejects the request with a stable business validation error and leaves the lesson unchanged

#### Scenario: Cancel scheduled lesson successfully
- **WHEN** an authorized operator cancels a lesson currently in `SCHEDULED` status with a valid cancel reason
- **THEN** the system marks the lesson `CANCELLED`, preserves the cancel reason, and prevents further edits

#### Scenario: Reject cancel for already cancelled lesson
- **WHEN** an authorized operator attempts to cancel a lesson already in `CANCELLED` status
- **THEN** the system rejects the request with a stable business validation error and does not alter the stored lesson

### Requirement: Lesson list and detail MUST be queryable without introducing finance or notification side effects
The system SHALL expose lesson list and detail retrieval for authorized operators and MUST NOT create notification jobs, finance ledger entries, attendance rows, or payout facts during lesson create, update, cancel, list, or detail operations in M4a.

#### Scenario: List lessons for operations review
- **WHEN** an authorized operator requests the lesson list with approved filters such as student, teacher, status, or time range
- **THEN** the system returns matching lesson records with their scheduling and status fields

#### Scenario: Read lesson detail without side effects
- **WHEN** an authorized operator requests lesson detail
- **THEN** the system returns the lesson record and does not write notification, finance, attendance, or payout data

#### Scenario: Write lesson without downstream side effects
- **WHEN** an authorized operator successfully creates, updates, or cancels a lesson
- **THEN** the system commits only the lesson-domain write and audit record, without creating payment ledger facts, evidence records, attendance rows, or notification outbox entries
