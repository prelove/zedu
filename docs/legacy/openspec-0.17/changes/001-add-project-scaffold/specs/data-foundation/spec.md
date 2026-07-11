## ADDED Requirements

### Requirement: 工程骨架可运行
系统必须提供一个可以本地启动的最小Go服务和前端脚手架，且严格遵循
PRD16.3节的目录分层约定。

#### Scenario: 本地启动
- **WHEN** 执行 `make dev`
- **THEN** 后端服务启动并监听配置端口，访问 `/healthz` 返回 `{"status":"ok"}`

#### Scenario: 目录结构符合规范
- **WHEN** 检查backend/目录结构
- **THEN** 存在cmd/zedu-server、internal/、pkg/、migrations/(含sqlite/mysql/postgres/seed四个子目录)四个顶层结构

### Requirement: 数据库结构完整
系统必须包含PRD第十二章定义的全部31张表及其索引、外键约束，且支持
正向迁移和回滚。

#### Scenario: 迁移执行
- **WHEN** 对全新的空SQLite文件执行goose迁移
- **THEN** 全部31张表被创建，且 `PRAGMA foreign_keys` 生效
  （插入违反外键的记录应报错）

#### Scenario: 迁移可回滚
- **WHEN** 对已执行迁移的数据库执行goose down
- **THEN** 全部31张表被正确删除，不因外键依赖顺序问题而报错

#### Scenario: PRAGMA配置生效
- **WHEN** 服务启动后连接数据库
- **THEN** journal_mode为WAL，foreign_keys为ON，busy_timeout≥5000

### Requirement: 初始种子数据完整且幂等
系统必须提供日语模板、K12模板（供未来复制部署使用）、支付方式字典、
出勤分类字典和基础系统配置的种子数据，且重复执行不产生重复记录。

#### Scenario: 日语模板种子数据校验
- **WHEN** 查询course_domain/track/level/skill_tag（应用日语模板后）
- **THEN** 能查到日语领域及其4个方向（JLPT备考/日常会话/商务日语/
  少儿日语）、JLPT6个等级+会话3个等级、9个能力标签

#### Scenario: K12模板种子数据存在但不默认应用
- **WHEN** 检查migrations/seed/目录
- **THEN** 存在K12模板的种子SQL文件，但本change不会自动执行它
  （留给002的初始化向导按需调用）

#### Scenario: 字典与配置数据校验
- **WHEN** 查询payment_method/attendance_outcome_type/system_config
- **THEN** 分别能查到6条支付方式、9条出勤分类（含建议值）、以及
  base_currency=JPY、base_currency_locked=0等初始配置

#### Scenario: 种子数据幂等
- **WHEN** 对同一数据库重复执行种子数据SQL
- **THEN** 不产生重复记录，也不报错（使用INSERT OR IGNORE模式）

### Requirement: 跨模块工具包不反向依赖业务包
系统的pkg/目录下的工具函数必须保持对internal/业务包零依赖。

#### Scenario: 依赖方向校验
- **WHEN** 检查pkg/money、pkg/datetime等包的import
- **THEN** 不存在任何对internal/下业务包的引用
