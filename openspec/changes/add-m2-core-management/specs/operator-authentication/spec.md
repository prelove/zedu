## ADDED Requirements

### Requirement: Operator authentication and session lifecycle
The system MUST authenticate only ACTIVE `OWNER` and `OPERATOR` accounts with a username and password, issue a 60-minute access token and a 14-day rotating refresh session, and never expose password hashes or refresh token values in JSON, logs, or audit details.

#### Scenario: Successful login
- **WHEN** an ACTIVE Owner or Operator submits valid credentials
- **THEN** the system SHALL return code `0`, current role and a short-lived access token, set a secure HttpOnly refresh cookie, and record the successful login without secrets

#### Scenario: Failed credentials do not disclose account existence
- **WHEN** a login username is unknown or its password is invalid
- **THEN** the system SHALL return HTTP 401 with code `40102`, increment the failure count only for an existing account, and use the same client message for both cases

#### Scenario: Lockout after repeated failure
- **WHEN** an account reaches five consecutive failed login attempts
- **THEN** the system SHALL lock the account for 15 minutes and reject further login attempts with HTTP 401 and code `40103`

#### Scenario: Refresh token rotation
- **WHEN** a valid, unrevoked refresh cookie is presented to the refresh endpoint
- **THEN** the system SHALL revoke that refresh session, issue exactly one replacement refresh session and access token, and reject reuse of the old refresh token

#### Scenario: Logout revokes the session
- **WHEN** an authenticated account logs out
- **THEN** the system SHALL revoke the presented refresh session, clear the refresh cookie, and reject later refresh attempts with code `40101`

### Requirement: Role-based access control
The system MUST require authentication for all M2 business APIs. `OWNER` SHALL include all `OPERATOR` permissions; only `OWNER` may manage Operator accounts or reset an initialized template.

#### Scenario: Unauthenticated business request
- **WHEN** a request without a valid access token calls an M2 business endpoint
- **THEN** the system SHALL return HTTP 401 with code `40101` and no resource data

#### Scenario: Operator is denied Owner-only action
- **WHEN** an Operator requests an Owner-only account-management or template-reset action
- **THEN** the system SHALL return HTTP 403 with code `40301` and SHALL not modify data

#### Scenario: Disabled account loses access
- **WHEN** an Owner disables an Operator account
- **THEN** the system SHALL revoke that account's refresh sessions and reject subsequent authenticated requests from it with code `40101`

### Requirement: Authentication observability
The system MUST attach the request ID to authentication logs and MUST redact passwords, authorization headers, access tokens, refresh tokens and password hashes.

#### Scenario: Login request logging
- **WHEN** a login request succeeds or fails
- **THEN** the emitted structured log SHALL include request ID and outcome but SHALL not contain the submitted password, token, or password hash
