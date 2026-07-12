# Zedu 总路线图

> 当前焦点：M2 核心资料与认证。更新日期：2026-07-12。进度以验收门禁而非主观百分比计算。

| 里程碑 | 能力 | 状态 | 退出条件 |
|---|---|---|---|
| M0 | 治理、PRD修订、OpenSpec 1.6迁移、规范、仓库基线 | ACCEPTED | strict通过、旧需求零孤儿、文档/证据齐全 |
| M1 | 工程骨架、迁移、CI、i18n与质量门禁 | ACCEPTED | Windows/Ubuntu CI全绿，Linux race、up/down/up及Win JP通过 |
| M2 | 认证初始化、人员课程、报名安排 | IN_PROGRESS | 已冻结契约；RBAC与核心资料E2E通过后验收 |
| M3 | 充值、流水、付款凭证 | BACKLOG | 事务/IDOR/恶意文件/补偿/恢复通过 |
| M4a | 排课 | BACKLOG | 排课与通知副作用解耦 |
| M4b | Resend通知outbox-lite | BACKLOG | 幂等/lease/失败/重放/三语通过 |
| M5 | 课后确认、学生流水、老师应付 | BACKLOG | 并发/回滚/核账差异0，无结款入口 |
| M6 | 工作台、备份恢复、MVP验收 | BACKLOG | MVP Go/No-Go与业务签字 |
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
