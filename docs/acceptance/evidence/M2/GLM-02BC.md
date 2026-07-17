# M2-GLM-02B/02C 验收证据

验收日期：2026-07-17
验收结论：ACCEPTED

## 范围与提交

- 实施链：`0a574e0`（人员资料）、`2c0f111`（课程字典）、`d8eb497`（报名与安排）、`af37523`（首轮复验整改）。
- 本次独立验收的局部收口：连续等级事件以前一条事件的目标等级为来源；课程选择变更与等级事件写入不可在同一 PATCH 混合；等级事件引用阻止层级重挂；`004_student_level_event` 枚举与 PRD §12 一致。
- 未扩展前端、认证、初始化、排课、课消、财务、通知、备份或报表范围；未新增依赖和 API。

## 验收结果

- OpenSpec：`openspec validate add-m2-core-management --strict` 通过。
- 后端：Go `1.23.3`、`GOTOOLCHAIN=local`；`gofmt`、`go vet ./...`、`go test ./... -count=1`、`go build ./...` 已执行。
- 本次针对性回归：
  - 先红：连续等级变更写成 `N5→N4`、`N5→N3`；跨课程方向与等级组合 PATCH 返回 200；仅被等级事件引用的等级可重挂。
  - 后绿：`TestSequentialLevelChangesFollowTheLatestRecordedLevel`、`TestCourseSelectionAndLevelChangeAreRejectedTogether`、`TestLevelEventSchemaUsesPRDEventTypesAndProtectsEventReferencedLevels` 通过。
  - 原有课程选择审计语义由 `TestCourseSelectionChangeAuditsBeforeAndAfter` 保留覆盖。

## 已知验证边界

- Windows 本机保持 CGO 禁用，未运行 `-race`；Linux race 继续由后续 CI/M2 集成验收执行。
- 20 次重复稳定性扫描仅在里程碑候选、迁移/并发基础设施变更或出现非确定性失败时执行，不作为每次局部修复门禁。
