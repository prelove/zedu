# Design: 多币种折算精度、作废冲正与退款处理

## 金额精度方案

- original_amount以decimal字符串存储，避免JSON数字精度丢失（JavaScript
  的Number类型和某些JSON库在解析大数字或高精度小数时会有精度损失，
  这是选择字符串传输金额的根本原因）
- fx_rate_to_base同样以decimal字符串存储，建议保留至少4位小数精度
  （如21.8000），避免汇率本身的精度不足导致折算结果有偏差
- 折算逻辑：`amount_base = round(decimal(original_amount) * decimal(fx_rate_to_base))`，
  使用shopspring/decimal计算，最终取整为本位币最小单位的整数。取整
  方式采用四舍五入（Round，非截断Truncate），因为截断会系统性地
  让运营者"吃亏"（折算结果总是偏小），四舍五入更符合一般商业惯例
- 若original_currency等于当前system_config.base_currency，fx_rate_to_base
  固定为"1"，后端强制覆盖前端传入值（防止前端bug或恶意请求误传导致
  金额计算错误——这是一条不信任前端输入的防御性设计）

## 作废冲正方案

作废操作不物理删除student_payment记录，而是：
1. student_payment.status = VOIDED，记录voided_at和void_reason
2. 生成一条student_account_ledger记录，biz_type=VOID，
   amount_delta = -原充值的amount_base，lesson_delta = -原充值的lessons_added
3. 更新enrollment.balance_amount和lesson_balance相应扣减
4. 以上三步在同一事务内完成

## 退款方案（区别于作废）

退款针对"充值已经生效一段时间，现在需要退回部分金额"的场景，与作废
的关键区别：
- 作废：默认退回全部金额，且暗示"这笔充值本身有问题"（录错了）
- 退款：可以是部分金额，且charge本身没有问题，只是运营者主动决定
  退还一部分（比如学生因为特殊原因中途不学了，退还未消耗的部分）

实现上，退款不修改student_payment记录本身（该笔充值仍然是
CONFIRMED状态，因为它确实生效过），而是直接生成一条
student_account_ledger记录，biz_type=REFUND，amount_delta为负数
（退还的金额），lesson_delta也可以是负数（如果连带扣减课时）。
退款金额和课时的具体数值由Operator在退款表单中手动输入，系统不
做"退款不能超过剩余可退额度"这类自动校验（信任运营者的人工判断，
避免系统过度介入本该由人决定的财务判断，呼应PRD原则四）。

## 边界情况

- 若作废时enrollment余额已经因为后续课后确认变成不足以扣减（即余额已经
  被花掉一部分），扣减后balance_amount允许为负数（如实反映"该学生欠费"的
  真实情况），不做额外拦截
- 已作废的充值记录不允许再次作废（状态机保护，见Requirement）
- 退款没有"作废"那样的终态限制，同一笔充值理论上可以有多次退款
  记录（虽然实际业务中很少见，但系统不应该人为限制这种可能性）
