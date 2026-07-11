# Proposal: 课程报名与师生安排

## Why
真实业务中一个学生可能同时学习多个课程方向、由不同老师负责，因此不能
用"学生绑定单一老师"的简化模型。核心关系链是：学生→课程报名→师生安排→课次。
这个多对多设计是PRD第九章9.3节反复强调的核心决策，如果在这个change里
偷懒做成简化模型，后续所有依赖enrollment的change（排课、账务、结款）
都要推倒重来。

## 业务背景
PRD举了一个具体例子：王同学同时报名"JLPT备考"和"日常会话"两个方向，
分别由老师A和老师B教。这不是一个边缘情况，而是小型教培机构的常态——
学生一旦建立信任关系，往往会追加报名其他课程。因此enrollment必须
支持"一学生多项目"，且每个项目独立维护自己的余额、课时和当前负责老师，
不能把这些字段挂在student表上。

同样重要的是"换老师"这个操作：PRD第九章9.3节明确要求"历史课次不受
影响，保留原老师快照"，这意味着换老师不是简单的UPDATE一个字段，
而是要通过"结束旧assignment、创建新assignment"的方式保留完整的
师生关系变更历史，这对于未来分析"这个学生换了几次老师、为什么换"
这类问题是有价值的审计数据。

## What Changes
- 新增student_course_enrollment CRUD
- 新增student_teacher_assignment创建与换老师接口
- 新增学生详情页"学习项目"Tab及换老师交互

## Non-Goals
- 不实现"批量更新未来课次"这个换老师的增强体验（留给
  009-add-payout-and-scheduling-enhancement）
- 不实现学习路径(student_learning_path)和等级变化事件
  (student_level_event)的完整管理界面（PRD第五章5.5节的数据结构
  已在001的DDL中预留，本change聚焦enrollment和assignment的核心
  CRUD，学习路径的详细跟踪属于锦上添花功能，不阻塞主链路）
- 不实现小班课（class_group）相关功能，数据结构已预留

## Impact
- Affected specs: enrollment（新增）
- Affected code: backend/internal/enrollment/、
  frontend/admin/src/views/student/(学习项目Tab)
- 依赖：003-add-student-teacher-profile（需要student/teacher/course表）
- 被依赖：005（lesson需要enrollment和assignment）、
  006（课后确认需要enrollment余额字段）、
  007（充值需要enrollment作为归属项目）
