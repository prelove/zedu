# Design: 报名与师生安排的关系设计

## 为什么余额字段挂在enrollment而不是assignment或student上

- 不挂在student上：因为一个学生可能有多个课程项目，各自的收费和
  课时消耗应该独立计算（PRD第六章第一层价格模型：
  charge_per_lesson_amount是enrollment级别的字段）
- 不挂在assignment上：因为换老师时"余额不应该随老师变动"
  （PRD9.3节异常分支明确要求），如果余额挂在assignment上，换老师
  就意味着要把余额从旧assignment搬到新assignment，这是不必要的
  复杂度，也容易在搬迁过程中出错。余额天然属于"学生对这个课程项目
  的投入"，跟具体是哪个老师教无关

## role_type的三种取值的语义边界
- MAIN：主责老师，一个enrollment在任意时刻只能有一条status=ACTIVE
  的MAIN记录，这是整个换老师逻辑的核心约束
- SUBSTITUTE：临时代课，不结束MAIN记录，代课期间MAIN记录仍然是
  ACTIVE状态，只是某几次具体课次(lesson)的teacher_id指向代课老师，
  这种场景下assignment的作用更多是"记录曾经有代课发生过"这个事实，
  具体哪次课是谁代课，最终以lesson表自己的teacher_id为准
- ASSISTANT：辅助老师（如助教），不参与主要授课，V1不深入使用，
  只预留这个枚举值供未来班级课场景使用

## enrollment_type的状态转换设计
TRIAL（试听）→ONE_TO_ONE（正式）是一个允许的合法转换，代表"试听
满意后转正式报名"，这个转换本质上是修改enrollment_type字段本身，
而不是创建一条新的enrollment记录——因为试听期间产生的课次和账务
记录应该延续在同一个enrollment下，不应该因为"转正"这个业务动作
而被割裂到两条不同的enrollment记录里，导致历史查询时出现断层。

## enrollment状态机
```
ACTIVE ⇄ PAUSED（可双向切换，对应"学生暂停学习一段时间"）
ACTIVE/PAUSED → COMPLETED（达成学习目标，终态）
ACTIVE/PAUSED → CANCELLED（学生终止，终态）
```
COMPLETED和CANCELLED都是终态，不允许从终态恢复到ACTIVE——如果
运营者操作失误把一个还在继续学习的项目标记成了COMPLETED，正确的
补救方式是新建一条新的enrollment，而不是"编辑"已有记录改回ACTIVE，
这是为了保证enrollment的状态转换历史本身也具有审计意义（不应该
出现"这个项目完成了又没完成"这种反复横跳的记录）。
