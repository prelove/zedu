# M2-KIMI-01 验收证据

验收日期：2026-07-17
验收结论：ACCEPTED

## 范围与提交

- 实现提交：`702e680`（登录、内存会话、refresh-once、受保护路由、Owner 初始化界面、三语）。
- 验收局部收口：恢复登录页全局语言切换；Operator 被 `/onboarding` 路由拒绝后重定向首页并显示本地化提示；可变更的 Node 冒烟脚本改为显式 disposable-environment 与变更确认门禁。
- 无后端、数据库、迁移、OpenSpec 或禁止业务页面改动；唯一新增依赖为冻结的 `vue-router@5.1.0`。

## 独立验证

- 前端门禁：`npm run lint`、`npm run typecheck`、`npm run test:unit`、`npm run build` 已执行；定向路由/首页回归为 2 个文件、8 个测试通过。
- 实际浏览器：Playwright CLI 访问 `http://localhost:5173/login`，确认用户名与密码的可访问 label、禁用提交按钮，以及 `zh-CN` 与 `ja-JP` 语言切换后的页面文本。
- OpenSpec：`openspec validate add-m2-core-management --strict` 已在前一进度更新中通过；本任务未改 OpenSpec 工件。

## 验证边界

- 完整的真实后端浏览器登录、cookie 自动管理、初始化和 reset 流在隔离数据库中由 M2-CODEX-02 执行；本机未以默认开发数据运行可变更冒烟脚本。
- Windows 本机未执行 Linux `-race`；该项继续由 CI/集成门禁负责。
