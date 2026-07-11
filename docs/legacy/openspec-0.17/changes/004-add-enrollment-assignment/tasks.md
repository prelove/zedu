## 1. 课程报名后端

- [ ] 1.1 编写失败测试：POST /enrollments 创建成功，charge_per_lesson_amount
      默认可为0，lesson_balance/balance_amount默认为0
      文件：backend/internal/enrollment/service_test.go
- [ ] 1.2 实现最小代码使测试通过
      文件：backend/internal/enrollment/model.go, dto.go, service.go, handler.go, repository.go
- [ ] 1.3 编写失败测试：同一学生新建第二条不同track的enrollment，两条均可查询
      且余额互不影响
- [ ] 1.4 验证1.3通过（若失败检查是否有多余的唯一约束）
- [ ] 1.5 编写失败测试：TRIAL转ONE_TO_ONE后，此前关联的课次记录仍指向
      同一enrollment_id
      文件：backend/internal/enrollment/trial_conversion_test.go
- [ ] 1.6 实现转换接口使测试通过
- [ ] 1.7 提交：git commit -m "feat(enrollment): course enrollment CRUD with trial conversion"

## 2. 课程报名状态机

- [ ] 2.1 编写失败测试：ACTIVE可暂停为PAUSED，PAUSED可恢复为ACTIVE
      文件：backend/internal/enrollment/status_test.go
- [ ] 2.2 实现状态变更接口使测试通过
- [ ] 2.3 编写失败测试：COMPLETED/CANCELLED状态不允许变回ACTIVE/PAUSED
- [ ] 2.4 实现状态机校验使测试通过（返回42201）
- [ ] 2.5 提交：git commit -m "feat(enrollment): status state machine with terminal states"

## 3. 师生安排与换老师

- [ ] 3.1 编写失败测试：给enrollment绑定老师应创建一条ACTIVE的assignment
      文件：backend/internal/enrollment/assignment_test.go
- [ ] 3.2 实现最小代码使测试通过
- [ ] 3.3 编写失败测试：换老师后旧assignment变ENDED有end_date，
      新assignment为ACTIVE，且同一时刻只有一条MAIN角色ACTIVE记录
- [ ] 3.4 实现换老师逻辑使测试通过（须在事务内完成两次写入）
- [ ] 3.5 编写失败测试：换老师后enrollment.balance_amount/lesson_balance不变
- [ ] 3.6 验证3.5通过
- [ ] 3.7 编写失败测试：新增SUBSTITUTE角色assignment不影响已有MAIN角色
      的ACTIVE状态
- [ ] 3.8 验证3.7通过
- [ ] 3.9 提交：git commit -m "feat(enrollment): teacher assignment, reassignment, and substitute cover"

## 4. 前端：学习项目Tab与换老师

- [ ] 4.1 学生详情页"学习项目"Tab，卡片展示enrollment列表（含状态标签）
- [ ] 4.2 新建课程报名表单
- [ ] 4.3 试听转正式操作入口
- [ ] 4.4 换老师弹窗（选新老师+填写原因）
- [ ] 4.5 提交：git commit -m "feat(frontend): enrollment tab and teacher reassignment"

## 5. 规格场景覆盖检查表

对照本change下specs/enrollment/spec.md的全部Scenario，逐条标注
验证task：

- [ ] 5.1 「新建课程报名」→ 1.1-1.2
- [ ] 5.2 「同一学生多项目并行」→ 1.3-1.4
- [ ] 5.3 「试听转正式」→ 1.5-1.6
- [ ] 5.4 「暂停与恢复」→ 2.1-2.2
- [ ] 5.5 「终止为终态」→ 2.3-2.4
- [ ] 5.6 「终态后不可排课」→ 2.3-2.4（状态字段校验，实际拦截由005验证）
- [ ] 5.7 「新建师生安排」→ 3.1-3.2
- [ ] 5.8 「换老师保留历史」→ 3.3-3.4
- [ ] 5.9 「余额不随老师变动」→ 3.5-3.6
- [ ] 5.10 「代课不影响主责老师记录」→ 3.7-3.8

全部勾选后才可执行`/opsx:archive add-enrollment-assignment`。
