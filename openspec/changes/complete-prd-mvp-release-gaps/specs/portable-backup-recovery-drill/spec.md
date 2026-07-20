## ADDED Requirements

### Requirement: Owner backup MUST contain database, attachments, configuration summary and integrity manifest
The system SHALL let only an OWNER create a staged local backup package containing a verified SQLite snapshot, the configured attachment upload tree, a non-sensitive configuration summary, and a SHA-256 manifest for every included file. Secrets MUST NOT be included.

#### Scenario: Successful package backup
- **WHEN** an Owner creates a backup with a database and uploaded payment evidence
- **THEN** the published package SHALL contain the SQLite snapshot, the evidence file, a manifest with matching SHA-256 values, and a BACKUP_CREATE audit record.

#### Scenario: Package creation fails
- **WHEN** snapshot, attachment copy, manifest generation, verification, or audit persistence fails
- **THEN** no published backup package or success audit record SHALL remain.

### Requirement: Backup recovery MUST be verified without an HTTP restore endpoint
The system SHALL provide a local controlled verification path that validates the manifest and restores a package only into a new temporary target. It MUST NOT expose an HTTP restore route or overwrite the active database.

#### Scenario: Recovery drill succeeds
- **WHEN** a valid backup package is verified into a fresh temporary location
- **THEN** manifest hashes, database readability, and attachment files SHALL match the package without changing the active instance.

#### Scenario: Tampered package
- **WHEN** a package file does not match its manifest hash
- **THEN** verification SHALL fail before any restored target is published.
