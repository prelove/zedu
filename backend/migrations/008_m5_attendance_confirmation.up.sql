CREATE TABLE attendance_outcome_type (
 code TEXT PRIMARY KEY, name TEXT NOT NULL, suggested_lesson_deducted TEXT, suggested_charge_ratio TEXT, suggested_teacher_pay_ratio TEXT, enabled INTEGER NOT NULL DEFAULT 1
);
INSERT OR IGNORE INTO attendance_outcome_type(code,name,suggested_lesson_deducted,suggested_charge_ratio,suggested_teacher_pay_ratio) VALUES
 ('ATTENDED','Attended','1','1','1'),('STUDENT_LEAVE','Student leave','0.5','0.5','1'),('STUDENT_NOSHOW','Student no-show','1','1','1'),('TEACHER_LEAVE','Teacher leave','0','0','0');
ALTER TABLE student_account_ledger RENAME TO student_account_ledger_m3;
CREATE TABLE student_account_ledger (
 id INTEGER PRIMARY KEY AUTOINCREMENT, student_id INTEGER NOT NULL REFERENCES student(id), enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
 biz_type TEXT NOT NULL CHECK(biz_type IN ('RECHARGE','VOID','LESSON_CONFIRM')), amount_delta INTEGER NOT NULL, lesson_delta REAL NOT NULL DEFAULT 0,
 balance_after INTEGER NOT NULL, lesson_balance_after REAL NOT NULL, related_payment_id INTEGER REFERENCES student_payment(id), operator_id INTEGER REFERENCES user_account(id), note TEXT, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO student_account_ledger(id,student_id,enrollment_id,biz_type,amount_delta,lesson_delta,balance_after,lesson_balance_after,related_payment_id,operator_id,note,created_at) SELECT id,student_id,enrollment_id,biz_type,amount_delta,lesson_delta,balance_after,lesson_balance_after,related_payment_id,operator_id,note,created_at FROM student_account_ledger_m3;
DROP TABLE student_account_ledger_m3;
CREATE INDEX idx_ledger_student ON student_account_ledger(student_id); CREATE INDEX idx_ledger_enrollment ON student_account_ledger(enrollment_id); CREATE INDEX idx_ledger_created ON student_account_ledger(created_at);
CREATE TABLE attendance (
 id INTEGER PRIMARY KEY AUTOINCREMENT, lesson_id INTEGER NOT NULL UNIQUE REFERENCES lesson(id), outcome_type TEXT NOT NULL REFERENCES attendance_outcome_type(code),
 suggested_lesson_deducted TEXT, suggested_charge_ratio TEXT, suggested_teacher_pay_ratio TEXT,
 actual_duration_min INTEGER, lesson_deducted TEXT NOT NULL, charge_amount INTEGER NOT NULL, teacher_pay_amount INTEGER NOT NULL, note TEXT, confirmed_by INTEGER REFERENCES user_account(id), confirmed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE lesson_finance (id INTEGER PRIMARY KEY AUTOINCREMENT, lesson_id INTEGER NOT NULL UNIQUE REFERENCES lesson(id), student_id INTEGER NOT NULL, teacher_id INTEGER NOT NULL, enrollment_id INTEGER NOT NULL, charge_amount INTEGER NOT NULL, teacher_pay_amount INTEGER NOT NULL, gross_profit_amount INTEGER NOT NULL, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE teacher_account_ledger (id INTEGER PRIMARY KEY AUTOINCREMENT, teacher_id INTEGER NOT NULL REFERENCES teacher(id), lesson_id INTEGER NOT NULL UNIQUE REFERENCES lesson(id), amount_delta INTEGER NOT NULL, balance_after INTEGER NOT NULL, operator_id INTEGER REFERENCES user_account(id), note TEXT, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
