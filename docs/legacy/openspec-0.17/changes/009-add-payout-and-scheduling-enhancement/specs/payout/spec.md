## ADDED Requirements

### Requirement: 结款预览
系统必须能预览某老师某周期内未结算的应付课次明细，且已结算记录
不会重复出现。

#### Scenario: 预览待结算明细
- **WHEN** 调用POST /finance/payouts/preview 指定老师和周期
- **THEN** 返回该周期内related_payout_id为空的LESSON_PAYABLE记录列表
  及汇总金额

#### Scenario: 空结果不报错
- **WHEN** 该周期内该老师没有任何未结算记录
- **THEN** 返回空列表和汇总金额0，不报错

### Requirement: 结款提交与去重
系统必须支持提交结款，可排除个别记录，且已结算记录不会在下次预览
中重复出现，排除的记录仍留在未结算池中。

#### Scenario: 提交结款
- **WHEN** 基于预览结果提交结款，可勾选排除个别记录
- **THEN** 创建teacher_payout记录，未被排除的关联ledger记录被标记
  related_payout_id，teacher的unpaid_amount正确扣减（不含被排除的部分）

#### Scenario: 已结算不重复出现
- **WHEN** 结款提交后再次调用预览接口（相同或更晚的周期）
- **THEN** 已被本次结算覆盖的课次不会再出现在预览结果中

#### Scenario: 被排除的记录仍可在下次预览中出现
- **WHEN** 某次结款预览中Operator排除了某条记录未提交结算
- **THEN** 该条记录的related_payout_id保持为空，下次预览该老师时
  仍会出现在待结算列表中

### Requirement: 实付金额可与应付金额不同
系统必须允许结款时的实付金额与系统计算的应付金额不一致，且不需要
额外的调整流水解释差异。

#### Scenario: 实付调整
- **WHEN** 提交结款时actual_amount_base与预览的应付汇总金额不同
- **THEN** teacher_payout记录同时保存amount_base(应付)和
  actual_amount_base(实付)两个值，差异体现在这两个字段本身
