# OpenSpec + Superpowers AI 开发流程

OpenSpec管理“构建什么”和事实追踪；Superpowers管理“如何可靠实现”。每个change执行：读取事实源→设计/Non-Goals检查→15-90分钟任务计划→红绿重构→规格审查→代码审查→新鲜验证→更新Evidence/状态。

OpenSpec固定1.6.0。Superpowers固定官方v6.1.1并按Claude/Codex/Kimi各自插件安装；没有统一`superpowers` CLI。Gemini的v6.1.1官方支持未证实，当前仅承诺OpenSpec技能，待单独验证后更新基线。

GLM、Devin等未有OpenSpec官方适配器时，使用 `docs/templates/task-brief.md` 传递相同约束，输出必须回写统一证据和追踪矩阵。

## 当前 MVP 收口状态（2026-07-20）

- OpenSpec change `complete-prd-mvp-release-gaps` 的实现任务和真实浏览器回归（5.2）已完成；浏览器证据为 M1–M6、20 组 × 三语、60/60 PASS。
- 任务 5.3 是 Product Owner 外部发布闸门，仍保持未勾选：人工账务核对、Resend 受控发送、备份恢复 runbook 演练、Win10 日文环境 UAT。
- 因此所有 AI 必须把当前状态表述为“`IN_REVIEW / MVP 技术发布候选`”，不得把 M5/M6 或 MVP 写为生产 `ACCEPTED`，不得为通过闸门而伪造外部证据、提交密钥或扩大 V1 范围。
- 任何后续 AI 会话先读取 `docs/status/PROJECT_STATUS.md`、`docs/roadmap/MASTER_ROADMAP.md`、当前 change 的 artifacts 和 `docs/acceptance/evidence/MVP/release-gate.md`；只处理被明确标记为 READY 的工作。
