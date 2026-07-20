# 老师应付与工作台只读计数 — 后端证据

## 范围

实现 PRD §23.2 范围内的只读老师应付与工作台运营指标 API，禁止结款/退款/导出/调账。

## 实现文件

- `backend/internal/app/payable/handler.go` — `GET /teachers/payable`（汇总）、`GET /teachers/{id}/payable`（明细）
- `backend/internal/app/dashboard/handler.go` — 扩展 `GET /dashboard`，新增 5 个只读字段
- `backend/cmd/zedu-server/main.go` — 路由挂载

## API 契约

### `GET /teachers/payable`
返回所有老师的应付汇总列表，按 `teacherId` 升序：

```json
{ "code": 0, "data": { "items": [
  { "teacherId": 1, "teacherName": "Sensei", "payableAmount": 3000, "currency": "JPY", "lessonCount": 3 }
] } }
```

### `GET /teachers/{id}/payable`
返回单个老师的应付明细，按 `lessonId` 升序：

```json
{ "code": 0, "data": {
  "teacherId": 1, "teacherName": "Sensei", "currency": "JPY",
  "items": [ { "lessonId": 10, "lessonNo": "L-0010", "amount": 1000, "balanceAfter": 1000, "createdAt": "2026-07-20T10:00:00Z" } ]
} }
```

### `GET /dashboard`（扩展）
新增字段（原 `todayLessons`、`pendingConfirmations` 保留）：

| 字段 | 类型 | 说明 |
|---|---|---|
| `todayLessons` | int | 今日（UTC 当天）`SCHEDULED` 课次 |
| `pendingConfirmations` | int | `SCHEDULED` 且 `scheduled_start_at <= now`（已到点待确认） |
| `renewalNeeded` | int | 剩余课时 ≤ 阈值的活跃报名 |
| `teacherPayable` | int | 所有老师应付总额（本位币整数） |
| `failedNotifications` | int | `notification_outbox.status='FAILED'` 行数 |

## 数据来源

- 老师应付：从 `teacher_account_ledger` 聚合 `entry_type='LESSON_PAY'` 的金额（整数，无 float）
- 工作台计数：`lesson`、`enrollment`、`notification_outbox` 表的只读 `COUNT`/`SUM`

## 测试

`backend/internal/app/payable/handler_test.go`、`backend/internal/app/dashboard/extended_test.go`：

- 未认证 → 40101
- Owner/Operator 均可读取
- 空数据返回零值
- 确认课次后金额正确
- 读操作零副作用（lesson 状态不变）
- 无 `/payout`、`/settlement` 路由

## 运行结果

```
go test ./internal/app/payable/... ./internal/app/dashboard/...
ok  payable   PASS
ok  dashboard PASS
```

## 门禁

- 禁止结款/退款/导出/调账：无 POST/PATCH/DELETE 路由
- 禁止绕过 RBAC：所有路由经 `AuthMiddleware`
- 禁止金额 float：所有金额字段为 `int64`
