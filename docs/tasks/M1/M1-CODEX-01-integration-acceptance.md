# M1-CODEX-01：集成与验收

Codex/PM负责根目录工具锁、跨平台命令入口、CI、证据与共享文档。仅在GLM/Kimi工单通过独立审查后集成。

## 验收清单

- OpenSpec strict 通过，任务无范围漂移。
- 后端fmt/vet/test/build及migration up/down/up通过。
- 前端lint/typecheck/unit/build通过。
- Windows 10 ja-JP和Ubuntu CI通过；UTF-8/key parity通过。
- 扫描确认没有认证/财务/通知/凭证/结款空壳。
- 依赖均固定版本且有理由，秘密扫描无发现。
- Evidence保存至`docs/acceptance/evidence/M1/`。
- 独立Reviewer批准后，Codex统一更新tasks、追踪矩阵、路线图、状态和风险。
