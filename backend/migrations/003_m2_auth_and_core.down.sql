-- Rollback: drop all M2 tables in reverse dependency order.
-- Only drops objects created by 003_m2_auth_and_core.up.sql.

DROP TABLE IF EXISTS student_teacher_assignment;
DROP TABLE IF EXISTS student_course_enrollment;
DROP TABLE IF EXISTS teacher_capability;
DROP TABLE IF EXISTS teacher_availability;
DROP TABLE IF EXISTS teacher;
DROP TABLE IF EXISTS parent;
DROP TABLE IF EXISTS student;
DROP INDEX IF EXISTS idx_student_email_unique;
DROP TABLE IF EXISTS skill_tag;
DROP TABLE IF EXISTS course_level;
DROP TABLE IF EXISTS course_track;
DROP TABLE IF EXISTS course_domain;
DROP TABLE IF EXISTS operation_log;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS refresh_session;
DROP TABLE IF EXISTS user_account;
