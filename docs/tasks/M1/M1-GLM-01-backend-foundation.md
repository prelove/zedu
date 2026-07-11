# M1-GLM-01：后端工程基础（给 GLM）

## 执行身份与目标

你是本工单唯一实现者。目标是建立最小、真实可运行的Go后端基础：`GET /healthz`、包依赖边界、modernc SQLite连接、可逆迁移、PRAGMA以及脱敏结构化日志。不得实现任何业务功能。

从`origin/main`创建分支`m1/glm-backend-foundation`。如与Kimi并行，使用独立clone/worktree；不得在对方工作目录切换分支。

## 开始前必须读取

1. `AGENTS.md`
2. `docs/status/PROJECT_STATUS.md`
3. `docs/governance/GOVERNANCE.md`
4. `docs/standards/coding-standard.md`
5. `docs/standards/testing-standard.md`
6. `openspec/changes/establish-engineering-foundation/{proposal.md,design.md,tasks.md}`
7. `openspec/changes/establish-engineering-foundation/specs/engineering-foundation/spec.md`

执行前运行并记录：

```powershell
openspec instructions apply --change establish-engineering-foundation --json
openspec validate --all --strict --no-interactive
go version
```

## 写入范围

只允许新增/修改：

```text
backend/go.mod
backend/go.sum
backend/cmd/zedu-server/**
backend/internal/platform/database/**
backend/internal/platform/httpserver/**
backend/internal/platform/logging/**
backend/pkg/**
backend/migrations/**
backend/**/*_test.go
```

禁止修改根目录、`frontend/`、`openspec/`、`docs/`、`.github/`、`scripts/`及任何任务勾选。

## 技术约束

- module暂定`github.com/prelove/zedu/backend`。
- 使用标准库`net/http`完成首个健康检查；Gin/GORM只在后续确有需求时引入，禁止为了目录好看增加空层。
- SQLite驱动必须是`modernc.org/sqlite`，禁止CGO和`mattn/go-sqlite3`。
- PRAGMA至少验证`foreign_keys=ON`、`journal_mode=WAL`、`busy_timeout>=5000`。
- migration必须增量、可up/down/up；本工单只创建运行所需最小基础结构，禁止一次创建全部业务表。
- 结构化日志不得记录token、密钥、完整邮箱或请求体；每个请求具备request ID/correlation ID。
- 代码、SQL和测试数据统一UTF-8；测试包含中文、日文和emoji往返。

## TDD执行步骤

### A. 健康检查与依赖方向

1. 先写失败测试：启动测试server后`GET /healthz`返回200、`Content-Type: application/json`及`{"status":"ok"}`。
2. 运行测试，保存失败输出。
3. 实现最小handler/server使测试通过。
4. 写依赖边界测试或静态脚本测试：`pkg/`不得import`internal/`；platform包不得import未来业务模块。

### B. SQLite与迁移

1. 先写失败测试：临时库执行up/down/up；外键违规插入失败；三个PRAGMA生效。
2. 运行并保存失败输出。
3. 实现最小连接和migration runner。
4. 不得通过删除断言或放宽PRAGMA标准使测试通过。

### C. 日志

1. 先写失败测试：请求日志含request/correlation ID，敏感字段不会输出。
2. 实现最小中间件和日志接口。

## 必须执行的验收命令

```powershell
Set-Location backend
go fmt ./...
go vet ./...
go test ./... -race -count=1
go test ./... -run "TestHealth|TestMigration|TestPragma|TestUTF8|TestRedaction" -count=20
go build ./cmd/zedu-server
```

全部退出码必须为0。若Windows上的`-race`因工具链限制不可用，不能假装通过：记录BLOCKED证据，并提供非race结果和CI验证建议。

## 禁止事项

- 不得实现登录、用户、学生、老师、课程、充值、排课、通知、上传、结款API或模型。
- 不得生成31张业务表或使用`AutoMigrate`替代版本化迁移。
- 不得新增Docker、消息队列、Redis或其他未批准依赖。
- 不得修改OpenSpec勾选状态、路线图或自称任务已验收。

## 交付格式

提交一个Lore commit，并回复：commit SHA、改动文件、红灯命令/摘要、绿灯命令/摘要、依赖清单及理由、已知风险、未测试项。若需要超出写入范围，停止并标记`BLOCKED`。
