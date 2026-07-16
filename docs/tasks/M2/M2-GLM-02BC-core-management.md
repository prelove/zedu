# M2-GLM-02B/02C：人员资料、课程字典、报名与师生安排后端

## 工单元数据

| 项目 | 内容 |
|---|---|
| 状态 | IN_PROGRESS |
| Owner | GLM |
| 工作方式 | 直接在 `main` 的当前 HEAD 上工作；不新建长期分支。每个阶段仅在本阶段门禁真实通过后才可作 Lore commit；不得推送未自测代码。 |
| 基线 | `d49c77cd88652d4e99536cb184aa97c6fd8685d5`（M2-GLM-01、M2-GLM-02A 已验收） |
| OpenSpec | `add-m2-core-management` 任务 2.2、2.3；`people-directory`、`course-enrollment-assignment` specs |
| 事实源优先级 | `docs/2_prd/Zedu-PRD-Final-v3.1.md` → OpenSpec specs/design/tasks → `docs/acceptance/evidence/M2/contract-freeze.md` → 本工单 |
| 验收人 | Codex + 独立 Reviewer |
| 交付目标 | 在不进入财务、课次或排课的前提下，完成 M2 全部人员资料、课程字典、报名与师生安排 HTTP API；为 Kimi 的 M2 页面提供冻结契约。 |

## 开始前必读与硬边界

依次阅读：

1. `AGENTS.md`；
2. `docs/status/PROJECT_STATUS.md`、`docs/roadmap/MASTER_ROADMAP.md`、`docs/governance/GOVERNANCE.md`；
3. `openspec/changes/add-m2-core-management/{proposal.md,design.md,tasks.md}`；
4. `openspec/changes/add-m2-core-management/specs/{people-directory,course-enrollment-assignment}/spec.md`；
5. `docs/acceptance/evidence/M2/contract-freeze.md`；
6. PRD v3.1：§5、§9.1–§9.3、§13.4–§13.7、§15.2、§15.4、§24.5；
7. `docs/standards/testing-standard.md` 与相关编码/安全/i18n 规范。

冻结路由以 `contract-freeze.md` 为准（使用 `PATCH` 与嵌套路由）；PRD 的历史 `PUT` 或旧路径不得自行恢复。请求/响应均使用既有 `{code,data}` / `{code,message,requestId}` 外层，列表必须是 `{items,page,pageSize,total}`。所有 M2 业务路由都要求 Bearer access token，Owner 与 Operator 均可使用；账号管理和 onboarding 的 Owner-only 规则不受本工单影响。

允许新增或修改：

- `backend/internal/app/directory/**`、`backend/internal/app/course/**`；
- 为新领域新增的 `backend/internal/application/**`、`backend/internal/repository/**` 或等价的领域内 `service.go` / `repository.go`，及对应测试；
- 必要的 `backend/internal/platform/httpserver/**` 路由装配、通用且可复用的错误/分页辅助代码及测试；
- `backend/cmd/zedu-server/main.go`，仅用于装配新领域路由；
- 仅在已有 003 schema 无法表达冻结规则时，最小新增 `004` up/down migration 和 migration 测试；先在交付报告中说明不可替代性。

禁止修改：

- `frontend/**`、`.github/**`、`openspec/**`、共享契约、状态、路线图、任务勾选和验收证据；
- `auth/**`、`onboarding/**` 的既有行为或将其大规模重构为分层架构；
- 新依赖、认证角色、生产可调用 test mode、全局故障 hook；
- lesson、attendance、payment、付款凭证、notification、backup、report、payout、正式结款、学生/老师/家长登录；
- 任何仅为“方便前端”而未冻结的路由、数据字段、软删除/删除 API 或导入导出 API。

首个 Owner 的安全引导属于部署/基础设施待规划事项，不在本工单添加默认账号、公开 bootstrap API 或绕过 Owner 校验。

## 架构决定（本任务必须遵守）

OpenSpec design 已规定 `HTTP → application → repository/database`。从本任务新增的领域代码起严格采用该方向：

```text
http handler (decode/context/encode)
  → application service (授权、校验、状态机、事务编排、审计)
    → repository (参数化 SQL、scan、查询结果)
```

- Handler 不得直接写业务 SQL、手写事务或业务状态机；repository 不依赖 `net/http`，不生成 HTTP 响应，也不决定权限。
- application service 是唯一多表事务边界；repository 接受 `*sql.DB` 或 `*sql.Tx` 的最小执行接口，使同一服务事务内的所有写入和审计可回滚。
- 不回溯重构已验收 `auth` 和 `onboarding`。若为共用审计/分页确实需要提取小型无业务依赖 helper，保持兼容并增加回归测试；否则在新领域内部实现。
- 所有 SQL 参数化；DB/BeginTx/Commit/Scan/rows.Err 失败一律 HTTP 500 / `50002`，禁止泄露 SQLite 文本。非 DB 的内部错误为 500 / `50001`。
- 成功业务写与 `operation_log` 必须同一事务。审计含 actor、action、target_type、target_id、request_id、可解析的非敏感 `detail_json`；失败、冲突、未授权、验证拒绝不得留下成功审计。

## 实施阶段与完成条件

本任务是一个长程工单，但必须按以下顺序开发、分阶段提交并在报告中分别列出证据；不等前一阶段被 Codex 接收才开始后续阶段，但任何阶段失败不得掩盖或跳过。

### 阶段 A：领域骨架与人员资料（OpenSpec 2.2）

实现并装配：

- `/students`、`/students/{id}`：GET/POST/PATCH；支持冻结分页与合理的只读搜索/状态筛选。
- `/students/{id}/parents`、`/students/{id}/parents/{parentId}`：GET/POST/PATCH；parent 必须被当前路径 student 所拥有，否则 `40401`。
- `/teachers`、`/teachers/{id}`：GET/POST/PATCH。
- `/teachers/{id}/capabilities`、`/teachers/{id}/availability`：GET/POST/PATCH。

业务规则：

1. 学生姓名必填；`email` 可空，非空时全局唯一（含软删除），创建、更新和并发冲突均 `40901`；不得有 bypass。
2. 学生/老师状态仅使用 migration 003 的合法值；对 ENDED 学生的新报名由阶段 C 拒绝。不得在本阶段自行增加 delete/restore 语义。
3. parent 跨 student 读取或更新返回 `40401`，不得泄露记录。
4. capability 必须引用存在且层级匹配的 domain/track/level；`(teacher_id, track_id, level_id)` 冲突为 `40901`。结束能力只写 `effective_to` / 合法状态，保留历史，不 delete。
5. availability 的 `weekday`、时间格式、起止顺序、有效期范围必须校验；不得写入无效区间。
6. 任何创建、编辑、状态变更、能力结束、availability 变更都在业务写和审计的同一事务。

### 阶段 B：课程字典（OpenSpec 2.3 的前置）

实现并装配：`/course-domains`、`/tracks`、`/levels`、`/capability-tags` 的 GET/POST/PATCH。

规则：

1. 保持四层可配置模型：domain → track → level，skill tag 归属 domain；不得以显示名（日语、N1、数学）决定业务逻辑。
2. code 唯一性由现有 schema 最终裁决：domain 全局唯一，track/level/tag 在父级唯一；唯一冲突 `40901`。
3. 创建或编辑 track/level/capability 时验证完整父子层级，错误状态/不匹配关系 `42201`，不存在资源 `40401`。
4. 本冻结 API 不含 DELETE。若 PATCH 有“禁用”语义，引用它的 capability/enrollment 仍完整保留；不得破坏关系或物理删除。
5. 所有成功字典写和审计同事务。

### 阶段 C：报名与师生安排（OpenSpec 2.3）

实现并装配：

- `/students/{id}/enrollments`：GET/POST；`/enrollments/{id}`：GET/PATCH；
- `/enrollments/{id}/assignments`：GET/POST；`/assignments/{id}/end`：POST。

规则：

1. enrollment 只能为存在、`ACTIVE` 学生建立；ENDED 学生或不存在学生一律 `42201`。需验证 domain/track/current/target level 的层级一致性。
2. 允许先创建无老师 assignment 的 enrollment；M2 不创建 lesson，也不写任何财务、通知、邮件或 payout 数据。
3. `enrollment.status` 仅允许 `ACTIVE → PAUSED → ACTIVE`、`ACTIVE → COMPLETED`、`ACTIVE/PAUSED → CANCELLED`；终态不得恢复。拒绝状态迁移为 `42201`。
4. assignment 只能连接 active enrollment 与 active teacher；一个 enrollment 最多一条 ACTIVE assignment。创建新 ACTIVE assignment 若已有 ACTIVE，必须以单事务结束旧记录、创建新记录并保留 reason/history（“替换”）；不得依赖应用层先查后写来规避唯一索引。
5. `/assignments/{id}/end` 只能结束该 enrollment 的当前合法 assignment；ENDED 不可恢复。并发替换必须始终保持至多一条 ACTIVE assignment；每个成功替换的旧记录结束、新记录创建与审计必须原子一致。当前冻结契约无版本号/If-Match 前置条件，不强行规定并发有效请求必须只有一个成功；不得泄露 DB 文本或留下半结束记录。
6. enrollment/assignment 每个成功写操作均含同事务审计；替换的旧结束、新建与审计要么全部提交，要么全部回滚。

## 必须先红后绿的测试矩阵

使用真实 SQLite 临时库、迁移和真实 HTTP handler。先保存每阶段至少一条红灯命令/输出，再最小实现。至少覆盖：

| 领域 | 必测行为 |
|---|---|
| 认证/RBAC | 未认证 `40101`；Owner 与 Operator 可访问本任务所有路由；请求/响应外层与 requestId 正确。 |
| 学生 | 无邮箱成功；创建/更新重复邮箱 `40901`；并发重复邮箱恰一成功；冲突后原记录与审计不变。 |
| 家长 | 多 parent；跨 student GET/PATCH `40401`；失败无审计。 |
| 老师 | capability 三元唯一 `40901`；结束 capability 留历史；availability 无效时间/范围拒绝。 |
| 审计/事务 | 每类成功写一条非敏感、可解析、request_id 非空的审计；审计或任一业务 SQL 失败时业务和审计均回滚。 |
| 字典 | 各层 code 冲突 `40901`；错误父子关系拒绝；被 capability/enrollment 引用的字典不可破坏性删除/清空（接口不存在也要负面路由检查）。 |
| 报名 | ACTIVE 学生无老师报名成功；不存在/ENDED 学生 `42201`；错误课程层级拒绝；终态 status 不可恢复。 |
| 安排 | 创建 active assignment；替换时旧记录 ENDED、新记录 ACTIVE、余额/课时不变；在结束旧记录后注入失败时全部回滚；并发替换不出现两条 ACTIVE。 |
| 负面范围 | 每一阶段均断言没有 lesson/attendance/payment/evidence/notification/payout/email 写入或新增路由。 |
| i18n/Windows | CJK、日文、emoji 往返；所有时间为 UTC，不依赖 Windows 系统语言。 |

## 门禁、提交和交付

每个阶段完成后至少运行：

```powershell
Set-Location backend
$env:GOTOOLCHAIN='local'
go version                 # 必须 go1.23.3
go fmt ./...
go vet ./...
go test ./... -count=1
go build ./cmd/zedu-server
```

再运行该阶段的 HTTP、并发、事务故障注入定向测试。`go test ./... -count=20` 不属于日常门禁；只在 M2 候选验收、迁移/并发基础设施变更或发现不确定性时执行。涉及共享状态/事务/并发，提交后由 CI 在 Linux 运行一次 `go test ./... -race -count=1`；Windows 不伪造 race 结果。

每个阶段只暂存允许范围内的文件，遵守 Lore Commit Protocol。最终交付报告发送给 Codex，必须包含：

1. 基线 SHA、每阶段 SHA、完整改动文件和新增路由；
2. 红灯→绿灯证据及上表逐项映射；
3. migration 是否新增、up/down/up 证据与安全回滚说明；
4. 单事务/审计/并发证明，及 DB 错误到 `50002` 的证明；
5. 全部门禁真实输出、未测试项、风险与不在范围的项目；
6. 不修改 OpenSpec 勾选、共享状态/路线图/证据，也不自行宣布 ACCEPTED 或推送未验收结果。

## 验收边界

本工单完成后，M2-GLM-02B/02C 才进入 `IN_REVIEW`，而不是 M2 整体 ACCEPTED。Kimi 的 M2-KIMI-02 仅可在 Codex 验收 API 后解除阻塞；M2-CODEX-02 仍需真实 HTTP、浏览器、CI 和全链路验收。
