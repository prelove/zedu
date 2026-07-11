## ADDED Requirements

### Requirement: 本位币首次锁定
系统必须在产生首条财务记录时自动锁定本位币。

#### Scenario: 首次课后确认触发锁定
- **WHEN** 系统此前从未有过任何student_account_ledger或
  teacher_account_ledger记录，本次课后确认产生了第一条
- **THEN** 同一事务内system_config.base_currency_locked被置为'1'

#### Scenario: 锁定后拒绝修改
- **WHEN** base_currency_locked='1'时调用PUT /system/base-currency
- **THEN** 返回42201，且system_config中的base_currency值未被修改

#### Scenario: 锁定判断性能优化
- **WHEN** base_currency_locked已经为'1'
- **THEN** 后续课后确认不再重复执行"是否为首条记录"的COUNT查询判断，
  直接跳过锁定逻辑
