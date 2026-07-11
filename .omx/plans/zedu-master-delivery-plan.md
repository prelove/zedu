# Zedu 总交付计划（共识批准版）

## RALPLAN-DR 摘要

### 原则

1. 唯一事实源与零孤儿追踪。
2. MVP 是可运营闭环，规格/测试/证据先于完成声明。
3. 财务正确性、最小权限和可恢复性不可让步。
4. 外部副作用与主事务隔离，采用最小可用设计。
5. 三语、Win10 JP 和多 AI 工具可复现。

### Top 3 驱动

1. 课后确认与账务的高风险一致性。
2. GLM/Devin/Kimi/Claude/Codex 协作中的上下文稳定性。
3. 快速取得真实 MVP 反馈且不丢失 V1 定位。

### 方案

| 方案 | 结论 | 原因 |
|---|---|---|
| 原地升级旧001-014 | 拒绝 | 迁移快但继续混合MVP/V1与多能力，独立验收差 |
| 按技术层重组 | 拒绝 | ownership清楚但切断用户价值，跨层Scenario易丢 |
| 按业务能力/里程碑重组 | 采用 | 追踪、独立验收和AI执行性最佳；以legacy映射控制迁移风险 |

## 执行顺序

`M0治理迁移 → M1工程基线 → M2认证资料报名 → M3充值凭证 → M4a排课 → M4b通知 → M5课后账务 → M6恢复与MVP验收 → V1结款/通知/字典/报表/移动/迁移`

M4b 适配器和模板可在 M4a 事件契约冻结后并行；任何并行任务必须单 owner、限定文件和独立验收。

## ADR

### Decision

冻结旧 OpenSpec 为迁移输入，在 OpenSpec 1.6 下按业务能力和里程碑重组；先交付包含 Resend 与付款凭证、排除正式结款的 MVP。

### Drivers

- MVP/V1 边界与真实运营频率。
- 财务、文件、通知的可靠性及安全风险。
- 多 AI 工具需要小而明确的可验证任务。

### Alternatives considered

- 原地升级旧 change。
- 按 database/backend/frontend 技术层重组。
- 完全丢弃旧文档重新生成。

### Why chosen

按能力重组保留了旧文档价值，又能让每个 change 独立归档、验收并映射用户旅程；迁移成本通过 hash、disposition 和零孤儿检查控制。

### Consequences

- M0 增加前置治理成本。
- 旧编号不再直接作为执行编号，必须持续维护映射。
- 通知采用 outbox-lite，文件系统采用显式补偿和恢复验证，代码略增但风险显著降低。

### Follow-ups

- 修订正式 PRD 23.2/23.3/24.1、风险表和变更记录。
- 升级并锁定 OpenSpec/Superpowers。
- 生成新 changes、规范、路线图、状态表及 CI 门禁。
- M3/M4b/M5 前分别完成运营参数确认。

## Pre-mortem

| 失败 | 指标/门槛 | Owner | 缓解与演练 |
|---|---|---|---|
| 迁移丢需求 | orphan>0 | PM/Architect | 全量矩阵；陌生AI抽样复述5个Scenario |
| 重复/漏邮件 | duplicate>0或pending>15m | Backend | DB唯一键、lease；双worker/kill/429/503演练 |
| 凭证泄露/孤儿 | IDOR或orphan>0 | Security | 对象授权、补偿；穿越/伪MIME/磁盘满/DB失败 |
| 财务漂移 | discrepancy!=0 | Backend/QA | 单事务幂等；全故障点及20次并发 |
| Win10 JP乱码 | missing key/replacement char>0 | Frontend/QA | UTF-8/key parity；日文环境路径/CSV/邮件 |
| 备份不可恢复 | checksum差异或backup>24h | Release | 临时恢复、双向核对、原子切换演练 |
| MVP被结款拖延 | payout入口>0 | PM | 负向扫描UI/API/route/flags |

## 进度治理

状态：`BACKLOG → READY → IN_PROGRESS → BLOCKED → IN_REVIEW → VERIFIED → ACCEPTED`。

- 只有 READY 可认领；一个 task 一个 owner。
- BLOCKED 记录原因、影响、替代方案和决策人。
- VERIFIED 必须有本次运行证据；ACCEPTED 必须通过里程碑门禁。
- 每次合并更新路线图、状态、风险、决策和追踪矩阵。

## Available-Agent-Types Roster

- `architect`：架构、事务、边界与 ADR，高推理。
- `executor`：实现与重构，高推理。
- `test-engineer`：测试设计和故障注入，中高推理。
- `security-reviewer`：认证、IDOR、上传、secret，中高推理。
- `verifier`：独立执行门禁和证据核验，高推理。
- `explore`：文件/符号映射，低推理。
- `style-reviewer`：格式、命名、编码，低推理。
- `writer`：中文规范和用户文档，中推理。

## 后续执行配置

### Ralph（顺序持久化）

适合 M0/M1 和单个高风险 change。建议每个 change 使用 executor 主责，test-engineer 与 verifier 独立复核；M3/M4b/M5 增加 security-reviewer/architect。

启动提示：`$ralph .omx/plans/zedu-master-delivery-plan.md`。

### Team（协调并行）

在规格和共享契约冻结后使用。建议 3 个 lane：backend executor、frontend executor、test-engineer；leader 负责共享文件和整合，security/verifier 作为后续独立门禁。

启动提示：`$team .omx/plans/zedu-master-delivery-plan.md`，或按本机 OMX 版本使用等价 `omx team` 命令。

### Team Verification Path

Team 关闭前证明：任务范围、代码审查、相关 unit/integration/E2E、追踪矩阵和 evidence 已完成。随后由独立 verifier/Ralph 重跑完整里程碑 Gate，检查无隐藏 TODO、无范围漂移、无未解决 P0/P1，再标记 ACCEPTED。

## 共识改进记录

- 将排课与通知拆为 M4a/M4b。
- 将通用事件框架收敛为 SQLite outbox-lite。
- 增加 recipient 级状态与“供应商接受≠送达”语义。
- 增加凭证对象授权、安全响应、原子文件流程与恢复演练。
- 增加三语/Win10 JP 矩阵、量化可观测性和无结款负向验收。
- 将7天从硬承诺改为不牺牲质量的目标。
