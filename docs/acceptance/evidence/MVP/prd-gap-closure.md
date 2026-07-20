# PRD §23.2 MVP 缺口闭合追踪矩阵

变更：`complete-prd-mvp-release-gaps`
基线：`main` @ `81e800a`，PRD v3.1 §23.2、§9.4、§9.5、§20、§21、§24
角色：Codex（契约）/ GLM（后端）/ Kimi（前端 + 浏览器回归）
门禁：不把正式结款、报表、冲突检测、SMTP fallback、移动端、HTTP 恢复纳入 MVP 实现。

## 1. PRD §23.2 与现有 M1–M6 范围差异

| PRD §23.2 要求 | 现有 M1–M6 状态 | 本变更闭合动作 | 实现位置 |
|---|---|---|---|
| 老师应付只读查询 | `teacher_account_ledger` 已写入但无查询 API/UI | 新增只读 summary/detail API 与老师详情页展示 | `backend/internal/app/payable/**`、`frontend/src/features/directory/**` |
| 课前提醒 | outbox 仅支持 `LESSON_CREATED/CANCELLED`，无调度扫描 | 新增 `LESSON_REMINDER` event、固定 30 分钟窗口扫描 runner、幂等键 | `backend/internal/app/notification/**`、`backend/cmd/zedu-reminder/**`、migration 009 |
| 失败通知有界自动重试 | 仅人工 `POST /notifications/outbox/{id}/retry` | runner 在 `available_at` 后自动重试，三次上限，失败不回滚课次 | `backend/internal/app/notification/runner.go` |
| 可携带备份包（DB + 附件 + manifest） | 仅 `VACUUM INTO` 单一 `.db` 文件 | staging→SQLite 快照 + uploads 副本 + 非敏感配置摘要 + SHA-256 manifest，原子 rename | `backend/internal/app/backup/**` |
| 本地恢复演练 | 无 | 新增 `zedu-backup-verify` CLI：校验 manifest、解压到临时目录、不覆盖活动 DB | `backend/cmd/zedu-backup-verify/**` |
| 工作台最小运营汇总 | 仅 `pendingLessonConfirmations` + `failedNotifications` | 扩展为今日课程、待确认、待续费、老师应付、失败通知五项只读计数 | `backend/internal/app/dashboard/**`、`frontend/src/features/dashboard/**` |

## 2. 路由矩阵（MVP 完整暴露面）

### 2.1 后端 HTTP 路由

| 方法 | 路径 | 权限 | 能力 | 备注 |
|---|---|---|---|---|
| POST | `/auth/login` | public | M1 | 40102 登录失败 |
| POST | `/auth/refresh` | cookie | M1 | |
| POST | `/auth/logout` | auth | M1 | |
| GET | `/auth/me` | auth | M1 | |
| POST | `/onboarding/initialize` | OWNER | M2 | |
| GET/POST/PATCH | `/students`, `/students/{id}`, `/students/{id}/parents[/{pid}]` | auth | M2 | |
| GET/POST/PATCH | `/teachers`, `/teachers/{id}`, `/teachers/{id}/capabilities[/{cid}]`, `/teachers/{id}/availability[/{aid}]` | auth | M2 | |
| GET/POST | `/course-domains`, `/tracks`, `/levels`, `/capability-tags` | auth | M2 | |
| GET/POST/PATCH | `/enrollments/{id}`, `/enrollments/{id}/assignments[/{aid}]` | auth | M2 | |
| GET/POST/PATCH | `/system/payment-methods`, `/finance/payments[/{id}]`, `/finance/payments/{id}/void`, `/finance/students/{id}/ledger` | auth/OWNER | M3 | |
| POST/GET | `/finance/payments/{id}/attachments[/{aid}]` | auth | M3 | 凭证 |
| GET/POST | `/system/attendance-outcomes`, `/lessons`, `/lessons/{id}`, `/lessons/{id}/cancel`, `/lessons/{id}/confirm` | auth | M4a/M5 | |
| PATCH | `/lessons/{id}` | auth | M4a | |
| GET/POST | `/notifications/outbox`, `/notifications/outbox/process`, `/notifications/outbox/{id}/retry` | auth | M4b | 人工重放保留 |
| GET | `/dashboard` | auth | M6 | 本变更扩展为五项只读计数 |
| POST | `/system/backups` | OWNER | M6 | 本变更为备份包 |
| **GET** | **`/teachers/payable`** | **auth** | **新增** | 老师应付 summary（分页） |
| **GET** | **`/teachers/{id}/payable`** | **auth** | **新增** | 老师应付 detail（按课次） |

### 2.2 后端 CLI 路由（非 HTTP）

| 命令 | 作用 | 权限 | 备注 |
|---|---|---|---|
| `zedu-server` | HTTP 服务 | 运维 | 不隐式启动 reminder runner |
| `zedu-reminder` | 扫描 SCHEDULED 课次 + 入队 LESSON_REMINDER + 处理可发送 outbox | 运维（计划任务） | 本变更新增 |
| `zedu-backup-verify` | 校验备份包 manifest 并恢复到临时目录 | 运维 | 本变更新增，不暴露 HTTP |

### 2.3 前端路由

| 路径 | 权限 | 能力 | 备注 |
|---|---|---|---|
| `/login` | public | M1 | |
| `/` | auth | M1 | |
| `/onboarding` | OWNER | M2 | |
| `/students`, `/students/{id}` | auth | M2 | 老师详情新增只读应付区段 |
| `/teachers`, `/teachers/{id}` | auth | M2 | |
| `/courses` | auth | M2 | |
| `/finance/config` | OWNER | M3 | |
| `/finance/payments` | auth | M3 | |
| `/lessons` | auth | M4a/M5 | |
| `/notifications` | auth | M4b | 本变更新增提醒/失败状态可视化 |
| `/dashboard` | auth | M6 | 本变更新增五项运营计数 + 备份包名/错误态 |

## 3. 负向范围（明确不实现）

| 不实现项 | 原因 | 防护 |
|---|---|---|
| 正式老师结款 / payout 表 / 已结算状态 | PRD §23.2 明确 V1 | 无 `payout` 路由、无 `settle` action、前端无结款按钮 |
| 退款 / 人工调账 | 非 MVP | 无 `refund`、`adjust` 路由 |
| 复杂报表 / 图表 | 非 MVP | 工作台仅 5 个数字，无图表组件 |
| 排课冲突检测 | 非 MVP | lesson 创建不做冲突检测 |
| SMTP fallback | 非 MVP | 仅 Resend sender |
| 晨报 / 周报 / 提醒规则配置页 | 非 MVP | reminder 窗口固定 30 分钟，无配置 UI |
| 移动端 / PWA | 非 MVP | 仅 Desktop Chrome |
| HTTP 恢复 | 高风险运维动作 | 无 `POST /system/backups/restore`，仅本地 CLI |
| 云备份 | 非 MVP | 备份包仅本地目录 |
| 导入导出 | 非 MVP | 无 export 路由 |

## 4. 验收门禁

- 后端：`go fmt` / `go vet` / `go test ./...` / `go build ./...` 全通过；migration 001–009 up/down/up 可重复；账务核对（teacher_account_ledger 聚合 = 工作台应付）；备份恢复演练哈希一致。
- 前端：三语 key parity；无结款/退款/restore 入口；真实 Vite 代理浏览器回归。
- 证据：本目录下 `payable-dashboard.md`、`payable-dashboard-ui.md`、`notification-operations.md`、`notification-ui.md`、`backup-package.md`、`backup-recovery-drill.md`、`backup-ui.md`、`backend-regression.md`、`browser-regression.md`、`release-gate.md`。

## 5. OpenSpec strict 校验

本变更的 specs 已通过 `openspec validate`；ADDED/MODIFIED Requirements 与 Scenario 与上表一一对应。
