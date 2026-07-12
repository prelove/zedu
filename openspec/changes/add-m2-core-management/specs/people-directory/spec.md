## ADDED Requirements

### Requirement: Student directory and unique email
The system MUST allow an authenticated Owner or Operator to create, view, update and change status of student records. `student.email` MAY be empty; a non-empty value MUST be globally unique, including against soft-deleted student records.

#### Scenario: Student without email is saved
- **WHEN** an authorized user creates a student with a valid name and no email
- **THEN** the system SHALL create the student and return code `0`

#### Scenario: Duplicate email creation is rejected
- **WHEN** an authorized user creates a student using a non-empty email already held by any student record
- **THEN** the system SHALL return HTTP 409 with code `40901`, create no student, and provide no bypass action

#### Scenario: Duplicate email update is rejected atomically
- **WHEN** an authorized user changes a student's email to another student's non-empty email
- **THEN** the system SHALL return HTTP 409 with code `40901` and preserve the original student record

#### Scenario: Concurrent duplicate create has one winner
- **WHEN** two requests concurrently create students with the same non-empty email
- **THEN** exactly one request SHALL succeed and every other request SHALL receive code `40901`

### Requirement: Parent contacts are scoped to a student
The system MUST allow multiple parent contacts for a student and MUST prevent a parent record from being read or modified through a different student's resource path.

#### Scenario: Create parent contact
- **WHEN** an authorized user creates a parent contact under an existing student
- **THEN** the system SHALL persist the contact with that student ID and return code `0`

#### Scenario: Cross-student parent access is denied
- **WHEN** an authorized user addresses a parent contact through a student ID that does not own it
- **THEN** the system SHALL return HTTP 404 with code `40401` and SHALL not disclose the contact

### Requirement: Teacher capability and availability history
The system MUST allow authorized users to maintain teachers, capabilities and availability. Capability identity MUST be unique by `(teacher_id, track_id, level_id)`; ending a capability MUST retain its historical record by setting an effective end time.

#### Scenario: Duplicate teacher capability is rejected
- **WHEN** an authorized user creates a capability with an existing teacher, track and level combination
- **THEN** the system SHALL return HTTP 409 with code `40901` and SHALL not create a duplicate row

#### Scenario: End capability preserves history
- **WHEN** an authorized user ends an active teacher capability
- **THEN** the system SHALL set its effective end time without deleting the capability record

### Requirement: Directory writes are auditable
The system MUST write successful create, update, status-change, capability-end and availability-change actions to `operation_log` in the same transaction as the business write.

#### Scenario: Failed validation produces no audit fact
- **WHEN** a directory write is rejected for validation, conflict or authorization failure
- **THEN** the system SHALL not create a successful operation log entry or partial business record
