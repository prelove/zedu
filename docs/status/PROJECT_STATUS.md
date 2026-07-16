# 项目状态

- 当前阶段：M2 核心资料与认证
- 当前状态：IN_PROGRESS
- 当前目标：在已验收认证与初始化基础上，按独立切片实现 M2 人员资料、课程、报名与安排；禁止扩展到财务或排课
- 最近更新：2026-07-16

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

1. 由 GLM 实现 M2-GLM-02B：人员资料 API；由 Kimi 实现 M2-KIMI-01：认证、登录、受保护路由与初始化界面。
2. 完成并验收人员资料后，发布课程/报名/安排后端切片；不得跳过真实 HTTP 与浏览器验收。
3. 稳定性 20 次重复扫描只在里程碑候选、迁移/并发基础设施变更或出现非确定性失败时执行。

## 阻塞

M2-GLM-01 与 M2-GLM-02A 已于 2026-07-16 经 Windows/Ubuntu CI（含 Linux race）验收并合并。学生邮箱唯一性与40901语义已由 ADR-007 冻结。Resend sender/test inbox、凭证限制和备份参数仍在相应 MVP 门禁前确认。
