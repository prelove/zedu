## ADDED Requirements

### Requirement: 工作台完整版图表数据
系统必须提供五类图表所需的聚合数据接口，且统计口径必须在数据库层
完成，避免应用层重复实现导致的口径不一致。

#### Scenario: 月度收入趋势
- **WHEN** 调用GET /reports/revenue-trend
- **THEN** 返回最近12个月每月的收入、课酬、毛利三条数据系列

#### Scenario: 学生增长曲线
- **WHEN** 调用GET /reports/student-growth
- **THEN** 返回最近6个月每月累计活跃学生数

#### Scenario: 课程方向占比
- **WHEN** 调用GET /reports/track-distribution
- **THEN** 返回本月各课程方向的课时收入占比，百分比总和允许因
  四舍五入存在微小误差（不强制精确等于100%）

#### Scenario: 老师带课分布
- **WHEN** 调用GET /reports/teacher-workload
- **THEN** 返回按课时数排序的各老师本月带课量

#### Scenario: 课时完成率
- **WHEN** 调用GET /reports/completion-rate
- **THEN** 返回按月分类的完成/取消/总计课次数

### Requirement: 财务报表与多格式导出
系统必须支持按月/周/日粒度查询财务报表，且Excel和PDF两种导出格式
的数据必须完全一致。

#### Scenario: 报表粒度切换
- **WHEN** 调用GET /finance/report 指定groupBy=month/week/day
- **THEN** 返回对应粒度汇总的收入/课酬/毛利数据

#### Scenario: 导出格式数据一致性
- **WHEN** 对同一时间范围分别导出Excel和PDF格式的报表
- **THEN** 两份文件中的汇总数字完全相同，只是文件格式不同
