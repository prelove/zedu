# Proposal: 学生、老师与课程体系基础档案

## Why
学生档案、老师档案、课程体系是后续排课/账务的基础数据，三者高度关联
（学生报名依赖课程体系，老师能力也依赖课程体系），放在同一个change中
统一实现更符合业务内聚性，且可以让评审人一次性看到"这三类基础数据
之间怎么互相引用"的完整图景。

## 业务背景
根据PRD第九章9.1/9.2节的业务流程描述，学生和老师建档不是简单的表单
录入，而是承载了真实运营场景的细节：
- 学生可能没有邮箱（PRD 9.1节异常分支），这种情况下应允许保存但要
  提示"将无法接收自动提醒"，而不是强制要求邮箱
- 邮箱重复不应阻止创建（双胞胎、家长用同一邮箱注册多个孩子等真实
  场景），只做提示不做拦截
- 老师能力不是老师档案上的一个静态字段，而是可以有多条、随时间变化
  的记录（PRD第五章5.5节），因为老师的教学能力天然是多维度的
- 课程体系必须保持"结构通用、不锁死学科"（PRD原则一），因此本change
  的课程体系维护功能设计成三层联动的通用配置界面，而不是针对日语
  写死的表单

## What Changes
- 新增student/parent CRUD
- 新增teacher/teacher_capability/teacher_availability CRUD
- 新增course_domain/track/level/skill_tag CRUD
- 新增学生列表/详情页(基础信息Tab)、老师列表/详情页(能力与时间Tab)、
  课程体系维护三栏联动页面

## Non-Goals
- 不实现学生/老师的物理删除（只支持软删除标记deleted_at，历史课次
  和账务记录必须能追溯到已"删除"的学生老师）
- 不实现Excel批量导入（留给012-add-reports-and-data-io）
- 不实现学生/老师的多字段模糊搜索之外的高级筛选（如按最近上课时间
  排序这类，留待后续需要时再加）

## Impact
- Affected specs: student（新增）、teacher（新增）、course-system（新增）
- Affected code: backend/internal/student/、backend/internal/parent/、
  backend/internal/teacher/、backend/internal/course/、
  frontend/admin/src/views/student/、views/teacher/、views/course/
- 依赖：001-add-project-scaffold
- 被依赖：004（enrollment需要student/teacher/course表）、
  005（lesson需要course表引用）
