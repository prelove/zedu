# Proposal: 课后确认与账务事务

## Why
课后确认是Zedu账务体系的唯一触发点，必须保证：
1. 出勤分类的"建议值"和Operator最终填写的"实际值"都要落库，互不覆盖
2. 涉及的六处写入必须原子完成，任一步失败全部回滚
3. 本位币锁定规则必须在此处正确触发
4. 同一课次不能被重复确认（并发或误操作场景下）

这是全项目正确性要求最高的一个capability，任何疏漏都会导致账目无法
信任。PRD第二章"产品设计五大原则"中的原则三"财务上落事实，不依赖
回算"和原则四"系统给建议，人做最终决定"，在整个系统里最集中体现的
地方就是这个change。

## 业务背景
根据PRD第九章9.5节的描述，课后确认表单要处理的真实场景远比"是否
出勤"这个二元判断复杂：学生可能提前请假（不扣费不扣课时）、当天
请假（按约定扣半节但老师仍应得课酬）、无故缺席（正常扣费）、老师
请假（不扣学生任何东西）。这些情况不能用一套自动规则完全覆盖，
所以设计上采用"分类给建议值，人工可覆盖"的模式（PRD第八章8.3节的
9种出勤分类字典）。

同时，本位币锁定这条规则（PRD第七章7.1节）之所以放在这个change里
而不是独立实现，是因为"锁定"这个动作的触发条件就是"产生首条财务
记录"，而课后确认正是最典型会产生首条财务记录的操作（充值虽然也会
产生financial record，但如果先做充值再做课后确认，锁定应该在充值
那个change就触发——本change和007-add-payment-and-ledger都需要
各自独立实现这个锁定判断逻辑，二者谁先执行触发都要正确）。

## What Changes
- 新增attendance_outcome_type字典表及其种子数据（含建议值三列，
  已在001中完成种子数据准备，本change负责读取接口）
- 新增attendance表（含建议值快照字段+实际值字段）
- 新增POST /lessons/{id}/confirm 事务实现
- 新增本位币锁定校验逻辑
- 新增课后确认前端表单

## Non-Goals
- 不实现课后确认的撤销/编辑功能（PRD业务规则R7：attendance一旦创建
  不可修改，如需修正走人工调整流水ADJUST流程，属于007的范围）
- 不实现基于历史出勤数据的自动化统计分析（如"这个学生请假率多高"
  这类洞察，属于012报表范围）

## Impact
- Affected specs: attendance-confirmation（新增）、base-currency-lock（新增）
- Affected code: backend/internal/lesson/(confirm相关)、
  backend/internal/finance/(ledger写入)、backend/internal/system/(本位币锁定)、
  frontend/admin/src/views/lesson/confirm/
- 依赖：005-add-lesson-scheduling（需要lesson表）、
  007-add-payment-and-ledger（需要ledger表结构，若尚未创建可在本change中
  一并建立最小表结构，007再补充充值相关的写操作）
- 被依赖：008（工作台需要展示待确认课次）、009（结款需要
  teacher_account_ledger数据）

## 风险与应对
本change不设时间压力，若与其他change冲突需要延后，优先保证本change的
完整性和测试覆盖，参照PRD第二十四章24.3节TC-01~07。任何AI工具生成的
本change代码，人工review时都应该重新走一遍design.md里描述的每一个
步骤，逐条对照代码是否真的做到了。
