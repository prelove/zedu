# M2-GLM-02A：Owner 显式初始化与受限重置

## 工单元数据

| 项目 | 内容 |
|---|---|
| 状态 | ACCEPTED |
| Owner | Codex |
| 基线 | `main` 的 M2-GLM-01 已验收状态（PR #2 / merge commit `15a9c0a`） |
| OpenSpec | `add-m2-core-management` 任务 2.1；`onboarding-initialization` spec |
| 依据 | PRD v3.1 §5.2–§5.4、OpenSpec design §4、`contract-freeze.md` |
| 交付目标 | 仅实现 Owner 的显式模板初始化与无业务数据时的受限重置 API |

## 固定边界

允许新增或修改：

- `backend/internal/app/onboarding/**` 及其测试；
- 必要的认证后路由装配（`backend/cmd/zedu-server/main.go` 或既有 HTTP 路由文件）；
- 为该切片所必需的现有测试辅助代码。

默认禁止新增 migration。`003_m2_auth_and_core` 已提供 `system_settings`、`operation_log`、课程字典和业务数据表。若发现确实缺少不可替代的字段、索引或约束，先提交最小变更建议与证据，等待 Codex 批准；不得自行创建 `004`。

禁止修改：`frontend/**`、`docs/**`、`openspec/**`、`.github/**`、共享状态/路线图/任务勾选、认证基础实现、人员目录、课程维护 API、报名、安排、lesson、attendance、payment、凭证、notification、backup、report、payout。不得新增依赖、路由以外的认证角色或生产测试旁路。

## 路由与行为

1. `POST /onboarding/initialize`，仅 Owner。请求显式选择 `japanese`、`k12` 或 `blank` 模板；未知模板返回 HTTP 422 / `42201`，不写任何数据。
2. 第一次初始化必须在单一事务中写入选定模板数据、`system_settings` 初始化标记及 `operation_log`。审计包含 actor、action、target_type、target_id、request_id 和不含敏感信息的 JSON 摘要。
3. 已初始化时重复同一请求返回既有初始化结果，不重复插入模板或再次写成功审计；不得静默切换到另一个模板。
4. `POST /onboarding/reset`，仅 Owner。仅当 `student`、`teacher`、`student_course_enrollment`、`student_teacher_assignment` 均无业务记录时允许；否则返回 HTTP 422 / `42201`，不改数据、不写成功审计。
5. 成功 reset 在单一事务中替换模板数据、更新初始化标记与写审计；不得删除账号、refresh session、operation log 或 foundation seed。
6. 服务启动不得自动写业务模板；`foundation_seed` 不是业务初始化标记。

模板数据以 PRD §5.3（日语）和 §5.4（K12）为准：日语模板含 1 个日语领域、4 个方向、9 个等级、9 个能力标签；K12 模板按 PRD 列出的学科/年级配置；blank 仅写初始化标记。不得让业务逻辑依赖“日语”“N1”“数学”等显示名称。

## 先红后绿的必测场景

先提交失败测试，再实现最小代码。至少覆盖：

1. Owner 首次选择三种模板各成功一次；日语模板计数/层级正确，blank 无课程字典数据；CJK 与 emoji 往返。
2. 同一初始化请求幂等：模板数据和成功审计均不重复；不同模板的后续请求不切换既有结果。
3. Operator 调 initialize/reset 得 HTTP 403 / `40301`，且无模板/标记/审计副作用；未认证为 40101。
4. 有任一受保护业务记录时 reset 得 HTTP 422 / `42201`，原数据和初始化标记保持不变，且无成功审计。
5. 初始化或 reset 中模板写入、标记写入或审计写入失败时，事务完整回滚；返回 HTTP 500 / `50002`，不暴露 SQLite 文本。
6. 每个成功写操作有且仅有一行对应审计，`request_id` 非空，且 detail JSON 可解析并无 password、password_hash、Authorization、access/refresh token 或哈希。
7. 不出现 lesson、attendance、payment、notification 或其他禁止范围的数据写入/路由。

## 门禁与交付

```powershell
Set-Location backend
$env:GOTOOLCHAIN='local'
go version                 # 必须 go1.23.3
go fmt ./...
go vet ./...
go test ./... -count=1
go build ./cmd/zedu-server
```

`go test ./... -count=20` 留待 M2 里程碑候选验收或出现非确定性失败时执行；本切片仍必须运行定向的并发、权限与事务故障注入测试。

提交必须遵循 Lore Commit Protocol，且只暂存允许范围文件。交付报告必须给出：基线 SHA、红灯→绿灯证据、完整改动文件、测试与上述场景的对应关系、未测试项、回滚方式与风险。不得更新 OpenSpec 勾选、状态、路线图或合并 `main`；等待 Codex 独立验收。

## 验收记录

- 实现与合并：`3bc40782e8b1b81880951faddbfa2cec51a49695`，2026-07-16。
- 独立审查：无未解决 P0/P1；补正审计全局目标 `system/1` 与 reset 审计失败回滚证据。
- GitHub Actions run `29500940531`：Windows/Ubuntu 治理与 foundation 全绿，Ubuntu 已通过 `go test ./... -race -count=1`。
