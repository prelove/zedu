# Proposal: 老师结款与排课/换老师增强

## Why
MVP阶段（005/004）实现的排课和换老师是基础版，缺少时间冲突检测和
"批量更新未来课次"的体验。同时老师应付款需要正式的结款闭环，避免
重复结算。本change既新增payout capability，也MODIFIES此前的
lesson-scheduling和enrollment capability，是本项目中第一次演示
"对已归档能力做增量修改"这种OpenSpec用法的change。

## 业务背景
根据PRD第九章9.7节，结款不是简单的"把应付金额转给老师"，而是要
经过"预览→勾选确认→提交"这个流程，因为真实场景中运营者可能需要
排除个别有争议的课次（比如老师中途请假但已经确认了课后记录，
运营者事后决定这次不计入本次结款，留到下次单独处理）。这个"预览+
可勾选排除"的设计，本质上是给运营者留出人工核查的机会，避免结款
金额出现运营者事后才发现的错误。

排课冲突检测和换老师的批量更新，都是PRD第九章异常分支表格里明确
提到、但在MVP阶段被判定"可以延后"的体验增强（PRD9.4/9.3节异常
分支）。现在补上，是因为MVP运行一段时间后，这类"重复手动处理"的
摩擦感会变得越来越明显——比如换了老师但忘记手动改掉后面5节已排好
的课，这种遗漏在人工操作下几乎必然发生。

## What Changes
- 新增POST /finance/payouts/preview + POST /finance/payouts
- 新增GET /lessons/{id}/conflicts 时间冲突检测
- 修改换老师接口，增加updateFutureLessons参数支持批量更新未来课次
- 新增结款页面和排课表单的冲突警示

## Non-Goals
- 不实现结款的自动定时提醒（"该给某老师结款了"这类主动提醒属于010
  通知系统范围，本change只负责结款操作本身能被正确执行）
- 不实现跨老师批量结款（一次只能对一位老师执行结款，不支持"批量给
  所有老师结款"这种批处理操作，因为不同老师的结算周期和实付金额
  调整通常各不相同，批量操作反而容易掩盖需要人工核实的差异）
- 不实现冲突检测的"强制阻止创建"模式（PRD明确要求只提示不阻止，
  尊重小机构业务的灵活性，例外情况确实存在，如老师连续带两个班）

## Impact
- Affected specs: payout（新增）、lesson-scheduling（MODIFIED）、
  enrollment（MODIFIED）
- Affected code: backend/internal/finance/(payout相关)、
  backend/internal/lesson/(冲突检测)、backend/internal/enrollment/(换老师增强)
- 依赖：005-add-lesson-scheduling、004-add-enrollment-assignment、
  007-add-payment-and-ledger（需要teacher_account_ledger表）
- 被依赖：010（通知系统的"待结款老师"晨报条目依赖本change的payout
  数据结构判断哪些老师已经历史结算过）
