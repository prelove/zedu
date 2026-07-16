# M2-GLM-01 独立验收记录

- 验收日期：2026-07-16
- 状态：ACCEPTED
- 实现分支：`m2/glm-auth-foundation`
- 合并记录：PR #2，merge commit `15a9c0a31678d91943a6b20ee6c5eabfe549be22`
- 最终修复：`22fc17c`（审计详情 JSON 编码与测试注释 UTF-8 修复）

## 验收范围

本记录仅覆盖 OpenSpec change `add-m2-core-management` 的任务 1.1、1.2、1.3：认证/RBAC、统一 HTTP 契约、M2 基础迁移与认证/账号审计原子性。不覆盖初始化、人员资料、课程、报名、安排、财务、排课、通知或前端页面。

## 独立审查结论

1. bcrypt、HS256 JWT、60 分钟 access token 与 14 天 HttpOnly/Secure/SameSite=Strict refresh cookie 符合冻结契约。
2. 登录、refresh 轮换、logout、创建 Operator、禁用账号的成功业务写入与 `operation_log` 在同一事务；禁用与 refresh 并发后的不变量为“DISABLED 账号无活跃 refresh session”。
3. `student.email` 非空唯一、`teacher_capability` 三元唯一及单 enrollment ACTIVE assignment 的数据库约束均有迁移回归。
4. 审计明细改为结构化 JSON 编码；含引号、反斜线和换行的用户名不会破坏审计 JSON，且密码/token 不写入日志、响应或审计。
5. 数据库/事务错误统一为 HTTP 500 / `50002`；非数据库内部错误为 HTTP 500 / `50001`。本记录同步修正冻结表对 PRD §14.3 既有错误码的遗漏。

## 新鲜门禁证据

- 本机（Go 1.23.3）：`gofmt`、`go vet ./...`、`go test ./... -count=1`、`go build ./cmd/zedu-server` 通过。
- GitHub Actions [run 29498664305](https://github.com/prelove/zedu/actions/runs/29498664305)：
  - governance：Ubuntu、Windows 成功；包含 OpenSpec strict、UTF-8 与追溯检查。
  - foundation：Ubuntu、Windows 成功；包含后端格式/vet/test/build、前端质量与构建。
  - Ubuntu：`go test ./... -race -count=1` 成功。

## 已知非阻塞项

- 真实 TLS 下浏览器对 `Secure` refresh cookie 的验证留待 M2-CODEX-02 真实 HTTP/浏览器验收。
- JWT 密钥轮换属于生产运维能力，不在本切片范围。
- OpenSpec 顶层任务勾选保留到整个 M2 里程碑验收，避免把子任务验收误记为里程碑完成。
