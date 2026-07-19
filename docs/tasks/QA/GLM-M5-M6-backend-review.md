# GLM：M5/M6 后端独立验收工单

基线：`main` `8559558`。只读审查与测试；不得修改、提交或推送。

## M5 必测

1. `POST /lessons/{id}/confirm`：Owner/Operator、SCHEDULED 仅一次、attendance/lesson_finance/学生流水/老师应付/lesson COMPLETED/audit 同事务。
2. 余额不足失败后 attendance、finance、ledger、lesson 状态均无半写；两个并发确认仅一个成功。
3. 实际值与建议值同时保存；出勤字典更新不改变历史 attendance。
4. 不存在自动确认、正式结款、退款/调整、通知副作用。

## M6 必测

1. `/dashboard` 未认证 40101，认证后只读返回待确认与失败通知数。
2. `POST /system/backups`：Operator 40301、Owner 仅在 `ZEDU_BACKUP_DIR` 配置时成功；生成有效 SQLite 文件与 BACKUP_CREATE 审计。
3. 没有 restore HTTP 路由；日志/API 不泄露 DB 路径、JWT 或 Resend 配置。

执行 Go 全量测试、vet、build；报告 P0/P1/P2、可复现命令、验证证据与未测项至隔离输出，不提交。
