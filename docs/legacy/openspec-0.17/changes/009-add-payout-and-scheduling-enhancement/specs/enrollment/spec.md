## MODIFIED Requirements

### Requirement: 师生安排与换老师
系统必须支持给enrollment绑定老师，并支持更换老师且保留历史，
换老师时可选择是否批量更新该enrollment下的未来课次（此前版本
不支持批量更新）。

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

## ADDED Requirements

### Requirement: 换老师批量更新未来课次
系统必须支持换老师时可选地批量更新未来课次的授课老师，且历史课次
和已确认课次不受影响。

#### Scenario: 批量更新未来课次
- **WHEN** 换老师时传入updateFutureLessons=true
- **THEN** 该enrollment下scheduled_start_at>当前时间且status为
  SCHEDULED/REMINDED的课次，其teacher_id被批量更新为新老师

#### Scenario: 历史课次不受影响
- **WHEN** 执行换老师操作（无论updateFutureLessons取值）
- **THEN** 已COMPLETED的课次的teacher_id保持不变

#### Scenario: 时间已过但仍挂起的课次也会被更新
- **WHEN** 某课次的scheduled_start_at已经过去，但status仍为
  SCHEDULED（尚未被确认或自动关闭），且updateFutureLessons=true
- **THEN** 该课次仍会被批量更新为新老师（因为它实质上还没有真正
  发生教学行为，理应由新老师负责）
