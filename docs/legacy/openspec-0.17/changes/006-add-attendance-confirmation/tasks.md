## 1. 出勤分类字典与建议值计算

- [ ] 1.1 编写失败测试：种子数据应包含9条attendance_outcome_type记录，
      NORMAL的三个建议值分别为1/1.0/1.0，RESCHEDULED的三个建议值均为0
      文件：backend/internal/lesson/outcome_test.go
- [ ] 1.2 实现读取接口使测试通过（种子数据已在001准备好）
      文件：backend/internal/lesson/outcome_service.go
- [ ] 1.3 编写失败测试：给定分类code+enrollment默认值+assignment默认值，
      calculateSuggestion()函数应返回正确的三个建议值（覆盖NORMAL/
      STUDENT_LEAVE_EARLY/RESCHEDULED/OTHER四种情况）
      文件：backend/internal/lesson/suggestion_test.go
- [ ] 1.4 实现calculateSuggestion()纯函数使测试通过
      文件：backend/internal/lesson/suggestion.go
- [ ] 1.5 提交：git commit -m "feat(attendance): outcome type suggestion calc covering four cases"

## 2. attendance表与课后确认事务（核心，逐步验证不可跳步）

- [ ] 2.1 编写失败测试：POST /lessons/{id}/confirm 提交NORMAL分类+完整实际值，
      应返回200且attendance表新增一条记录，suggested_*和实际字段都正确
      文件：backend/internal/lesson/confirm_test.go
- [ ] 2.2 实现最小代码：仅写attendance，暂不涉及ledger（先让这一步测试通过）
      文件：backend/internal/lesson/model.go(attendance struct), service.go, handler.go
- [ ] 2.3 编写失败测试：确认后student_account_ledger应新增一条LESSON_DEDUCT记录，
      enrollment.balance_amount和lesson_balance正确扣减
- [ ] 2.4 实现代码使2.3测试通过，注意仍在同一tx对象上操作
- [ ] 2.5 编写失败测试：确认后teacher_account_ledger应新增一条
      LESSON_PAYABLE记录
- [ ] 2.6 实现代码使2.5测试通过
- [ ] 2.7 编写失败测试：确认后lesson_finance应新增一条记录，
      gross_profit_amount = charge_amount - teacher_pay_amount计算正确
- [ ] 2.8 实现代码使2.7测试通过，同时更新lesson.status=COMPLETED
- [ ] 2.9 编写失败测试（关键）：mock repository层在写teacher_account_ledger
      这一步返回error，验证attendance和student_account_ledger
      的写入也不会真正提交到数据库
      文件：backend/internal/lesson/confirm_transaction_test.go
      验证方式：用sqlmock或内存sqlite，故意在第4步注入错误，
      查询数据库应看不到本次操作的任何记录
- [ ] 2.10 若2.9失败，检查是否所有写入都用了同一个tx *gorm.DB对象，修正后重测
- [ ] 2.11 编写失败测试：STUDENT_LEAVE_EARLY分类提交后，
      lesson_deducted=0、teacher_pay_amount=0，验证不产生ledger记录
- [ ] 2.12 实现代码使2.11测试通过
- [ ] 2.13 编写失败测试：对status已为COMPLETED的课次再次调用确认接口
      应返回42201，不产生新记录
      文件：backend/internal/lesson/confirm_duplicate_test.go
- [ ] 2.14 实现状态前置校验使测试通过
- [ ] 2.15 编写失败测试：模拟并发场景（两个goroutine几乎同时调用confirm），
      验证只有一个成功写入attendance，另一个因唯一约束冲突而回滚
      文件：backend/internal/lesson/confirm_concurrent_test.go
- [ ] 2.16 验证2.15通过（若失败检查attendance.lesson_id是否有UNIQUE约束）
- [ ] 2.17 提交：git commit -m "feat(attendance): confirm transaction with duplicate and concurrency protection"

## 3. 本位币锁定

- [ ] 3.1 编写失败测试：全新环境（无任何ledger记录）执行一次课后确认后，
      system_config.base_currency_locked应变为'1'
      文件：backend/internal/lesson/currency_lock_test.go
- [ ] 3.2 实现代码：在2.2-2.8的事务内增加锁定判断逻辑，使3.1测试通过
      要点：先判断base_currency_locked是否已为'1'，是则跳过COUNT查询
- [ ] 3.3 编写失败测试：base_currency_locked='1'时调用
      PUT /system/base-currency 应返回42201
      文件：backend/internal/system/currency_test.go
- [ ] 3.4 实现代码使3.3测试通过
- [ ] 3.5 提交：git commit -m "feat(system): base currency lock on first financial record with perf shortcut"

## 4. 前端：课后确认表单

- [ ] 4.1 出勤分类下拉组件，选中后调用建议值接口自动填充四个"建议"只读字段
- [ ] 4.2 "实际"字段默认带入建议值，Operator可编辑
- [ ] 4.3 提交前二次确认弹窗（"提交后不可撤销，确认继续？"）
- [ ] 4.4 提交时若后端返回"课次已被确认"类错误，前端应友好提示并刷新
      课次状态，而非展示原始错误信息
- [ ] 4.5 提交：git commit -m "feat(frontend): attendance confirmation form with conflict handling"

## 5. 集成验收（对照PRD 24.3节TC-01~07，逐条勾选后才能archive）

- [ ] 5.1 TC-01 确认后学生余额和老师应付正确变化（人工用计算器核对至少3笔）
- [ ] 5.2 TC-02 中途失败全部回滚（已在2.9-2.10覆盖，此处做一次端到端手动复测）
- [ ] 5.3 TC-03 STUDENT_LEAVE_EARLY不扣课时不产生课酬（已在2.11-2.12覆盖）
- [ ] 5.4 TC-04 TEACHER_LEAVE不扣学生课时不产生应付（补充测试，模式同2.11）
- [ ] 5.5 TC-07 suggested_*字段与实际字段可以不同且都正确落库（已在2.1覆盖，
      此处做一次人工查库确认）
- [ ] 5.6 本位币锁定TC-08（已在3.1-3.4覆盖，此处做一次端到端手动复测）
- [ ] 5.7 重复确认保护（2.13-2.16覆盖，此处做一次端到端手动复测：真的
      在浏览器里对同一课次点两次确认按钮，验证第二次得到友好错误提示）

**只有第5节全部勾选，才运行 `/opsx:archive add-attendance-confirmation`**，
这是本change区别于其他change的特殊要求。

## 6. 规格场景覆盖检查表

对照本change下specs/attendance-confirmation/spec.md和
specs/base-currency-lock/spec.md的全部Scenario，逐条标注验证task
（与第5节的TC编号验收互为补充，第5节验证PRD测试用例口径，本节
验证spec文件本身的场景口径，两者应完全对应）：

- [ ] 6.1 「正常上课分类」→ 1.1-1.2
- [ ] 6.2 「学生提前请假分类」→ 1.3-1.4
- [ ] 6.3 「已改期分类」→ 1.3-1.4
- [ ] 6.4 「其他分类无建议」→ 1.3-1.4
- [ ] 6.5 「正常提交」→ 2.1-2.8
- [ ] 6.6 「事务中途失败」→ 2.9-2.10
- [ ] 6.7 「已确认课次拒绝再次确认」→ 2.13-2.14
- [ ] 6.8 「并发提交只有一个成功」→ 2.15-2.16
- [ ] 6.9 「实际值覆盖建议值」→ 2.1-2.2
- [ ] 6.10 「建议值与实际值可完全不同」→ 2.1-2.2、2.11-2.12
- [ ] 6.11 「首次课后确认触发锁定」→ 3.1-3.2
- [ ] 6.12 「锁定后拒绝修改」→ 3.3-3.4
- [ ] 6.13 「锁定判断性能优化」→ 3.2（实现要点包含在内）

全部勾选（含第5节）后才可执行`/opsx:archive add-attendance-confirmation`。
