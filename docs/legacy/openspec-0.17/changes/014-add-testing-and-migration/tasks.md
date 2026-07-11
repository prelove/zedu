## 1. 端到端测试环境准备

- [ ] 1.1 在预发布环境部署完整编译的二进制（非开发环境的go run，
      而是真实的dist/产物）
- [ ] 1.2 准备真实规模的测试数据（参考PRD目标规模：约100学生、
      20老师，比生产初期规模略大以留出观察余量）
- [ ] 1.3 提交：git commit -m "chore: staging environment setup for e2e testing"

## 2. 端到端测试执行

- [ ] 2.1 通过HTTP客户端（非直接调用service）逐项执行TC-01~TC-07，
      记录每项结果
      文件：docs/test-reports/tc-01-07-report.md
- [ ] 2.2 若发现问题，定位到具体change并修复，修复后重新执行该
      测试项直至通过
- [ ] 2.3 逐项执行TC-08~TC-11，记录结果
      文件：docs/test-reports/tc-08-11-report.md
- [ ] 2.4 修复发现的问题并重测
- [ ] 2.5 逐项执行TC-12~TC-14，记录结果
      文件：docs/test-reports/tc-12-14-report.md
- [ ] 2.6 修复发现的问题并重测
- [ ] 2.7 逐项执行TC-15~TC-17，记录结果
      文件：docs/test-reports/tc-15-17-report.md
- [ ] 2.8 修复发现的问题并重测
- [ ] 2.9 逐项执行TC-18~TC-21，记录结果
      文件：docs/test-reports/tc-18-21-report.md
- [ ] 2.10 修复发现的问题并重测
- [ ] 2.11 提交：git commit -m "test: full TC-01 to TC-21 end-to-end verification report"

## 3. 真实数据迁移

- [ ] 3.1 获取运营者提供的脱敏Excel（学生+老师）
- [ ] 3.2 使用012实现的导入接口执行正式导入，保存导入报告
      文件：docs/migration/import-report.md
- [ ] 3.3 人工核对导入报告中的跳过/未匹配行，与运营者逐条确认
      处理方式并完成手动补充
- [ ] 3.4 核对导入总数与原Excel记录数一致
- [ ] 3.5 随机抽取5~10个学生核对关键字段与原Excel一致
- [ ] 3.6 提交：git commit -m "chore: real data migration with verification"

## 4. 并行运行观察期

- [ ] 4.1 启动并行运行期（建议2周），运营者继续用Excel记录，
      同时在Zedu里同步录入
- [ ] 4.2 每周核对一次两边数据（学生数/课次数/账务汇总）是否一致，
      记录差异及原因
      文件：docs/migration/parallel-run-week1.md, week2.md
- [ ] 4.3 运营者确认可以完全切换到Zedu

## 5. 备份恢复演练

- [ ] 5.1 执行一次手动备份
- [ ] 5.2 人为破坏当前数据库文件（在预发布/演练环境，非生产环境）
- [ ] 5.3 执行恢复流程，验证恢复前是否先对损坏状态做了快照
- [ ] 5.4 验证恢复后系统正常启动，数据与备份时刻一致
- [ ] 5.5 记录整个演练耗时
      文件：docs/migration/backup-restore-drill.md
- [ ] 5.6 提交：git commit -m "test: backup and restore drill documented"

## 6. 24小时稳定性观察

- [ ] 6.1 部署到预发布/生产环境后持续观察24小时
- [ ] 6.2 检查六类定时任务在此期间均有实际触发记录
      （查询notification_log和应用日志）
- [ ] 6.3 检查进程运行时长和日志，确认无未捕获异常导致的退出
- [ ] 6.4 记录观察结果
      文件：docs/migration/stability-observation.md
- [ ] 6.5 提交：git commit -m "test: 24-hour stability observation report"

## 7. 正式部署上线

- [ ] 7.1 确认第2-6节全部通过
- [ ] 7.2 选定最终部署模式（本地Windows/Linux云端/Wails桌面）
- [ ] 7.3 执行正式部署
- [ ] 7.4 提交：git commit -m "chore: V1 production deployment"

## 8. 规格场景覆盖检查表

- [ ] 8.1 「账务事务类测试通过」→ 2.1-2.2
- [ ] 8.2 「配置化边界测试通过」→ 2.3-2.4
- [ ] 8.3 「幂等性与状态机测试通过」→ 2.5-2.6
- [ ] 8.4 「数据安全测试通过」→ 2.7-2.8
- [ ] 8.5 「跨平台部署测试通过」→ 2.9-2.10
- [ ] 8.6 「导入总数核对」→ 3.4
- [ ] 8.7 「抽样字段核对」→ 3.5
- [ ] 8.8 「恢复后数据一致」→ 5.1-5.4
- [ ] 8.9 「定时任务触发验证」→ 6.2
- [ ] 8.10 「进程无异常退出」→ 6.3

全部勾选后才可执行`/opsx:archive add-testing-and-migration`，
本change archive后即视为Zedu V1正式发布。
