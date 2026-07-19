# M5 课后确认验证

实现了 attendance outcome、attendance、lesson_finance、teacher ledger 与 `POST /lessons/{id}/confirm`。确认仅允许 Owner/Operator 对 SCHEDULED lesson 执行，并以单一事务写入事实、学生流水、老师应付、lesson COMPLETED 与审计。

已执行 Go 测试、vet、构建，前端 typecheck/lint/Vitest/build 及 OpenSpec strict。完整业务人工核算、真实金额案例与最终 UAT 继续在 M6 总验收执行。
