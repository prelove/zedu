# M2-KIMI-02：人员、课程、报名与安排页面

## 工单元数据

| 项目 | 内容 |
|---|---|
| 状态 | READY |
| Owner | Kimi |
| 基线 | `main` 的 `57d08f5`（M2 后端核心 API 与 M2-KIMI-01 已验收） |
| OpenSpec | `add-m2-core-management` 任务 4.1、4.2；`people-directory`、`course-enrollment-assignment` specs |
| 依据 | PRD v3.1 §3、§5、§10、§13–§15；`docs/acceptance/evidence/M2/contract-freeze.md` |
| 交付目标 | 在既有认证、路由、API adapter 和三语基础上交付 M2 已批准的人/课主数据操作页面 |

## 开工前必读

1. `docs/status/PROJECT_STATUS.md`、`docs/roadmap/MASTER_ROADMAP.md`、`docs/governance/GOVERNANCE.md`
2. `openspec/changes/add-m2-core-management/{proposal.md,design.md,tasks.md}`
3. `openspec/changes/add-m2-core-management/specs/people-directory/spec.md`
4. `openspec/changes/add-m2-core-management/specs/course-enrollment-assignment/spec.md`
5. `docs/acceptance/evidence/M2/contract-freeze.md`
6. `docs/acceptance/evidence/M2/GLM-02BC.md`、`docs/acceptance/evidence/M2/KIMI-01.md`
7. 本工单。

后端 handler、已有前端认证 adapter 与冻结契约共同构成唯一 API 事实源。请求路径没有 `/api` 前缀；所有业务请求必须通过 `authStore.authedRequest`，不得自行复制 refresh 或 Bearer 重试逻辑。

## 允许与禁止范围

允许新增或修改：

- `frontend/src/features/directory/**`、`frontend/src/features/course/**`；
- `frontend/src/api/**`（仅本工单所需的目录/课程 adapter 与类型）；
- `frontend/src/router/**`、`frontend/src/features/auth/HomeView.vue`（只为已批准导航）；
- `frontend/src/i18n/**`、`frontend/tests/**`、必要的共享展示组件；
- 已有 `frontend/vite.config.ts` 的同路径开发代理维护。

禁止修改：`backend/**`、`docs/**`、`openspec/**`、`.github/**`、认证/初始化业务行为、依赖清单（不得新增任何依赖）、OpenSpec 勾选、状态/路线图、共享执行看板。不得实现或注册 lesson、attendance、payment、payment evidence、notification、backup、report、payout、正式结款、学生/老师/家长登录、移动端/PWA；不得创建 DELETE UI 或 API 请求。

Owner 与 Operator 均可访问本工单所有业务页面；保留 M2-KIMI-01 的 Owner-only onboarding 限制，不新增角色判断。

## 页面与路由冻结

| 路由 | 页面 | 数据与允许动作 |
|---|---|---|
| `/students` | 学生列表 | `GET/POST /students`，分页、空态、创建；邮箱可空。 |
| `/students/:id` | 学生详情 | `GET/PATCH /students/{id}`；嵌入家长和报名区。 |
| `/students/:id/parents` | 家长区 | `GET/POST/PATCH /students/{id}/parents...`；仅在当前学生上下文编辑。 |
| `/students/:id/enrollments` | 报名区 | `GET/POST /students/{id}/enrollments`；选择课程、创建无老师报名。 |
| `/enrollments/:id` | 报名与安排详情 | `GET/PATCH /enrollments/{id}`、`GET/POST /enrollments/{id}/assignments`、`POST /assignments/{id}/end`。 |
| `/teachers` | 老师列表 | `GET/POST /teachers`，分页、空态、创建。 |
| `/teachers/:id` | 老师详情 | `GET/PATCH /teachers/{id}`；嵌入能力与可授时间区。 |
| `/courses` | 课程字典 | 领域、方向、等级、能力标签的 `GET/POST/PATCH`；明确启用/停用而非删除。 |

导航只能包含上述已批准入口、首页与 onboarding；不得出现“财务”“通知”“排课”“课消”“报表”或占位入口。

## API 映射与领域不变量

1. 所有列表采用 `{items,page,pageSize,total}`，`page` 从 1 开始、`pageSize` 1–100。加载、空、错误、分页和重试均须可访问且三语。
2. 学生写入只发送已编辑字段；邮箱为空允许保存；非空重复统一显示 `40901` 冲突并阻断提交，绝不提供“仍然新建”或 warning 旁路。`40401` 的家长/学生上下文错误不得展示或猜测其他学生信息。
3. 老师能力的唯一键为 `(teacherId, trackId, levelId)`；重复显示 `40901`。结束能力仅 PATCH `effectiveTo` 或状态，历史记录保持可见，绝不 DELETE。可授时间须校验 weekday 与 `startTime < endTime` 后再请求。
4. 课程字典层级为 domain → track → level；能力标签归属 domain。被引用条目只允许 PATCH `enabled=false`，`42201` 时展示稳定错误并保留表单状态；不得以 DELETE 或本地绕过重挂层级。
5. 创建报名必须选择有效 domain/track，且 `currentLevelId`、`targetLevelId`（若提供）属于该 track；活动学生可创建**无老师**报名。报名 status 使用后端返回的 `ACTIVE`、`PAUSED`、`COMPLETED`、`CANCELLED`，前端不得自造状态或财务字段。
6. **等级历史约束：** 后端保留 enrollment 初始 `current_level_id`，后续等级变化写 `student_level_event`。页面必须将“课程选择（domain/track/target level）”与“记录当前等级变化（currentLevelId）”拆成互斥的两次 PATCH，禁止在同一请求混合 currentLevelId 与实际 domain/track 改动；同等级无变更显示 `42201`。本 M2 API 没有等级事件列表，页面不得伪造“完整等级历史”。
7. 安排页面允许创建第一个 ACTIVE assignment 或用一次 `POST /enrollments/{id}/assignments` 原子替换老师；**不得**先单独结束旧安排再创建新安排。结束使用 `POST /assignments/{id}/end`，重复结束的 `42201` 要稳定呈现。安排动作不得展示“已排课”“已通知”“已收款”等副作用。
8. 所有后端错误统一通过已有 stable key/i18n 映射；不得显示 SQLite、原始响应体、token、密码、cookie 或 requestId 调试内容。

## UX 与可访问性

- 表格或列表提供语义化标题、加载/空/错误状态、键盘可达编辑入口；所有输入有 label、关联 `for/id`，错误使用 `role=alert`。
- 写操作进行中禁用重复提交；成功后刷新相关列表/详情；失败保留用户输入。
- 所有新增可见文本同步 `zh-CN`、`ja-JP`、`en-US`，保持 key parity、UTF-8、CJK/emoji 与 Windows 日文环境兼容。
- 金额/费率若展示老师默认费率，使用现有 `formatJPY`；不实现或编辑任何支付、余额、课时或结款字段。

## 必须先红后绿的测试

1. API adapter：精确真实路径、JSON 信封、Bearer 经既有 wrapper、分页查询参数；无 `/api` 前缀、无禁止 API 请求。
2. 学生：无邮箱创建成功；重复邮箱 `40901` 阻断且无旁路；编辑保存、分页/空态；跨学生家长 `40401` 错误态。
3. 老师：老师创建/编辑；能力三元组冲突 `40901`；结束能力后仍展示历史；无效可授时段在客户端拦截。
4. 课程：domain/track/level/tag 创建与启用/停用；引用冲突 `42201`；三层选择联动不得发送不匹配 ID。
5. 报名：活动学生无老师报名；非活动学生/层级不匹配 `42201`；课程选择 PATCH 与等级变更 PATCH 分离；同等级 `42201`；不出现财务字段或请求。
6. 安排：首次安排、原子替换、结束、重复结束 `42201`；不发送 lesson/payment/notification 请求或导航。
7. 三语 key parity、无禁止路由/菜单/API、原有认证与 onboarding 回归。

## 门禁与交付

```powershell
Set-Location frontend
npm ci
npm run lint
npm run typecheck
npm run test:unit
npm run test:coverage
npm run build
```

对新增或修改的关键用户流使用实际浏览器验证：登录、学生无邮箱创建、重复邮箱阻断、无老师报名、安排替换、三语切换。若涉及写数据，必须使用 disposable 环境并遵守既有 `ZEDU_SMOKE_ALLOW_MUTATION=1` 门禁；不得对默认本机/共享数据执行可变更冒烟。

提交遵循 Lore Commit Protocol，只暂存允许范围。交付报告必须包含基线 SHA、完整改动文件、红灯→绿灯、每项测试映射、浏览器证据、三语/负面范围证据、未测试项、回滚方式和风险。不得更新 docs/OpenSpec/状态/路线图、不得合并或推送 `main`；等待 Codex 独立验收。
