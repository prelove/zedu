## 1. Contract freeze and migration baseline

- [ ] 1.1 [Codex/PM；依赖：proposal+design+spec；输出：M4a contract freeze；测试：OpenSpec strict/人工逐条比对 PRD 8、9.4、10.7、13.8、15 与 legacy 005；证据：`docs/acceptance/evidence/M4a/contract-freeze.md`；门禁：无通知/冲突增强/结款/attendance 范围混入] 冻结 M4a 业务边界、错误语义、字段清单和 Non-Goals。
- [ ] 1.2 [GLM；依赖：1.1；输出：lesson migration 红灯测试与 up/down/up 证据；测试：先写 migration focused 失败测试，再执行 migration up/down/up；证据：`docs/acceptance/evidence/M4a/GLM-04A.md`；门禁：Windows/SQLite 通过，表结构与状态枚举冻结一致] 建立 `lesson` 基础表、索引、状态字段与审计字段迁移。

## 2. Backend lesson domain

- [ ] 2.1 [GLM；依赖：1.2；输出：创建课次红灯测试与实现；测试：先写 create 成功、终态 enrollment 拒绝、非 ACTIVE assignment 拒绝、未授权拒绝的失败测试，再跑 go test focused suite；证据：`docs/acceptance/evidence/M4a/GLM-04B-create.md`；门禁：后端只写 lesson+audit，不触发财务/通知] 实现 lesson create service/repository/handler 与唯一 `lesson_no` 生成。
- [ ] 2.2 [GLM；依赖：2.1；输出：时区与输入校验红灯测试与实现；测试：先写 Asia/Tokyo→UTC、duration 越界、WECHAT 非法链接失败测试，再跑 focused suite；证据：`docs/acceptance/evidence/M4a/GLM-04C-time-validation.md`；门禁：统一由服务层做 timezone 归一化] 实现 lesson 时间换算、duration/meeting_link 校验和稳定错误码映射。
- [ ] 2.3 [GLM；依赖：2.2；输出：更新/取消/详情/列表红灯测试与实现；测试：先写 SCHEDULED 可更新、COMPLETED 不可更新、SCHEDULED 可取消、重复取消拒绝、list/detail 无副作用测试，再跑 focused suite；证据：`docs/acceptance/evidence/M4a/GLM-04D-mutate-query.md`；门禁：仅 `SCHEDULED` 可变更，查询和写入都不产生下游副作用] 实现 lesson update/cancel/detail/list API 与审计日志。

## 3. Frontend lesson pages

- [ ] 3.1 [Kimi；依赖：1.1、2.1 contract 可用；输出：lesson 路由/菜单/三语红灯测试与实现；测试：先写仅 Owner/Operator 可见菜单与无通知/无结款入口红灯测试，再跑 Vitest focused suite；证据：`docs/acceptance/evidence/M4a/KIMI-04A-shell.md`；门禁：三语 key parity，路由仅包含 M4a 批准页面] 建立 lesson 列表页、详情入口与导航壳层。
- [ ] 3.2 [Kimi；依赖：2.2、3.1；输出：lesson 创建/编辑表单红灯测试与实现；测试：先写 duration、meeting_link、timezone、终态 enrollment/无 ACTIVE assignment 错误态红灯测试，再跑 Vitest focused suite；证据：`docs/acceptance/evidence/M4a/KIMI-04B-form.md`；门禁：错误态与后端契约一致，不出现通知/课消/结款字段] 实现 lesson 创建与编辑表单、错误提示和详情展示。
- [ ] 3.3 [Kimi；依赖：2.3、3.1；输出：取消与列表筛选红灯测试与实现；测试：先写 SCHEDULED 可取消、已取消不可再次取消、按学生/老师/状态/时间筛选红灯测试，再跑 Vitest focused suite；证据：`docs/acceptance/evidence/M4a/KIMI-04C-list-cancel.md`；门禁：列表不展示财务副作用字段，只展示 M4a 冻结字段] 实现 lesson 列表筛选、取消动作和详情只读态。

## 4. Focused verification and regression gates

- [ ] 4.1 [Codex/Test；依赖：2.3；输出：后端 focused 验收报告；测试：go fmt、go vet、go test ./...、lesson HTTP focused suite、migration up/down/up；证据：`docs/acceptance/evidence/M4a/backend-verification.md`；门禁：无未解决 P0/P1，时区与状态守卫全部覆盖] 执行 M4a 后端契约验证并补齐最小回归。
- [ ] 4.2 [Codex/Test；依赖：3.3；输出：前端 focused 验收报告；测试：pnpm lint、pnpm typecheck、Vitest focused suite、lesson coverage、pnpm build；证据：`docs/acceptance/evidence/M4a/frontend-verification.md`；门禁：三语 parity、Windows 环境可复现、无未解决 P0/P1] 执行 M4a 前端契约验证并补齐最小回归。
- [ ] 4.3 [Codex/Reviewer；依赖：4.1、4.2；输出：M4a release gate；测试：串联 create→detail→update→cancel→list focused 验证与负向用例复核；证据：`docs/acceptance/evidence/M4a/release-gate.md`；门禁：确认未引入通知、attendance、账务或结款副作用] 形成 M4a 集成验收结论，决定是否进入 ACCEPTED。

## 5. Project tracking and handoff

- [ ] 5.1 [Codex/PM；依赖：4.3；输出：状态面板与追踪矩阵更新；测试：人工核对 requirements-matrix / roadmap / project-status / evidence 索引；证据：`docs/acceptance/evidence/M4a/status-sync.md`；门禁：Requirement→Scenario→Task→Test→Evidence 全链路闭合] 更新项目状态、路线图、追踪矩阵和 M4a 证据索引。
- [ ] 5.2 [Codex/PM；依赖：5.1；输出：M4b 启动前接口冻结说明；测试：人工核对 lesson 契约是否足够支撑通知 outbox-lite；证据：`docs/acceptance/evidence/M4a/m4b-handoff.md`；门禁：无开放式 lesson 字段争议残留] 输出 M4a→M4b handoff，明确通知能力只基于 lesson 契约继续演进。
