# Proposal: 充值管理与账务流水

## Why
充值是学生余额的唯一正向入口，必须保证多币种折算精度正确、作废时
余额能被正确冲正、且历史充值记录永不物理删除以保证审计可信。这是
PRD第二章原则三"财务上落事实，不依赖回算"在充值这一侧的具体体现——
充值记录一旦产生，就是不可篡改的历史事实，任何后续的调整都必须通过
新增流水（VOID/REFUND/ADJUST）来体现，而不是直接修改原记录。

## 业务背景
根据PRD第七章7.2节，本项目支持多币种收款是一个刻意的设计决策：小型
教培机构的家长可能用微信支付人民币、也可能用PayPay支付日元，运营者
不应该被迫先手动换算好金额再录入系统，而应该是"录入实际收到的金额
和币种，系统负责折算"。这意味着每笔充值必须完整保留四个字段：原始
金额、原始币种、当时汇率、折算后金额——四者缺一都会导致未来审计时
无法还原"这笔钱到底是怎么算出来的"。

同时，PRD第九章9.6节的异常分支列出了"充值录错需要作废重录"、"学生
要求部分退款"两种真实场景，这两种场景在数据模型上是不同的：作废
（VOID）是"这笔充值从未真正生效"的语义，退款（REFUND）是"这笔充值
生效过，但现在要退回一部分"的语义，两者对应的ledger biz_type不同，
不能混用同一套逻辑处理。

## What Changes
- 新增student_payment CRUD + 作废接口
- 新增支持退款调整（REFUND类型的ledger记录）
- 新增payment_method字典读取接口
- 新增充值表单和学生/老师账务Tab（流水展示）

## Non-Goals
- 不实现付款凭证上传（见011-add-configurable-dictionaries，本change
  的充值表单暂时没有"上传截图"按钮）
- 不实现老师结款（见009-add-payout-and-scheduling-enhancement）
- 不实现自动汇率抓取（PRD明确V1只做人工录入汇率，不依赖外部汇率API，
  这是刻意的简化，避免系统对外部服务产生新的依赖）
- 不实现课时套餐(package)的独立管理页面（本change的充值表单可以
  手动输入课时数和套餐名称的自由文本，结构化的套餐管理留待012或
  按需再评估）

## Impact
- Affected specs: payment-and-ledger（新增）
- Affected code: backend/internal/finance/(payment相关)、
  frontend/admin/src/views/finance/payments/
- 依赖：004-add-enrollment-assignment（需要enrollment表）、
  001-add-project-scaffold（需要payment_method字典种子数据）
- 被依赖：006（课后确认需要student_account_ledger/
  teacher_account_ledger表结构，若006先执行则本change需要确认表结构
  一致，不重复定义）、008（工作台展示待续费学生需要enrollment余额
  数据）、009（结款需要teacher_account_ledger数据的写入方式与本change
  保持一致）
