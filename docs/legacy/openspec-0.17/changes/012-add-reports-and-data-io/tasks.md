## 1. 报表数据接口（完整版）

- [ ] 1.1 编写失败测试：GET /reports/revenue-trend 返回最近12个月的
      收入/课酬/毛利三条数据系列
      文件：backend/internal/report/revenue_trend_test.go
- [ ] 1.2 实现最小代码使测试通过（数据库层GROUP BY聚合）
      文件：backend/internal/report/revenue_trend_service.go
- [ ] 1.3 编写失败测试：GET /reports/student-growth 返回最近6个月
      累计活跃学生数
- [ ] 1.4 实现代码使测试通过
- [ ] 1.5 编写失败测试：GET /reports/track-distribution 返回本月各
      方向课时收入占比
- [ ] 1.6 实现代码使测试通过
- [ ] 1.7 编写失败测试：GET /reports/teacher-workload 返回按课时数
      排序的老师带课分布
- [ ] 1.8 实现代码使测试通过
- [ ] 1.9 编写失败测试：GET /reports/completion-rate 返回按月分类的
      完成/取消/总计课次数
- [ ] 1.10 实现代码使测试通过
- [ ] 1.11 提交：git commit -m "feat(report): full dashboard chart data endpoints"

## 2. 财务报表与导出

- [ ] 2.1 编写失败测试：GET /finance/report 按groupBy=month/week/day
      返回对应粒度汇总数据
      文件：backend/internal/finance/report_test.go
- [ ] 2.2 实现最小代码使测试通过，提取共享的"生成报表数据"service函数
      文件：backend/internal/finance/report_service.go
- [ ] 2.3 编写失败测试：导出Excel和PDF格式的同一份报表，汇总数字
      应完全一致
      文件：backend/internal/finance/report_export_test.go
- [ ] 2.4 实现导出接口使测试通过，确保Excel和PDF导出复用同一份
      report_service.go生成的数据结构
      文件：backend/internal/finance/report_export.go
- [ ] 2.5 提交：git commit -m "feat(finance): multi-granularity report with consistent excel/pdf export"

## 3. Excel批量导入

- [ ] 3.1 编写失败测试：全部合法数据的Excel导入应全部成功
      文件：backend/internal/student/import_test.go
- [ ] 3.2 实现最小代码使测试通过
      文件：backend/internal/student/import_service.go, import_handler.go
      要点：逐行独立处理，不使用单一大事务
- [ ] 3.3 编写失败测试：某行邮箱重复时，该行跳过但其余行正常导入
- [ ] 3.4 实现代码使测试通过
- [ ] 3.5 编写失败测试：邮箱大小写不同应被判定为重复（不敏感匹配）
- [ ] 3.6 实现大小写不敏感比较逻辑使测试通过
- [ ] 3.7 编写失败测试：学习方向列填写系统中不存在的名称时，
      学生仍导入成功但不创建enrollment，导入报告标注该行
- [ ] 3.8 实现降级处理逻辑使测试通过
- [ ] 3.9 编写失败测试：GET /students/import-template 返回含正确
      列名和示例数据的模板文件
- [ ] 3.10 实现代码使测试通过
- [ ] 3.11 对teacher导入重复上述3.1-3.8的测试与实现模式
      文件：backend/internal/teacher/import_test.go, import_service.go
- [ ] 3.12 提交：git commit -m "feat(import): row-independent excel import with detailed report"

## 4. 数据导出

- [ ] 4.1 编写失败测试：GET /students/export 附带筛选参数，导出内容
      应只包含符合条件的记录
      文件：backend/internal/student/export_test.go
- [ ] 4.2 实现代码使测试通过
- [ ] 4.3 提交：git commit -m "feat(export): filtered student list export"

## 5. 前端：工作台图表与报表页

- [ ] 5.1 工作台接入ECharts完整版（5个图表）
- [ ] 5.2 数据图表页
- [ ] 5.3 财务报表页（粒度切换+导出按钮）
- [ ] 5.4 学生/老师列表页增加"导入Excel"入口和"导出Excel"按钮
- [ ] 5.5 操作日志查询页面
- [ ] 5.6 提交：git commit -m "feat(frontend): dashboard charts, reports, and import/export UI"

## 6. 规格场景覆盖检查表

对照本change下specs/reports/spec.md和specs/data-io/spec.md的全部
Scenario，逐条标注验证task：

- [ ] 6.1 「月度收入趋势」→ 1.1-1.2
- [ ] 6.2 「学生增长曲线」→ 1.3-1.4
- [ ] 6.3 「课程方向占比」→ 1.5-1.6
- [ ] 6.4 「老师带课分布」→ 1.7-1.8
- [ ] 6.5 「课时完成率」→ 1.9-1.10
- [ ] 6.6 「报表粒度切换」→ 2.1-2.2
- [ ] 6.7 「导出格式数据一致性」→ 2.3-2.4
- [ ] 6.8 「全部成功导入」→ 3.1-3.2
- [ ] 6.9 「部分行失败不影响其他行」→ 3.3-3.4
- [ ] 6.10 「邮箱大小写不敏感去重」→ 3.5-3.6
- [ ] 6.11 「学习方向不匹配时的降级处理」→ 3.7-3.8
- [ ] 6.12 「下载导入模板」→ 3.9-3.10
- [ ] 6.13 「按筛选条件导出学生列表」→ 4.1-4.2

全部勾选后才可执行`/opsx:archive add-reports-and-data-io`。
