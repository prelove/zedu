-- 004: Student level event history (PRD §5.5)
-- Records each level transition for a student's enrollment. The enrollment's
-- current_level_id is NOT overwritten on level change; instead an event row is
-- written so the full level history is queryable. This is the minimal schema
-- addition needed to satisfy the frozen course-selection history rule that
-- migration 003 alone cannot express.

CREATE TABLE student_level_event (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  student_id INTEGER NOT NULL REFERENCES student(id),
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  from_level_id INTEGER REFERENCES course_level(id),
  to_level_id INTEGER NOT NULL REFERENCES course_level(id),
  event_type TEXT NOT NULL DEFAULT 'MANUAL' CHECK(event_type IN ('ASSESSMENT','EXAM_PASS','HOURS_REACHED','AGE_REACHED','MANUAL')),
  event_date TEXT NOT NULL,
  evidence_note TEXT,
  operator_id INTEGER NOT NULL REFERENCES user_account(id),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_level_event_student ON student_level_event(student_id);
CREATE INDEX idx_level_event_enrollment ON student_level_event(enrollment_id);
