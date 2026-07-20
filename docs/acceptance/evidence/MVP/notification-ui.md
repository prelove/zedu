# 通知提醒/失败状态 — 前端证据

## 范围

补充提醒/失败状态的只读可视化与人工重放提示（三语）。

## 实现文件

- `frontend/src/features/notification/NotificationsView.vue` — 扩展事件类型显示、状态徽章、人工重放提示
- `frontend/src/i18n/locales/{zh-CN,ja-JP,en-US}.ts` — 新增 `notifications` 键组

## 新增 i18n 键

| 键 | zh-CN | ja-JP | en-US |
|---|---|---|---|
| `notifications.reminder` | 课前提醒 | レッスン前リマインダー | Lesson reminder |
| `notifications.failed` | 发送失败 | 送信失敗 | Failed |
| `notifications.sent` | 已发送 | 送信済み | Sent |
| `notifications.pending` | 待发送 | 送信待ち | Pending |
| `notifications.manualRetryHint` | 失败通知可点击重试；超过三次上限的失败需人工排查后重放。 | 失敗通知は再試行できます。3回上限を超えた失敗は手動で調査して再送してください。 | Failed notifications can be retried. Failures exceeding the 3-attempt cap require manual investigation and replay. |

## 测试

- `tests/i18n.test.ts`、`tests/i18n-m2.test.ts`、`tests/i18n-m2-kimi02.test.ts` — 三语 key parity PASS
- `tests/negative-scope.test.ts` — 无规则配置页、SMTP、晨报、周报路由

## 门禁

- 不提供规则配置页：仅 `GET /notifications` 列表
- 不提供 SMTP 配置：无 SMTP 相关 UI 或 API
- 不提供晨报/周报：仅 `LESSON_*` 事件显示
- 无敏感信息：`lastError` 由后端 `sanitize()` 脱敏为 "delivery failed"
