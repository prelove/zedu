## ADDED Requirements

### Requirement: Operators MUST have a read-only operational dashboard
The system SHALL show pending lesson confirmations and failed notification counts without modifying business facts.

#### Scenario: Dashboard query
- **WHEN** an authenticated operator opens the dashboard
- **THEN** it receives current read-only operational counts

### Requirement: Only Owner MUST trigger an audited local backup
The system SHALL let an Owner create a verified SQLite backup in the configured local backup directory and record the operation. It MUST NOT expose an HTTP restore action in MVP.

#### Scenario: Owner backup
- **WHEN** an Owner requests a backup with a valid configured directory
- **THEN** a backup artifact and audit record are created
