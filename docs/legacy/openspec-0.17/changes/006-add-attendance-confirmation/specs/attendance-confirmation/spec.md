## ADDED Requirements

### Requirement: 出勤分类建议值计算
系统必须能根据出勤分类和课程/师生安排的默认值，计算出建议扣课时、
建议收费、建议课酬三个值。

#### Scenario: 正常上课分类
- **WHEN** 出勤分类为NORMAL，enrollment.charge_per_lesson_amount=1000，
  assignment.rate_amount=1500
- **THEN** 建议扣课时=1，建议收费=1000，建议课酬=1500

#### Scenario: 学生提前请假分类
- **WHEN** 出勤分类为STUDENT_LEAVE_EARLY
- **THEN** 建议扣课时=0，建议收费=0，建议课酬=0

#### Scenario: 已改期分类
- **WHEN** 出勤分类为RESCHEDULED
- **THEN** 建议扣课时=0，建议收费=0，建议课酬=0，但仍生成attendance
  记录以关闭该课次的挂起状态

#### Scenario: 其他分类无建议
- **WHEN** 出勤分类为OTHER
- **THEN** 三个建议值均为null，前端不自动填充

### Requirement: 课后确认事务原子性
系统必须保证课后确认涉及的全部写入在单一事务内完成。

#### Scenario: 正常提交
- **WHEN** Operator提交合法的课后确认请求
- **THEN** attendance/student_account_ledger/teacher_account_ledger/lesson_finance
  四张表的相关记录同时写入成功，且lesson.status变为COMPLETED

#### Scenario: 事务中途失败
- **WHEN** 写入teacher_account_ledger时发生错误（模拟）
- **THEN** 此前已尝试写入的attendance和student_account_ledger记录必须
  连同回滚，数据库中不应查到本次操作产生的任何记录

### Requirement: 课次确认的并发与重复保护
系统必须防止同一课次被重复确认，无论是误操作还是并发请求。

#### Scenario: 已确认课次拒绝再次确认
- **WHEN** 对status已经是COMPLETED的课次再次调用确认接口
- **THEN** 返回42201，不产生任何新的attendance记录

#### Scenario: 并发提交只有一个成功
- **WHEN** 两个请求几乎同时对同一课次调用确认接口
- **THEN** 数据库的attendance.lesson_id唯一约束确保只有一个事务
  成功提交，另一个因约束冲突而完整回滚

### Requirement: 建议值与实际值分离存储
attendance表必须同时保存"建议值快照"和"实际值"两组独立字段。

#### Scenario: 实际值覆盖建议值
- **WHEN** 系统建议扣课时为1，但Operator手动改为0.5提交
- **THEN** attendance.suggested_deduct_lessons保存为1（审计用），
  attendance.lesson_deducted保存为0.5（真正生效并用于后续流水计算的值）

#### Scenario: 建议值与实际值可完全不同
- **WHEN** 出勤分类为NORMAL(建议扣课时1)，但Operator因某种约定改为
  扣0课时
- **THEN** 两组字段都正确落库且不互相覆盖，系统不因"实际值偏离建议值
  较大"而产生任何额外拦截或警告（人工判断优先于系统建议）
