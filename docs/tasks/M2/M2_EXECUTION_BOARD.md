# M2 执行看板

> OpenSpec change：`add-m2-core-management`。共享进度由 Codex/PM 唯一维护；执行工具不得自行勾选 OpenSpec tasks 或修改共享契约、治理、CI、路线图、状态与证据。

| 工单 | Owner | 范围 | 状态 | 依赖 | 允许写入 | 验收人 |
|---|---|---|---|---|---|---|
| M2-CODEX-01 | Codex | 契约、依赖与负面范围冻结 | ACCEPTED | M1 ACCEPTED | docs/tasks/M2、docs/acceptance/evidence/M2、OpenSpec task状态 | Codex/PM |
| M2-GLM-01 | GLM | 安全基础、统一契约、M2 迁移、认证/RBAC | ACCEPTED | M2-CODEX-01 | 已合并 `main`（PR #2）；证据 `docs/acceptance/evidence/M2/GLM-01.md` | Codex + 独立 Reviewer |
| M2-GLM-02A | Codex | 初始化、受限重置 API（M2-GLM-02 的第一个切片） | ACCEPTED | M2-GLM-01 ACCEPTED | 已合并 `main`（`3bc4078`）；证据 `docs/acceptance/evidence/M2/GLM-02A.md` | Codex + 独立 Reviewer |
| M2-GLM-02B/02C | GLM | 人员资料；课程/报名/安排 API | READY | M2-GLM-02A ACCEPTED | `backend/internal/app/directory/**`、`course/**`、必要路由装配及测试 | Codex + 独立 Reviewer |
| M2-KIMI-01 | Kimi | 前端路由、认证、登录与初始化界面 | READY | M2-GLM-01 的认证契约 ACCEPTED | `frontend/package*.json`、`frontend/src/router/**`、`api/**`、`stores/**`、`features/auth/**`、`features/onboarding/**`、i18n与测试 | Codex + 独立 Reviewer |
| M2-KIMI-02 | Kimi | 人员、课程、报名、安排页面 | BLOCKED | M2-GLM-02 API ACCEPTED、M2-KIMI-01 ACCEPTED | `frontend/src/features/directory/**`、`course/**`、i18n与测试 | Codex + 独立 Reviewer |
| M2-CODEX-02 | Codex | 真实 HTTP、浏览器、CI 与发布验收 | BLOCKED | GLM-02、KIMI-02 ACCEPTED | 集成测试、CI、证据、追踪矩阵、状态与路线图 | 独立 Reviewer |

## 统一交付要求

1. 先提交红灯证据，再提交最小实现与绿灯证据；不得把 commit 当作验收。
2. 每份报告必须列出 commit SHA、允许范围内的文件、未测试项、回滚方式和执行命令输出。
3. 所有新用户可见字符串同步维护 `zh-CN`、`ja-JP`、`en-US`；CJK/emoji 与 Windows 日文环境必须回归。
4. 任何需求、依赖、路由或数据库约束变更均停止任务并提交变更建议，不得自行扩权。
5. `docs/acceptance/evidence/M2/contract-freeze.md` 是编码契约；违反其邮箱、权限、负面范围或依赖规则的交付直接拒收。
