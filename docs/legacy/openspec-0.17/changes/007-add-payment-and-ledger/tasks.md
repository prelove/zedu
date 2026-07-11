## 1. 充值创建与折算

- [ ] 1.1 编写失败测试：CNY 500、汇率21.8充值，amount_base应为10900(JPY)
      文件：backend/internal/finance/payment_test.go
- [ ] 1.2 实现最小代码使测试通过
      文件：backend/internal/finance/model.go(payment), dto.go, service.go,
      handler.go, repository.go
      要点：用shopspring/decimal计算折算金额
- [ ] 1.3 编写失败测试：金额汇率相乘结果为123.456时，amount_base应
      四舍五入为123
- [ ] 1.4 实现四舍五入逻辑使测试通过（明确使用Round而非Truncate）
- [ ] 1.5 编写失败测试：original_currency等于base_currency时，
      fx_rate_to_base应被强制设为"1"（即使前端传了其他值）
- [ ] 1.6 实现代码使测试通过
- [ ] 1.7 编写失败测试：lessons_added为0或负数应返回40001
- [ ] 1.8 实现校验逻辑使测试通过
- [ ] 1.9 编写失败测试：充值成功后enrollment.balance_amount和
      lesson_balance正确增加
- [ ] 1.10 实现代码使测试通过
- [ ] 1.11 提交：git commit -m "feat(payment): multi-currency recharge with rounding and validation"

## 2. 充值作废

- [ ] 2.1 编写失败测试：作废CONFIRMED充值后，status变VOIDED，
      生成VOID类型ledger记录，enrollment余额正确冲正
      文件：backend/internal/finance/void_test.go
- [ ] 2.2 实现最小代码使测试通过（同一事务内完成三步写入）
- [ ] 2.3 编写失败测试：对已VOIDED记录再次作废应返回42201
- [ ] 2.4 实现状态校验使测试通过
- [ ] 2.5 编写失败测试：作废导致余额为负时应允许，不拦截
- [ ] 2.6 验证2.5通过
- [ ] 2.7 提交：git commit -m "feat(payment): void with ledger reversal allowing negative balance"

## 3. 部分退款

- [ ] 3.1 编写失败测试：对CONFIRMED充值提交部分退款，应生成REFUND
      类型ledger记录，原payment记录status不变
      文件：backend/internal/finance/refund_test.go
- [ ] 3.2 实现最小代码使测试通过
      文件：backend/internal/finance/refund_service.go, refund_handler.go
- [ ] 3.3 编写失败测试：退款金额超过剩余余额时仍应成功执行
- [ ] 3.4 验证3.3通过（确认没有误加上限校验）
- [ ] 3.5 提交：git commit -m "feat(payment): partial refund without payment status change"

## 4. 支付方式字典读取接口

- [ ] 4.1 编写失败测试：GET /system/payment-methods 返回种子数据的6条记录
      文件：backend/internal/system/payment_method_test.go
- [ ] 4.2 实现代码使测试通过
- [ ] 4.3 编写失败测试：充值时引用不存在的payment_method_code应被拒绝
- [ ] 4.4 实现外键或应用层校验使测试通过
- [ ] 4.5 提交：git commit -m "feat(system): payment method dictionary read API with reference validation"

## 5. 前端：充值表单与流水展示

- [ ] 5.1 充值记录列表页
- [ ] 5.2 新建充值表单（暂不含上传凭证按钮，留给011）
- [ ] 5.3 学生详情页账务Tab：充值记录表格+账户流水表格
- [ ] 5.4 老师详情页账务Tab：应付概览+课时流水（暂不做正式结款提交）
- [ ] 5.5 退款操作入口（在充值详情页）
- [ ] 5.6 提交：git commit -m "feat(frontend): payment form, refund entry, and ledger views"

## 6. 规格场景覆盖检查表

对照本change下specs/payment-and-ledger/spec.md的全部Scenario，逐条
标注验证task：

- [ ] 6.1 「非本位币充值折算」→ 1.1-1.2
- [ ] 6.2 「折算结果四舍五入」→ 1.3-1.4
- [ ] 6.3 「本位币充值汇率固定为1」→ 1.5-1.6
- [ ] 6.4 「课时数必须为正」→ 1.7-1.8
- [ ] 6.5 「作废冲正」→ 2.1-2.2
- [ ] 6.6 「已作废记录不可重复作废」→ 2.3-2.4
- [ ] 6.7 「作废导致余额为负也允许」→ 2.5-2.6
- [ ] 6.8 「部分退款」→ 3.1-3.2
- [ ] 6.9 「退款金额由人工判断，系统不做上限校验」→ 3.3-3.4
- [ ] 6.10 「使用字典中的支付方式」→ 4.1-4.2
- [ ] 6.11 「引用不存在的支付方式被拒绝」→ 4.3-4.4

全部勾选后才可执行`/opsx:archive add-payment-and-ledger`。
