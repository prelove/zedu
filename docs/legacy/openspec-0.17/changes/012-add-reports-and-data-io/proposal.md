# Proposal: 报表图表与数据导入导出

## Why
工作台在008中只做了极简版（三个列表，无图表），运营者需要更直观的
经营数据可视化来做决策；同时001~011的所有change都假设数据是通过
UI一条条手动录入的，但真实场景下运营者手上已经有一份Excel在维护
现有学生老师信息，如果不支持批量导入，"从Excel迁移到Zedu"这个
关键的上线前置动作就无法完成。

## 业务背景
根据PRD第十章10.2/10.10/10.11节，工作台完整版需要展示月度收入趋势、
课程方向占比、学生增长趋势三类图表，财务报表需要支持按月/周/日切换
粒度并导出Excel/PDF。这些图表数据本质上都是对已有账务和排课数据的
聚合统计，不需要新增任何业务表，纯粹是查询和展示层面的能力。

数据导入导出方面，PRD第十四章业务规则R14明确了"Excel导入时若邮箱
已存在则整行跳过并在导入报告中标注，不做自动合并"这条规则——这是
一条刻意保守的设计：宁可要求运营者手动处理冲突数据，也不做自动
合并猜测，避免把两个实际上不同的学生的数据错误地合并到一起。

## What Changes
- 新增报表数据接口：/reports/revenue-trend、student-growth、
  track-distribution、teacher-workload、completion-rate（PRD13.10节）
- 新增财务报表接口 GET /finance/report + 导出接口(Excel/PDF)
- 新增Excel批量导入学生/老师接口（含导入报告）
- 新增工作台完整版（接入ECharts）、数据图表页、财务报表页前端
- 新增操作日志查询页面

## Non-Goals
- 不实现自定义报表构建器（运营者不能自己拖拽字段生成新图表，
  五个图表的维度是PRD预先定义好的，不做可配置的BI工具）
- 不实现导入历史课次/充值记录（PRD附录C问题31已建议只导入当前
  余额快照，不做历史流水的批量导入，这是与运营者business确认过
  的范围边界，历史流水从系统上线日起重新产生）
- 不实现导入模板之外的自定义字段映射（导入模板的列是固定的，
  不支持运营者自己上传任意格式的Excel并做字段映射配置）
- 不在本change中实现PDF模板的可视化编辑（导出的PDF/Excel格式是
  代码里固定的模板，不提供后台可编辑的报表样式配置）

## Impact
- Affected specs: reports（新增）、data-io（新增）
- Affected code: backend/internal/report/(完整版)、
  backend/internal/student/(import相关)、
  backend/internal/teacher/(import相关)、
  frontend/admin/src/views/dashboard/(完整版)、views/reports/、
  views/finance/report/、views/system/operation-logs/
- 依赖：008-add-dashboard-and-backup（工作台极简版已存在，本change
  是在其基础上补充图表）、007-add-payment-and-ledger（报表需要
  ledger数据）、003-add-student-teacher-profile（导入功能需要
  student/teacher的CRUD接口已存在）
- 被依赖：无（本change是MVP之后V1完善阶段的功能补充，不阻塞其他
  change的核心链路）
