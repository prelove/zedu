-- 005: M3 payment, student ledger and financial configuration (PRD §7).
-- The values are data defaults, not database enums: Owner may maintain the
-- payment-method dictionary while historical payments retain stable codes.

CREATE TABLE payment_method (
  code TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE student_payment (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  payment_no TEXT NOT NULL UNIQUE,
  student_id INTEGER NOT NULL REFERENCES student(id),
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  original_amount TEXT NOT NULL,
  original_currency TEXT NOT NULL,
  fx_rate_to_base TEXT NOT NULL DEFAULT '1',
  amount_base INTEGER NOT NULL,
  lessons_added INTEGER NOT NULL DEFAULT 0,
  package_name TEXT,
  payment_method_code TEXT NOT NULL REFERENCES payment_method(code),
  payment_method_name TEXT NOT NULL,
  paid_at DATETIME NOT NULL,
  operator_id INTEGER REFERENCES user_account(id),
  status TEXT NOT NULL DEFAULT 'CONFIRMED' CHECK(status IN ('CONFIRMED', 'VOIDED')),
  voided_at DATETIME,
  void_reason TEXT,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_payment_student ON student_payment(student_id);
CREATE INDEX idx_payment_enrollment ON student_payment(enrollment_id);
CREATE INDEX idx_payment_status ON student_payment(status);

CREATE TABLE payment_attachment (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  payment_id INTEGER NOT NULL REFERENCES student_payment(id),
  file_name TEXT NOT NULL,
  file_path TEXT NOT NULL,
  file_type TEXT,
  file_size INTEGER,
  uploaded_by INTEGER REFERENCES user_account(id),
  uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_attachment_payment ON payment_attachment(payment_id);

CREATE TABLE student_account_ledger (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  student_id INTEGER NOT NULL REFERENCES student(id),
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  biz_type TEXT NOT NULL CHECK(biz_type IN ('RECHARGE', 'VOID')),
  amount_delta INTEGER NOT NULL,
  lesson_delta INTEGER NOT NULL DEFAULT 0,
  balance_after INTEGER NOT NULL,
  lesson_balance_after INTEGER NOT NULL,
  related_payment_id INTEGER REFERENCES student_payment(id),
  operator_id INTEGER REFERENCES user_account(id),
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_ledger_student ON student_account_ledger(student_id);
CREATE INDEX idx_ledger_enrollment ON student_account_ledger(enrollment_id);
CREATE INDEX idx_ledger_created ON student_account_ledger(created_at);

INSERT OR IGNORE INTO system_settings (config_key, config_value, description)
VALUES
  ('base_currency', 'JPY', 'system base currency before the first financial record'),
  ('base_currency_locked', 'false', 'whether the system base currency is locked by financial facts');

INSERT OR IGNORE INTO payment_method (code, name, sort_order, enabled)
VALUES
  ('WECHAT', '微信支付', 10, 1),
  ('ALIPAY', '支付宝', 20, 1),
  ('PAYPAY', 'PayPay', 30, 1),
  ('BANK', '银行转账', 40, 1),
  ('CASH', '现金', 50, 1),
  ('OTHER', '其他', 99, 1);
