# M3-GLM-03A：财务配置基础（005 migration、本位币、支付方式）

| 项 | 内容 |
|---|---|
| 状态 | READY |
| Owner | GLM |
| 前置 | M3-CODEX-01 VERIFIED；不得等待或修改 M2 未验收前端内容 |
| 基线 | 当前本地 `main`，以 `docs/status/PROJECT_STATUS.md` 为准 |
| 目标 | 只完成 M3 财务配置基础，为后续充值、流水和凭证提供稳定契约 |
| 交付后状态 | IN_REVIEW；不得自行合并 main、勾选 OpenSpec 或称 M3 ACCEPTED |

## 0. 必读上下文（按顺序）

1. `AGENTS.md`、`docs/status/PROJECT_STATUS.md`、`docs/roadmap/MASTER_ROADMAP.md`、`docs/governance/GOVERNANCE.md`。
2. `docs/standards/implementation-contract.md`、`docs/tasks/M3/M3_EXECUTION_BOARD.md`。
3. `openspec/changes/add-m3-recharge-ledger-evidence/{proposal.md,design.md,tasks.md}` 与 `specs/financial-configuration/spec.md`。
4. PRD `docs/2_prd/Zedu-PRD-Final-v3.1.md`：§7.1、§7.3、§13.12、§14.1–14.3（R16）、§16.3–16.5、§23.2、§24.4。
5. 已验收样例：`backend/internal/app/{directory,course}/` 的 HTTP → application → repository 分层、`backend/internal/app/onboarding/` 的设置/审计写法、`backend/internal/platform/httpserver/` 的错误契约。

只做上述定向阅读；不要做全仓扫描。发现契约缺失或与 PRD 冲突时停止并报告，不自行补需求。

## 1. 允许与禁止范围

允许新增或修改：

- `backend/migrations/005_m3_financial_configuration.{up,down}.sql`
- `backend/internal/platform/database/**`（仅 migration/测试辅助）
- `backend/internal/app/finance/**`（新领域，必须 HTTP → application → repository）
- `backend/cmd/zedu-server/main.go`（仅 M3 finance routes 挂载）
- 必要的后端测试文件。

禁止：`frontend/**`、`.github/**`、`openspec/**`、共享状态/路线图/证据、既有 auth/onboarding/directory/course 行为、新依赖、新错误码、新角色、`float32/float64` 金额运算、充值/流水/作废/附件实现、退款、人工调整、课次、出勤、老师流水、结款、报表、通知、备份，以及生产测试旁路/全局 hook/test mode/静默吞错。

## 2. 必须交付的行为

### 2.1 005 migration

- 幂等初始化 `system_settings.base_currency=JPY` 和 `system_settings.base_currency_locked=false`；up/down/up 后默认值稳定。
- 幂等初始化 `payment_method`：`WECHAT`、`ALIPAY`、`PAYPAY`、`BANK`、`CASH`、`OTHER`；稳定 code、可编辑 name、sort_order、enabled；不得用 CHECK 写死支付方式。
- 仅补 M3 必需的索引/约束；不得提前建立 payment、ledger、attachment 的业务写路径，也不得改变 003 已验收语义。
- down 只回退 005 自己创建/插入的可逆对象；不得破坏真实财务数据。

### 2.2 API（无 `/api` 前缀，沿用既有 JSON 信封）

- `GET /system/base-currency`：Owner/Operator，返回 `{currency, locked}`。
- `PUT /system/base-currency`：仅 Owner，`{currency}` 只接受 `JPY`/`CNY`/`USD`。只有不存在任何 `student_payment`、`student_account_ledger`、`teacher_account_ledger`、`lesson_finance`、`teacher_payout`，且 lock 不为 true 时可改；否则 `42201` 且不改值。
- `GET /system/payment-methods`：Owner/Operator；Operator 只看 enabled，Owner 可显式请求含 disabled 的管理列表。
- `POST /system/payment-methods`：仅 Owner；创建 uppercase code、name、sortOrder、enabled。重复 code 为 `40901`。
- `PATCH /system/payment-methods/{code}`：仅 Owner；只允许 name/sortOrder/enabled，不允许改 code。未知为 `40401`，空 patch/非法字段为 `42201`；禁用不删除历史。

所有成功写操作与 `operation_log` MUST 同一 tx；审计包含 actor/action/target/request_id 和无敏感 detail。DB/BeginTx/Commit/audit failure MUST 是 `50002`，不得泄露 SQLite 文本；权限拒绝不得产生成功审计。

### 2.3 分层与可观测性

- 新包按 `finance/handler.go`、`service.go`、`repository.go`（可按职责拆分）组织，handler 不直接写 SQL。
- 本任务不计算充值金额；任何保留的金额代码均不可用 float。
- 错误日志带 request_id，不得记录 password、token、Authorization、个人联系方式或完整 SQLite 文本。

## 3. 先红后绿的测试矩阵

1. 005 up/down/up、默认设置/六个方式、UTF-8（中文/日文/emoji）往返。
2. Owner/Operator/未认证的 base currency 权限；合法修改；存在每一种财务表记录或 lock 时拒绝；失败不改值。
3. Owner/Operator/未认证的 payment method 权限；重复 40901；禁用后管理列表保留、Operator 列表隐藏；不可改 code。
4. base currency 与 payment create/update 的审计存在且有 request_id，detail/log 不含敏感字段。
5. DB/事务/审计故障注入无半写，统一 `50002`。
6. 负面范围：没有 `/finance/payments`、ledger、attachment、payout、refund、adjust、lesson、attendance、notification、backup 路由。

执行（不跑 20 次重复；该门禁已移至里程碑质量批次）：

```powershell
Set-Location backend
$env:GOTOOLCHAIN='local'
go version
go fmt ./...
go vet ./...
go test ./... -count=1
go build ./cmd/zedu-server
```

## 4. 交付与回滚

每个提交均遵循 Lore protocol，至少含 `Constraint`、`Rejected`、`Confidence`、`Scope-risk`、`Directive`、`Tested`、`Not-tested` trailers。不得提交未通过门禁的实现。

交付报告包含 commit SHA、完整改动文件、红灯→绿灯证据、上述六类契约映射、命令输出摘要、未测试项、风险和 migration 回滚说明。等待 Codex 独立验收。

回滚：无真实财务数据的临时/开发库可执行 005 down；已有真实财务事实的环境不得 destructive down，须停止并升级。
