-- 004 down: remove student_level_event table and its indexes.
DROP INDEX IF EXISTS idx_level_event_enrollment;
DROP INDEX IF EXISTS idx_level_event_student;
DROP TABLE IF EXISTS student_level_event;
