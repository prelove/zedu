## MODIFIED Requirements

### Requirement: Operators MUST have a read-only operational dashboard
The system SHALL show today lesson count, pending lesson confirmation count, renewal-needed student count, teacher payable aggregate, and failed notification count without modifying business facts.

#### Scenario: Dashboard query
- **WHEN** an authenticated operator opens the dashboard
- **THEN** it receives current read-only operational counts and no ledger, lesson, outbox, or audit record is written.

#### Scenario: Empty dashboard
- **WHEN** no operational records exist
- **THEN** each dashboard count SHALL be zero.

### Requirement: Only Owner MUST trigger an audited local backup
The system SHALL let an Owner create a verified local backup package containing the SQLite snapshot, attachment upload tree, non-sensitive configuration summary, and integrity manifest, and record the operation. It MUST NOT expose an HTTP restore action in MVP.

#### Scenario: Owner backup
- **WHEN** an Owner requests a backup with valid configured directories
- **THEN** a verified package and audit record are created.

#### Scenario: Non-Owner backup
- **WHEN** an Operator requests a backup
- **THEN** the system SHALL return HTTP 403 with code 40301 and create no package or audit record.
