-- ========== M2: Auth, System, and Core Master Data ==========
-- All times stored as UTC (SQLite DATETIME defaults to UTC).
-- All FKs enforced via PRAGMA foreign_keys=ON (set in database.Open).

-- ========== Auth & Session ==========

CREATE TABLE user_account (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK(role IN ('OWNER','OPERATOR')),
  display_name TEXT NOT NULL,
  email TEXT,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','DISABLED')),
  last_login_at DATETIME,
  login_fail_count INTEGER NOT NULL DEFAULT 0,
  locked_until DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);

CREATE TABLE refresh_session (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL REFERENCES user_account(id),
  token_hash TEXT NOT NULL UNIQUE,
  expires_at DATETIME NOT NULL,
  revoked_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_refresh_session_user ON refresh_session(user_id);
CREATE INDEX idx_refresh_session_expires ON refresh_session(expires_at);

-- ========== System Settings (key-value) ==========

CREATE TABLE system_settings (
  config_key TEXT PRIMARY KEY,
  config_value TEXT NOT NULL,
  description TEXT,
  updated_by INTEGER REFERENCES user_account(id),
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ========== Operation Log (audit) ==========

CREATE TABLE operation_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  operator_id INTEGER REFERENCES user_account(id),
  operator_name TEXT,
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id INTEGER,
  detail_json TEXT,
  ip_addr TEXT,
  request_id TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_oplog_target ON operation_log(target_type, target_id);
CREATE INDEX idx_oplog_created ON operation_log(created_at);
CREATE INDEX idx_oplog_request_id ON operation_log(request_id);

-- ========== Course System (configurable dictionary) ==========

CREATE TABLE course_domain (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  code TEXT NOT NULL UNIQUE,
  type TEXT NOT NULL CHECK(type IN ('LANGUAGE','K12','SPORT','ART','ACADEMIC','CERTIFICATE','OTHER')),
  sort_order INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE course_track (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  domain_id INTEGER NOT NULL REFERENCES course_domain(id),
  name TEXT NOT NULL,
  code TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(domain_id, code)
);
CREATE INDEX idx_track_domain ON course_track(domain_id);

CREATE TABLE course_level (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  track_id INTEGER NOT NULL REFERENCES course_track(id),
  name TEXT NOT NULL,
  code TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  min_age INTEGER,
  max_age INTEGER,
  min_lesson_hours REAL,
  recommended_lesson_hours REAL,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(track_id, code)
);
CREATE INDEX idx_level_track ON course_level(track_id);

CREATE TABLE skill_tag (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  domain_id INTEGER NOT NULL REFERENCES course_domain(id),
  name TEXT NOT NULL,
  code TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(domain_id, code)
);

-- ========== People ==========

CREATE TABLE student (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  name_local TEXT,
  email TEXT,
  phone TEXT,
  nationality TEXT,
  timezone TEXT NOT NULL DEFAULT 'Asia/Tokyo',
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','ENDED')),
  source_channel TEXT,
  note TEXT,
  extra_json TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX idx_student_status ON student(status);
CREATE INDEX idx_student_name ON student(name);
-- Partial unique index: non-NULL emails are globally unique.
-- Soft-deleted records retain their email (deleted_at does not release uniqueness).
CREATE UNIQUE INDEX idx_student_email_unique ON student(email) WHERE email IS NOT NULL;

CREATE TABLE parent (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  student_id INTEGER NOT NULL REFERENCES student(id),
  name TEXT NOT NULL,
  email TEXT,
  phone TEXT,
  relationship TEXT CHECK(relationship IN ('FATHER','MOTHER','OTHER')),
  is_primary INTEGER NOT NULL DEFAULT 0,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_parent_student ON parent(student_id);

CREATE TABLE teacher (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  name_local TEXT,
  email TEXT,
  phone TEXT,
  bio TEXT,
  default_rate_amount INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','ENDED')),
  note TEXT,
  extra_json TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX idx_teacher_email ON teacher(email);
CREATE INDEX idx_teacher_status ON teacher(status);

CREATE TABLE teacher_availability (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  weekday INTEGER NOT NULL CHECK(weekday BETWEEN 0 AND 6),
  start_time TEXT NOT NULL,
  end_time TEXT NOT NULL,
  effective_from DATE,
  effective_to DATE,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_avail_teacher ON teacher_availability(teacher_id);

CREATE TABLE teacher_capability (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  domain_id INTEGER NOT NULL REFERENCES course_domain(id),
  track_id INTEGER NOT NULL REFERENCES course_track(id),
  level_id INTEGER NOT NULL REFERENCES course_level(id),
  skill_tag_codes TEXT,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','ENDED')),
  verified INTEGER NOT NULL DEFAULT 0,
  effective_from DATE,
  effective_to DATE,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(teacher_id, track_id, level_id)
);
CREATE INDEX idx_cap_teacher ON teacher_capability(teacher_id);

-- ========== Enrollment & Assignment ==========

CREATE TABLE student_course_enrollment (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  student_id INTEGER NOT NULL REFERENCES student(id),
  domain_id INTEGER NOT NULL REFERENCES course_domain(id),
  track_id INTEGER NOT NULL REFERENCES course_track(id),
  current_level_id INTEGER REFERENCES course_level(id),
  target_level_id INTEGER REFERENCES course_level(id),
  enrollment_type TEXT NOT NULL DEFAULT 'ONE_TO_ONE' CHECK(enrollment_type IN ('ONE_TO_ONE','GROUP','TRIAL')),
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','COMPLETED','CANCELLED')),
  charge_per_lesson_amount INTEGER NOT NULL DEFAULT 0,
  lesson_balance REAL NOT NULL DEFAULT 0,
  balance_amount INTEGER NOT NULL DEFAULT 0,
  started_at DATE,
  ended_at DATE,
  note TEXT,
  extra_json TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX idx_enrollment_student ON student_course_enrollment(student_id);
CREATE INDEX idx_enrollment_status ON student_course_enrollment(status);

CREATE TABLE student_teacher_assignment (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  student_id INTEGER NOT NULL REFERENCES student(id),
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  role_type TEXT NOT NULL DEFAULT 'MAIN' CHECK(role_type IN ('MAIN','SUBSTITUTE','ASSISTANT')),
  rate_amount INTEGER,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','ENDED')),
  start_date DATE NOT NULL,
  end_date DATE,
  reason TEXT,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_assign_enrollment ON student_teacher_assignment(enrollment_id);
CREATE INDEX idx_assign_teacher ON student_teacher_assignment(teacher_id);
CREATE INDEX idx_assign_status ON student_teacher_assignment(status);
-- Each enrollment may have at most one ACTIVE assignment at any time.
CREATE UNIQUE INDEX idx_assign_enrollment_active ON student_teacher_assignment(enrollment_id) WHERE status = 'ACTIVE';
