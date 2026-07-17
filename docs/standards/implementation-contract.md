# Zedu 统一实现契约

版本：2026-07-17（M2 基线）
适用：数据库、Go 后端、Vue 前端及 AI 编码工具。

## 1. 用途、事实来源与最小阅读集

本文件固化**已独立验收**的共用实现模式，减少重复全仓扫描。它不创造业务需求：事实优先级始终是 Decision Record → PRD → 已批准 OpenSpec Requirement/Scenario → 当前工单 → 本契约 → 代码。

每次实现的默认阅读集只有：

1. `docs/status/PROJECT_STATUS.md`、当前工单、`docs/acceptance/evidence/M2/contract-freeze.md`；
2. 本文件；
3. 当前领域目录、同类测试，以及该工单明确引用的 migration/handler/adapter。

仅在下列情况扩大为定向扫描：API 或 schema 与工单不一致；需要新增依赖、角色、路由、迁移；跨领域事务；定向测试无法解释失败。不得为“确认风格”重复扫描整个仓库。

## 2. 仓库与职责边界

| 区域 | 职责 | 不得承担 |
|---|---|---|
| `backend/cmd/zedu-server` | 组装依赖、路由和运行配置 | 业务规则、SQL、测试旁路 |
| `backend/internal/platform` | 数据库打开/迁移、认证、HTTP 信封、日志 | 领域业务规则 |
| `backend/internal/repository` | `database/sql` 的 DB/Tx/Executor 接口与数据库哨兵错误 | HTTP、业务状态机 |
| `backend/internal/app/<domain>` | handler → service → repository；领域 DTO、校验、事务、审计 | 跨领域临时复制、直接修改无关领域 |
| `backend/migrations` | 递增、可回滚 SQL schema | 业务模板或运行时数据 |
| `frontend/src/api` | typed API adapter、统一信封 | 组件内散落 fetch |
| `frontend/src/stores` | 跨页面响应式状态 | 领域页面细节 |
| `frontend/src/router` | 已批准路由和守卫 | 未批准功能入口 |
| `frontend/src/features/<domain>` | 页面、表单、领域 UI | 认证重试、底层 HTTP |
| `frontend/src/i18n` | 三语资源与格式化 | 业务英文常量重命名 |

当前后端数据访问是 `database/sql`，不是 GORM。前端是 Vue 3、Vite、TypeScript strict、vue-i18n、`vue-router@5.1.0`；M2 禁止 Pinia、UI 框架和未批准的新依赖。

## 3. HTTP、错误与认证契约

### 3.1 JSON 外层

成功：`{ "code": 0, "data": <payload> }`。失败：`{ "code": <stable-code>, "message": <stable-key>, "requestId": <string> }`。列表 data 固定为 `{items,page,pageSize,total}`，page 从 1 开始，pageSize 为 1–100。所有路径均不带 `/api` 前缀。

| code | HTTP | 含义 |
|---:|---:|---|
| 40101 | 401 | 未认证、token 或 session 无效 |
| 40102 | 401 | 登录凭据失败 |
| 40103 | 401 | 账号锁定 |
| 40301 | 403 | 权限不足 |
| 40401 | 404 | 不存在（含不属于父资源） |
| 40901 | 409 | 唯一约束/数据冲突 |
| 42201 | 422 | 状态、层级或输入组合不允许 |
| 50001 | 500 | 非数据库内部错误 |
| 50002 | 500 | 数据库、BeginTx、Commit、Rollback、查询或写入失败 |

不得自造错误码、将数据库错误伪装成 40101/42201，或向 UI/日志返回 SQLite 文本。

### 3.2 认证与前端会话

- access token 为 60 分钟 JWT HS256；前端仅存内存，绝不写 localStorage、sessionStorage、cookie、日志或 URL。
- refresh token 为 14 天、仅 `HttpOnly; Secure; SameSite=Strict` cookie；浏览器处理，JavaScript 不读取、不写入。
- 认证 API 位于 `/auth/login`、`/auth/refresh`、`/auth/logout`、`/auth/me`；前端统一使用 `authStore`/`authApi`。
- 任一受保护前端请求通过 `authStore.authedRequest`：40101 时最多 refresh 一次并重放一次；并发请求共享 refresh；login/refresh 自身不得递归重试；刷新/重放仍失败则清内存会话并回登录。
- Owner 包含 Operator 权限；`/onboarding` 仅 Owner。所有业务路由要求 Bearer token。

## 4. 数据库与迁移契约

- 使用 `modernc.org/sqlite`，禁止 CGO；运行时 `database.Open` 固定 `foreign_keys=ON`、`journal_mode=WAL`、`busy_timeout>=5000ms`、`MaxOpenConns=1`。
- migration 文件按编号递增；一对 `NNN_name.up.sql`/`down.sql`；必须参数化运行时 SQL，并验证 up/down/up、外键、UTF-8、约束。已被验收/推送的 migration 不得改写，需新增 migration。
- 所有金额使用整数最小货币单位，禁止 float；时间存 UTC，UI 以 Asia/Tokyo 展示。等级的非货币课时字段按领域 DTO，不得误作金额。
- DB 约束是并发最终防线，应用层校验不得替代 UNIQUE、FK、CHECK 或部分唯一索引。
- `repository.Executor` 可由 `*sql.DB` 或 `*sql.Tx` 实现；多表写入只能由 service 创建一个 transaction，所有 repository 写和 audit INSERT 使用同一个 Tx。
- 任何 BeginTx/查询/写入/Commit/Rollback 失败映射 `repository.ErrDatabase` → 50002；Rollback 失败不得吞掉。

### 4.1 M2 已冻结数据不变量

- `student.email` 可空；非空全局唯一，软删除不释放；冲突为 40901，前端没有继续创建旁路。
- `teacher_capability(teacher_id,track_id,level_id)` 唯一；结束能力写 `effective_to`，不删除历史。
- `student → enrollment → assignment`；一个 enrollment 最多一个 ACTIVE assignment；替换老师在一个事务中结束旧记录并建立新记录。
- 课程层级为 domain → track → level；引用中的字典只可停用，不能破坏性重挂/删除。
- 等级变化写 `student_level_event`，不覆写 enrollment 初始 `current_level_id`；课程选择与 `currentLevelId` 变化必须分开 PATCH；等级事件引用也阻止 level/track 重挂。

## 5. 后端实现、审计和安全

- handler：解析 HTTP/DTO、取认证上下文、调用 service、将领域错误映射稳定 HTTP；不写领域 SQL、不持有多表事务。
- service：校验状态机/关系、开启事务、调用 repository、写审计并提交；失败路径完整回滚。
- repository：只做参数化 SQL 与扫描；不依赖 `net/http`，不决定 HTTP 响应。
- 每个成功业务写与 `operation_log` 同事务，至少有 actor、action、target_type、target_id、request_id、无敏感 detail JSON。失败、冲突、未授权不能留下成功审计。
- 日志、audit、响应不得出现 password、password_hash、Authorization、access/refresh token 或其 hash、完整邮箱、凭证内容。
- 资源访问同时验证认证和对象授权，跨学生 parent 等 IDOR 返回 40401，不泄露资源存在性。

## 6. 前端实现与可访问性

- API adapter 复用 `httpRequest`；页面不得直接 fetch。稳定错误经 `ApiError`/`NetworkError` 和 `errorToI18nKey` 转成本地化 key；不显示原始 body、requestId 或调试异常。
- 新用户文本在 `zh-CN`、`ja-JP`、`en-US` 三个 locale 同步、严格同构且非空；编码 UTF-8/LF，CJK/emoji 与 Windows 日文环境须有测试。
- 页面/表单必须有加载、空、成功和错误状态；输入有可见 label 和 `for/id`；错误使用 `role=alert`/`aria-live`；写入期间禁重复提交，失败保留用户输入。
- PATCH 只发送实际编辑字段，禁止空 body；写成功后刷新相关详情/列表。
- 页面、路由、导航只由当前工单批准。M2 禁止 lesson、attendance、payment、payment evidence、notification、backup、report、payout、正式结款及学生/老师/家长登录；禁止 DELETE UI/API。

## 7. 测试、浏览器与交付

| 层级 | 必须证明 |
|---|---|
| Unit | 状态机、校验、i18n key parity、错误映射、权限、空/加载/错误状态、禁止范围 |
| 后端 integration | 临时 SQLite、migration、事务/审计原子性、并发约束、故障注入、IDOR |
| 前端 integration | adapter 精确路径/信封/401 重试、路由守卫、表单 payload、稳定错误显示 |
| 浏览器 | 关键用户流、键盘/label、三语；可变更流仅 disposable 数据库 |

小切片门禁：后端 `go fmt ./...`、`go vet ./...`、`go test ./... -count=1`、`go build ./cmd/zedu-server`；前端 `npm ci`、lint、typecheck、unit、build。当前工单要求 coverage 时才跑 coverage。20 次稳定性扫描只在里程碑、迁移/并发基础设施变更或非确定性失败后执行。Linux CI 负责 `go test ./... -race -count=1`。

浏览器可变更测试必须显式指定 disposable 环境和 `ZEDU_SMOKE_ALLOW_MUTATION=1`，不得默认访问本机或共享数据。交付只暂存允许文件，遵循 Lore Commit Protocol，报告基线 SHA、改动、红绿、门禁、未测项、风险和回滚；执行者不更新共享状态、路线图、证据或 OpenSpec 勾选。

## 8. 代码规模、命名与注释约束

这些是**新增或实质修改文件**的默认上限，用于保持 AI 和人工均可快速阅读、审查和定位。它们是拆分触发器，不是要求在功能任务中重构全部既有代码的理由；既有超限文件须在工单中登记，只有修改触及同一职责时才做最小拆分。

| 对象 | 推荐上限 | 超限时的动作 |
|---|---:|---|
| Go handler / service / repository 单文件 | 500 非空代码行 | 按资源、查询/写入、状态机或审计职责拆分；不跨领域搬迁。 |
| Go 单个函数 | 60 非空代码行 | 提取命名明确的私有步骤；事务边界仍由 service 保持可见。 |
| TypeScript adapter / store / utility 单文件 | 250 非空代码行 | 按 API 资源、状态职责或纯函数拆分；不得复制 HTTP/认证逻辑。 |
| Vue 页面组件 | 350 非空代码行（script+template+style） | 提取领域子组件、表单段或展示组件；页面保留路由编排和用户流。 |
| Vue 通用组件 | 200 非空代码行 | 拆出无状态展示组件或 composable；不得为一处使用引入泛化框架。 |
| 单个测试文件 | 600 非空代码行 | 按资源/场景拆分；共享 fixture 放 `helpers`，不可把断言隐藏进不透明 helper。 |
| 单个测试用例 | 80 非空代码行 | 以 Given/When/Then 抽取 setup；一个测试只证明一个关键行为。 |

以下例外必须在交付报告中说明：迁移 SQL、locale 资源、可读性更好的表驱动测试、不可拆的事务状态机、以及有明确生成来源的文件。禁止为了满足行数删除断言、注释或验证。

### 8.1 命名

- Go：package 用简短小写领域名；导出标识使用 PascalCase 并有以名称开头的注释；未导出标识 camelCase。缩写统一 `ID`、`URL`、`API`、`HTTP`、`JWT`、`DB`、`JSON`；错误使用 `ErrXxx`，布尔值用 `is/has/can/should`，避免 `data`、`info`、`util`、`manager` 等无职责名。
- Go DTO 名称表达边界：`StudentWrite`、`LoginResponse`；repository 方法表达意图：`Get...`、`List...`、`Create...`、`Update...`、`Count...`、`Insert...`。不得让 handler 直接拼 SQL。
- TypeScript：变量/函数 camelCase，类型/interface/class PascalCase，常量和稳定错误键使用约定的 UPPER_SNAKE_CASE。API 文件使用资源小写或 kebab-case（如 `error-mapping.ts`），Vue 组件 PascalCase（如 `StudentListView.vue`），测试为同名 `*.test.ts`。
- 路由名称用稳定英文领域名，URL 用 kebab-case；i18n key 按 `domain.intent` 分组，不把中文、日文或展示句子作为 key。
- 禁止无语义缩写（`tmp`、`obj`、`res2`、`handleData`）；循环索引、短生命周期的错误变量和标准 `ctx/tx/db` 可例外。

### 8.2 注释与结构

- 注释解释**为什么、约束、边界或失败原因**，不逐行复述代码。业务不变量、并发条件、审计/安全决策、不可直观的时区和兼容性必须有注释并引用 PRD/OpenSpec/契约来源。
- 公开 Go API、跨领域 adapter、通用 composable 和复杂状态机必须有简短文档注释；私有的直观代码不为凑注释率添加噪声。
- TODO 必须包含可追踪任务/原因；禁止以 TODO、注释掉的旧代码、空 catch 或吞错代替实现。暂不支持的能力应由工单 Non-Goal 和负面测试表达。
- 每个文件只承担一个可命名职责；import 按语言工具格式化，不手工维持无意义分组。重复三次以上或已出现跨页面状态时，再提取共享组件/composable；不要过早抽象。
- 新增逻辑应让失败路径与成功路径同样可读：数据库/网络错误不得被 `_`、空 catch、默认成功值或 UI 静默吞掉。

### 8.3 审查触发器

新增或实质修改代码出现以下任一情况，提交前必须在交付报告说明拆分决定：超过上表上限；函数嵌套超过三层；相同请求/错误映射/表单校验复制两处；引入跨领域 import；或为了通过测试增加 test-only 生产分支。审查者可要求最小拆分，但不得借此扩大为无关重构。

## 9. 维护与变更触发

仅在独立验收后由 Codex/PM 更新本契约。以下情况必须先更新契约或创建 OpenSpec change：新依赖、迁移、角色、错误码、通用 HTTP/认证模式、跨领域事务、可变更浏览器策略或新的产品领域。领域字段、页面细节和一次性验收场景只写入当前工单，避免本文件膨胀为 PRD。
