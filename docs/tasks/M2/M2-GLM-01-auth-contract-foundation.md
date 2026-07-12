# M2-GLM-01：认证、统一契约与 M2 数据基础

## 工单元数据

| 项目 | 内容 |
|---|---|
| 状态 | READY |
| Owner | GLM |
| 建议分支 | `m2/glm-auth-foundation`，从 `main` 的 `ad9c661` 创建 |
| OpenSpec | `add-m2-core-management` 的 1.1、1.2、1.3 |
| 依据 | `docs/acceptance/evidence/M2/contract-freeze.md`、四项 M2 specs、PRD v3.1-r1、ADR-007 |
| 交付目标 | 可由前端消费的认证 API 与可迁移的 M2 数据基础；不实现初始化、人员、课程、报名或页面 |

## 允许与禁止范围

允许修改或新增：

- `backend/go.mod`、`backend/go.sum`：只能新增 `github.com/golang-jwt/jwt/v5 v5.3.1` 与 `golang.org/x/crypto v0.54.0`，不得升级或替换其他依赖。
- `backend/migrations/003_m2_auth_and_core.up.sql`、`backend/migrations/003_m2_auth_and_core.down.sql`。
- `backend/internal/platform/auth/**`、`backend/internal/app/auth/**`。
- `backend/internal/platform/httpserver/**`、必要的 `backend/cmd/zedu-server/main.go` 路由装配。
- 上述目录的 `*_test.go`。

禁止修改：`frontend/**`、`docs/**`、`openspec/**`、`.github/**`、`scripts/**`、路线图、状态、看板、任务勾选；禁止新增业务页面、lesson、attendance、payment、evidence、notification、backup、report、payout，或 Student/Teacher/Parent 登录。

## 必须实现的契约

1. 严格使用 `contract-freeze.md` 的 JSON 外层、错误码、角色矩阵、路由和 token 规则。
2. `POST /auth/login`：仅 ACTIVE Owner/Operator；成功返回 access token 并设置 `HttpOnly; Secure; SameSite=Strict` refresh cookie；密码和 token 绝不输出到日志、JSON 或审计摘要。
3. `POST /auth/refresh`：refresh token 只保存 SHA-256 哈希，14 天有效，轮换后旧 token 必须失效。
4. `POST /auth/logout`、`GET /auth/me`、Owner-only `/users` 与 `/users/{id}/disable`。
5. 连续 5 次失败锁定 15 分钟；失败消息不得暴露用户名是否存在。
6. M2 迁移至少创建 user account、refresh session、system settings、operation log 与后续 M2 所需主数据表；每个表有明确外键、必要索引和 down migration。
7. `student.email`：允许 NULL；非空全局唯一且软删除不释放；数据库约束最终裁决，冲突统一 HTTP 409 / `40901`。
8. `teacher_capability`：数据库唯一键严格为 `(teacher_id, track_id, level_id)`。

## 先红后绿的测试清单

先写并运行失败测试，再实现最小代码。测试至少覆盖：

- login 成功、未知用户名与错误密码同一外部响应、五次失败锁定；
- refresh 轮换、旧 refresh token 重放失败、logout 后 refresh 失败、禁用账号后失效；
- 未认证 `40101`、Operator 调 Owner-only API `40301`；
- 成功/失败 JSON 外层和 request ID；敏感日志脱敏；
- migration up/down/up、外键、CJK/emoji 往返；
- NULL student email 可多行、非空 email 只能一行、并发重复写仅一个成功且其他 `40901`；
- teacher capability 三元重复 `40901`。

## 本机门禁

```powershell
Set-Location backend
go mod tidy
go fmt ./...
go vet ./...
go test ./... -count=1
go test ./... -count=20
go build ./cmd/zedu-server
```

不得在本机 CGO 受限时伪造 `-race` 结果；Linux race 由 Codex 的 CI 验收执行。

## 交付与回滚

提交前只暂存允许范围内的文件。提交信息遵循 Lore 协议。交付报告必须包含：commit SHA、红灯/绿灯命令输出、改动文件、精确新增依赖、所有未测试项、migration down 回滚命令与已知风险。

未经 Codex 的 diff 审查、独立 Reviewer 与新鲜门禁证据，不得合入 `main`、不得勾选 OpenSpec 任务、不得宣称 ACCEPTED。
