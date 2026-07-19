# M5：课后确认与不可变课次账务事实

M5 在 M3 的充值/学生流水与 M4 lesson 事实基础上，交付 Operator 明确确认后的 attendance、学生扣课/扣费与老师应付快照。确认是唯一触发财务事实的入口，绝不自动完成。

范围：出勤分类建议字典、attendance 与 lesson_finance、`POST /lessons/{id}/confirm`、单事务账务写入、只读确认页面。

非目标：正式结款、退款/调整入口、自动确认、通知、报表、备份、移动端。
