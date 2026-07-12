## ADDED Requirements

### Requirement: Course dictionary hierarchy
The system MUST allow authorized users to maintain course domain, track, level and capability-tag dictionaries while preserving referential integrity for existing enrollment and capability records.

#### Scenario: Dictionary code is unique at its hierarchy level
- **WHEN** an authorized user creates a dictionary item using a code already used at the same hierarchy level
- **THEN** the system SHALL reject the write with HTTP 409 and code `40901`

#### Scenario: Referenced dictionary item is not destructively deleted
- **WHEN** an authorized user attempts to remove a dictionary item referenced by a teacher capability or enrollment
- **THEN** the system SHALL reject the operation with HTTP 422 and code `42201` and preserve existing relations

### Requirement: Student enrollment
The system MUST allow an authorized user to create and update an enrollment only for an existing active student and a valid course hierarchy. An enrollment MUST retain its status and course selection history; M2 MUST NOT create lessons.

#### Scenario: Create enrollment without teacher assignment
- **WHEN** an authorized user creates a valid enrollment for an active student without selecting a teacher
- **THEN** the system SHALL persist the enrollment with no assignment and return code `0`

#### Scenario: Enrollment for missing or ended student is rejected
- **WHEN** an authorized user creates an enrollment for a nonexistent or ended student
- **THEN** the system SHALL return HTTP 422 with code `42201` and SHALL not create an enrollment

### Requirement: Teacher-student assignment lifecycle
The system MUST allow an authorized user to add, end or replace assignments under an enrollment. At most one assignment MAY be ACTIVE for an enrollment at a time; replacement MUST end the former assignment and activate the replacement in one transaction.

#### Scenario: Create active assignment
- **WHEN** an authorized user assigns an active teacher to an active enrollment
- **THEN** the system SHALL create one ACTIVE assignment and return code `0`

#### Scenario: Replace active assignment atomically
- **WHEN** an authorized user replaces an enrollment's active assignment
- **THEN** the system SHALL end the old assignment, create the new ACTIVE assignment, and write audit records in one transaction

#### Scenario: Assignment does not create a lesson or notification
- **WHEN** an assignment is created, ended or replaced
- **THEN** the system SHALL not create a lesson, attendance, payment, notification, payout or email record

### Requirement: Enrollment and assignment writes are auditable
The system MUST record successful enrollment and assignment writes with actor and request ID in `operation_log` within the same transaction.

#### Scenario: Assignment transaction failure rolls back all changes
- **WHEN** an assignment replacement fails after ending the old assignment but before the new assignment commits
- **THEN** the system SHALL roll back the old assignment end, create no new assignment, and create no successful audit entry
