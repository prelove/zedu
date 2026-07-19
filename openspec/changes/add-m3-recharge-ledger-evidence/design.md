## Context

M2 已提供认证、审计、学生、报名和余额字段，但实际 003 migration 尚未创建 `payment_method`、`student_payment`、`payment_attachment` 与 `student_account_ledger`。005 会创建这四张 M3 学生财务表；老师账务、课次和结款表仍留在后续能力，不能提前创建。M3 要在单实例 SQLite、Windows 10（日文）开发环境中建立第一条财务事实，满足 PRD v3.1 的“人工确认、历史不可覆盖、可追溯、可回滚”原则。

参与者为 Owner 和 Operator；学生、家长、老师均无登录能力。当前部署的本位币为 JPY（PRD §7.1），原始收款币种可不同。正式老师结款明确属于 V1，不能被本 change 通过隐藏路由、占位菜单或共享账务接口提前暴露。

## Goals / Non-Goals

**Goals:**

- 让授权 Owner/Operator 以幂等的人工操作确认充值、查询其明细和作废事实。
- 将 payment、RECHARGE/VOID ledger、报名余额快照、审计和首次财务记录后的本位币锁定，纳入一个可失败回滚的 SQLite 事务。
- 实现受控的支付方式配置、本位币状态和最多三份可鉴权凭证。
- 提供与后端冻结契约一致的三语前端主路径，而不让前端自行计算或信任余额。

**Non-Goals:**

- 不做老师结款/老师流水、退款或人工调整、课次/出勤、报表、支付网关、实时汇率、OCR、导出、通知、备份与恢复。
- 不做充值编辑、删除、匿名文件 URL、文件预览转码、跨实例分布式事务或文件云存储。
- 不回溯重构 M2 已验收 auth/onboarding；新增领域遵循既定 `HTTP → application → repository` 分层。

## Decisions

### 1. 金额以字符串输入、精确整数最小单位落库

请求中的 `originalAmount` 与 `fxRateToBase` 是非科学计数法十进制字符串；服务端 MUST 以标准库 `math/big.Rat`（或等价无 float 的精确算法）计算，禁止 `float32/float64`。本部署本位币 JPY，`amount_base` 是日元整数；结果采用半向上取整并在响应中同时返回原始快照和折算值。`originalCurrency == baseCurrency` 时汇率 MUST 是 `"1"`。

被拒方案：前端计算/`float64`，因为会让不同客户端或二进制浮点舍入造成账务差异；接入实时汇率 API，因为 PRD 要求事实快照且 MVP 不需要持续费用。

### 2. 充值以客户端 paymentNo 作为业务幂等键

创建充值请求 MUST 提供 UUID 格式 `paymentNo`。服务端在事务中把它写入既有唯一列；同一操作重试若所有业务字段一致，返回原先已确认的资源而不新增 ledger、余额或审计；同一号码配不同字段则返回既有 `40901`，不覆盖历史。前端在首次提交前生成并在网络重试期间复用该值。

被拒方案：仅依赖 HTTP 请求或随机服务端编号，无法区分网络超时后的安全重试与第二笔真实收款。

### 3. 单一事务是充值与作废的唯一事实边界

确认充值按顺序执行：校验 ACTIVE student/enrollment、支付方式 enabled 且存在、校验 enrollment 属于 student → 插入 payment → 锁定 `system_settings.base_currency_locked` → 读取该 enrollment 当前余额 → 插入 RECHARGE ledger（含 after snapshot）→ 原子更新 enrollment balance/lesson balance → 写 operation_log → commit。所有读写均在同一 tx；任一步 DB/审计失败返回 `50002`，不泄露 SQLite 文本。

作废用条件更新 `WHERE status='CONFIRMED'`，要求非空作废原因；同一 tx 中写负向 VOID ledger、更新余额、写审计。重复作废或无效状态返回 `42201`，不得第二次冲正。并发确认/作废依赖 SQLite 写事务与条件更新；测试必须证明余额与 ledger 只有一个赢家且不出现半写。

被拒方案：payment、ledger、余额分开提交，或“失败后由定时任务补偿”，因为会短暂或永久制造不可核对余额。

### 4. 本位币配置在第一个财务事实前受 Owner 控制

`system_settings` 使用稳定键 `base_currency`（初始默认 `JPY`）与 `base_currency_locked`（初始 `false`）。Owner 可读取状态，并且仅在不存在 M3 已落库的 payment 或 student ledger 且未锁定时修改为 PRD 已列 MVP 代码 `JPY`、`CNY` 或 `USD`；后续老师/课次财务 change 必须将各自的财务事实加入同一禁止条件。M3 以当前部署 JPY 作为精确整数最小单位实现。首次成功充值将锁定标志置为 `true`，与财务写入同一事务。修改已锁定或已有财务事实时返回 `42201`。

被拒方案：以 UI disabled 代替服务端约束，或首次 payment commit 后异步锁定，二者均可能改变既有余额的统计口径。其它本位币精度模型需要单独 ADR/Change，不能在本 M3 猜测实现。

### 5. 支付方式是可审计字典，不做物理删除

首次 M3 初始化幂等写入 `WECHAT`、`ALIPAY`、`PAYPAY`、`BANK`、`CASH`、`OTHER`。Owner 可列举、创建、更新显示名/排序/启用状态；code 创建后不可改，禁用仅影响新充值，历史 payment 仍能显示快照。Operator 仅能读取 enabled 方法用于录入。

被拒方案：数据库 CHECK 或前端常量，因为它们无法符合 PRD 的后台可维护要求；物理删除会断开历史外键。

### 6. 附件采用“临时文件 → DB 事务 → 发布/补偿”的受控存储

上传端点只接受 multipart 单文件，请求先在 `data/uploads/.tmp` 写入随机临时文件，流式限制 5 MiB 并从文件魔数/受限 MIME 推导真实类型；文件名不参与路径。事务内锁定 payment，确认其状态为 CONFIRMED，计数 `<3`，插入 attachment 元数据与审计；commit 后把临时文件原子 rename 到 `data/uploads/payments/{paymentID}/{attachmentID}.{ext}`。若 rename 失败，删除 metadata 的补偿事务并删除临时文件，响应 `50002`；启动时/测试清理由本 change 的受控 orphan-cleaner 处理只位于 `.tmp` 的残留。

下载使用 `GET /finance/payments/{paymentId}/attachments/{attachmentId}/content`，先经现有 AuthMiddleware，再校验 attachment 属于 payment、payment 可见；仅返回受控路径中登记的文件，设置安全 Content-Type 与 attachment disposition。不存在、跨 payment 或路径越界均不泄露文件系统信息，返回现有 `40401`/`50002`。

被拒方案：直接保存客户端文件名、静态目录暴露、base64 落 SQLite、先发布文件再写 DB；这些会造成路径穿越、匿名泄露、数据库膨胀或失败孤儿。

### 7. API 和前端边界

- 保持既有无 `/api` 前缀、`{code,data}` 成功信封和 `{code,message,requestId}` 错误信封。
- M3 路由：`GET/POST /finance/payments`、`GET /finance/payments/{id}`、`POST /finance/payments/{id}/void`、`POST/GET /finance/payments/{id}/attachments`、`GET /finance/payments/{id}/attachments/{attachmentId}/content`、`GET /finance/ledger/student/{studentId}`、`GET/PUT /system/base-currency`、`GET/POST/PATCH /system/payment-methods[/{code}]`。
- Owner 与 Operator 可进行充值、作废、流水与凭证操作；本位币和支付方式写操作仅 Owner。所有写操作写 operation_log，detail 不得包含 token、密码、文件内容、原始文件系统路径或敏感联系方式。
- 前端仅将后端的余额、金额、状态作为事实展示；金额输入保持字符串并提交 `paymentNo`，任何预估金额仅为 UI 辅助且不替代服务器校验。

## Risks / Trade-offs

- [文件 publish 与 DB commit 不可跨资源原子] → 临时文件、原子 rename、反向 metadata 补偿、残留清理和失败注入测试；不会声称跨系统 ACID。
- [SQLite 单写者与充值并发] → 事务内读取/更新、条件作废、busy timeout、并发测试；MVP 不尝试多节点并行写。
- [金额原币有小数、JPY 本位为整数] → 服务端精确计算和固定舍入规则；响应与 ledger 显示最小单位语义。
- [凭证属于敏感个人数据] → 认证下载、不可预测路径、日志脱敏、无匿名 URL；备份包含 uploads 的验证留给 M6，不在本 change 假装完成。
- [M2 前端总回归延后] → 状态中明确为“质量债务/里程碑门禁”，M3 的 focused test 与 typecheck 仍必须各自新鲜通过。

## Migration Plan

1. 新建 005 migration：仅补 M3 所需的约束、索引、初始 settings/payment methods，保留 003 的历史表；up/down/up 测试必须通过。
2. 先落后端迁移、repository/service/handler 与红灯事务/安全测试，再落前端 adapter、页面和三语测试。
3. 在隔离临时 data root 做真实 HTTP 测试：充值、附件、作废、鉴权下载、失败补偿和并发。
4. 发布前运行 M3 focused gates；M2/M3 全量前端回归与浏览器 E2E 汇入下一次里程碑候选总门禁，不提前标记 ACCEPTED。
5. 回滚仅在没有真实财务数据时执行 migration down；已有财务记录禁止破坏性 down，采用部署回退并保留 data/uploads。代码回滚不会删除真实 payment 或凭证。

## Open Questions

- 无阻塞问题。本 M3 按 PRD 的当前部署 JPY 和默认支付方式执行；“其它本位币”及其最小单位规则属于需要独立产品/ADR 决策的扩展，不能在财务实现中自行猜测。
