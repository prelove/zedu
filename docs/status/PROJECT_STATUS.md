# 项目状态

> **状态覆盖（2026-07-20，优先于本文较早的里程碑记录）**：MVP 的 M2–M6 已完成技术实现和技术验证；当前为 **IN_REVIEW / MVP 技术发布候选**，并非生产 `ACCEPTED`。真实浏览器回归为 20 组 × zh-CN/ja-JP/en-US = **60/60 PASS**；OpenSpec `complete-prd-mvp-release-gaps` 仅剩 Product Owner 外部发布闸门（任务 5.3）。

## 当前定位（2026-07-20）

| 层级 | 状态 | 说明 |
|---|---|---|
| M0 治理与需求 | ACCEPTED | PRD、OpenSpec、工程与协作规范已建立。 |
| M1 工程基础 | ACCEPTED | CI、迁移、i18n、基础质量门禁已验收。 |
| M2 人员、课程、认证 | VERIFIED | 后端、前端及浏览器主路径已验证。 |
| M3 充值、流水、凭证 | VERIFIED | 技术实现与回归已完成。 |
| M4a 排课 / M4b 通知 | VERIFIED | 课次、outbox、提醒 runner 与 UI 已实现并回归。 |
| M5 课后确认 / 教师应付展示 | VERIFIED | 余额、确认、只读应付展示已实现；结款非 MVP。 |
| M6 工作台、备份、MVP 闸门 | IN_REVIEW | 技术门禁通过；等待外部 UAT/发布确认。 |

## 唯一剩余 MVP 发布闸门

1. Product Owner 人工账务样本核对；
2. Resend 受控发件域/测试收件人真实发送与失败重试证据（不得提交密钥）；
3. 授权操作者独立目录备份恢复 runbook 演练；
4. Win10 日文环境人工 UAT（含三语、CJK/emoji、日期和 JPY 格式）。

详细证据见 `docs/acceptance/evidence/MVP/browser-regression.md` 与 `docs/acceptance/evidence/MVP/release-gate.md`。上述外部闸门完成前，不得将 M5/M6 或 MVP 整体标记为生产 `ACCEPTED`。

- 当前阶段：MVP 技术发布候选收口
- 当前状态：IN_REVIEW
- 当前目标：仅完成 Product Owner 外部发布闸门；不得以技术证据替代人工 UAT 或扩大范围
- 最近更新：2026-07-20

## 已完成

- 确认正式 PRD v3.1 为事实源并完成MVP范围修订。
- 批准 MVP/V1 边界：通知和凭证进 MVP，正式结款留 V1。
- 完成 RALPLAN-DR 共识规划及测试规格。
- 核验 OpenSpec 1.6.0、Superpowers v6.1.1。
- 为 Codex/Claude/Gemini/Kimi 初始化 OpenSpec 1.6 技能。
- 将旧 001-014 移入 legacy 区，未标记为已完成。
- 完成Claude/OpenSpec评审、治理/编码/测试/验收/安全/i18n规范及首版追踪矩阵。
- 新change `establish-engineering-foundation` 已完成四类规划工件并通过OpenSpec strict。
- M0独立架构复验APPROVED；M0状态ACCEPTED。

## 已完成（M1）

- M1-GLM-01、M1-KIMI-01、M1-GLM-02 与集成任务均已独立审查并合入 `main`。
- 本机后端 14 个测试、20 次稳定性回归、构建；前端 57/57 单测、100% 覆盖率、构建与生产依赖审计均通过。
- GitHub Actions run `29153829469` 的 Windows/Ubuntu 治理与 foundation 四个 job 全绿；Ubuntu 已实际通过 `go test -race`。
- M1 验收证据见 `docs/acceptance/evidence/M1/verification-report.md`；OpenSpec 任务已由验收人勾选。

## 下一步

1. M2-KIMI-02 已交付、局部验收与可维护性收口进行中；M2 仍为 IN_PROGRESS，等待延后全量前端回归和浏览器集成验收，不能称 ACCEPTED。
2. M3 OpenSpec `add-m3-recharge-ledger-evidence` 已 strict 通过；M3-GLM-03A 已进入实施，范围为 005 migration、本位币和支付方式，不得提前实现充值、凭证或 M4+。
3. 实施者先阅读 `docs/standards/implementation-contract.md`、`docs/tasks/M3/M3_EXECUTION_BOARD.md`、M3 OpenSpec 全部工件及 PRD §7/§9.6/§13.9/§14/§16.5；仅在契约列出的触发条件下做定向扩展扫描。
4. 稳定性 20 次重复扫描与 Windows 前端全量 Vitest 均纳入 M3-CODEX-02 的里程碑质量批次；单工单仍必须跑其 focused lint/typecheck/unit/integration 证据。

## 阻塞

M2 后端、认证、初始化和认证前端已验收；M2-KIMI-02 正在独立验收与总回归收口。M3-GLM-03A 正在实施，其余实现任务均因依赖阻塞。学生邮箱唯一性与40901语义已由 ADR-007 冻结。Resend sender/test inbox、凭证限制和备份参数仍在相应 MVP 门禁前确认。
