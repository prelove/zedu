## ADDED Requirements

### Requirement: 老师档案管理
系统必须支持老师档案的创建、查询、编辑和状态变更。

#### Scenario: 新建老师
- **WHEN** 提交合法的老师姓名和默认课酬(≥0整数)
- **THEN** 创建成功，status默认为ACTIVE

#### Scenario: 默认课酬校验
- **WHEN** 提交负数的default_rate_amount
- **THEN** 返回40001参数校验失败

#### Scenario: 老师暂停后排课候选列表排除
- **WHEN** 老师status变更为PAUSED
- **THEN** 该老师不应出现在排课时的可选老师候选列表中（此规则的具体
  排课接口实现属于005-add-lesson-scheduling范围，本change只需保证
  status字段可正确变更并被后续change读取）

### Requirement: 老师能力记录
系统必须支持为老师维护多条能力记录，且同一(老师,方向,等级)组合唯一，
同时允许等级为空以覆盖"暂不分级"的场景。

#### Scenario: 新增能力记录
- **WHEN** 给老师添加一条(领域,方向,等级)能力记录
- **THEN** 创建成功

#### Scenario: 重复能力记录被拒绝
- **WHEN** 对同一老师提交相同(teacher_id,track_id,level_id)组合的能力记录
- **THEN** 返回40901冲突错误

#### Scenario: 等级为空时不受唯一约束限制
- **WHEN** 对同一老师同一方向提交两条level_id均为NULL的能力记录
- **THEN** 两条记录都能创建成功（SQL的NULL不参与唯一性判断为相等），
  但前端应在提交前提示"该方向已有一条不分级的能力记录，是否确认
  仍要新增"，避免用户误操作产生冗余数据

#### Scenario: 能力记录状态变更不删除历史
- **WHEN** 老师的某条能力记录从ACTIVE变为ENDED（如培训暂停）
- **THEN** 记录保留在数据库中，effective_to被设置，而非被删除

### Requirement: 老师可授时间管理
系统必须支持维护老师的可授时间段（周几+起止时间）。

#### Scenario: 新增可授时间
- **WHEN** 给老师添加一条周几+时间段记录
- **THEN** 创建成功，可设置生效起止日期

#### Scenario: 时间段格式校验
- **WHEN** 提交的start_time晚于或等于end_time
- **THEN** 返回40001参数校验失败
