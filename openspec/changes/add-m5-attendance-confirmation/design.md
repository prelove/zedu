# M5 设计

确认仅允许 Owner/Operator 对 `SCHEDULED` lesson 执行一次。一个事务写入 attendance（建议值和实际值快照）、学生 ledger、老师应付事实与 lesson 状态；任一失败全部回滚。数据库 `UNIQUE(lesson_id)` 作为并发最终防线。

金额使用现有整数本位币约定；课时使用明确的 decimal text/整数最小单位，禁止 float。实际扣课、收费和老师课酬由 Operator 明确提交，建议值只用于表单默认值。

M5 不实现结款；老师应付只作为不可变课次事实保存，M6/V1 再聚合展示或结算。
