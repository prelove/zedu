# Zedu 总路线图

> **最新路线图状态（2026-07-20，优先于本文较早的执行记录）**：M2–M5 已 `VERIFIED`，M6 为 `IN_REVIEW`。MVP 已进入技术发布候选，不扩展 V1 范围；只收口 OpenSpec `complete-prd-mvp-release-gaps` 的 Product Owner 外部闸门。

## MVP 收口看板（最新）

| 里程碑 | 状态 | 技术完成物 | 下一步 |
|---|---|---|---|
| M2 | VERIFIED | 认证、初始化、人员、课程、报名、师生安排 | 保持回归，不追加范围。 |
| M3 | VERIFIED | 充值、流水、凭证上传/下载/限制、作废反向流水 | 人工账务核对。 |
| M4a | VERIFIED | 课次创建、Asia/Tokyo 排期、权限与审计 | 浏览器回归已覆盖。 |
| M4b | VERIFIED | outbox、通知 UI、提醒 runner/CLI | Resend 受控真实发送验证。 |
| M5 | VERIFIED | 课后确认、余额约束、教师应付只读展示 | 结款/付款仍为非 MVP。 |
| M6 | IN_REVIEW | 工作台、备份包、校验 CLI、发布闸门证据 | 完成外部 UAT 后由 Product Owner 决定 Go/No-Go。 |

技术回归记录：真实 Vite 浏览器回归 20 组 × 三语 = **60/60 PASS**；详见 `docs/acceptance/evidence/MVP/browser-regression.md`。M6 不能因技术通过自动转为 `ACCEPTED`，外部闸门见 `docs/acceptance/evidence/MVP/release-gate.md`。

> 当前焦点：MVP 外部发布闸门（Product Owner UAT）。更新日期：2026-07-20。进度以验收门禁而非主观百分比计算。

| 里程碑 | 能力 | 状态 | 退出条件 |
|---|---|---|---|
| M0 | 治理、PRD修订、OpenSpec 1.6迁移、规范、仓库基线 | ACCEPTED | strict通过、旧需求零孤儿、文档/证据齐全 |
| M1 | 工程骨架、迁移、CI、i18n与质量门禁 | ACCEPTED | Windows/Ubuntu CI全绿，Linux race、up/down/up及Win JP通过 |
| M2 | 认证初始化、人员课程、报名安排 | VERIFIED | RBAC 与核心资料 E2E 已通过 |
| M3 | 充值、流水、付款凭证 | VERIFIED | 事务/IDOR/恶意文件/补偿/恢复技术验证通过 |
| M4a | 排课 | VERIFIED | 排课与通知副作用已解耦并回归 |
| M4b | Resend通知outbox-lite | VERIFIED | 幂等/lease/失败/重放/三语技术验证通过；等待受控真实发送 |
| M5 | 课后确认、学生流水、老师应付 | VERIFIED | 并发/回滚/核账技术验证通过，无结款入口 |
| M6 | 工作台、备份恢复、MVP验收 | IN_REVIEW | 技术 Go；等待 Product Owner UAT/业务签字 |
| V1 | 正式结款、完整通知、字典、报表、移动/打包、迁移 | BACKLOG | 分能力验收后V1签字 |
| V1.5/V2 | PWA/Wails、规则增强、门户、支付API、AI辅助 | BACKLOG | PRD触发条件满足后立项 |

## M0 当前任务

- [x] 需求深访和共识计划
- [x] OpenSpec 1.6 工具核验及多工具初始化
- [x] 旧 001-014 只读迁移
- [x] 修订正式 PRD MVP/验收口径（v3.1）
- [x] 完成 Claude/OpenSpec 评审和首版追踪矩阵
- [x] 建立编码、测试、验收、AI协作、i18n、安全规范
- [x] 生成并严格校验首个 READY change（后续能力按依赖逐个创建，禁止空壳）
- [x] 初始化 Git 并配置 GitHub remote
- [x] 按Lore协议提交并推送GitHub

## M1 验收结论

- 结论：ACCEPTED；GitHub Actions run `29153829469` 全绿。
- 证据：`docs/acceptance/evidence/M1/verification-report.md`。
- 下一关：M2 OpenSpec 规格冻结，不允许绕过该关直接实现业务功能。

## M2 当前进度

- M2-CODEX-01（契约冻结）：ACCEPTED。
- M2-GLM-01（认证/RBAC/M2 基础迁移）：ACCEPTED；PR #2 已于 2026-07-16 合并至 `main`。Windows 与 Ubuntu CI 全绿，Ubuntu 已通过 `go test ./... -race -count=1`；独立验收证据见 `docs/acceptance/evidence/M2/GLM-01.md`。
- M2-GLM-02A（显式初始化与受限重置）：ACCEPTED；`3bc4078` 已合并 `main`，Windows/Ubuntu CI 与 Linux race 全绿，证据见 `docs/acceptance/evidence/M2/GLM-02A.md`。
- M2-GLM-02B/02C（人员资料、课程、报名与安排 API）：ACCEPTED。复验已覆盖分页、ACTIVE 学生约束、字典引用完整性、等级历史、错误/事务语义；额外收口了连续等级事件链、课程选择与等级变更的互斥写入、等级事件引用保护及 PRD 等级事件枚举。证据见 `docs/acceptance/evidence/M2/GLM-02BC.md`。新领域保持 `HTTP → application → repository` 分层，不回溯重构已验收认证/初始化。
- M2-KIMI-01（认证前端、受保护路由、Owner 初始化界面）：ACCEPTED。复验补足了 Operator 被 Owner 路由拒绝后的可见三语提示，并将可变更的冒烟脚本改为显式的 disposable-environment 门禁；证据见 `docs/acceptance/evidence/M2/KIMI-01.md`。
- M2-KIMI-02 已交付并处于 IN_REVIEW：核心页面、三语与组件拆分已完成；Windows 前端全量 Vitest 作为延后质量批次，M2-CODEX-02 仍等待该批次与浏览器集成验收，不能提前标记 M2 ACCEPTED。

## M3 当前进度

- M3-CODEX-01（契约冻结）：VERIFIED。`add-m3-recharge-ledger-evidence` 已具备 proposal/design/specs/tasks 并通过 strict；冻结证据见 `docs/acceptance/evidence/M3/contract-freeze.md`。
- M3-GLM-03A（005 migration、本位币、支付方式）：READY；后续账务与附件实现受其依赖约束。
- M3-GLM-03B/03C、M3-KIMI-03A/03B、M3-CODEX-02：按 `docs/tasks/M3/M3_EXECUTION_BOARD.md` 受依赖阻塞，尚未开始。
