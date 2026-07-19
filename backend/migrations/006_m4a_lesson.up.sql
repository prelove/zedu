-- 006: M4a only creates the scheduling fact. Attendance, notifications,
-- financial settlement and conflict-resolution tables remain later changes.

CREATE TABLE lesson (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  lesson_no TEXT NOT NULL UNIQUE,
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  assignment_id INTEGER NOT NULL REFERENCES student_teacher_assignment(id),
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  student_id INTEGER NOT NULL REFERENCES student(id),
  scheduled_start_at DATETIME NOT NULL,
  scheduled_end_at DATETIME NOT NULL,
  duration_min INTEGER NOT NULL CHECK(duration_min BETWEEN 10 AND 480),
  timezone TEXT NOT NULL,
  meeting_type TEXT NOT NULL,
  meeting_link TEXT,
  lesson_topic TEXT,
  note TEXT,
  status TEXT NOT NULL DEFAULT 'SCHEDULED' CHECK(status IN ('SCHEDULED', 'COMPLETED', 'CANCELLED')),
  cancel_reason TEXT,
  created_by INTEGER REFERENCES user_account(id),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_lesson_student ON lesson(student_id);
CREATE INDEX idx_lesson_teacher ON lesson(teacher_id);
CREATE INDEX idx_lesson_scheduled ON lesson(scheduled_start_at);
CREATE INDEX idx_lesson_status ON lesson(status);
