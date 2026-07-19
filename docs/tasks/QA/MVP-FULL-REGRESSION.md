# MVP 全量回归与 Go/No-Go 工单

## 目标与基线

本工单只验证已经合入 `main` 的 M1–M6 MVP，不新增功能、不修改产品代码、不调整 OpenSpec 任务勾选。开始前必须执行 `git pull --ff-only`，记录实际 HEAD、Go/Node 版本和隔离数据库目录。任何失败都以可复现步骤、请求/响应（脱敏）和截图/trace 记录；不得用改测试或降低断言方式绕过。

## 共同环境

- Windows 10 JP 可运行，仓库和测试数据均采用 UTF-8。
- 使用全新临时 `ZEDU_DB_PATH`、`ZEDU_UPLOAD_DIR`、`ZEDU_BACKUP_DIR`；不得操作共享或真实数据。
- 设置长度不少于 32 字符的临时 `ZEDU_JWT_SECRET`；使用 Resend 测试 key 或明确验证 outbox 失败重试路径，禁止发送真实收件人。
- 启动 Go 后端与 Vite 前端；所有 HTTP 请求走真实前后端，不接受 mock 作为验收证据。

## GLM：后端、迁移与账务一致性

1. 执行 `go fmt ./...`、`go vet ./...`、`go test ./... -count=1`、`go build ./cmd/zedu-server`，报告每项输出与耗时；不执行历史性的 20 次重复门禁。
2. 验证 migrations 001–008 的 up/down/up（在隔离库）；确认 student 非空邮箱唯一、M3 充值/作废冲正、M5 attendance/lesson_finance/student/teacher ledger 和 operation_log 的原子性。
3. 特别验证 M5：`STUDENT_LEAVE` 的 `0.5` 课时确认后，学生流水查询仍可读，返回 `lessonDelta=-0.5` 和 `lessonBalanceAfter=-0.5`；负数、非数字和超过三位小数的扣课值必须 42201 且不留事实。
4. 验证 M4/M5 边界：排课/取消不会产生财务或 attendance；仅一次课后确认可成功；余额不足、重复确认、并发确认全部无半写。
5. 验证 M4b outbox：创建/取消课次生成预期邮件任务；发送失败状态可重试；日志、响应、审计不出现 token、密码、邮箱正文以外的敏感凭证。
6. 验证 M6：`GET /dashboard` 只读；Owner 能创建可打开的 SQLite 备份并产生 `BACKUP_CREATE` 审计；Operator 40301；不存在 restore HTTP 路由。

## Kimi：真实浏览器人工模拟与三语

1. 使用 Playwright 或等价真实浏览器，在同一隔离环境跑 zh-CN、ja-JP、en-US；保存截图、trace 和 `results.json`。
2. 完整主链路：登录 → Owner 初始化 → 学生/家长/老师/课程 → 报名/师生安排 → 充值/凭证 → 排课 → 通知 outbox → 课后确认 → 工作台 → Owner 备份。
3. M5 表单必须验证：选择结果类型后可编辑建议值；提交的是实际输入值，不是固定 `1/0/0`；小数 `0.5` 可提交；负数在浏览器约束或后端 42201 被阻止。
4. 验证 Owner/Operator 的路由与操作权限、未认证 40101、冲突 40901、无删除/退款/结款/restore UI 或 API 入口。
5. 验证页面中 M4–M6 新增文字在三种 locale 下可显示，无硬编码英文、乱码或空翻译；重点检查课程确认、通知、工作台和备份。

## 交付与判定

- GLM 输出 `docs/acceptance/evidence/MVP/backend-regression.md`；Kimi 输出 `docs/acceptance/evidence/MVP/browser-regression.md`，另附截图、trace、JSON 结果。
- P0/P1 为阻断：账务/权限/原子性/敏感信息/数据损坏/三语不可用。P2 可带风险进入人工 UAT，但必须说明缓解措施。
- 两份独立报告完成后，由 Codex 汇总自动化结果、人工核算结果和发布风险，才可勾选 M5.6、M6.4 并作 MVP Go/No-Go 决策。
