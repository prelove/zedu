# Zedu 项目治理与统一完成定义

## 事实层级

1. 经批准的 Decision Record（仅修正明确条款）
2. `docs/2_prd/Zedu-PRD-Final-v3.1.md`
3. 已批准 OpenSpec Requirement/Scenario
4. Task Brief、实现与测试证据

发生冲突必须先修事实层，禁止让代码自行成为新需求。

## 状态流

`BACKLOG → READY → IN_PROGRESS → BLOCKED → IN_REVIEW → VERIFIED → ACCEPTED`

- 只有 READY 可认领，一个任务一个 Owner。
- BLOCKED 必须记录原因、影响、绕行方案与决策人。
- VERIFIED 需要本次运行的测试证据。
- ACCEPTED 需要通过里程碑 Release Gate。

## Definition of Done

1. 有 PRD、Requirement、Scenario、Task、Test、Evidence 完整追踪。
2. Task 写明 In/Out、依赖、允许修改文件和回滚方法。
3. 先记录失败测试，再记录最小实现通过和重构后复验。
4. lint、typecheck、unit 及适用 integration/E2E 均通过。
5. 适用的三语、安全、权限、并发、失败恢复与可观测性均验证。
6. 独立 Reviewer 无未解决 P0/P1。
7. 追踪矩阵、项目状态、风险和证据同步更新。
8. 不以 TODO、空壳、隐藏开关或 commit 代替交付。

## 防跑偏、防缩水、防蔓延

- 新表、字段、API 必须引用 PRD/Decision；错误码不得自造。
- 测试红绿顺序和失败注入不可省略；困难不得静默降标。
- Non-Goals 是硬边界；超出 Impact 必须拆 change 或重新审批。

## 前端快速上下文契约

- 所有实现任务除本治理文件外，必须阅读 `docs/standards/implementation-contract.md`、当前工单和适用的冻结契约。
- 该契约只固化已经验收的数据库、后端、前端、HTTP、测试与负面范围；它不能覆盖 PRD、OpenSpec 或真实运行时行为。
- 只有 API/Schema 缺失或不一致、认证或依赖变更、跨领域新页面、测试失败无法定位时，才扩大为相关目录或 handler/repository/migration 的定向扫描；不得把每次全仓扫描当作编码前置条件。

## 人工门禁

产品范围、财务语义、角色权限、个人信息处理、持续费用、生产部署、真实数据迁移必须由 Product Owner 确认。
