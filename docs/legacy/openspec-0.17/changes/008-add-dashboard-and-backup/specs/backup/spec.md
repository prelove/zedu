## ADDED Requirements

### Requirement: 手动数据备份
系统必须支持一键触发数据库备份，且不能简单复制正在写入的文件，
备份结果必须留痕。

#### Scenario: 触发备份
- **WHEN** 调用POST /backup/trigger
- **THEN** 使用SQLite的VACUUM INTO方式在backup/目录生成一个带时间戳的.db文件，
  该文件可被独立打开且数据与主库一致

#### Scenario: 备份记录留痕
- **WHEN** 备份操作完成（无论成功失败）
- **THEN** backup_log新增一条记录，成功时记录文件名和大小，
  失败时记录error_msg

#### Scenario: 备份文件数据一致性验证
- **WHEN** 备份完成后，独立打开生成的备份文件
- **THEN** 其表数量和关键表(student/lesson等)的记录数与主库完全一致
