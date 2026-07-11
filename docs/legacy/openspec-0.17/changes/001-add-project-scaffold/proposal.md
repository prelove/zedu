# Proposal: 工程脚手架与数据基础

## Why
所有后续capability都依赖一个能跑起来的工程骨架和已建好的数据库结构。
这是全项目唯一没有直接"业务价值"但阻塞一切的change，同时也是后续所有
change能否遵守PRD架构约束的第一道关卡——如果这里的目录规范、迁移机制、
种子数据没有打好，后面每个change都会在这些基础问题上反复踩坑。

## 业务背景
根据PRD第十六章16.1/16.2节，本项目明确不采用任何现成Go后台管理框架
（gin-vue-admin、go-admin、soybean-admin-go等），原因是它们普遍默认绑定
MySQL/PostgreSQL+Redis，与"SQLite+单文件All-in-One"的部署目标直接冲突。
这个决策的代价是：我们不能"抄"一个现成脚手架，必须自己把这几十张表、
种子数据、迁移机制从零搭对。

同时PRD第一章1.3节明确了首个部署实例的业务背景：日语教育场景，本位币
JPY，这意味着种子数据不是随便造几条测试数据，而是要精确对应PRD第五章
5.3节列出的完整日语模板结构（4个方向、9个等级、9个能力标签），后续
Sprint的所有开发和测试都依赖这份种子数据的准确性。

## What Changes
- 建立backend/frontend/docs三层仓库目录，严格遵循PRD16.3节的目录约定
- 建立全部31张表的goose迁移文件（见PRD第十二章12.2节完整DDL），
  包含正向迁移和回滚迁移
- 写入日语模板种子数据、K12模板种子数据（供未来复制部署使用，本次
  不应用）、支付方式字典、出勤分类字典、基础系统配置
- 建立三条PRAGMA配置（WAL/foreign_keys/busy_timeout）作为服务启动的
  强制前置步骤，不是可选项

## Non-Goals（本change明确不做的事）
- 不实现任何业务API（学生/老师/课程CRUD留给后续change）
- 不实现认证机制（留给002）
- 不做MySQL/PostgreSQL的迁移文件（PRD预留了目录结构migrations/mysql、
  migrations/postgres，但内容留到真正需要迁移数据库时再补，本change
  只做sqlite目录下的内容）

## Impact
- Affected specs: data-foundation（新增）
- Affected code: backend/cmd/、backend/internal/、backend/migrations/、
  backend/pkg/、frontend/admin/（拉取Soybean Admin脚手架）
- 依赖：无（这是第一个change）
- 被依赖：002~014全部change都直接依赖本change产出的表结构和种子数据
