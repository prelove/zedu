# M3 执行看板：充值、学生流水与付款凭证

> OpenSpec change：`add-m3-recharge-ledger-evidence`。共享进度仅由 Codex/PM 更新；状态流遵循 `BACKLOG → READY → IN_PROGRESS → IN_REVIEW → VERIFIED → ACCEPTED`。

| 工单 | Owner | 范围 | 状态 | 前置条件 | 验收者 |
|---|---|---|---|---|---|
| M3-CODEX-01 | Codex/PM | 契约冻结、任务拆分与追踪 | VERIFIED | M2 后端已验收 | Codex/PM |
| M3-GLM-03A | Codex（暂代） | 005 migration、本位币、支付方式字典 | IN_PROGRESS | M3-CODEX-01 | Codex + 独立 Reviewer |
| M3-GLM-03B | GLM | 充值、流水、作废的精确账务事务 | BLOCKED | GLM-03A ACCEPTED | Codex + 独立 Reviewer |
| M3-GLM-03C | GLM | 付款凭证上传、鉴权下载、文件补偿 | BLOCKED | GLM-03B ACCEPTED | Codex + 独立 Reviewer |
| M3-KIMI-03A | Kimi | 配置 adapter 与 Owner 配置页 | IN_REVIEW | 后端契约已提供但 GLM-03A 尚未 ACCEPTED | 已交付 focused gate；不得因提前实现而绕过后端独立验收 | Codex + 独立 Reviewer |
| M3-KIMI-03B | Kimi | 充值、流水、作废与凭证 UI | BLOCKED | GLM-03B/03C ACCEPTED | Codex + 独立 Reviewer |
| M3-CODEX-02 | Codex | HTTP/浏览器/质量批次/Release Gate | BLOCKED | 所有 M3 实施工单 ACCEPTED | 独立 Reviewer |

## 固定边界

- 本轮仅包含本位币状态、支付方式、学生充值、学生流水、充值作废和付款凭证。
- 正式老师结款、老师流水、退款、人工调整、课次/出勤、报表、通知、备份与恢复均不在 M3；不得创建 API、路由、菜单、隐藏开关或占位实现。
- 后端遵循 `HTTP → application → repository`；金额禁止 float；确认充值/作废必须同一数据库事务；所有写操作均有审计。
- 三语、UTF-8、Windows JP 兼容、既有 HTTP 信封与错误码是不可变基线；不得新增依赖或错误码。

## 延后质量门禁

当前 Windows 前端全量 Vitest 有未完成的收尾运行，已登记到 `M3-CODEX-02 / tasks 5.2`，不会阻塞 M3 的 focused 红绿测试、typecheck、lint 或后端实施。它也不构成 M2/M3 ACCEPTED 证据。
