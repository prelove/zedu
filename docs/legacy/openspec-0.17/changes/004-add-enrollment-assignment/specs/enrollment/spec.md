## ADDED Requirements

### Requirement: 课程报名创建
系统必须支持一个学生同时拥有多条课程报名记录，覆盖不同课程方向，
且各自独立维护余额和课时。

#### Scenario: 新建课程报名
- **WHEN** 给学生新建一条enrollment，指定domain/track/level
- **THEN** 创建成功，status默认ACTIVE，charge_per_lesson_amount可为0，
  lesson_balance和balance_amount默认为0

#### Scenario: 同一学生多项目并行
- **WHEN** 同一学生已有一条ACTIVE的enrollment，再新建一条不同track的enrollment
- **THEN** 两条记录都能独立查询，各自的余额和课时互不影响

#### Scenario: 试听转正式
- **WHEN** 对一条enrollment_type=TRIAL的记录执行"转为正式"操作
- **THEN** enrollment_type变为ONE_TO_ONE，该enrollment下此前产生的
  课次和账务记录保持关联不变（不会被割裂到新记录）

### Requirement: 课程报名状态机
系统必须限制enrollment的状态转换路径，终态不可逆转。

#### Scenario: 暂停与恢复
- **WHEN** 对ACTIVE状态的enrollment执行暂停
- **THEN** 状态变为PAUSED，之后可再次执行恢复变回ACTIVE

#### Scenario: 终止为终态
- **WHEN** enrollment变为COMPLETED或CANCELLED
- **THEN** 不允许再变更为ACTIVE或PAUSED，如需继续学习须新建enrollment

#### Scenario: 终态后不可排课
- **WHEN** enrollment状态为COMPLETED或CANCELLED
- **THEN** 尝试基于该enrollment创建新课次应被拒绝（该校验的具体实现
  属于005-add-lesson-scheduling范围，本change确保状态字段本身
  可被正确读取判断）

### Requirement: 师生安排与换老师
系统必须支持给enrollment绑定老师，并支持更换老师且保留历史，同一
enrollment任意时刻只能有一条MAIN角色的ACTIVE安排。

#### Scenario: 新建师生安排
- **WHEN** 给enrollment绑定一位老师，role_type=MAIN
- **THEN** 创建一条status=ACTIVE的assignment

#### Scenario: 换老师保留历史
- **WHEN** 对某enrollment执行换老师操作，指定新老师
- **THEN** 旧assignment.status变为ENDED并记录end_date，新assignment
  创建为status=ACTIVE，同一时刻同一enrollment下只有一条MAIN角色的
  ACTIVE记录

#### Scenario: 余额不随老师变动
- **WHEN** 执行换老师操作
- **THEN** enrollment.balance_amount和lesson_balance数值不变

#### Scenario: 代课不影响主责老师记录
- **WHEN** 给enrollment新增一条role_type=SUBSTITUTE的assignment记录
- **THEN** 此前role_type=MAIN的ACTIVE记录不受影响，仍保持ACTIVE状态
