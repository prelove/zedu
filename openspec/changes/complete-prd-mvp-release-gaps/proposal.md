## Why

现有 M1–M6 实现已经覆盖认证、主数据、充值、基础排课、outbox 和课后确认主链路，但 M6 收口时把 PRD v3.1 第 23.2 章的若干 MVP 能力过度压缩：老师应付只写入未展示、通知没有课前调度和自动失败重试、备份只包含 SQLite 数据库、工作台只显示两个计数。这会让“确认后查看应付→备份并恢复演练”的 MVP 验收主链路无法完整走通。

本变更在进入 V1 前补齐这些明确属于 MVP 的最小运营能力，并维持“正式结款、复杂报表和自动排课不进入 MVP”的边界。

## What Changes

- 增加老师应付只读查询与最小前端展示；只展示 lesson confirmation 产生的不可变应付事实，不创建结款、调整或支付入口。
- 增加课前提醒的受控扫描与 outbox 入队，以及 FAILED 通知的有界自动重试；保留人工重放，外部邮件失败不得回滚课次事实。
- 将 Owner 手动备份扩展为受控备份包：SQLite、上传凭证、必要配置摘要、SHA-256 manifest；提供无 HTTP restore 的本地恢复演练命令/服务并写审计证据。
- 将工作台扩展为最小只读运营汇总：今日课程、待确认、待续费、老师应付、失败通知；不增加图表或报表。
- 将 M4/M5/M6 浏览器与后端回归更新为真实 Vite 代理、真实备份包和上述运营路径的发布证据。

## Capabilities

### New Capabilities

- `teacher-payable-view`: 查询并展示老师未结应付事实，不包含结款。
- `scheduled-notification-operations`: 课前提醒扫描、失败通知有界自动重试与可追踪 outbox 操作。
- `portable-backup-recovery-drill`: 包含数据库、凭证和 manifest 的本地备份包，以及受控恢复演练。

### Modified Capabilities

- `dashboard-backup`: 工作台从两个基础计数扩展为 PRD MVP 所需的只读运营汇总，备份要求改为可验证的备份包。

## Scope and Non-Goals

本变更只完成 PRD §23.2 的遗漏项，参考 §9.4、§9.5、§20、§21、§24。不实现正式老师结款、退款/人工调账、SMTP fallback、晨报/周报、提醒规则管理页面、冲突检测、报表/图表、导入导出、移动端、HTTP 恢复或云备份。

## Impact

- 后端：notification、backup、dashboard 新增受控服务/API（恢复演练仅本地命令）；新增 migration/测试仅在确需持久化提醒幂等标记时引入。
- 前端：老师详情/工作台/通知页面新增只读或显式手动操作入口，三语同步。
- 运维：需要 `ZEDU_RESEND_API_KEY`、已验证 `ZEDU_RESEND_FROM`、`ZEDU_BACKUP_DIR`、`ZEDU_UPLOAD_DIR`；真实发送仅使用测试收件人。
- 风险：备份与提醒会触及外部文件/邮件，必须通过隔离环境、manifest 哈希和恢复演练证明，不得使用真实业务数据。
