## Why

Zedu 尚无可运行框架，而后续认证、账务、通知和凭证均依赖一致的工程、迁移、质量与编码基础。必须先建立可在 Windows 10 日文环境复现的最小骨架，避免各 AI 工具各自发明结构。

## What Changes

- 建立 Go 后端和 Vue/TypeScript 前端的最小可运行骨架及健康检查。
- 建立 modernc SQLite 连接、增量 migration 和日语模板种子机制。
- 建立 UTF-8/三语基础、结构化日志、request/correlation ID。
- 建立 Windows/Ubuntu CI 的 lint、typecheck、unit、migration、build 和 OpenSpec strict 门禁。

Non-Goals：本 change 不实现认证、业务档案、充值、排课、通知、凭证或正式结款；不创建这些能力的空壳 API。

依赖：M0 治理基线和 `docs/2_prd/Zedu-PRD-Final-v3.1.md`。PRD引用：12、16、17、19、22、23.2、24章。

## Capabilities

### New Capabilities

- `engineering-foundation`: 可运行、可迁移、可构建、可国际化且可在 Win10 JP 验证的工程基础。

### Modified Capabilities

无。

## Impact

新增 `backend/`、`frontend/`、迁移、脚本和 CI；锁定依赖版本。主要风险是脚手架漂移、编码差异和一次性大迁移，均通过版本锁、增量迁移及跨平台门禁控制。
