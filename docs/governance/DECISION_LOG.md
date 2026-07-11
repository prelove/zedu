# 决策记录

| ID | 日期 | 决策 | 状态 | 决策人 | 最晚复核 |
|---|---|---|---|---|---|
| ADR-001 | 2026-07-11 | 先验收 MVP，再继续 V1；总路线图持续追踪全部阶段 | ACCEPTED | Product Owner | M6 |
| ADR-002 | 2026-07-11 | Resend 通知与付款凭证进入 MVP | ACCEPTED | Product Owner | M4b/M3 |
| ADR-003 | 2026-07-11 | 正式老师结款留到 V1，MVP 无可执行入口 | ACCEPTED | Product Owner | V1启动前 |
| ADR-004 | 2026-07-11 | 旧 OpenSpec 冻结，按业务能力迁移到 1.6 | ACCEPTED | PM/Architect | M0 |
| ADR-005 | 2026-07-11 | 通知使用 SQLite outbox-lite；供应商接受不等于送达 | ACCEPTED | PM/Architect | M4b |
| ADR-006 | 2026-07-11 | 7天是目标，不得越过质量门禁 | ACCEPTED | Product Owner | 每周 |

## 暂定运营参数

| 项目 | 暂定值 | Owner | 状态 | 截止门禁 | 影响 |
|---|---|---|---|---|---|
| 本位币 | JPY | Product Owner | PROVISIONAL | M3 | 首笔财务事实后锁定 |
| 初始化模板 | 日语培训 | Product Owner | PROVISIONAL | M2 | 种子数据 |
| 凭证限制 | 5MB/文件、3个/充值 | Product Owner | PROVISIONAL | M3 | 上传与备份 |
| 通知语言 | 用户locale；缺失时ja-JP | Product Owner | PROVISIONAL | M4b | 模板与收件人体验 |
| Resend sender/test inbox | 待提供安全配置 | Product Owner | OPEN | M4b | 真实 smoke 测试 |
| 备份位置/保留 | 本地、30天 | Product Owner | PROVISIONAL | M6 | 容量与恢复 |
