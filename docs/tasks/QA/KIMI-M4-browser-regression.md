# Kimi M4 浏览器模拟人工验收工单

## 基线与目标

基线：`main` 的 M4a `abdcd61` 与 M4b `5c0ae77`。仅验证浏览器行为，不修改产品代码、OpenSpec、状态/路线图或依赖，不提交。

目标：补充 M4a 排课与 M4b 通知 outbox 的真实浏览器验收，并复核上一轮报告的 P1/P2 风险。

## 启动方式

使用独立临时 SQLite DB 与 `ZEDU_JWT_SECRET`；不得使用、读取或输出 Resend key。可使用 fake/未配置 sender 验证稳定错误，禁止发送真实邮件。

## 必测场景

1. Owner/Operator 登录后可从导航进入 `/lessons`；未登录访问被导向登录。
2. 创建 Asia/Tokyo 课程、筛选、取消；取消后不可再次编辑；验证页面不显示财务、考勤、结款入口。
3. 课程创建后，使用 Owner/Operator 进入 `/notifications`，确认 outbox 日志可见；未配置 Resend 时“处理队列”显示稳定失败而不泄露 key。
4. 验证 40101、40301、42201 的用户可见错误不含 token、密码、数据库文本或 Resend 密钥。
5. 以 `zh-CN`、`ja-JP`、`en-US` 各走一次 M4 页面可达性和主要按钮检查，记录缺失翻译。
6. 复核上一轮风险：bcrypt 高负载超时与 Vite 深链接 404。分别记录可重复条件、影响与建议；不得把开发服务器 deep-link 行为误判为 vue-router 客户端守卫失败。

## 交付物

在独立输出目录写 `REGRESSION_REPORT.md`、`results.json`、截图/trace。报告必须列出环境、通过/失败、风险分级、复现步骤、未测项。停止临时进程并保留隔离数据；不提交任何文件。
