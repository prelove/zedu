# Kimi：M5/M6 浏览器模拟人工验收工单

基线：`main` `8559558`。使用隔离 DB 与临时服务，不修改产品代码或文档、不提交、不使用真实 Resend key。

1. Owner/Operator 创建 lesson 后，从课程页执行课后确认；验证已完成状态不可再次确认。
2. 验证余额不足时页面报稳定错误，刷新后不存在错误 attendance/finance 结果。
3. 验证 `/dashboard` 展示待确认和失败通知计数，且只读无业务写入。
4. Owner 看到并能触发备份；Operator 不可触发；验证没有 restore UI 或路由。
5. zh-CN/ja-JP/en-US 都访问 M5/M6 页面，记录实际缺失翻译（禁止自行修复）。
6. 对 M1–M6 跑一次最小关键路径：登录→初始化/主数据→充值→凭证→排课→通知→确认→dashboard→backup。

交付 `REGRESSION_REPORT.md`、`results.json`、截图/trace、风险分级和未测项；停止临时进程，所有产物保持未提交。
