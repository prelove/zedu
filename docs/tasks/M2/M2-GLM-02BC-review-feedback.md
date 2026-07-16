# M2-GLM-02B/02C 独立验收反馈（r1）

| 项目 | 结论 |
|---|---|
| 评审范围 | `9427534..d8eb497`（Stage A `0a574e0`、B `2c0f111`、C `d8eb497`） |
| 评审结论 | CHANGES_REQUESTED；P0：0，P1：6，P2：2 |
| 当前状态 | 不得推送或标记 ACCEPTED；M2-KIMI-02 继续 BLOCKED |
| 已确认优点 | 新领域已形成 handler → service → repository 方向；未改 auth/onboarding；无新增依赖；禁止范围未见新增业务路由。 |

## 一次性修复范围（必须全部完成）

### P1-1：所有列表必须遵守冻结分页外层

`contract-freeze.md` 与本工单均规定任何列表的 `data` 是 `{items,page,pageSize,total}`。目前以下 handler 直接返回数组，现有测试也错误地断言 `data.([]any)`：

- `directory.ListParents`、`ListCapabilities`、`ListAvailability`；
- `course.ListEnrollments`、`ListAssignments`。

修复：为上述 service/repository 增加已验证 `page/pageSize` 参数，handler 使用 `httpserver.ParsePage`，所有响应返回 `httpserver.ListData`。空列表必须是 `items: []`，不得为 `null`。补真实 HTTP 回归，覆盖每条嵌套路由的分页外层、页码边界和不存在父资源的稳定错误。

### P1-2：ENDED/PAUSED 学生不得更新既有报名

`course/service.go` 的 `UpdateEnrollment` 只读取 enrollment，未再次检查其 student 是否仍为 `ACTIVE`。这违反 OpenSpec：enrollment 仅可为现有 ACTIVE student 创建或更新。

修复：在同一事务中用现有 enrollment 的 `student_id` 调用 active 校验，且必须在任何报名写入、审计写入之前完成。学生变为 PAUSED/ENDED 或不存在时返回 `42201`，不改 enrollment、不写成功审计。补 HTTP 回归。

### P1-3：禁止被引用的字典重挂而破坏层级事实

`UpdateTrack` 可改 `domain_id`，`UpdateLevel` 可改 `track_id`，但未检查既有 `teacher_capability` / `student_course_enrollment` 引用。这样会让冗余保存的 domain/track/level 语义失配，违反 course dictionary spec 的引用完整性。

修复：若 track 或 level 已被 capability 或 enrollment 引用，拒绝改变其父级，返回 `42201`，不写审计；未被引用时可保持现有可编辑行为，但必须完整校验新父级。补 capability 引用、enrollment 引用、无引用成功重挂三类真实 HTTP 测试。不得用删除/重建或静默级联更新规避。

### P1-4：保留课程选择与等级变化历史

`UpdateEnrollment` 直接覆盖 `domain_id/track_id/current_level_id/target_level_id`，而未保存历史。这违反 OpenSpec 的 course-selection history 要求和 PRD v3.1 §5.5：等级变化必须记录 `student_level_event`，不直接覆盖 `enrollment.current_level_id`。

修复边界：这是 003 schema 无法表达的已冻结规则，允许新增最小 `004` migration（up/down 与 up/down/up 测试）。

1. 新建最小的 `student_level_event` 表与索引，字段与 PRD §12 的 `student_id`、`enrollment_id`、`from_level_id`、`to_level_id`、`event_type`、`event_date`、`evidence_note`、`operator_id` 相一致；不引入 lesson、学习路径 API 或其他非目标表。
2. PATCH 的 `currentLevelId` 不得直接改 enrollment；当且仅当有效等级实际变化时，在同一事务写 level event（`event_type=MANUAL`、UTC 日期、actor），并保留 enrollment 的原 current level。课程选择（domain/track/target level）变更的审计 detail 必须包含结构化 `before` 与 `after` ID 快照，形成可查询的审计历史。
3. level event、业务更新和 operation_log 必须同一事务；任一步失败完整回滚。

补测试：等级变化不覆盖 enrollment current level 且写一条正确 event；课程选择变更审计包含 before/after；event/audit 故障均无半写；migration up/down/up。

### P1-5：非法输入不得通过数据库 CHECK 变成 50002

- `currentLevelId` / `targetLevelId` 为 `0` 或负值会绕过 hierarchy 校验，随后触发 FK 错误并返回 `50002`；
- `AssignmentWrite.roleType` 未校验，非法值命中 SQLite CHECK 后也返回 `50002`。

修复：在 service 输入校验阶段拒绝非正 level ID 和非枚举 role type（`MAIN`、`SUBSTITUTE`、`ASSISTANT`），返回 HTTP 422 / `42201`；无数据库写、无审计、无 SQLite 文本。补每项 HTTP 回归。

### P1-6：Rollback 失败必须稳定映射 50002

`course.inTx` 与 `directory.inTx` 均丢弃 `Rollback()` 错误；这违反冻结错误契约的“Rollback 相关失败为 50002”。

修复：重写两个 helper，使 BeginTx、业务函数、Commit、Rollback 任一数据库错误均向上返回可映射的数据库错误；在业务函数返回验证/冲突错误后若 rollback 失败，最终必须为 50002。补确定性 fault-injection 回归，证明不会返回 40901/42201，也不泄露底层文本。

## P2（本次一并修复，避免下轮返工）

1. `UpdateAvailability` 只校验请求带来的 effective range，未合并既有另一端日期；先合并既有值后完整验证，禁止把 `effective_from` 更新到既有 `effective_to` 之后。补测试。
2. 所有 PATCH 空 body / 空对象应返回 `42201`，不更新数据、不写成功审计；补至少学生、课程字典、报名、assignment end 的回归。

## 并发验收口径勘误

原工单“并发替换只有一个提交者成功”的要求缺少客户端版本号/If-Match 前置条件，且未被冻结为路由契约；在当前“POST 即原子替换”模型中，串行化的两个有效请求均可成功并形成历史。故验收改为：任一时刻最多一条 ACTIVE assignment；每次成功替换的旧 END、新 CREATE 与审计原子一致；发生 DB/审计故障不留半写。保留并增强该不变量的并发测试，不要求人为制造未冻结的冲突策略。

## 提交与复验

- 允许修改原工单授权范围及本反馈指定的最小 `004` migration、migration 测试；不要修改 `auth/**`、`onboarding/**`、前端、OpenSpec、共享状态/路线图、任务勾选或验收证据。
- 先为每项 P1/P2 补红灯，再最小修复；把红绿证据按编号写入交付报告。
- 运行 Go 1.23.3 下 `go fmt ./...`、`go vet ./...`、`go test ./... -count=1`、`go build ./cmd/zedu-server` 及定向故障/并发/migration 测试。不要运行常规 `-count=20`。
- 不推送；提交所有修复的 Lore commit 后，报告 SHA、完整文件、门禁输出、迁移回滚方式与未测项，等待 Codex 再次独立验收。
