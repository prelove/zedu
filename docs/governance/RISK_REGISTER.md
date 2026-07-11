# 风险登记表

| ID | 风险 | 等级 | 预警 | 缓解/验收 | Owner | 状态 |
|---|---|---|---|---|---|---|
| R-001 | 迁移遗漏旧需求 | 高 | orphan>0 | 全量映射、hash、陌生AI抽样 | PM/Architect | OPEN |
| R-002 | 课后确认产生部分账务 | 极高 | 核账差异非0 | 单事务、幂等、全写点故障注入 | Backend/QA | OPEN |
| R-003 | 重复/遗漏邮件 | 高 | duplicate>0或pending>15m | DB唯一键、lease、重试/崩溃演练 | Backend | OPEN |
| R-004 | 凭证越权/孤儿/恶意内容 | 高 | IDOR或orphan>0 | 对象授权、内容检测、补偿、恢复 | Security | OPEN |
| R-005 | Win10 JP乱码 | 中 | U+FFFD或缺key | UTF-8/三语矩阵/日文路径测试 | Frontend/QA | OPEN |
| R-006 | 备份存在但不可恢复 | 极高 | checksum差异 | 临时恢复、双向核对、原子切换 | Release | OPEN |
| R-007 | MVP被V1结款拖偏 | 中 | payout入口>0 | UI/API/route负向断言 | PM | OPEN |
| R-008 | Superpowers Gemini基线未证实 | 低 | 安装/行为不一致 | v6.1.1不宣称Gemini支持，单独验证 | PM | OPEN |
