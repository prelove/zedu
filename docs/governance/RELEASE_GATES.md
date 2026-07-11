# 发布与里程碑门禁

每个里程碑必须运行 OpenSpec strict、格式/静态检查、相关 unit/integration/E2E，并把输出保存至 `docs/acceptance/evidence/<M>/`。

## MVP Go/No-Go

- OpenSpec、lint、typecheck、测试、构建全部退出码0。
- P0/P1为0；核账差异0；重复通知0；孤儿凭证0。
- DB/附件恢复 SHA-256 一致率100%。
- zh-CN/ja-JP/en-US 与 Win10 JP 核心矩阵100%。
- UI/API/route/feature flag 中正式结款入口为0。
- 暂定运营参数已确认或由 Product Owner 明确接受。
- Product Owner 完成完整业务演练并签字。
