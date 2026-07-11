## ADDED Requirements

### Requirement: 工作台聚合数据（极简版）
系统必须能一次性返回今日课程、待续费学生、待结款老师三类列表，
且判断口径与后续的通知任务保持一致。

#### Scenario: 查询工作台数据
- **WHEN** 调用GET /reports/dashboard
- **THEN** 返回今日排课列表、lesson_balance≤阈值或balance_amount不足
  的enrollment列表、unpaid_amount>0的老师列表

#### Scenario: 今日课程按系统时区判断
- **WHEN** 系统时区为Asia/Tokyo，当前UTC时间对应日本时间是次日凌晨
- **THEN** "今日"的判断以Asia/Tokyo时区的日期为准，而非UTC日期

#### Scenario: 待续费口径与余额预警任务一致
- **WHEN** 某enrollment满足lesson_balance≤阈值或
  balance_amount<charge_per_lesson_amount任一条件
- **THEN** 该enrollment同时出现在工作台待续费列表和（未来010实现的）
  余额预警邮件名单中，两处判断条件必须一致
