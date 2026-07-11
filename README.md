# Zedu

Zedu 是面向小型教培机构的轻量级教务与经营闭环系统。当前项目处于 **M0 治理与规格迁移阶段**：先完成并验收 MVP，再继续 V1；路线图持续覆盖 V1.5/V2，后续任务不会因聚焦 MVP 而丢失。

## 当前基准

- 唯一业务事实源：[正式 PRD v3.1](docs/2_prd/Zedu-PRD-Final-v3.1.md)
- 总路线图：[MASTER_ROADMAP](docs/roadmap/MASTER_ROADMAP.md)
- 实时状态：[PROJECT_STATUS](docs/status/PROJECT_STATUS.md)
- 治理与 DoD：[GOVERNANCE](docs/governance/GOVERNANCE.md)
- Claude/OpenSpec 评审：[claude-spec-review](docs/reviews/claude-spec-review.md)
- 决策与风险：[DECISION_LOG](docs/governance/DECISION_LOG.md) / [RISK_REGISTER](docs/governance/RISK_REGISTER.md)

## MVP 边界

MVP 包含登录初始化、人员课程、报名安排、充值与付款凭证、排课、Resend 邮件通知、课后确认、学生流水/老师应付、极简工作台和完整备份恢复。正式老师结款属于 V1，MVP 不提供入口。

## 工具基线

- OpenSpec：`@fission-ai/openspec@1.6.0`
- Superpowers：官方 `v6.1.1`；按 Claude/Codex/Kimi 各自插件机制安装。Gemini 的 v6.1.1 官方基线尚未确认，暂使用 OpenSpec 原生技能，Superpowers 支持进入待验证事项。
- Node.js：满足 OpenSpec `>=20.19.0`

```powershell
npm install -g @fission-ai/openspec@1.6.0
openspec --version
openspec validate --all --strict --no-interactive
```

## AI 开发入口

Codex、Claude、Gemini、Kimi 已分别生成 OpenSpec 1.6 技能。所有 AI 工具必须先读 `AGENTS.md`、当前 change、治理 DoD 和项目状态，只能认领 `READY` 任务，并提交可复现测试证据。

## 编码环境

仓库文本统一 UTF-8 和 LF，兼容 Windows 10 日文版；产品支持 `zh-CN`、`ja-JP`、`en-US`。不得依赖系统 ANSI/CP932 保存源码或文档。
