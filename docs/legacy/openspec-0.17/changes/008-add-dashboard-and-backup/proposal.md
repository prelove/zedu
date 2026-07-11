# Proposal: 工作台（极简版）与手动备份

## Why
这是MVP收尾的最后两块拼图：一个能看到今日待办的极简工作台，
以及保证数据不丢失的手动备份能力。图表化的完整工作台留到012。
本change同时是MVP整体验收的落脚点——完成本change意味着可以让
运营者开始用真实数据小范围试用。

## 业务背景
根据PRD第十章10.2节，工作台是运营者每天打开系统第一眼看到的页面，
即使是极简版，也要能立刻回答三个问题："今天有哪些课"、"谁快没
课时了该催续费"、"哪个老师该结款了"。这三个问题对应的正是Excel时代
最容易被遗漏的三类信息——本change虽然不做图表，但这三个列表本身
就已经在交付"信息不再靠人脑记忆"这个核心价值。

备份的设计依据PRD第二十章20.3节，特别要注意"不能简单粗暴复制正在
写入的db文件"这条要求——SQLite文件在WAL模式下，直接cp文件可能会
拿到一个处于中间状态、不一致的快照，必须用SQLite官方支持的
VACUUM INTO语法来保证备份文件本身的完整性。

## What Changes
- 新增GET /reports/dashboard极简版（今日课程+待续费+待结款，不含图表数据）
- 新增POST /backup/trigger手动备份（VACUUM INTO方式）
- 新增backup_log记录每次备份的结果
- 新增工作台页面（列表代替图表）和系统设置备份Tab

## Non-Goals
- 不实现ECharts图表（留给012）
- 不实现自动定时备份（本change只做手动触发，Litestream云备份和
  每日自动备份的cron任务留给012-add-mobile-and-packaging对应章节，
  或视优先级在013中实现）
- 不实现备份文件的自动清理/保留策略（30天保留期，本change的备份
  文件会持续累积，清理逻辑留待后续按需实现）

## Impact
- Affected specs: dashboard（新增）、backup（新增）
- Affected code: backend/internal/report/、backend/internal/backup/、
  frontend/admin/src/views/dashboard/、views/system/backup/
- 依赖：005-add-lesson-scheduling、007-add-payment-and-ledger
- 特殊说明：本change完成后是MVP的整体验收点，验收清单见本文档
  tasks.md第4节
