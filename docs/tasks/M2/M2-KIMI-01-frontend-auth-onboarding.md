# M2-KIMI-01：前端认证、受保护路由与 Owner 初始化界面

## 工单元数据

| 项目 | 内容 |
|---|---|
| 状态 | READY |
| Owner | Kimi |
| 基线 | `main` 的 `808d25e`（M2-GLM-01、02A、02B/02C 已验收） |
| OpenSpec | `add-m2-core-management` 任务 3.1、3.2；`contract-freeze.md` |
| 依据 | PRD v3.1 §5.2–§5.4、§13–§15，`docs/acceptance/evidence/M2/contract-freeze.md` |
| 交付目标 | 建立可复用的前端认证基础，并实现 Owner 可操作的显式模板初始化界面 |

## 开工前必读

1. `docs/status/PROJECT_STATUS.md`
2. `docs/roadmap/MASTER_ROADMAP.md`
3. `docs/governance/GOVERNANCE.md`
4. `openspec/changes/add-m2-core-management/{proposal.md,design.md,tasks.md}`
5. `docs/acceptance/evidence/M2/contract-freeze.md`
6. 本工单及 `docs/acceptance/evidence/M2/GLM-02A.md`

后端已验收的真实契约是唯一事实源；不得采用预审报告中的候选 `/api/*` 路径、候选字段或“邮箱 warning”语义。前端请求路径不带 `/api` 前缀。

## 固定范围

允许新增或修改：

- `frontend/package.json`、`frontend/package-lock.json`（仅新增冻结的 `vue-router@5.1.0`）；
- `frontend/vite.config.ts`（仅开发代理）；
- `frontend/src/router/**`、`frontend/src/api/**`、`frontend/src/stores/**`；
- `frontend/src/features/auth/**`、`frontend/src/features/onboarding/**`；
- `frontend/src/App.vue`、必要的共享组件与 `frontend/src/i18n/**`；
- `frontend/tests/**` 或同等前端单元测试目录。

禁止修改：`backend/**`、`docs/**`、`openspec/**`、`.github/**`、状态/路线图/任务勾选；不得新增 Pinia、UI 框架或其他依赖。不得实现学生/老师/课程/报名/安排页面、lesson、attendance、payment、凭证、notification、backup、report、payout、正式结款、学生/老师/家长登录，也不得注册这些业务路由或导航入口。

## 冻结 API 与交互契约

所有成功响应为 `{ code: 0, data }`，失败响应为 `{ code, message, requestId }`；错误消息使用稳定键本地化，页面不得展示底层错误文本或 requestId 以外的调试信息。

| 请求 | 请求体 / 成功 `data` | 前端规则 |
|---|---|---|
| `POST /auth/login` | `{username,password}` → `{accessToken,role}` | access token 仅存内存；refresh cookie 由浏览器处理，绝不从 JSON、storage 或日志读取/写入。 |
| `POST /auth/refresh` | 无 body → `{accessToken,role}` | 仅遇到一次受保护请求的 `40101` 时刷新一次并重放原请求；刷新或重放仍为 `40101` 时清理内存会话并跳转登录。禁止递归、并发风暴和对 login/refresh 请求自身重试。 |
| `POST /auth/logout` | 无 body | 成功后清理内存会话并跳转登录；失败不伪装为成功。 |
| `GET /auth/me` | → `{id,username,role,displayName}` | 应用初始化时用现有内存 token 恢复用户；无 token 不请求。 |
| `POST /onboarding/initialize` | `{template:"japanese"|"k12"|"blank"}` → `{template,reused}` | 仅 Owner 路由与按钮可达；成功显示所选模板或“已使用既有模板”结果。 |
| `POST /onboarding/reset` | 同上 | 本任务只提供明确的 Owner 操作入口及二次确认；`42201/RESET_NOT_ALLOWED` 显示稳定本地化提示，不自行规避限制。 |

`/onboarding` 没有状态查询 API：不得假设或伪造“系统已初始化”。Owner 可主动打开初始化页；重复 initialize 的 `reused` 响应是唯一的既有状态反馈。

## 路由和最小页面

1. `/login`：公开。已认证用户访问时重定向至 `/`。
2. `/`：受保护的最小已登录首页，显示当前用户名/角色、退出操作、语言切换；不得承诺或链接 M2-KIMI-02 的业务页面。
3. `/onboarding`：受保护且仅 Owner；Operator 访问应显示本地化无权限状态并重定向至 `/`，不能向后端发初始化请求。
4. 未认证访问 `/`、`/onboarding`：保留目标地址，登录成功后回跳；无效或失效会话必须回到 `/login`。
5. `/login` 页面提供用户名、密码、提交中禁用和本地化错误；不得记录或渲染密码、token、cookie。

开发代理须保持真实路径：为 `/auth` 与 `/onboarding` 指向 `http://localhost:8080`，不做 `/api` 重写；已有 `/healthz` 检查不回退。

## 三语与可访问性

- `zh-CN`、`ja-JP`、`en-US` 键严格同构；新增登录、会话、权限、初始化、重置确认、模板和 API 错误键。
- 所有表单字段有可见 label、关联 `for/id`、错误可被辅助技术读取；键盘可提交和操作。
- CJK/emoji 在文本和测试数据中保持 UTF-8；不得依赖 Windows 系统语言。

## 必须先红后绿的测试

1. 登录成功只将 access token 保存在内存；登录失败 `40102`、锁定 `40103` 和网络/500 错误均显示三语稳定文案，不泄露原始响应。
2. 未认证路由守卫、已认证跳过登录、登录回跳、Owner/Operator 初始化访问控制。
3. 受保护请求 `40101` 只 refresh 一次并重放一次；refresh 失败、重复 `40101`、login/refresh 请求自身均不会循环重试；并发 `40101` 共享一次 refresh。
4. logout 成功清理会话；logout 失败保留当前页面并显示错误。
5. 三种模板请求体准确；`reused=true`、`42201/INVALID_TEMPLATE`、`42201/RESET_NOT_ALLOWED` 和 `40301` 的 UI 状态正确；reset 必须经二次确认。
6. 三 locale key parity、新增 key 非空；路由与组件测试不得发出 lesson/payment/notification 等禁止范围请求。

## 门禁与交付

```powershell
Set-Location frontend
npm ci
npm run lint
npm run typecheck
npm run test:unit
npm run build
```

执行一次 Playwright 或等价浏览器冒烟：登录、刷新重试、Owner 初始化、Operator 被拒绝、三语切换。若仓库尚未具备可复用浏览器设施，先提交最小可运行方案和证据，不新增无关测试框架。

提交遵循 Lore Commit Protocol；只暂存允许范围的文件。不得勾选 OpenSpec、更新状态/路线图、合并或推送 `main`。交付报告必须包含：基线 SHA、改动文件、红灯→绿灯证据、每条测试与上述场景映射、浏览器证据、依赖锁定、未测试项、回滚方式与风险，等待 Codex 独立验收。

## 验收记录

- 实现提交：`702e680`，2026-07-17；仅前端认证、路由、初始化、三语和测试范围。
- 独立复验：无未解决 P0/P1。Codex 额外修正了 Operator 访问 `/onboarding` 后缺少可见本地化权限提示的问题；路由现重定向至首页并携带受控提示状态。
- 冒烟脚本不再默认访问本机后端或使用内置账号；必须显式提供 disposable-environment 地址、凭据和 `ZEDU_SMOKE_ALLOW_MUTATION=1`。
- 实际浏览器复验：Playwright CLI 打开 `/login`，确认表单可访问名称/label、禁用提交状态和 `zh-CN → ja-JP` 语言切换；完整真实后端浏览器流留待 M2-CODEX-02 的隔离环境执行。
