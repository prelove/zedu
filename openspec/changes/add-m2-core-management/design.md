## Context

M1 已提供 `net/http`、`database/sql`、modernc SQLite、迁移、结构化日志、Vue 3/Vite 和三语基础，但没有业务领域、身份、会话或页面路由。M2 要建立 Owner/Operator 的受控入口及其主数据，不得把财务、课次、排课、通知、备份、报表或结款提前带入。

业务事实源为 PRD v3.1-r1 与 ADR-007：学生/老师/家长不是登录主体；学生邮箱选填，填写时全局唯一，冲突一律为 `40901`；老师能力由 `(teacher_id, track_id, level_id)` 唯一。所有时间存 UTC，展示使用应用 locale 与系统时区；金额字段不在 M2 范围内。

## Goals / Non-Goals

**Goals:**

- 以现有 `net/http` + `database/sql` 基础实现清晰的 HTTP、应用服务和数据库边界。
- 提供 Owner/Operator 登录、短期 access token、可撤销 refresh token、退出和最小 RBAC。
- 提供受限的首次初始化，以及学生、家长、老师、能力、可授时间、课程字典、报名和师生安排的主数据链路。
- 让每个写操作可鉴权、审计、原子提交，并给前端提供稳定的成功、错误和分页契约。

**Non-Goals:**

- 不提供 Student/Teacher/Parent 登录、注册、自助改资料或公开 API。
- 不创建 lesson、attendance、payment、payment evidence、notification、backup、report、payout 或其可点击前端入口。
- 不在 M2 实现复杂价格、课酬或自动排课规则。

## Decisions

### 1. 单体分层和依赖方向

HTTP handler 只做请求解码、身份上下文和响应编码；每个领域的 application service 承担授权、状态转换和事务编排；repository 只执行参数化 SQL。依赖单向为 `http → application → repository/database`，领域实体不得依赖 HTTP。

拒绝方案：一次性引入完整 DDD/CQRS。它会为当前单 SQLite 进程增加事件总线、聚合和投影的维护成本，不能改善 M2 的验收能力。

### 2. 会话与密码

仅 `OWNER`、`OPERATOR` 可认证。密码只保存 bcrypt 哈希；连续失败 5 次锁定 15 分钟。access token 为短期 JWT（60 分钟），refresh token 为 14 天、随机生成、仅保存 SHA-256 哈希、单设备会话记录可撤销。刷新时轮换 refresh token；登出或禁用账号立即撤销对应会话。

access token 经 `Authorization: Bearer` 传递；refresh token 使用 `HttpOnly`、`Secure`、`SameSite=Strict` cookie。前端不得读取或持久化 refresh token。JWT 库与 bcrypt 实现必须在首个编码任务中经过依赖审查并锁定精确版本；禁止自行实现密码算法或 JWT 签名。

拒绝方案：只用长寿命 JWT。它无法可靠执行登出、禁用和泄漏 token 的即时失效。

### 3. 统一 HTTP 契约与权限

所有 JSON 返回统一外层：成功为 `{ "code": 0, "data": ... }`；失败为 `{ "code": <业务码>, "message": <三语可映射的稳定错误键>, "requestId": "..." }`。HTTP status 与 PRD 对应：未认证 `40101`/401、登录失败 `40102`/401、锁定 `40103`/401、权限不足 `40301`/403、缺失 `40401`/404、数据冲突 `40901`/409、非法状态 `42201`/422。

路由按资源分组；写请求由 auth middleware 写入 actor，再由 service 进行 Owner/Operator 检查。Owner 包含 Operator 权限；账号管理、模板重置仅 Owner。分页使用 `page`（从 1 开始）和 `pageSize`（1–100），响应 `items`、`page`、`pageSize`、`total`。

拒绝方案：让前端按 HTTP 文本或数据库错误字符串分支。它不可国际化、不可稳定测试，且会泄露 SQLite 实现细节。

### 4. 初始化与种子

`foundation_seed` 继续只表示工程基础。M2 的业务模板通过显式 `POST /onboarding/initialize` 应用，并在单事务内写入 `system_settings.initialized_at` 与模板数据。首次初始化仅允许未初始化状态；模板重置仅 Owner 且不存在学生、老师、报名、安排等业务数据时允许，否则返回 `42201`。重复请求读取既有结果，不重复插入字典。

拒绝方案：服务启动时自动写业务模板。启动副作用会掩盖操作人、无法选择模板，也不能安全控制重置。

### 5. 主数据、唯一约束与并发

学生 `email` 允许 NULL；非空值在 `student.email UNIQUE` 下唯一。创建和更新均由数据库约束最终裁决：前端预检仅改善体验，后端在唯一错误时映射 `40901`，不提供 bypass。软删除不释放邮箱，以满足 ADR-007 的“全局唯一”。

老师能力表保存 `teacher_id`、`track_id`、`level_id`、标签展示字段和有效期；数据库唯一约束严格为 `(teacher_id, track_id, level_id)`。能力结束通过 `effective_to` 记录，不覆盖历史；同一组合再次启用走编辑/恢复，不新增重复行。

报名与安排使用链路 `student → enrollment → assignment`。assignment 的激活/结束由单事务服务控制；M2 不产生 lesson。所有多表写（学生与家长、报名与安排、初始化、会话轮换）均以单一数据库事务完成，失败即回滚。

### 6. 审计、日志和前端隔离

所有成功的写操作在同一事务写入 `operation_log`，至少包含 actor、动作、资源类型/ID、时间、request ID 与不含密码/token 的摘要。认证失败只记录安全日志，不泄露用户名是否存在、密码、token 或完整邮箱。M1 的 request/correlation ID 贯穿响应和日志。

前端使用 API adapter 与 Pinia 风格的最小 auth/app store（具体库须在编码任务依赖审查后确定）。路由守卫只依据 `/auth/me` 或内存 access token 状态；未认证跳转登录，Owner-only 页面隐藏并由后端再次拒绝。三语资源采用现有 `LocaleSchema` 校验，错误键不可直接写为单语言文本。

## Risks / Trade-offs

- [Token 被盗用] → refresh token 仅存哈希、轮换、可撤销；cookie 属性和 CSRF 风险在 HTTP 集成测试中验证。
- [SQLite 并发写冲突] → 保持 M1 WAL/busy timeout 设置，所有跨表写单事务，映射唯一/锁冲突为稳定错误而非原始驱动错误。
- [邮箱唯一性与历史数据冲突] → M2 尚无生产导入；未来导入必须预扫描、报告重复且不得自动合并。
- [初始化误重置] → Owner 限制、业务数据存在即拒绝、审计操作；不提供静默清库。
- [范围蔓延] → 路由、迁移和前端导航均以 Non-Goals 负面检查；验收测试断言不存在财务、排课、通知与结款入口。

## Migration Plan

1. 新增顺序迁移：账户/会话/设置/审计表，随后课程字典与主数据表，最后报名与安排表及索引/约束。
2. 每个迁移必须具备 up/down/up、外键、唯一约束和 UTF-8 往返测试；down 仅删除本 change 新建对象。
3. 部署时先备份数据库，再运行迁移；首次业务模板仅由 Owner 显式触发。
4. 回滚仅在尚无 M2 业务数据时执行 migration down；已有数据时采取前向修复迁移，禁止破坏性回滚。

## Open Questions

- 无阻断性业务问题。实现前的技术性依赖审查必须锁定 JWT、bcrypt、前端路由和状态管理的精确版本；未经审查不得新增依赖。
