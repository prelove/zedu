## M4b Resend 通知 outbox-lite

- [x] 1. 新建 outbox migration、up/down/up 与幂等唯一约束测试。
- [x] 2. 在 lesson 创建/取消事务中写入收件人 outbox；无 email 不阻断排课。
- [x] 3. 实现 Resend sender、原子领取、发送结果、三次限制和失败清理。
- [x] 4. 实现通知日志、处理队列和失败重试 API，覆盖权限与并发领取。
- [x] 5. 实现最小前端日志页与重试入口。
- [x] 6. 通过 focused Go/前端门禁、OpenSpec strict，并写入 M4b 证据。
