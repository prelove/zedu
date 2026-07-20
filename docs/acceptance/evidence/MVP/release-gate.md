# MVP 发布闸门

## 当前决策

**CONDITIONAL GO（MVP 技术发布候选）**，日期：2026-07-20。

M2–M6 的实现、后端回归和真实浏览器回归均已完成技术验证。M5/M6 不标记为 `ACCEPTED`：正式可用仍需要 Product Owner 完成外部 UAT 与运维确认。

## 已通过的技术闸门

| 闸门 | 结果 | 证据 |
|---|---|---|
| 后端 fmt / vet / build / 回归、001–009 migration up/down/up、备份恢复演练 | PASS | `backend-regression.md`、`backup-recovery-drill.md` |
| 前端 lint / 类型检查 / Vitest / 构建 | PASS，40 文件 / 365 测试 | 2026-07-20 前端门禁与组件回归 |
| M1–M6 真实 Vite 浏览器回归 | PASS，60/60 | `browser-regression.md` |
| PRD §23.2 缺口追踪、payable、提醒、备份包/验证 | PASS | `prd-gap-closure.md` 及本目录各功能证据 |

## 发布前仍需 Product Owner 完成的外部闸门

1. **人工账务核对**：以受控样本核对充值、作废反向流水、课后确认余额和教师应付展示；结款/付款不属于 MVP。
2. **Resend 受控发送**：在不提交 API key 的前提下配置受控发件域与测试收件人，验证一次成功、一次可重试失败，并保留发送日志证据。
3. **备份恢复 runbook**：由被授权操作者执行一次独立目录恢复与 `zedu-backup-verify` 演练，确认不会覆盖活动数据库。
4. **Win10 日文环境人工 UAT**：检查三语切换、日期/JPY 格式、CJK/emoji 输入、深链接与主要表单交互。

## 明确的非目标

- 不启用未批准的自动付款、结款、真实 HTTP restore、财务报告或排班外功能。
- 不在文档、源码、测试产物或 Git 中写入 Resend API key、JWT secret、访问令牌或真实个人数据。

## 状态约束

- OpenSpec 任务 5.2 已完成：真实浏览器回归已留存可复核证据。
- OpenSpec 任务 5.3 保持未完成，直到上述四项外部闸门均有 Product Owner 证据与明确 Go/No-Go 决定。
- 在此之前，项目状态应为 `IN_REVIEW` / 技术发布候选，而不是生产 `ACCEPTED`。
