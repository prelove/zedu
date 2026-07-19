## 1. 契约冻结与迁移准备

- [x] 1.1 [Codex/PM；依赖：M2 当前集成候选；输出：`docs/tasks/M3/M3_EXECUTION_BOARD.md` 与冻结路由/角色/错误映射；测试：OpenSpec strict；证据：`docs/acceptance/evidence/M3/contract-freeze.md`；门禁：无未决财务语义] 将 proposal/spec/design 的 API、金额、幂等、文件边界和 Non-Goals 转为实施工单；不得开始 M3 代码前新增未批准字段、错误码或依赖。
- [ ] 1.2 [GLM；依赖：1.1；输出：005 migration 与迁移测试；测试：先失败的 up/down/up、默认设置/支付方式、约束与索引测试；证据：`docs/acceptance/evidence/M3/GLM-03A.md`；门禁：Go fmt/vet/test/build] 只补足 M3 所需 system settings、payment/attachment/ledger 的约束、索引和幂等默认数据；禁止改变 M2 已验收语义和 teacher payout。

## 2. 财务配置与充值账务后端

- [ ] 2.1 [GLM；依赖：1.2；输出：`backend/internal/app/finance/**` 的配置层；测试：先失败的 Owner/Operator RBAC、base currency lock、payment method defaults/禁用/历史显示、审计与 DB failure 测试；证据：`docs/acceptance/evidence/M3/GLM-03A.md`；门禁：所有写操作审计且无敏感 detail] 实现本位币读取/受限修改和支付方式字典 API；不得为 CNY/USD 以外的 base currency 猜测精度模型。
- [ ] 2.2 [GLM；依赖：2.1；输出：充值 create/list/detail/void 和学生 ledger API；测试：先失败的精确金额、paymentNo 重放/冲突、ACTIVE/归属校验、事务回滚、并发、作废冲正、错误码与脱敏日志；证据：`docs/acceptance/evidence/M3/GLM-03B.md`；门禁：无 float、确认/作废多表写入单 tx] 实现 CONFIRMED 充值与不可变 VOID 冲正；严禁退款、人工调整、课次、老师流水或结款入口。

## 3. 付款凭证后端

- [ ] 3.1 [GLM；依赖：2.2；输出：受控上传存储、attachment list/download API；测试：先失败的类型/大小/三份上限、未认证/错误归属/路径穿越、临时文件与 publish failure 补偿、并发上限和审计脱敏；证据：`docs/acceptance/evidence/M3/GLM-03C.md`；门禁：无匿名直链、失败后无可访问孤儿] 实现临时文件→metadata transaction→原子发布的附件路径；仅使用标准库与现有依赖。
- [ ] 3.2 [Codex/Reviewer；依赖：3.1；输出：后端独立审查结论；测试：迁移、权限、财务原子性、恶意文件和真实 HTTP focused suite；证据：`docs/acceptance/evidence/M3/backend-review.md`；门禁：无未解决 P0/P1] 一次性审查 M3 后端所有已交付切片；仅列与冻结规格直接相关的问题，不扩展到 M4+。

## 4. M3 前端主路径

- [ ] 4.1 [Kimi；依赖：2.1、M3 contract freeze；输出：finance/config adapter、三语 key、Owner 本位币/支付方式最小界面；测试：先失败的 API body/角色/错误映射/key parity；证据：`docs/acceptance/evidence/M3/KIMI-03A.md`；门禁：不新增状态库或 UI 依赖] 实现配置读取和 Owner-only 写入；不创建报表或结款导航。
- [ ] 4.2 [Kimi；依赖：2.2、3.1；输出：学生账务 Tab、充值记录/表单/作废、凭证上传与鉴权下载入口；测试：先失败的金额字符串、paymentNo 重试、课程归属、三附件限制、作废确认、负面范围和三语；证据：`docs/acceptance/evidence/M3/KIMI-03B.md`；门禁：不在前端计算账务事实] 实现 M3 MVP 页面和 API 映射；不得新增 lesson、refund、adjust、payout、report、notification、backup 路由或菜单。

## 5. 集成验收与里程碑收口

- [ ] 5.1 [Codex；依赖：3.2、4.1、4.2；输出：真实 HTTP/浏览器财务链路证据；测试：创建学生/报名→充值/凭证→流水→作废、跨角色/未认证下载、失败补偿；证据：`docs/acceptance/evidence/M3/verification-report.md`；门禁：账务核对零差异] 在隔离数据目录运行端到端验收；不得对真实或共享开发数据执行破坏性 migration down。
- [ ] 5.2 [Codex；依赖：5.1；输出：M2+M3 全量质量批次、状态/路线图/追踪矩阵；测试：前端全量 unit/coverage、后端全量、适用 browser/CI；证据：`docs/acceptance/evidence/M3/release-gate.md`；门禁：无未解决 P0/P1] 执行已登记的延后全量前端回归并决定 M2/M3 是否能进入 ACCEPTED；在此之前任务不得自行勾选或声称完成。
