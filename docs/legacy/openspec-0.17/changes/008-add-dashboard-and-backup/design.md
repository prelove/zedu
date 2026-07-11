# Design: 工作台聚合查询与备份机制

## 工作台数据聚合的查询设计

三类列表分别对应三个独立查询，不强求合并成一个巨大的SQL：
1. 今日课程：`WHERE DATE(scheduled_start_at) = DATE(NOW())`（注意时区
   转换，应该按系统配置时区判断"今天"而非UTC的今天，否则在日本时间
   凌晨时段可能出现日期判断偏差）
2. 待续费学生：`WHERE lesson_balance <= 配置阈值 OR balance_amount <
   charge_per_lesson_amount`（与010余额预警任务的判断条件保持一致，
   避免工作台显示的名单和邮件预警的名单出现不一致的观感）
3. 待结款老师：通过teacher_account_ledger聚合计算每个老师的
   unpaid_amount（= LESSON_PAYABLE的amount_delta总和 - PAYOUT的
   amount_delta绝对值总和），筛选出>0的老师

极简版工作台不需要为这些聚合查询做额外的物化视图或缓存表，V1数据
量级下直接查询即可，性能不是这个阶段的关注点。

## VACUUM INTO备份机制

SQLite的`VACUUM INTO 'backup_path.db'`语法会创建一个全新的、经过
整理（无碎片）的数据库文件副本，且这个操作在SQLite内部是事务安全的
——即使备份过程中主库仍在被其他连接写入，VACUUM INTO也能保证生成
的文件是某个一致时间点的完整快照，不会出现"表结构对不上"或"外键
引用悬空"这类中间状态问题。这是选择VACUUM INTO而非操作系统层面
`cp`命令的根本原因。

备份文件命名规则：`zedu_{YYYYMMDD}_{HHmmss}.db`，存放于配置的
backup目录下（默认./backup/）。每次备份无论成功失败都在backup_log
表留一条记录，失败时记录error_msg，便于后续（012阶段）在晨报里
展示"最近备份状态异常"这类提醒。

## 备份验证

备份完成后，本change要求验证生成的文件"可被独立打开且数据与主库
一致"——具体验证方式是用一个独立的数据库连接打开备份文件，查询
表数量和某几张关键表（如student、lesson）的记录数，与主库对比
一致，这个验证逻辑本身也应该写成自动化测试，而不是仅靠人工用
SQLite客户端点开看一眼。
