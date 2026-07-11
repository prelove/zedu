# Zedu AI 协作规范

所有 AI 工具在修改项目前必须依次阅读：

1. `docs/status/PROJECT_STATUS.md`
2. `docs/roadmap/MASTER_ROADMAP.md`
3. `docs/governance/GOVERNANCE.md`
4. 当前 OpenSpec change 的 proposal/spec/design/tasks
5. 相关标准与 PRD 章节

只允许认领 `READY` 任务。实现必须保持 PRD→Scenario→Test→Evidence 追踪；先失败测试后实现；不得自行扩大范围、降低验收或创建未批准错误码/依赖。正式老师结款不属于 MVP。涉及产品范围、财务语义、权限、隐私、持续费用、生产部署或真实数据时停止并升级给 Product Owner。

仓库统一 UTF-8；产品 locale 为 zh-CN/ja-JP/en-US；所有金额禁止 float；账务多表写入必须同一事务。完成后更新项目状态、路线图、追踪矩阵和证据。
