# M2-GLM-01 验收反馈：拒收 `9bdc42e`

状态：**CHANGES_REQUESTED**。本反馈替代先前 READY 工单中的依赖版本；不得合并或推送到 `main`。

## P0（必须修复）

1. `backend/cmd/zedu-server/main.go` 缺少 `ZEDU_JWT_SECRET` 时使用公开默认值 `dev-only-change-me-in-production`。攻击者可据此签发 HS256 Owner token。必须在非测试运行中缺少、过短或弱 secret 时启动失败；测试必须显式注入安全测试 secret。不得保留可预测默认 secret。

## P1（必须修复）

1. `backend/go.mod` 将 `go 1.23.3` 改为 `1.25.0`，违反 M1/CI 基线和允许范围。`golang.org/x/crypto` 改用已修订冻结的 `v0.41.0`，恢复 `go 1.23.3`，恢复无关间接依赖；以 `GOTOOLCHAIN=local` 运行全部门禁，禁止借自动 toolchain 通过。
2. `AuthMiddleware` 仅验 JWT，不查 `user_account.status`。禁用账号后，既有 Bearer token 仍可访问。每次受保护请求必须确认账号仍 ACTIVE，并从数据库取得有效角色；禁用后立即返回 `40101`。补回归测试。
3. refresh 轮换使用多个独立 `Exec`，没有事务或条件撤销；并发请求可各签发新 session，插入失败可使旧 session 已失效。使用单 `BeginTx`、`UPDATE ... WHERE revoked_at IS NULL` 并检查受影响行数、插入新 session、签发/提交的原子序列。补并发“仅一个成功”和故障回滚测试。
4. `POST /users` 仅检查 password 非空，未满足 PRD 的至少 8 位且包含字母和数字。实现校验并为非法输入使用冻结的参数/状态错误语义；补测试。`/users` 只能创建 Operator，不得通过请求创建 Owner。
5. `teacher_capability.track_id` 和 `level_id` 可为 NULL，SQLite 的 UNIQUE 对 NULL 不冲突，无法保证三元唯一。两列必须 NOT NULL，或在明确的独立能力模型下使用等价的 NULL-safe 约束；按当前冻结契约采用 NOT NULL，补 NULL 与并发重复测试。
6. `student_teacher_assignment` 缺少 `enrollment_id WHERE status='ACTIVE'` 的部分唯一约束。新增数据库层限制并补迁移与并发测试，确保每个 enrollment 最多一个 ACTIVE assignment。
7. 登录失败计数不得采用“先 SELECT `login_fail_count` 再 SET 新值”的读改写。并发错误密码可以同时读到同一计数并绕过第五次锁定。使用原子 SQL 更新或单事务条件更新，保证第 5 个失败请求设置 15 分钟锁定；补 5/6 个并发失败请求的锁定回归测试。

## P2（本次一并修复）

1. 输入错误不得以 `40102`、`40301` 或 `40401` 搭配 HTTP 400 表示；统一使用冻结的参数/状态语义并让 HTTP status 与业务码一致。
2. 所有认证成功/失败日志必须带 request ID；未知用户名不得记录完整敏感标识。特别是 `Refresh` 的 query、begin transaction、revoke、insert、commit 错误日志均须附 `slog.String("request_id", rid)`；补刷新失败日志断言。
3. JWT 验证必须严格只接受 `HS256`，不能接受任意 HMAC 算法；未知用户名路径执行 dummy bcrypt compare，降低可观测时序枚举。

## 必须重新执行

```powershell
Set-Location backend
$env:GOTOOLCHAIN = 'local'
go version                    # 必须为 go1.23.3
go fmt ./...
go vet ./...
go test ./... -count=1
go test ./... -count=20
go build ./cmd/zedu-server
```

交付报告必须包含修复 commit SHA、红灯/绿灯证据、完整改动文件、所有未测试项及 migration down 回滚方式。仅修复上述项，不得扩展到 onboarding、目录、课程、前端或共享治理文件。
