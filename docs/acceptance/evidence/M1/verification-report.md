# M1 工程基础验收报告

- 验收日期：2026-07-11
- 验收结论：**ACCEPTED**
- 范围：OpenSpec change `establish-engineering-foundation`
- 发布基线：`main` / `2402b14`
- 远端 CI：[GitHub Actions run 29153829469](https://github.com/prelove/zedu/actions/runs/29153829469)

## 独立审查结论

`M1-GLM-02` 的幂等 seed 实现经两轮复审后接受。生产公开 API `ApplyFoundationSeed` 不暴露故障注入能力；测试通过包内未导出的 `applyFoundationSeed(..., hook)` 验证“写入已发生、提交前故障、事务回滚、持久化记录为零”的完整序列。无全局 hook、setter 或并发共享状态。

## 本机验收证据（Windows 10 JP）

运行时：Go 1.23.3、Node 24.8.0、npm 11.7.0；所有文档与源文件均执行 UTF-8 校验。

| 项目 | 新鲜验证结果 |
|---|---|
| Go 格式与静态检查 | `go fmt ./...`、`go vet ./...` 成功 |
| Go 常规测试 | `go test ./... -count=1` 成功；14 个命名测试 |
| Go 稳定性 | `go test ./... -count=20` 成功 |
| Go 构建 | `go build ./cmd/zedu-server` 成功 |
| 前端依赖 | `npm ci` 成功；锁文件可重建 |
| 前端质量 | lint 零 warning、`vue-tsc --noEmit` 成功 |
| 前端测试 | 8 个测试文件、57/57 通过 |
| 覆盖率 | statements/branches/functions/lines 均为 100%（26/26、19/19、8/8、26/26） |
| 前端构建与审计 | Vite 8.1.4 构建成功；`npm audit --omit=dev --audit-level=high` 为 0 漏洞 |
| OpenSpec 与治理 | `openspec validate --all --strict --no-interactive`、编码校验、追溯校验均通过 |

三语资源 `zh-CN`、`ja-JP`、`en-US` 由类型约束及递归比对测试校验 key 一致；CJK 与 emoji 往返测试通过。日期与 JPY 格式化明确使用 locale 和 `Asia/Tokyo`，不依赖 Windows 系统显示语言。

## CI 验收证据

GitHub Actions run `29153829469` 的四个 job 全部成功：

| Job | 结论 | 关键覆盖 |
|---|---|---|
| governance / Ubuntu | 成功 | `npm ci`、OpenSpec strict、编码与追溯 |
| governance / Windows | 成功 | 同上，验证 Windows 路径与 shell |
| foundation / Ubuntu | 成功 | Go 格式、vet、测试、构建、**`go test ./... -race -count=1`**、前端全量门禁 |
| foundation / Windows | 成功 | Go 格式、vet、测试、构建、前端全量门禁；Linux-only race 步骤按工作流设计跳过 |

本机 Windows 默认 `CGO_ENABLED=0`，故本机不运行 `-race`；该风险已由 Ubuntu CI 的实际 race 成功结果关闭。

## 过程修复记录

1. 统一 CI 两个 job 的 Node 版本为 24.8.0，避免主版本浮动。
2. 修复历史归档文件被根目录 `backup/` 忽略规则误排除的问题：强制追踪经 `MANIFEST.sha256` 验证的 `legacy/.../specs/backup/spec.md`。此前 Ubuntu 干净 checkout 的追溯校验失败；修复后 Windows/Ubuntu 均通过。
3. 早期 run `29153775308` 失败仅作为诊断证据，不用于验收；最终结论只依据 `29153829469`。

## 范围与遗留项

- M1 只提供工程基础，不包含认证、人员、课程、报名、财务、排课或正式结款业务实现。
- M2 实现前必须先冻结学生邮箱重复语义及 API 失败/警告响应模型；该决策已登记为下一阶段前置条件。
