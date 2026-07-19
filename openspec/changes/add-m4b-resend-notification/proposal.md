# M4b：Resend 通知 outbox-lite

## 为什么现在做

M4a 已经提供稳定 lesson 事实。M4b 以该事实为来源，通过 SQLite outbox 将邮件发送从业务事务中解耦，交付 MVP 必需的课程创建/取消通知与可追踪的发送结果。

## 范围

- `notification_outbox` 迁移与唯一幂等键；
- lesson 创建和取消时，与 lesson/audit 同一事务写入待发送记录；
- Resend HTTP API sender（仅读取环境变量），领取、发送、成功/失败状态和最多三次尝试；
- Owner/Operator 可查询日志、手动处理队列或重试失败记录；
- 最小前端通知日志入口与三语状态词条。

## 非目标

- 不在请求事务内调用 Resend；
- 不发送真实邮件作为自动化测试；
- 不实现 SMTP fallback、可编辑模板、定时课前提醒、晨报、Webhook 送达回执、考勤、扣课、财务或老师结款。
