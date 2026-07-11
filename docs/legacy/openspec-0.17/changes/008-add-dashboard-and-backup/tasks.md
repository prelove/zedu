## 1. 工作台聚合接口

- [ ] 1.1 编写失败测试：GET /reports/dashboard 返回今日课程列表
      文件：backend/internal/report/dashboard_test.go
- [ ] 1.2 实现最小代码使测试通过
      文件：backend/internal/report/service.go, handler.go
- [ ] 1.3 编写失败测试：跨日边界时，今日课程判断应按系统配置时区
      而非UTC判断
- [ ] 1.4 实现时区转换逻辑使测试通过
- [ ] 1.5 编写失败测试：返回lesson_balance≤配置阈值或balance_amount
      不足的enrollment列表
- [ ] 1.6 实现代码使测试通过
- [ ] 1.7 编写失败测试：返回unpaid_amount>0的老师列表（通过ledger
      聚合计算）
- [ ] 1.8 实现代码使测试通过
- [ ] 1.9 提交：git commit -m "feat(report): minimal dashboard aggregation with timezone-aware date logic"

## 2. 手动备份接口

- [ ] 2.1 编写失败测试：触发备份后backup/目录生成一个.db文件，
      命名符合zedu_{YYYYMMDD}_{HHmmss}.db格式
      文件：backend/internal/backup/service_test.go
- [ ] 2.2 实现最小代码使用VACUUM INTO方式使测试通过
      文件：backend/internal/backup/service.go, handler.go
- [ ] 2.3 编写失败测试：生成的备份文件可独立打开且表数量、关键表
      记录数与主库一致
- [ ] 2.4 实现验证逻辑使测试通过（备份后自动做一次一致性校验）
- [ ] 2.5 编写失败测试：备份操作无论成功失败都在backup_log留一条记录
- [ ] 2.6 实现代码使测试通过
- [ ] 2.7 编写失败测试：模拟VACUUM INTO因磁盘空间不足等原因失败，
      backup_log应记录status=FAILED和error_msg
- [ ] 2.8 实现错误处理使测试通过
- [ ] 2.9 提交：git commit -m "feat(backup): manual backup via VACUUM INTO with consistency check and logging"

## 3. 前端：工作台与备份设置

- [ ] 3.1 工作台页面（极简版，三个列表卡片）
- [ ] 3.2 系统设置"数据备份"Tab，"立即备份"按钮+备份历史列表
- [ ] 3.3 提交：git commit -m "feat(frontend): minimal dashboard and backup settings"

## 4. MVP整体验收（Checkpoint，不是Task，是独立场景测试）

- [ ] 4.1 完整走一遍：新建学生→新建老师→配置课程→新建学生课程报名
      →绑定老师→录入充值(含非本位币折算)→创建课次→执行课后确认
      (至少测试"正常上课"和"学生请假"两种出勤分类)→查看学生流水
      →查看老师应付流水→查看单课财务(通过数据库直接查lesson_finance)
      →查看工作台汇总→触发一次手动备份并验证备份文件可用
- [ ] 4.2 账务人工核对：至少3笔课后确认，用计算器验证余额变化数字
      完全正确
- [ ] 4.3 尝试在产生财务记录后修改本位币，验证被拒绝
- [ ] 4.4 尝试对同一课次重复确认，验证被拒绝且提示友好

**通过标准**：以上全部打勾，且中途不需要手动改数据库或重启服务来
"绕过"某个卡点。通过后可以让运营者开始用真实数据小范围试用，
同时启动Sprint 8（对应change 009）。

## 5. 规格场景覆盖检查表

对照本change下specs/dashboard/spec.md和specs/backup/spec.md的全部
Scenario，逐条标注验证task（本节独立于第4节的MVP整体验收场景测试，
第4节是端到端场景走查，本节是spec文件的场景级追溯）：

- [ ] 5.1 「查询工作台数据」→ 1.1-1.2、1.5-1.8
- [ ] 5.2 「今日课程按系统时区判断」→ 1.3-1.4
- [ ] 5.3 「待续费口径与余额预警任务一致」→ 1.5-1.6
- [ ] 5.4 「触发备份」→ 2.1-2.2
- [ ] 5.5 「备份记录留痕」→ 2.5-2.6
- [ ] 5.6 「备份文件数据一致性验证」→ 2.3-2.4

全部勾选（含第4节MVP整体验收）后才可执行
`/opsx:archive add-dashboard-and-backup`。
