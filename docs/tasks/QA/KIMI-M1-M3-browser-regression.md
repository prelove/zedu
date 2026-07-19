# KIMI 浏览器回归工单：M1–M3 模拟人工验收

## 目标与边界

在隔离的本地测试数据库和测试上传目录中，使用真实浏览器验证 M1–M3 已交付主链路；这是自动化回归证据，不是最终人工 UAT。不得修改产品代码、迁移、OpenSpec、状态文档或提交代码。

产品范围仅到 M3：认证、初始化、学生/老师/课程/报名/安排、财务配置、充值、VOID、凭证。禁止测试或实现排课、通知、课消、退款、调整、老师结款、报表、备份。

## 必读资料

1. `docs/governance/GOVERNANCE.md`
2. `docs/standards/implementation-contract.md`
3. `openspec/changes/add-m2-core-management/**`
4. `openspec/changes/add-m3-recharge-ledger-evidence/**`
5. `docs/acceptance/evidence/M1/verification-report.md`

## 环境隔离

- 后端以独立 `ZEDU_DATABASE_DSN`、`ZEDU_DATA_ROOT`、测试 JWT secret 启动；严禁使用共享 `zedu.db`。
- 前端用 Vite 开发服务器访问后端；确认代理含 `/auth`、`/onboarding`、`/students`、`/teachers`、`/course-*`、`/enrollments`、`/system`、`/finance`。
- 所有截图、trace、浏览器输出仅写入 `output/playwright/m1-m3/`；不得纳入提交。
- 首先确认 `npx` 可用，使用仓库/系统提供的 Playwright CLI；每次交互前 snapshot。

## 测试数据与账号

- 创建唯一 Owner 与 Operator 测试账号；密码与 token 不得写入日志、截图说明或证据正文。
- 用中文、日文、英文各至少一条文本；包含 CJK 和 emoji 的备注/名称，确认显示无乱码。
- 浏览器 locale 至少验证 zh-CN、ja-JP、en-US 切换；金额显示使用 JPY，日期显示使用 Asia/Tokyo。

## 必测主链路

### A. M1 基础

1. 健康页能打开，前端可访问后端。
2. 三语切换后已有页面文本可读，无 `undefined`、缺 key 或乱码。

### B. M2 身份与教务资料

1. Owner 登录、退出；受保护路由未登录时回到登录页。
2. Operator 登录后不可访问 Owner-only 初始化/财务配置。
3. Owner 初始化一次；再次初始化展示既有结果而非重复创建。
4. 创建学生（邮箱可空）；重复非空邮箱显示冲突错误。
5. 创建老师、课程领域/方向/等级、报名、ACTIVE 师生安排；列表、详情和分页可用。

### C. M3 财务与凭证

1. Owner 查看/修改未锁定本位币；创建或编辑支付方式；禁用方式仍在历史配置可见。
2. Operator 可创建确认充值：输入金额和汇率为字符串，生成一次 paymentNo；列表、详情、学生流水一致。
3. 同一 paymentNo 重试不产生重复充值；相同号码改字段显示冲突。
4. 上传合法 PNG/PDF 凭证，列表可见且鉴权下载成功；错误 paymentId/attachmentId 不泄露内容。
5. 同一付款最多三份附件；第四份得到稳定校验错误。
6. 作废已确认充值必须填写原因；付款变为 VOIDED，流水出现反向记录，余额回到作废前；再次作废被拒绝。

## 必测负面与安全路径

- 无 token 访问受保护 M2/M3 路由：40101。
- Operator 访问 Owner-only：40301。
- 上传伪造 MIME、超 5MiB、路径穿越文件名：被拒绝且无可下载孤儿文件。
- 页面和网络响应中不得显示 password、password_hash、Authorization、access/refresh token 或其哈希。
- M3 页面与菜单中不得出现 refund、adjust、lesson、attendance、payout、report、notification、backup。

## 交付格式（只报告，不提交）

1. 环境命令、版本、隔离路径（不含 secret）。
2. 每个场景：步骤、预期、实际、PASS/FAIL、截图或 trace 路径。
3. 缺陷按 P0/P1/P2 分级；每项给最小复现、URL、角色、浏览器和 console/network 证据。
4. 单列“非阻塞、留待统一 UAT”的项目。
5. 最终结论只能为 `PASS`、`PASS_WITH_RISKS` 或 `FAIL`；不得修改任务状态。
