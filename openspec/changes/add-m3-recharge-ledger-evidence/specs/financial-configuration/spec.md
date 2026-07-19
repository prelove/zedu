## ADDED Requirements

### Requirement: 本位币状态与锁定
系统 MUST 提供本位币及锁定状态的读取；Owner MUST 只能在没有任何财务事实且未锁定时设置本位币。M3 MUST 支持 `JPY`、`CNY`、`USD`，首次成功确认充值 MUST 在同一事务内锁定本位币；已锁定或已有财务事实时拒绝修改并返回 `42201`。

#### Scenario: 首笔充值锁定本位币
- **WHEN** Owner 或 Operator 成功确认系统中的第一笔充值
- **THEN** 该 payment、ledger、余额更新和 `base_currency_locked=true` MUST 一并提交

#### Scenario: 有财务事实后修改本位币
- **WHEN** Owner 在已有充值或学生流水后提交不同本位币
- **THEN** 系统 MUST 返回 `42201` 且不修改现有设置或任何财务事实

### Requirement: 支付方式字典
系统 MUST 幂等提供 `WECHAT`、`ALIPAY`、`PAYPAY`、`BANK`、`CASH`、`OTHER` 默认支付方式。Owner MUST 能创建、更新名称/排序/启用状态但不能修改 code 或物理删除；Operator MUST 只读取启用方式。禁用方式 MUST 不影响历史 payment 的显示。

#### Scenario: Owner 禁用已被历史引用的支付方式
- **WHEN** Owner 禁用已有充值引用的支付方式
- **THEN** 系统 MUST 保留历史 payment 的方式代码和显示信息，并拒绝使用该方式创建新充值

#### Scenario: Operator 修改支付方式被拒绝
- **WHEN** Operator 请求创建或修改支付方式
- **THEN** 系统 MUST 返回现有 `40301` 且不产生字典或审计写入
