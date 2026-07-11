## MODIFIED Requirements

### Requirement: 课次创建
系统必须支持基于课程报名和师生安排创建具体课次，且在创建时提供
时间冲突检测能力（此前版本不含冲突检测）。

#### Scenario: 新建课次
- **WHEN** 指定enrollment、老师（自动带入当前assignment对应老师，可改选）、
  上课时间、时长(10~480分钟)、上课方式
- **THEN** 创建成功，status默认SCHEDULED

#### Scenario: 时长超出范围被拒绝
- **WHEN** 提交duration_min=5(小于10)或600(大于480)
- **THEN** 返回40001参数校验失败

### Requirement: 课次编辑限制
系统必须限制课次的可编辑状态。

#### Scenario: SCHEDULED状态可编辑
- **WHEN** 课次status=SCHEDULED时提交编辑
- **THEN** 编辑成功

#### Scenario: COMPLETED状态不可编辑基础信息
- **WHEN** 课次status=COMPLETED时尝试编辑上课时间或老师
- **THEN** 返回42201，仅note字段允许修改

## ADDED Requirements

### Requirement: 时间冲突检测
系统必须能检测同一学生或同一老师在某时段是否已有其他课次安排，
且首尾相接不算冲突。

#### Scenario: 检测到冲突
- **WHEN** 调用GET /lessons/{id}/conflicts，该学生或老师在此时段
  已有其他status非CANCELLED的课次（时间区间有实质重叠）
- **THEN** 返回冲突详情列表，前端显示黄色警示但不阻止创建

#### Scenario: 首尾相接不算冲突
- **WHEN** 新课次的开始时间恰好等于已有课次的结束时间（或反之）
- **THEN** 不视为冲突，不出现在冲突列表中

#### Scenario: 排除自身
- **WHEN** 对某课次调用冲突检测接口
- **THEN** 该课次本身不会出现在自己的冲突列表里
