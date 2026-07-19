# M4b 设计

## 事务边界

lesson 创建/取消在既有事务内写入 `notification_outbox(PENDING)`。出站 HTTP 永远发生在事务提交之后；发送失败不能回滚已提交的 lesson。

## 幂等与状态

`idempotency_key` 唯一，格式为 `lesson:{id}:{event}:{recipient}`。状态为 `PENDING`、`PROCESSING`、`SENT`、`FAILED`。处理器原子领取一个记录；每条最多三次。手动重试仅将 `FAILED` 重置到 `PENDING`，不直接发送。

## 配置与安全

读取 `ZEDU_RESEND_API_KEY` 与 `ZEDU_RESEND_FROM`；缺失时处理器返回稳定 50001，绝不在日志/API/审计中泄露 key。测试使用内存 fake sender，禁止真实 Resend 调用。

## 收件人与内容

仅收集 lesson 对应学生和家长的非空 email；无 email 时不建 outbox 行。默认 locale 为 `ja-JP`，邮件使用固定纯文本转义 HTML 内容，包含 lesson 编号、UTC 时间和时区，不含财务或认证数据。
