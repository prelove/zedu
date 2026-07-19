-- 005 down: remove only M3 default configuration rows.
-- Migration down is for empty development/test databases only; production
-- financial data must never be removed by a destructive migration rollback.

DELETE FROM system_settings
WHERE config_key IN ('base_currency', 'base_currency_locked');

DROP INDEX IF EXISTS idx_ledger_created;
DROP INDEX IF EXISTS idx_ledger_enrollment;
DROP INDEX IF EXISTS idx_ledger_student;
DROP TABLE IF EXISTS student_account_ledger;
DROP INDEX IF EXISTS idx_attachment_payment;
DROP TABLE IF EXISTS payment_attachment;
DROP INDEX IF EXISTS idx_payment_status;
DROP INDEX IF EXISTS idx_payment_enrollment;
DROP INDEX IF EXISTS idx_payment_student;
DROP TABLE IF EXISTS student_payment;
DROP TABLE IF EXISTS payment_method;
