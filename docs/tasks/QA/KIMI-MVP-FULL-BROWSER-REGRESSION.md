# Kimi 执行指令：MVP 全量真实浏览器回归

你是独立 QA 执行者。目标是对 `main` 当前 HEAD 进行 M1–M6 MVP 的真实浏览器回归，并产生可供 Codex Go/No-Go 审查的证据。你没有修改产品代码、OpenSpec、任务勾选、路线图或状态文件的权限；发现问题只写报告，不自行修复。

## 先读上下文

依次阅读：

1. `AGENTS.md`
2. `docs/status/PROJECT_STATUS.md` 与 `docs/roadmap/MASTER_ROADMAP.md`
3. `docs/standards/implementation-contract.md`
4. `docs/acceptance/evidence/MVP/IMPLEMENTATION_AUDIT.md`
5. `docs/tasks/QA/MVP-FULL-REGRESSION.md`
6. M2–M6 的 OpenSpec `proposal.md`、`design.md`、`specs/**/spec.md`、`tasks.md`

执行前运行 `git pull --ff-only`，报告 HEAD SHA。当前回归必须基于含 `frontend/vite.config.ts` 的 `/lessons`、`/notifications`、`/dashboard` 代理配置的 HEAD。

## 环境与硬约束

- 使用新的临时数据库、上传目录、备份目录；不得使用或删除共享/真实数据。
- 使用真实后端 `:8080` 和真实 Vite `:5173`；浏览器请求必须经过 Vite 代理。
- **禁止** 对 `/lessons`、`/notifications`、`/dashboard` 使用 `page.route`、mock、直接替换 fetch 或手工 HTTP 转发。若代理失败，测试应失败并报告。
- 可使用 Playwright，保留 `trace: on`、失败截图和 HTML/JSON 报告。
- 每种 locale 使用独立上下文或重置状态：zh-CN、ja-JP、en-US。
- 仅使用 disposable 邮箱/Resend 测试环境；不得向真实收件人发信。

## 必须实现的用例（每个 locale 都执行）

### A. 认证与初始化（M1/M2）

1. 未登录访问 `/students`、`/lessons`、`/dashboard` 被重定向至 `/login`。
2. Owner 登录成功；错误密码为统一失败提示且页面、控制台、网络记录不泄漏密码/token。
3. Owner 初始化模板一次成功，重复执行显示幂等结果；Operator 无初始化入口且后端返回 40301。
4. Owner 创建 Operator；Operator 可以运行日常业务，不能访问 Owner 限制配置/备份。

### B. 主数据与排课（M2/M4a）

5. 创建无邮箱学生成功；重复非空邮箱被 40901 阻断，没有“仍然新建”绕过。
6. 创建家长、老师、能力、可用时间；验证能力层级和时间输入错误被阻断。
7. 创建课程领域/方向/等级；创建学生报名及 ACTIVE 师生安排。
8. 创建 Asia/Tokyo 课次，列表显示；修改仍为 SCHEDULED 的课次；取消后不可再修改。
9. 确认页面从 Vite 实际代理加载 `/lessons`，不得有 404 或 route 拦截。

### C. 充值与凭证（M3）

10. Owner 配置本位币/支付方式；Operator 创建一笔充值并能查看学生流水。
11. 重复 paymentNo 被稳定阻断；上传、列出并下载凭证；超过三份被阻断。
12. 作废充值后出现不可变反向流水；页面无退款、人工调账或老师结款入口。

### D. 通知与课后确认（M4b/M5）

13. 创建/取消课次后通过实际代理打开 `/notifications`，验证存在预期 outbox 行；无 Resend key 时失败状态不泄漏 key/token，符合重试资格的记录可重试。
14. 进入课后确认表单，选择 `STUDENT_LEAVE` 并提交实际值：`lessonDeducted=0.5`、正整数 charge/teacher pay、实际时长。验证不是固定 `1/0/0`。
15. 提交后 lesson 完成，重复确认被阻断；尝试负数、非数字、四位小数扣课值被页面或 42201 阻断。访问学生流水，确认小数课时记录可显示。
16. 余额不足的确认失败，刷新后课次仍 SCHEDULED 且没有半完成事实。

### E. 工作台与备份（M6）

17. `/dashboard` 通过真实代理显示待确认课程与失败通知计数，且不改变业务事实。
18. Owner 可创建备份并展示成功文件名；Operator 看不到或被 40301 阻断；浏览器、网络和菜单中都不存在 restore 操作。

### F. 范围、国际化和安全负面项

19. 三种 locale 均显示 M4–M6 的课程确认、通知、工作台、备份文案；不得有乱码、空文案或新增硬编码英文。
20. 检查网络与 DOM：不存在 payment card/Authorization/refresh token/password hash/Resend API key；不存在 lesson DELETE、refund、payout、settlement、report 或 restore HTTP 请求/入口。

## 交付格式

输出到 `output/playwright/mvp-full/`：

- `REGRESSION_REPORT.md`：环境、HEAD、每个用例/locale 的 PASS/FAIL、缺陷优先级、复现步骤、明确说明“未使用 route/mock”。
- `playwright-report/index.html`、`playwright-report/results.json`。
- `test-results/` 下失败截图和 trace；关键成功路径至少各 locale 一张截图。

最终只报告：总通过数、失败数、阻断 P0/P1、已知 P2 风险及产物绝对路径。不得提交、推送或修改产品文件；等待 Codex 独立汇总。
