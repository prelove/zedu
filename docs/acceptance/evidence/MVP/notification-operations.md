# 课前提醒与失败重试 — 后端证据

## 范围

扩展 outbox 支持 `LESSON_REMINDER`、固定 30 分钟窗口扫描、三次上限和延迟重试。

## 实现文件

- `backend/migrations/009_m4b_lesson_reminder.up.sql` — 扩展 `event_type` CHECK 约束
- `backend/migrations/009_m4b_lesson_reminder.down.sql` — 回滚（删除 `LESSON_REMINDER` 行）
- `backend/internal/app/notification/runner.go` — `ReminderRunner`、`ScanReminders`
- `backend/cmd/zedu-reminder/main.go` — 独立 CLI（不启动 HTTP 服务）

## 设计

### 30 分钟窗口扫描

`ReminderRunner.ScanReminders` 查询 `lesson` 表中 `status='SCHEDULED'` 且 `scheduled_start_at` 在 `[now, now+30min]` 区间的课次，对每个收件人（学生 + 家长邮箱）插入 `LESSON_REMINDER` outbox 行。

### 幂等

`idempotency_key = "lesson:{lessonID}:LESSON_REMINDER:{email}"` 唯一约束 + `INSERT OR IGNORE`。重复扫描同一课次不会产生重复 outbox 行。

### 三次上限延迟重试

复用现有 `ClaimAndSend`：
- `WHERE status IN ('PENDING','FAILED') AND attempts<3 AND available_at<=CURRENT_TIMESTAMP`
- 失败后 `available_at = CURRENT_TIMESTAMP + 5 minutes`（延迟重试）
- `attempts` 达到 3 后不再被选中（人工排查）

### 不在 HTTP 请求或服务启动时隐式发送

提醒扫描与发送仅在 `zedu-reminder` CLI 中显式触发，`zedu-server` 启动时不调用。

## 测试

`backend/internal/app/notification/runner_test.go`：

| 测试 | 验证 |
|---|---|
| `TestReminderScanIsIdempotent` | 重复扫描无重复 outbox 行 |
| `TestReminderScanWindowBoundary` | 窗口外课次不入队 |
| `TestReminderScanSkipsNonScheduledLessons` | `CANCELLED` 课次不入队 |
| `TestFailedNotificationRetriesAfterAvailableAt` | 未来 `available_at` 不重试；过去则重试；课次状态不变 |
| `TestNotificationRetryStopsAtThree` | `attempts=3` 后不再被选中 |
| `TestNotificationErrorIsSanitized` | `last_error` 不含原始错误细节 |

## 运行结果

```
go test ./internal/app/notification/... -v
=== RUN   TestClaimAndSendRecordsSuccessAndFailure --- PASS
=== RUN   TestOutboxIdempotencyKeyIsUnique --- PASS
=== RUN   TestReminderScanIsIdempotent --- PASS
=== RUN   TestReminderScanWindowBoundary --- PASS
=== RUN   TestReminderScanSkipsNonScheduledLessons --- PASS
=== RUN   TestFailedNotificationRetriesAfterAvailableAt --- PASS
=== RUN   TestNotificationRetryStopsAtThree --- PASS
=== RUN   TestNotificationErrorIsSanitized --- PASS
PASS
ok  notification  14.920s
```

## Migration 验证

`009_m4b_lesson_reminder.up.sql` 通过重建表扩展 CHECK 约束（SQLite 不支持 ALTER CHECK），保留所有现有数据。

## 门禁

- 不在 HTTP 请求或服务启动时隐式发送：仅 CLI 触发
- 不新增依赖：复用现有 `database`、`repository` 包
- 失败不回滚课次：`ScanReminders` 仅读 `lesson` 表
- 日志脱敏：`sanitize()` 将 `last_error` 替换为 "delivery failed"
