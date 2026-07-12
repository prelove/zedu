## ADDED Requirements

### Requirement: Explicit first-run initialization
The system MUST expose initialization only to an authenticated Owner and MUST apply the selected approved template in one transaction. Service startup MUST NOT create business template data implicitly.

#### Scenario: First initialization succeeds
- **WHEN** an Owner selects an approved template while the system is uninitialized
- **THEN** the system SHALL create the template data and initialization marker atomically, return code `0`, and write an operation log entry

#### Scenario: Duplicate initialization is idempotent
- **WHEN** an Owner repeats the same initialization request after successful initialization
- **THEN** the system SHALL not duplicate dictionary or template records and SHALL return the existing initialization result

#### Scenario: Non-Owner cannot initialize
- **WHEN** an Operator requests initialization
- **THEN** the system SHALL return HTTP 403 with code `40301` and SHALL not create template data

### Requirement: Safe template reset
The system MUST permit template reset only to an Owner and only when no student, teacher, enrollment, or assignment business records exist.

#### Scenario: Reset is rejected when business data exists
- **WHEN** an Owner requests template reset after any protected business record exists
- **THEN** the system SHALL return HTTP 422 with code `42201`, preserve all data, and write no reset marker

#### Scenario: Empty-system reset succeeds
- **WHEN** an Owner requests reset and no protected business record exists
- **THEN** the system SHALL replace only template data in one transaction and write an operation log entry
