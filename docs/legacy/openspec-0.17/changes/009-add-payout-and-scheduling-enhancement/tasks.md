## 1. 老师结款后端

- [ ] 1.1 编写失败测试：POST /finance/payouts/preview 返回未结算的
      LESSON_PAYABLE记录及汇总金额
      文件：backend/internal/finance/payout_test.go
- [ ] 1.2 实现最小代码使测试通过
      文件：backend/internal/finance/payout_model.go, payout_service.go, payout_handler.go
- [ ] 1.3 编写失败测试：该周期内无未结算记录时，预览应返回空列表和0，
      不报错
- [ ] 1.4 验证1.3通过
- [ ] 1.5 编写失败测试：提交结款后teacher.unpaid_amount正确扣减，
      关联ledger记录被标记related_payout_id
- [ ] 1.6 实现代码使测试通过（同一事务内完成结款单+ledger标记）
- [ ] 1.7 编写失败测试：结款后再次预览，已结算课次不重复出现，
      被排除的课次仍出现
- [ ] 1.8 验证1.7通过（检查预览查询是否正确过滤related_payout_id IS NULL，
      且排除逻辑不误标记）
- [ ] 1.9 编写失败测试：实付金额与应付金额不同时，两个字段各自
      正确保存，不产生额外流水
- [ ] 1.10 验证1.9通过
- [ ] 1.11 提交：git commit -m "feat(payout): preview and settlement with dedup and exclusion support"

## 2. 排课时间冲突检测

- [ ] 2.1 编写失败测试：同一老师在重叠时段已有课次时，
      GET /lessons/{id}/conflicts应返回冲突详情
      文件：backend/internal/lesson/conflict_test.go
- [ ] 2.2 实现最小代码使测试通过
      文件：backend/internal/lesson/conflict_service.go
- [ ] 2.3 编写失败测试：不重叠时段应返回空冲突列表
- [ ] 2.4 验证2.3通过
- [ ] 2.5 编写失败测试：首尾相接（existing.end==new.start）不应
      判定为冲突
- [ ] 2.6 实现严格不等号的区间判断逻辑使测试通过
- [ ] 2.7 编写失败测试：冲突检测应排除课次自身
- [ ] 2.8 验证2.7通过
- [ ] 2.9 提交：git commit -m "feat(lesson): time conflict detection with boundary handling"

## 3. 换老师批量更新未来课次

- [ ] 3.1 编写失败测试：换老师时updateFutureLessons=true，
      未来SCHEDULED课次的teacher_id应批量更新
      文件：backend/internal/enrollment/assignment_test.go
- [ ] 3.2 实现代码使测试通过（同一事务内完成换老师+批量更新）
- [ ] 3.3 编写失败测试：已COMPLETED课次的teacher_id不受影响
- [ ] 3.4 验证3.3通过
- [ ] 3.5 编写失败测试：时间已过但状态仍为SCHEDULED的挂起课次，
      在updateFutureLessons=true时也应被更新
- [ ] 3.6 验证3.5通过
- [ ] 3.7 提交：git commit -m "feat(enrollment): bulk update future and pending lessons on reassignment"

## 4. 前端：结款页面与冲突提示

- [ ] 4.1 结款记录页（预览流程+勾选排除+提交）
- [ ] 4.2 排课表单增加冲突警示条（黄色，不阻止提交）
- [ ] 4.3 换老师弹窗增加"是否批量更新未来课次"选项
- [ ] 4.4 提交：git commit -m "feat(frontend): payout page and conflict warnings"

## 5. 规格场景覆盖检查表

对照本change下specs/payout、specs/lesson-scheduling（MODIFIED）、
specs/enrollment（MODIFIED）三份spec.md的全部Scenario，逐条标注
验证task：

- [ ] 5.1 「预览待结算明细」→ 1.1-1.2
- [ ] 5.2 「空结果不报错」→ 1.3-1.4
- [ ] 5.3 「提交结款」→ 1.5-1.6
- [ ] 5.4 「已结算不重复出现」→ 1.7-1.8
- [ ] 5.5 「被排除的记录仍可在下次预览中出现」→ 1.7-1.8
- [ ] 5.6 「实付调整」→ 1.9-1.10
- [ ] 5.7 「检测到冲突」→ 2.1-2.2
- [ ] 5.8 「首尾相接不算冲突」→ 2.5-2.6
- [ ] 5.9 「排除自身」→ 2.7-2.8
- [ ] 5.10 「批量更新未来课次」→ 3.1-3.2
- [ ] 5.11 「历史课次不受影响」→ 3.3-3.4
- [ ] 5.12 「时间已过但仍挂起的课次也会被更新」→ 3.5-3.6

全部勾选后才可执行`/opsx:archive add-payout-and-scheduling-enhancement`。
