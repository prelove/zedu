## Context

M1–M6 已有的 `lesson`、`notification_outbox`、`teacher_account_ledger`、`student_course_enrollment` 和受控附件存储可以承载 PRD §23.2 的遗漏能力。当前缺口不是重新定义账务，而是让已落库事实可被运营者查看、让通知/备份成为可验证的运营操作。正式老师结款仍是 V1，恢复仍不能暴露 HTTP。

## Goals / Non-Goals

**Goals:**

- 以只读、可审计方式展示老师待付金额和工作台运营计数。
- 基于现有 outbox 提供幂等课前提醒和固定间隔、有三次上限的失败自动重试；手动重放保持显式。
- 产生可验证、可携带的本地备份包，并通过命令行恢复演练验证数据库、附件和 manifest，而不向 Web 暴露恢复。
- 对所有新增 UI 同步 zh-CN/ja-JP/en-US，并在真实 Vite 代理中覆盖浏览器路径。

**Non-Goals:**

- 结款、退款、人工调账、报表、图表、冲突检测、SMTP fallback、晨报/周报、可配置提醒规则、移动端、云备份和 HTTP 恢复。

## Decisions

### 1. 老师应付由不可变流水聚合，不创建新余额真相

`teacher_account_ledger` 是唯一应付事实源。只读查询按 teacher 聚合 `amount_delta`，并可按老师查看课次明细。不会创建 `payout` 表或“已结算”状态。

**Rejected:** 先做最小结款按钮 | 会越过 PRD §23.2 的“暂不做正式结款”边界并改变账务语义。

### 2. 提醒仍使用 outbox，调度由显式受控 runner 执行

新增 `LESSON_REMINDER` event type。扫描只选择处于 SCHEDULED、开始时间落入固定提醒窗口的课次；idempotency key 包含 lesson、event 和收件人。runner 每次执行扫描、入队并处理可发送任务；FAILED 记录在 `available_at` 之后重试，三次后不再自动发送。服务启动时不隐式发送，管理员通过受控命令/计划任务运行 runner，HTTP 仅保留现有人工处理入口。

**Rejected:** 在每个 HTTP 请求中后台扫描 | 请求副作用不可预测且无法可靠地保证周期。
**Rejected:** 引入新的 scheduler 依赖 | MVP 使用现有 Go 标准库与平台计划任务即可。

### 3. 备份包必须用 staging 后原子发布

备份先在 `ZEDU_BACKUP_DIR/.tmp/<timestamp>-<nonce>/` 生成 SQLite 快照、复制 `ZEDU_DATA_ROOT/uploads/`、写入仅含非敏感配置摘要的 manifest，并对每个文件写 SHA-256。验证完成后原子 rename 到不可变目录；若任一阶段失败，删除 staging，不写成功审计。`ZEDU_JWT_SECRET`、Resend key 及任何 token 绝不进入包或 manifest。

**Rejected:** 单一 `.db` 文件 | 不符合 PRD §20/§23.2 的附件与 manifest 要求。
**Rejected:** HTTP restore | 数据恢复是高风险、需要停机/确认的运维动作。

### 4. 恢复演练是本地 CLI/测试路径，不改写活动实例

新增 `zedu-backup-verify` 命令或等价受控 package：将备份解压/复制到新的临时目录，先校验 manifest SHA-256，再打开恢复 SQLite 并验证附件路径和计数。它不覆盖运行中数据库；实际恢复仍需受控 runbook 和 Product Owner 确认。

### 5. 工作台使用单个只读聚合响应

`GET /dashboard` 增加 today lessons、待确认、待续费、教师应付、失败通知。待续费以 `lesson_balance <= 0` 表示，避免在 MVP 猜测价格阈值；所有计数从已有表实时查询，不新增缓存或写入。

## Risks / Trade-offs

- [Outbox runner 未被定期执行] → 运维 runbook 和浏览器/后端证据必须明确执行命令；V1 再引入完整自动化运营。
- [备份中附件复制失败] → staging、manifest 校验和失败清理；不生成成功审计。
- [本地恢复误操作] → verifier 仅写临时目录，生产恢复无 HTTP 路由且需要独立 runbook。
- [老师应付被误认为可结款] → 前端明确“待付事实”，没有任何结款、导出或调整入口。

## Migration Plan

1. 新增 migration，仅扩展 notification event/status 所需约束；up/down/up 覆盖。
2. 部署前配置 `ZEDU_DATA_ROOT`、`ZEDU_BACKUP_DIR`、Resend 发送环境变量和平台计划任务。
3. 在隔离数据运行备份→verify drill→哈希/附件/数据库核对。
4. 若回滚，停用 runner 并回退应用；已有 outbox/备份包保留为可审计历史，不删除业务事实。

## Open Questions

- 提醒窗口固定为课前 30 分钟，作为 MVP 默认值；无需配置 UI。若 Product Owner 需要其他窗口，应作为 V1 规则管理变更。
- `lesson_balance <= 0` 作为待续费的 MVP 口径；金额余额阈值不在本变更中推断。
