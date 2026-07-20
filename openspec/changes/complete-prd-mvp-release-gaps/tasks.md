## 1. 契约与测试基线

- [x] 1.1 [Codex；依赖：proposal/design/specs；允许：本 change、`docs/acceptance/evidence/MVP/**`] 固化 PRD §23.2 与现有 M1–M6 范围差异、路由矩阵和负向范围；输出：追踪矩阵；测试：OpenSpec strict；证据：`docs/acceptance/evidence/MVP/prd-gap-closure.md`；门禁：不把结款、报表、冲突检测、HTTP restore 纳入实现。
- [x] 1.2 [Codex；依赖：1.1；允许：`backend/**` 对应测试、`frontend/tests/**`] 先为老师应付、提醒幂等/重试、备份包/篡改校验、工作台只读计数写红灯测试；输出：可复现失败测试；证据：同上；门禁：禁止 mock 代替数据库/文件系统事务证明。

## 2. 老师应付与工作台

- [x] 2.1 [GLM；依赖：1.2；允许：`backend/internal/app/payable/**`、`backend/internal/app/dashboard/**`、对应测试、必要 migration] 实现从 `teacher_account_ledger` 聚合的只读老师应付 summary/detail 和完整工作台只读计数；输出：受认证保护的 GET API；测试：未认证40101、Owner/Operator成功、空数据零值、确认课次后金额正确、读操作零副作用、无 payout/settlement 路由；证据：`docs/acceptance/evidence/MVP/payable-dashboard.md`；门禁：不新增结款、调整、导出、写操作或金额 float。
- [x] 2.2 [Kimi；依赖：2.1；允许：`frontend/src/features/{dashboard,directory}/**`、`frontend/src/api/**`、`frontend/src/i18n/**`、`frontend/tests/**`] 展示工作台最小运营指标和老师详情的只读待付事实；输出：三语页面/API adapter；测试：Owner/Operator、零状态、无结款按钮/菜单/请求、三语 key parity；证据：`docs/acceptance/evidence/MVP/payable-dashboard-ui.md`；门禁：不在客户端计算账务事实。

## 3. 课前提醒与失败重试

- [x] 3.1 [GLM；依赖：1.2；允许：`backend/migrations/009_*`、`backend/internal/app/notification/**`、`backend/cmd/**`、对应测试] 扩展 outbox 支持 `LESSON_REMINDER`、固定30分钟窗口扫描、三次上限和延迟重试；输出：受控 runner/CLI 与 migration；测试：扫描幂等、窗口边界、失败不回滚课次、retry after available_at、三次停止、日志脱敏；证据：`docs/acceptance/evidence/MVP/notification-operations.md`；门禁：不在 HTTP 请求或服务启动时隐式发送，不新增依赖。
- [x] 3.2 [Kimi；依赖：3.1；允许：`frontend/src/features/notification/**`、`frontend/src/api/**`、`frontend/src/i18n/**`、`frontend/tests/**`] 补充提醒/失败状态的只读可视化与人工重放提示；测试：真实代理、无密钥错误态、失败重试、三语及无敏感信息；证据：`docs/acceptance/evidence/MVP/notification-ui.md`；门禁：不提供规则配置页、SMTP、晨报或周报。

## 4. 可携带备份与恢复演练

- [x] 4.1 [GLM；依赖：1.2；允许：`backend/internal/app/backup/**`、`backend/cmd/**`、对应测试、必要 migration] 实现 staging→SQLite/上传附件/配置摘要→manifest SHA-256→原子发布的 Owner 备份包；测试：含附件成功、无 secrets、故障清理无成功审计、Operator40301、同秒不冲突；证据：`docs/acceptance/evidence/MVP/backup-package.md`；门禁：不得暴露 HTTP restore。
- [x] 4.2 [GLM；依赖：4.1；允许：`backend/internal/app/backup/**`、`backend/cmd/**`、对应测试] 实现只恢复到新临时目录的本地 verify/drill 命令；测试：哈希、SQLite、附件一致；篡改包失败且不改变活动数据库；证据：`docs/acceptance/evidence/MVP/backup-recovery-drill.md`；门禁：恢复覆盖操作仍需独立运维 runbook 和 Product Owner 确认。
- [x] 4.3 [Kimi；依赖：4.1；允许：`frontend/src/features/dashboard/**`、`frontend/src/api/**`、`frontend/src/i18n/**`、`frontend/tests/**`] 更新 Owner 备份 UI 展示包名/错误态；测试：Owner成功、Operator无入口、无 restore 控件、三语；证据：`docs/acceptance/evidence/MVP/backup-ui.md`。

## 5. 全量验收与发布决策

- [x] 5.1 [GLM；依赖：2.1、3.1、4.2；允许：测试、证据] 在隔离环境运行 Go fmt/vet/test/build、migrations 001–009 up/down/up、账务核对、备份恢复演练；输出：后端报告；证据：`docs/acceptance/evidence/MVP/backend-regression.md`；门禁：P0/P1=0、DB/附件 hash一致。
- [x] 5.2 [Kimi；依赖：2.2、3.2、4.3；允许：`output/playwright/mvp-full/**`、证据] 在真实 Vite 代理下重跑 M1–M6 全链路和三语浏览器回归；输出：截图、trace、HTML、JSON、报告；证据：`docs/acceptance/evidence/MVP/browser-regression.md`；门禁：禁止 `page.route` 拦截或直连绕过 `/lessons`、`/notifications`、`/dashboard`。2026-07-20：隔离数据环境下 20 组 × zh-CN/ja-JP/en-US = 60/60 通过，`unexpected=0`。
- [ ] 5.3 [Codex + Product Owner；依赖：5.1、5.2；允许：状态/路线图/追踪/证据] 汇总人工核账、Resend 测试收件人、备份恢复、Win10 JP 证据并执行 Go/No-Go；测试：治理 `RELEASE_GATES.md`；证据：`docs/acceptance/evidence/MVP/release-gate.md`；门禁：仅在所有条件通过后更新 M5/M6 为 ACCEPTED。
