# M1 执行看板

> OpenSpec change：`establish-engineering-foundation`。状态：READY。共享进度由 Codex/PM 唯一维护。

| 工单 | 执行工具 | 范围 | 状态 | 依赖 | 验收人 |
|---|---|---|---|---|---|
| M1-GLM-01 | GLM | Go后端、健康检查、SQLite迁移、日志 | IN_REVIEW | M0 ACCEPTED | Codex + 独立Reviewer |
| M1-KIMI-01 | Kimi | Vue/TS前端、三语、健康页与测试 | IN_REVIEW | M0 ACCEPTED | Codex + 独立Reviewer |
| M1-CODEX-01 | Codex | 工具锁、共享脚本、CI、集成与证据 | READY | 两工单契约 | 独立Architect |
| M1-GLM-02 | GLM | 幂等模板seed框架与测试 | READY | 当前M1集成基线 | Codex + 独立Reviewer |

建议分支：`m1/glm-backend-foundation`、`m1/kimi-frontend-foundation`。并行时必须使用独立clone或Git worktree，禁止两个工具在同一工作目录切换分支。

## 并发写入规则

- GLM只写`backend/`及其内部测试；Kimi只写`frontend/`及其内部测试。
- 根目录、`openspec/`、`docs/`、`.github/`、`scripts/`、路线图和任务勾选仅由Codex/PM修改。
- 任一工具发现需要改共享契约时停止，在交付报告中提交变更建议，不能自行扩权。
- 完成状态必须由验收者基于新鲜测试证据更新，执行工具不得自行声明`ACCEPTED`。

## 集成顺序

1. 两工具分别在独立分支/工作树完成并提交Lore commit。
2. Codex逐项审查测试红灯证据、diff和依赖，再按GLM→Kimi顺序集成。
3. Codex运行OpenSpec strict、Go/前端测试、迁移往返、构建和编码检查。
4. 独立Reviewer签署后更新OpenSpec tasks、追踪矩阵、路线图和M1 evidence。
