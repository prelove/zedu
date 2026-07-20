# 老师应付与工作台 — 前端证据

## 范围

展示工作台最小运营指标和老师详情的只读待付事实（三语）。

## 实现文件

- `frontend/src/api/dashboard.ts` — 扩展 5 字段
- `frontend/src/api/payable.ts` — 新增 `getTeachersPayable`、`getTeacherPayableDetail`
- `frontend/src/features/dashboard/DashboardView.vue` — 5 个运营指标 + 备份创建按钮
- `frontend/src/features/directory/components/PayableSection.vue` — 只读应付明细
- `frontend/src/features/directory/TeacherDetailView.vue` — 集成 PayableSection
- `frontend/src/i18n/locales/{zh-CN,ja-JP,en-US}.ts` — 新增 `dashboard`、`notifications`、`teachers.payable*` 键

## 工作台指标（Owner 可见备份按钮）

| i18n 键 | zh-CN | ja-JP | en-US |
|---|---|---|---|
| `dashboard.todayLessons` | 今日课程：{count} | 本日のレッスン：{count} | Today's lessons: {count} |
| `dashboard.pendingConfirmations` | 待确认课程：{count} | 確認待ちレッスン：{count} | Pending confirmations: {count} |
| `dashboard.renewalNeeded` | 待续费学员：{count} | 更新が必要な生徒：{count} | Renewals needed: {count} |
| `dashboard.teacherPayable` | 老师应付总额（本位币）：{amount} | 講師未払合計（基本通貨）：{amount} | Teacher payable total (base currency): {amount} |
| `dashboard.failedNotifications` | 发送失败通知：{count} | 送信失敗の通知：{count} | Failed notifications: {count} |

## 老师详情应付区（只读）

- 标题：`teachers.payableTitle` — "老师应付（只读）" / "講師未払（読み取り専用）" / "Teacher payable (read-only)"
- 提示：`teachers.payableHint` — 明确"仅查询不结款"
- 表头：`payableLessonNo`、`payableAmount`、`payableBalanceAfter`、`payableCreatedAt`
- 空态：`payableEmpty`
- 无结款按钮、无菜单、无写请求

## 测试

`frontend/tests/`：

- `tests/teachers-view.test.ts` — 15 tests PASS（PayableSection 防御性处理 undefined）
- `tests/dashboard-api.test.ts` — 1 test PASS
- `tests/i18n.test.ts`、`tests/i18n-m2.test.ts`、`tests/i18n-m2-kimi02.test.ts` — 三语 key parity PASS
- `tests/negative-scope.test.ts` — 无 `/payout`、`/settlement`、`/backup`、`/report` 路由

## 运行结果

```
npx vitest run
Test Files  39 passed (39)
     Tests  362 passed (362)
```

## 门禁

- 不在客户端计算账务事实：所有金额来自后端 `int` 字段
- 无结款按钮/菜单/请求：仅 `GET` 调用
- 三语 key parity：所有新增键在 zh-CN/ja-JP/en-US 均存在
