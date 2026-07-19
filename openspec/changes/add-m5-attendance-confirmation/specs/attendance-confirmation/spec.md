## ADDED Requirements

### Requirement: Post-lesson confirmation MUST create an immutable attendance and financial fact atomically
The system SHALL allow an OWNER or OPERATOR to confirm a SCHEDULED lesson once. It MUST write attendance, financial snapshots and applicable ledgers in one transaction, then mark the lesson COMPLETED.

#### Scenario: Confirm a lesson
- **WHEN** an authorized operator submits a valid attendance outcome and actual values for a SCHEDULED lesson
- **THEN** attendance, immutable financial facts and lesson completion commit together

#### Scenario: Confirmation failure
- **WHEN** any persistence operation fails during confirmation
- **THEN** no attendance, ledger, finance fact or lesson status change remains

#### Scenario: Duplicate confirmation
- **WHEN** the lesson was already confirmed or two requests race to confirm it
- **THEN** exactly one confirmation may succeed and the other returns a stable validation/conflict response

### Requirement: Suggested values MUST be auditable but never override actual values
The system SHALL snapshot the selected outcome's suggested values alongside the operator's submitted actual values. Subsequent dictionary changes MUST NOT alter attendance history.

#### Scenario: Later dictionary change
- **WHEN** an outcome type's suggested values are changed after a lesson confirmation
- **THEN** the existing attendance retains both original suggested and actual snapshots
