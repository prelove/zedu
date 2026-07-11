## MODIFIED Requirements

### Requirement: 课次状态自动流转
系统必须支持课次在无人工确认的情况下，超时后自动流转为COMPLETED状态，
但不触发任何账务写入（此前版本课次状态只能由Operator手动确认或取消
来推进）。

#### Scenario: 超时自动关闭
- **WHEN** 课次status为SCHEDULED或REMINDED，且scheduled_end_at早于
  当前时间4小时以上
- **THEN** 定时任务将其status更新为COMPLETED，不生成attendance、
  不生成任何ledger记录

#### Scenario: 缓冲期内不误关
- **WHEN** 课次scheduled_end_at刚过去不到4小时
- **THEN** 该课次不会被自动关闭任务选中，仍保持原状态供Operator
  手动确认
