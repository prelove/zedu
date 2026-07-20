# MVP 实装覆盖审计

审计日期：2026-07-20
审计基线：`main`，以 PRD v3.1 与 M1–M6 已冻结 OpenSpec 的 Requirement/Scenario 为准。

## 结论

MVP 的已批准业务能力均已有后端路由、前端入口或明确的受控非暴露决定；不存在需要新增业务功能的缺口。MVP 当前处于 **IMPLEMENTED / PENDING_RELEASE_VERIFICATION**，而不是 ACCEPTED：M5、M6 的最后一项均是独立回归、人工 UAT 与 Go/No-Go，不是待编码功能。

M4 浏览器回归发现的开发代理缺口（`/lessons`、`/notifications`）已在 `frontend/vite.config.ts` 补齐；新增 `/dashboard` 代理以覆盖 M6。此前 M4 的 18/18 用例可作为行为证据，但涉及 `page.route` 的部分不作为真实 Vite 代理证据，必须在本次全量回归中以未拦截请求重跑。

## 能力覆盖

| 里程碑 | 已实现能力 | 关键实现位置 | 状态 |
|---|---|---|---|
| M1 | Go/Vue 工程骨架、SQLite migration、统一响应、日志脱敏、三语基础 | `backend/internal/platform/**`、`frontend/src/i18n/**` | 已实装 |
| M2 | 登录/刷新/登出、RBAC、Owner 初始化、学生/家长/老师、课程字典、报名与师生安排 | `backend/internal/app/{auth,onboarding,directory,course}`、对应前端 feature | 已实装 |
| M3 | 本位币、支付方式、充值、学生流水、作废冲正、受控凭证上传/下载 | `backend/internal/app/{finance,evidence}`、`frontend/src/features/finance` | 已实装 |
| M4a | 课次创建、UTC/时区、查询、修改、取消、审计 | `backend/internal/app/lesson`、`frontend/src/features/lesson` | 已实装 |
| M4b | 课次创建/取消通知 outbox、Resend sender、失败重试、查看与处理 | `backend/internal/app/notification`、`frontend/src/features/notification` | 已实装 |
| M5 | 出勤结果、一次性课后确认、attendance/学生流水/老师应付/课次财务快照同事务、小数课时 | `backend/internal/app/attendance`、`frontend/src/features/lesson/LessonsView.vue` | 已实装 |
| M6 | 只读工作台、Owner SQLite 备份与审计、无 HTTP 恢复入口 | `backend/internal/app/{dashboard,backup}`、`frontend/src/features/dashboard` | 已实装 |

## 明确不属于 MVP 的能力

- 正式老师结款、退款、人工调账、报表、复杂排课冲突检测、自动提醒、PWA/移动端。
- HTTP 恢复数据库操作；MVP 仅允许 Owner 创建受控备份，恢复留给受控运维/UAT。

## 验收前阻断条件

1. 全量后端/前端门禁必须通过；Linux race 由 CI 负责。
2. Kimi 必须以真实 Vite 代理重跑 M1–M6，不得为 `/lessons`、`/notifications`、`/dashboard` 设置 `page.route` 转发。
3. M5 必须确认 0.5 课时的 ledger API 可读，且负数/非法精度不写入事实。
4. Owner/Operator/未认证三种权限、zh-CN/ja-JP/en-US 三语、备份的 Owner-only 行为必须有新鲜浏览器证据。
