# 项目状态

- 当前阶段：M2 核心资料与认证
- 当前状态：IN_PROGRESS
- 当前目标：按已冻结的 M2 契约完成认证、初始化与核心主数据，禁止扩展到财务或排课
- 最近更新：2026-07-11

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

1. Codex 基于 PRD v3.1 与 M2 只读预审创建正式 OpenSpec change、设计与任务。
2. 产品负责人确认“学生邮箱重复”语义后冻结 M2 API 契约。
3. 仅在契约冻结后，向 GLM/Kimi 下达范围隔离的 M2 实现工单。

## 阻塞

M2 尚未开始编码；学生邮箱唯一性与40901语义已由 ADR-007 冻结。Resend sender/test inbox、凭证限制和备份参数仍在相应 MVP 门禁前确认。
