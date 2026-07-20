## ADDED Requirements

### Requirement: Operators MUST be able to read teacher payable facts without settlement actions
The system SHALL provide authenticated OWNER and OPERATOR users a paginated teacher payable summary and teacher-specific lesson payable entries derived solely from `teacher_account_ledger`. It MUST NOT expose payout, settlement, adjustment, export, or write actions in this capability.

#### Scenario: Read payable summary
- **WHEN** an authenticated operator requests the teacher payable summary after one confirmed lesson with a positive teacher pay
- **THEN** the response SHALL show the teacher and the aggregated unpaid payable amount derived from immutable ledger entries.

#### Scenario: No settlement surface
- **WHEN** a user inspects the payable API and UI routes
- **THEN** no payout, settlement, adjustment, or teacher-ledger write action SHALL be available.

### Requirement: Payable reads MUST be permission-safe and financially read-only
The system SHALL return 40101 to unauthenticated requests and SHALL not create or alter ledger, lesson, audit, or payment facts while serving a payable read.

#### Scenario: Unauthenticated payable request
- **WHEN** an unauthenticated caller requests teacher payable data
- **THEN** the system SHALL return HTTP 401 with code 40101 and no ledger data.
