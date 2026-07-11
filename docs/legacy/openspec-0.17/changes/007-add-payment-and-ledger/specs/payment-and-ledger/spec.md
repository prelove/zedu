## ADDED Requirements

### Requirement: 充值记录创建与折算
系统必须支持多币种充值，保留原始金额和汇率快照，并折算为本位币金额，
使用四舍五入而非截断。

#### Scenario: 非本位币充值折算
- **WHEN** 提交original_amount="500.00"、original_currency="CNY"、
  fx_rate_to_base="21.80"，且当前base_currency为JPY
- **THEN** amount_base计算为10900，且原始金额/币种/汇率均被保留

#### Scenario: 折算结果四舍五入
- **WHEN** 提交的金额和汇率相乘结果为小数（如123.456）
- **THEN** amount_base按四舍五入取整为123，而非截断为123或错误地
  向上取整为124（验证具体的.5边界舍入行为符合四舍五入规则）

#### Scenario: 本位币充值汇率固定为1
- **WHEN** original_currency等于当前base_currency
- **THEN** fx_rate_to_base被后端强制设为"1"，忽略前端传入的其他值

#### Scenario: 课时数必须为正
- **WHEN** 提交lessons_added=0或负数
- **THEN** 返回40001参数校验失败

### Requirement: 充值作废与余额冲正
系统必须支持作废已确认的充值记录，且不允许物理删除，作废后不可
再次作废。

#### Scenario: 作废冲正
- **WHEN** 对一条CONFIRMED状态的充值执行作废操作并填写原因
- **THEN** 记录状态变为VOIDED，同时生成一条VOID类型的账户流水，
  enrollment余额和课时正确扣减回作废前的水平

#### Scenario: 已作废记录不可重复作废
- **WHEN** 对状态已为VOIDED的记录再次调用作废接口
- **THEN** 返回42201

#### Scenario: 作废导致余额为负也允许
- **WHEN** 作废一笔充值时，该enrollment的余额已经因后续消费不足以
  完全扣减
- **THEN** 允许balance_amount变为负数，如实反映欠费状态，不做拦截

### Requirement: 部分退款
系统必须支持对已生效的充值进行部分或全部退款，且退款不改变原充值
记录的状态。

#### Scenario: 部分退款
- **WHEN** 对一笔CONFIRMED状态的充值提交退款请求，指定退款金额小于
  原充值金额
- **THEN** 生成一条REFUND类型的账户流水，amount_delta为负数，
  enrollment余额相应减少，原student_payment记录状态保持CONFIRMED不变

#### Scenario: 退款金额由人工判断，系统不做上限校验
- **WHEN** 提交的退款金额大于该enrollment当前剩余余额
- **THEN** 系统仍然接受该操作（信任Operator的人工判断），余额可以
  变为负数

### Requirement: 支付方式字典引用
充值记录的支付方式必须引用payment_method字典，不使用硬编码枚举。

#### Scenario: 使用字典中的支付方式
- **WHEN** 提交payment_method_code为字典中已存在的code
- **THEN** 充值创建成功

#### Scenario: 引用不存在的支付方式被拒绝
- **WHEN** 提交payment_method_code不存在于字典中
- **THEN** 返回40001或40401（视具体实现，明确不允许自由文本绕过字典）
