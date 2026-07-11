# OpenSpec + Superpowers AI 开发流程

OpenSpec管理“构建什么”和事实追踪；Superpowers管理“如何可靠实现”。每个change执行：读取事实源→设计/Non-Goals检查→15-90分钟任务计划→红绿重构→规格审查→代码审查→新鲜验证→更新Evidence/状态。

OpenSpec固定1.6.0。Superpowers固定官方v6.1.1并按Claude/Codex/Kimi各自插件安装；没有统一`superpowers` CLI。Gemini的v6.1.1官方支持未证实，当前仅承诺OpenSpec技能，待单独验证后更新基线。

GLM、Devin等未有OpenSpec官方适配器时，使用 `docs/templates/task-brief.md` 传递相同约束，输出必须回写统一证据和追踪矩阵。
