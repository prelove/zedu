# Design: 课后确认事务

## 数据流

POST /lessons/{id}/confirm 收到请求后，在单一db.Transaction()内顺序执行：

1. 校验lesson.status为SCHEDULED或REMINDED，否则返回42201（这一步同时
   防止了对同一课次重复确认的问题——第一次确认成功后lesson.status
   已经变为COMPLETED，第二次尝试确认会在这一步就被拦截）
2. 写入attendance记录：
   - suggested_deduct_lessons/suggested_charge_amount/suggested_teacher_pay_amount
     ← 根据出勤分类的建议比例 × enrollment/assignment的默认值计算得出（仅做快照，
     不参与后续任何计算）
   - lesson_deducted/charge_amount/teacher_pay_amount
     ← 直接取请求体中Operator提交的实际值（这才是真正生效的数字）
   - attendance.lesson_id有UNIQUE约束，这是第二道并发保护——即使两个
     并发请求都通过了第1步的状态校验（理论上不太可能但要防御性设计），
     数据库层面的UNIQUE约束会让其中一个写入失败，触发事务回滚
3. 若lesson_deducted > 0：写student_account_ledger(biz_type=LESSON_DEDUCT)，
   更新enrollment.balance_amount -= charge_amount、lesson_balance -= lesson_deducted
4. 若teacher_pay_amount > 0：写teacher_account_ledger(biz_type=LESSON_PAYABLE)，
   更新teacher待结算缓存（若通过ledger聚合计算则无需额外字段更新）
5. 写lesson_finance：gross_profit_amount = charge_amount - teacher_pay_amount
6. 更新lesson.status = COMPLETED
7. 判断：本次操作前，student_account_ledger和teacher_account_ledger是否均为空
   （即这是系统第一条财务记录）？若是，同一事务内将
   system_config.base_currency_locked 置为 '1'

## 并发安全设计

课后确认理论上存在"两个Operator同时打开同一课次的确认页面并几乎
同时提交"的场景（虽然V1只有个位数Operator账号，概率很低，但仍需
防御）。本设计依赖两层保护：
- 应用层：第1步的status校验（SCHEDULED/REMINDED才允许进入后续流程）
- 数据库层：attendance.lesson_id的UNIQUE约束作为最后防线，即使
  应用层因为读写间隙（TOCTOU）判断失误，数据库约束仍会阻止产生
  两条attendance记录

两个并发请求中，先提交事务的一个会成功，后提交的一个会因UNIQUE
约束冲突而失败并整体回滚，前端应捕获这类错误并提示"该课次可能已
被确认，请刷新页面查看最新状态"，而不是展示一个生硬的500错误。

## 关键约束（重申）

- 步骤2-7必须在同一个*gorm.DB事务对象上执行，使用tx.Create()/tx.Model().Update()，
  不能有任何一步用外层未开启事务的db对象
- 任一步返回error，整个函数return err，让GORM的Transaction()自动回滚
- outcome_type的建议值计算逻辑放在service层的一个独立纯函数里，
  便于单元测试（输入：分类code+enrollment/assignment默认值；输出：三个建议值）

## 边界情况

- 出勤分类为OTHER时，建议值三列均为NULL，前端不自动带出任何默认值，
  完全由Operator手填
- attendance一旦创建不可修改（业务规则R7），本change不实现编辑/删除接口
- RESCHEDULED（已改期）分类：建议值三列均为0，表示这节课因故改期，
  不产生任何费用变化，但仍然需要生成一条attendance记录来"关闭"这个
  课次的挂起状态，配合改期后另行创建的新课次一起构成完整的补课记录

## 本位币锁定判断的性能考虑

"是否是首条财务记录"的判断，最简单的实现是在事务内执行
`SELECT COUNT(*) FROM student_account_ledger` 和对teacher_account_ledger
同样查询，如果两者都为0则触发锁定。考虑到这个COUNT查询只在
base_currency_locked仍为'0'时才需要执行（一旦锁定后就不需要再判断），
应该先读取system_config.base_currency_locked，只有在其为'0'时才
执行这两个COUNT查询，避免每次课后确认都对全表做COUNT扫描——虽然
V1数据量级很小这个性能问题并不明显，但这是一个值得在实现时就
养成的良好习惯。
