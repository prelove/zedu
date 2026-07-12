## 0. 执行规则与冻结点

- [x] 0.1 [Codex/PM，依赖：M1 ACCEPTED；输出：`docs/tasks/M2/M2_EXECUTION_BOARD.md`；测试：OpenSpec strict；证据：`docs/acceptance/evidence/M2/contract-freeze.md`；门禁：READY] 冻结本 change 的路由、JSON 外层、错误码、角色矩阵、分页字段和精确依赖版本；确认 `student.email` 非空全局唯一（含软删除）、重复为`40901`，以及`teacher_capability(teacher_id,track_id,level_id)`唯一。未经该冻结，GLM/Kimi不得编码或新增依赖。
- [x] 0.2 [Codex/PM，依赖：0.1；输出：范围负面清单；测试：路由/迁移/前端导航审查；证据：M2 contract-freeze；门禁：READY] 明确 M2 禁止 lesson、attendance、payment、payment evidence、notification、backup、report、payout、学生/老师/家长登录及其 API、路由、菜单；将禁止项加入后续验收清单。

## 1. 后端通用契约与安全基础（GLM）

- [ ] 1.1 [GLM，依赖：0.1；允许：`backend/go.mod`、`backend/go.sum`、`backend/internal/platform/auth/**`、`backend/internal/platform/httpserver/**`及对应测试；输出：经审查锁定的 bcrypt/JWT 实现、统一 JSON 成功/错误编码、request ID 响应；测试：先写 40101/40102/40103/40301/40901/42201 红灯测试，再实现；证据：`docs/acceptance/evidence/M2/GLM-01.md`；门禁：Go fmt/vet/test/build] 以最小依赖实现密码哈希、JWT 验签、认证/角色 middleware 与稳定错误响应；日志和响应不得泄露 password、hash、access/refresh token 或 Authorization header。
- [ ] 1.2 [GLM，依赖：1.1；允许：`backend/migrations/003_*`、`backend/internal/platform/database/**`及对应测试；输出：用户、refresh_session、system_settings、operation_log 与 M2 主数据迁移；测试：先写 up/down/up、外键、UTF-8、唯一约束、NULL 邮箱多行和非空邮箱并发唯一红灯测试；证据：M2/GLM-01；门禁：20 次数据库稳定性回归] 实现所有 M2 表和索引，强制 student email 全局唯一（软删除不释放邮箱）、teacher capability 三元唯一，并保持 SQL 参数化和迁移可回滚。
- [ ] 1.3 [GLM，依赖：1.1、1.2；允许：`backend/internal/app/auth/**`、`backend/internal/platform/httpserver/**`及对应测试；输出：登录/刷新/登出/me、Owner 管理 Operator 账号 API；测试：先写登录成功/通用失败/5次锁定/refresh 轮换旧 token 拒绝/登出和禁用撤销/RBAC 红灯测试；证据：M2/GLM-01；门禁：HTTP 集成测试与脱敏日志测试] 实现 60 分钟 access token、14 天随机 refresh token 哈希存储与轮换；refresh 仅通过 Secure HttpOnly SameSite=Strict cookie，不得写入 JSON。

## 2. 后端初始化与核心资料（GLM）

- [ ] 2.1 [GLM，依赖：1.1、1.2、1.3；允许：`backend/internal/app/onboarding/**`、`backend/internal/platform/httpserver/**`及对应测试；输出：Owner 初始化/受限重置 API；测试：先写未初始化成功、重复幂等、Operator 40301、业务数据存在 42201、事务故障回滚红灯测试；证据：`docs/acceptance/evidence/M2/GLM-02.md`；门禁：迁移+HTTP 集成测试] 实现显式模板初始化和初始化标记；启动过程不得写业务模板，重置必须在无 student/teacher/enrollment/assignment 时才允许。
- [ ] 2.2 [GLM，依赖：1.1、1.2、1.3；允许：`backend/internal/app/directory/**`、`backend/internal/platform/httpserver/**`及对应测试；输出：student/parent/teacher/capability/availability API；测试：先写学生无邮箱成功、创建/更新/并发重复邮箱40901、跨学生 parent 40401、能力重复40901、结束能力保留历史、失败不写审计红灯测试；证据：M2/GLM-02；门禁：事务、授权、审计和并发测试] 实现人员资料服务；成功写操作与 operation_log 必须同事务，数据库唯一错误必须映射40901而不是原始 SQLite 文本。
- [ ] 2.3 [GLM，依赖：1.1、1.2、1.3；允许：`backend/internal/app/course/**`、`backend/internal/platform/httpserver/**`及对应测试；输出：课程字典、enrollment、assignment API；测试：先写层级 code 冲突40901、被引用字典删除42201、无老师报名成功、结束学生报名42201、替换 assignment 的原子回滚以及无 lesson/payment/notification 副作用红灯测试；证据：M2/GLM-02；门禁：服务+HTTP 集成测试] 实现`student → enrollment → assignment`主数据链路；每个 enrollment 至多一个 ACTIVE assignment，替换时旧记录结束和新记录激活必须原子完成。

## 3. 前端基础、认证和初始化（Kimi）

- [ ] 3.1 [Kimi，依赖：0.1、1.3 API契约已冻结；允许：`frontend/src/router/**`、`frontend/src/api/**`、`frontend/src/stores/**`、`frontend/src/i18n/**`、`frontend/tests/**`及必要精确依赖；输出：API adapter、认证状态、受保护路由和登录页面；测试：先写未登录重定向、登录成功、401刷新一次后重试、刷新失败清空状态、Owner-only 导航隐藏红灯测试；证据：`docs/acceptance/evidence/M2/KIMI-01.md`；门禁：lint/typecheck/unit/coverage/build] 实现不读取或持久化 refresh cookie 的认证前端；所有新中文文案必须同步 `zh-CN`/`ja-JP`/`en-US`。
- [ ] 3.2 [Kimi，依赖：3.1、2.1 API可用；允许：`frontend/src/features/onboarding/**`、`frontend/src/i18n/**`、`frontend/tests/**`；输出：首次初始化与 Owner 受限重置界面；测试：先写 Owner 初始化、Operator 无入口且后端403提示、重复初始化展示既有结果、业务数据存在时42201错误态红灯测试；证据：M2/KIMI-01；门禁：三语 key parity、unit、build] 实现显式模板选择与状态展示，不提供静默初始化或危险确认绕过。

## 4. 前端人员与课程主数据（Kimi）

- [ ] 4.1 [Kimi，依赖：3.1、2.2 API可用；允许：`frontend/src/features/directory/**`、`frontend/src/i18n/**`、`frontend/tests/**`；输出：学生/家长/老师/能力/可授时间列表与详情；测试：先写学生无邮箱保存、40901重复邮箱阻断、没有“仍然新建”按钮、跨资源40401、能力结束历史展示红灯测试；证据：`docs/acceptance/evidence/M2/KIMI-02.md`；门禁：三语 key parity、unit、coverage、build] 实现人员资料页面和可访问的加载/空/错误状态；重复邮箱必须阻断提交而不是 warning。
- [ ] 4.2 [Kimi，依赖：3.1、2.3 API可用；允许：`frontend/src/features/course/**`、`frontend/src/i18n/**`、`frontend/tests/**`；输出：课程字典、报名和师生安排页面；测试：先写无老师报名、替换老师、42201/40901错误态、无财务/排课/通知菜单与路由红灯测试；证据：M2/KIMI-02；门禁：三语 key parity、unit、coverage、build] 实现主数据操作与分页，路由只包含 M2 已批准页面。

## 5. 集成验收与发布控制（Codex/PM）

- [ ] 5.1 [Codex，依赖：1.1–2.3、3.1–4.2；允许：共享契约、CI、测试、证据、状态与路线图文件；输出：端到端契约和浏览器验收；测试：Playwright 覆盖登录、RBAC、初始化、邮箱40901、报名/安排原子性与负面范围检查；证据：`docs/acceptance/evidence/M2/verification-report.md`；门禁：Windows/Ubuntu CI全绿、Linux race、浏览器测试通过] 在真实 HTTP 前后端联调后验收，不接受仅 mock 通过的交付。
- [ ] 5.2 [Codex + 独立 Reviewer，依赖：5.1；输出：M2 审查结论、追踪矩阵、状态、风险、路线图更新；测试：OpenSpec strict、UTF-8、追溯、依赖审计、全量回归；证据：M2 verification-report；门禁：无未解决P0/P1、Release Gate签署] 仅在所有场景和 Non-Goals 均有新鲜证据后，才勾选 OpenSpec tasks 并将 M2 标记 ACCEPTED。
