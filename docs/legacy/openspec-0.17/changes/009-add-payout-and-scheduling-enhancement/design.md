# Design: 结款去重、冲突检测算法与批量更新

## 结款去重方案

teacher_account_ledger的LESSON_PAYABLE记录，一旦被某次payout关联结算，
必须标记related_payout_id，避免下次preview重复出现。方案：
- teacher_payout创建时，同一事务内对本次结算范围内的
  teacher_account_ledger(biz_type=LESSON_PAYABLE, related_payout_id IS NULL)
  记录批量更新related_payout_id为新建的payout.id
- preview接口的查询条件必须包含related_payout_id IS NULL，
  确保已结算记录不会重复出现
- 若Operator在预览界面勾选排除了某几条记录，这几条记录**不会**被
  标记related_payout_id，它们会继续留在"未结算"池中，供下次结款
  周期再次被预览到

## 冲突检测算法

GET /lessons/{id}/conflicts 查询同一学生或同一老师在
[scheduled_start_at, scheduled_end_at]时间区间内是否存在其他status
不为CANCELLED的课次，区间重叠判断用标准的区间相交公式：
```
existing.scheduled_start_at < new.scheduled_end_at
  AND existing.scheduled_end_at > new.scheduled_start_at
```
这个判断需要注意边界情况：如果新课次恰好在旧课次结束的那一刻开始
（existing.end == new.start），不应算作冲突（首尾相接是合理的连续
排课场景，比如老师连续教两节课中间没有休息也是常见情况）。因此
判断条件用严格不等号(< 和 >)而非≤/≥，这个细节容易被误写成闭区间
判断导致误报"首尾相接"为冲突。

冲突检测同时检查学生侧和老师侧：查询条件是
`(student_id = ? OR teacher_id = ?) AND lesson_id != 当前编辑的课次ID`
（排除自身，避免编辑课次时把自己检测成冲突）。

## 换老师批量更新未来课次的实现顺序

POST /enrollments/{id}/assignments/change-teacher 增加updateFutureLessons
布尔参数：
1. 在同一事务内先完成"结束旧assignment、创建新assignment"这两步
   （复用004已实现的核心逻辑）
2. 若updateFutureLessons=true，紧接着在同一事务内执行批量更新：
   ```sql
   UPDATE lesson SET teacher_id = 新老师ID
   WHERE enrollment_id = ? AND scheduled_start_at > NOW()
     AND status IN ('SCHEDULED', 'REMINDED')
   ```
3. 历史（已完成）课次不受影响，永远保留原老师快照——这意味着查询
   条件必须精确限定status IN (SCHEDULED, REMINDED)，不能用简单的
   "时间大于现在"来判断，因为理论上可能存在"时间已过但状态仍是
   SCHEDULED"的挂起课次（Operator还没来得及处理），这类课次是否
   应该被批量更新是一个值得注意的边界，本设计选择"仍然更新"，因为
   它们本质上还没有真正上课，换老师后理应由新老师负责

## 结款金额调整的记录方式

预览返回的是系统计算的应付总额，Operator提交时的actual_amount_base
若与预览金额不同（比如老师提出异议后达成的实际结算金额），差额
本身不需要额外生成一条ADJUST流水去"解释"这个差异——teacher_payout
记录本身的amount_base(应付)和actual_amount_base(实付)两个字段
就已经完整记录了差异，这是一种更简洁的设计，不需要为每次金额调整
都额外造一条流水记录。
