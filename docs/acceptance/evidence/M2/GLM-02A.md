# M2-GLM-02A 初始化切片验收记录

- 验收日期：2026-07-16
- 状态：ACCEPTED
- 实现提交：`3bc40782e8b1b81880951faddbfa2cec51a49695`
- OpenSpec：`add-m2-core-management` 任务 2.1（顶层勾选保留至 M2 整体验收）

## 范围与结论

实现 Owner-only 的 `POST /onboarding/initialize` 和 `POST /onboarding/reset`。支持日语、K12、空白模板；首次初始化原子写入模板、系统标记和审计，重复请求返回既有结果且不重复审计。重置仅在 student、teacher、enrollment、assignment 均不存在时进行。

未实现人员、课程维护、报名、安排、lesson、attendance、payment、notification、backup、report 或 payout。

## 红绿与审查证据

- 红灯：审计 target_id 断言在实现前因 NULL 失败；随后固定为全局 onboarding 目标 `system/1`。
- HTTP 场景：三模板、日语 1/4/9/9 层级、幂等、未知模板无副作用、Owner/Operator/未认证、业务数据阻止 reset、成功 reset 审计、初始化与 reset 的 SQLite trigger 审计故障回滚。
- 独立复审：无 P0/P1；复核确认 audit target 与 reset 回滚两项修正已关闭。

## 门禁

- 本机：Go 1.23.3；定向 onboarding HTTP 测试、`go vet`、单次全仓 Go 测试、构建、OpenSpec strict、UTF-8 与追溯检查通过。
- GitHub Actions [run 29500940531](https://github.com/prelove/zedu/actions/runs/29500940531)：治理与 foundation 的 Windows/Ubuntu 四项 job 全绿；Ubuntu `go test ./... -race -count=1` 成功。

## 风险与后续

- 20 次全仓稳定性扫描已按更新后的测试规范移至里程碑/并发基础设施门禁；本切片不以其替代事务故障和 race 测试。
- 首个 Owner 的安全创建属于部署引导，不在本切片临时新增；后续部署设计需明确其受控流程。
