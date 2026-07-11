## ADDED Requirements

### Requirement: 课次创建
系统必须支持基于课程报名和师生安排创建具体课次，且课次编号具有
业务可读性和唯一性。

#### Scenario: 新建课次
- **WHEN** 指定enrollment、老师（自动带入当前assignment对应老师，可改选）、
  上课时间、时长(10~480分钟)、上课方式
- **THEN** 创建成功，status默认SCHEDULED，lesson_no按当日序号规则生成

#### Scenario: 时长超出范围被拒绝
- **WHEN** 提交duration_min=5(小于10)或600(大于480)
- **THEN** 返回40001参数校验失败

#### Scenario: lesson_no唯一且可读
- **WHEN** 同一天内连续创建多个课次
- **THEN** 每条记录的lesson_no各不相同，且符合L+日期+序号的可读格式

#### Scenario: 上课链接选填
- **WHEN** 创建课次时meeting_type=WECHAT且meeting_link留空或填写群名而非URL
- **THEN** 创建成功，系统不对该字段做URL格式校验

#### Scenario: 基于终态enrollment创建课次被拒绝
- **WHEN** enrollment状态为COMPLETED或CANCELLED，尝试基于其创建新课次
- **THEN** 返回42201

### Requirement: 课次编辑限制
系统必须限制课次的可编辑状态，保护已完成课次的历史完整性。

#### Scenario: SCHEDULED状态可编辑
- **WHEN** 课次status=SCHEDULED时提交编辑
- **THEN** 编辑成功，可修改时间/老师/上课方式等字段

#### Scenario: COMPLETED状态不可编辑基础信息
- **WHEN** 课次status=COMPLETED时尝试编辑上课时间或老师
- **THEN** 返回42201，仅note字段允许修改

### Requirement: 时区一致性
系统必须统一以UTC存储课次时间，同时保留显示时区信息。

#### Scenario: 存储与显示分离
- **WHEN** 以Asia/Tokyo时区创建一个显示为"19:00"的课次
- **THEN** 数据库中scheduled_start_at存储为对应的UTC时间，timezone
  字段记录为Asia/Tokyo，供前端正确还原显示
