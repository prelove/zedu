## ADDED Requirements

### Requirement: 精确且幂等的充值确认
授权 Owner/Operator MUST 使用 `paymentNo`、student/enrollment、原始金额与币种、汇率、课时、支付方式和付款时间创建充值。金额与汇率 MUST 用十进制字符串精确校验与计算，禁止 float；payment MUST 永久保留原始快照。相同 `paymentNo` 的相同请求 MUST 安全重放且不重复记账；相同号码不同内容 MUST 返回 `40901`。

#### Scenario: 非本位币充值被确认
- **WHEN** Operator 用有效的 CNY 原始金额和汇率创建 JPY 本位币充值
- **THEN** 系统 MUST 保存原始金额、币种、汇率和精确折算金额，并返回确认后的同一 payment

#### Scenario: 网络重试同一充值
- **WHEN** 同一授权用户以相同 `paymentNo` 和相同业务字段再次提交
- **THEN** 系统 MUST 返回原 payment，且只存在一条 RECHARGE ledger、一次余额增加和一次成功审计

#### Scenario: 报名不属于学生
- **WHEN** 请求中的 enrollment 不属于请求中的 student，或任一对象不是 ACTIVE
- **THEN** 系统 MUST 返回 `42201` 且不写 payment、ledger、余额或成功审计

### Requirement: 充值确认的原子账务事实
确认充值 MUST 在单一数据库事务中写 `student_payment(CONFIRMED)`、`student_account_ledger(RECHARGE)`、报名余额/课时余额、首次锁定标志和 operation_log。任何数据库、审计或 commit 失败 MUST 返回 `50002`，不得泄露底层错误文本或留下半写。

#### Scenario: 审计写入失败
- **WHEN** 已通过充值业务校验后 operation_log 写入失败
- **THEN** 系统 MUST 返回 `50002`，且 payment、ledger、报名余额和锁定标志均保持提交前状态

#### Scenario: 并发确认同一业务编号
- **WHEN** 两个并发请求以同一有效 `paymentNo` 创建充值
- **THEN** 系统 MUST 最终只有一条 payment 和一条 RECHARGE ledger，余额只增加一次

### Requirement: 充值查询、流水与不可变作废
授权 Owner/Operator MUST 能分页筛选充值、读取单笔详情和读取学生流水。确认充值不得编辑或删除；作废 MUST 要求原因，在单一事务中将 payment 置为 VOIDED、写负向 VOID ledger、回滚报名余额并写审计。已作废记录再次作废 MUST 返回 `42201`。

#### Scenario: 作废确认充值
- **WHEN** Operator 为 CONFIRMED payment 提交非空作废原因
- **THEN** 系统 MUST 写一条相反金额/课时的 VOID ledger，并使报名余额回到该充值前的事实状态

#### Scenario: 作废事务失败
- **WHEN** 作废过程中的 ledger、余额或审计写入失败
- **THEN** 系统 MUST 返回 `50002`，payment 仍为 CONFIRMED，且不存在部分 VOID ledger 或余额变化

#### Scenario: 未认证读取财务记录
- **WHEN** 请求没有有效认证访问 payment 或 student ledger
- **THEN** 系统 MUST 返回 `40101` 且不返回任何财务数据
