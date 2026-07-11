# M1-GLM-02：幂等模板 Seed 基础（给 GLM）

## 目标

补齐 `establish-engineering-foundation` 的 OpenSpec tasks 2.3：建立**最小、可测试、幂等**的模板 seed 框架。此任务只提供框架和一个无业务语义的 foundation 样例，禁止提前写入日语课程、学生、老师、账号、支付或结款数据。

## 开始前

从当前集成基线创建分支 `m1/glm-seed-foundation`，完整读取：

1. `AGENTS.md`
2. `docs/governance/GOVERNANCE.md`
3. `docs/status/PROJECT_STATUS.md`
4. `docs/standards/coding-standard.md`
5. `docs/standards/testing-standard.md`
6. `openspec/changes/establish-engineering-foundation/{proposal.md,design.md,tasks.md}`
7. `openspec/changes/establish-engineering-foundation/specs/engineering-foundation/spec.md`

## 允许修改范围

```text
backend/internal/platform/database/**
backend/migrations/**
backend/**/*_test.go
backend/cmd/zedu-server/main.go
```

不得修改前端、根目录、OpenSpec、docs、CI、共享任务状态；不得增加新的第三方依赖。

## 技术契约

- seed 必须在 migration 完成后显式调用；不得把 seed 藏在连接初始化副作用中。
- 设计一个清晰的 `ApplyFoundationSeed(ctx, db)` 或等价 API，返回错误且可重复调用。
- 样例数据仅可使用非业务能力标记，例如 `foundation_marker`；不得模拟课程、账户或财务事实。
- 第二次调用不得新增重复记录，且应保留第一次数据不变。
- 用数据库唯一约束和明确SQL保证幂等，不能只在Go内先查再插。
- 失败时不得留下半写入：即使当前只有一个标记，也必须使用事务并写故障注入/回滚测试。
- 中/日/emoji 文本作为seed元数据往返必须正确。

## TDD步骤

1. 写失败测试：首次seed创建一个foundation标记；运行并保存红灯。
2. 写失败测试：重复seed后记录数量仍为1、内容不变；运行红灯。
3. 写失败测试：故障注入时事务回滚，无半写入；运行红灯。
4. 实现最小migration和seed runner使全部通过。
5. 不得勾选OpenSpec task；验收者会统一更新。

## 必须执行

```powershell
Set-Location backend
go fmt ./...
go vet ./...
go test ./... -count=1
go test ./... -run "Test(FoundationSeed|Migration|Pragma|UTF8)" -count=20
go build ./cmd/zedu-server
```

## 交付

提交Lore commit并返回：commit SHA、文件列表、红灯与绿灯输出、幂等机制说明、故障注入证明、未测试项。遇到需要业务seed或共享文件改动时立即 BLOCKED。
