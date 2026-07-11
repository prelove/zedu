# 需求追踪矩阵

| ID | PRD/Decision | Capability | Scenario | Task | Test | Evidence | Milestone | Status |
|---|---|---|---|---|---|---|---|---|
| Z-M0-001 | ADR-004 | openspec-migration | 旧change全部有处置 | M0-TRACE | T-M0-TRACE | evidence/M0 | M0 | IN_PROGRESS |
| Z-M1-001 | M1 OpenSpec 1.3/3.1 | engineering-foundation | `/healthz` 在后端与代理路径均返回 200 | M1-GLM-01/M1-KIMI-01 | Go health + Vitest health | evidence/M1 | M1 | ACCEPTED |
| Z-M1-002 | M1 OpenSpec 2.1-2.3 | database-foundation | 迁移往返、PRAGMA、幂等 seed 和故障回滚 | M1-GLM-01/M1-GLM-02 | Go database tests + Linux race | evidence/M1 | M1 | ACCEPTED |
| Z-M1-003 | M1 OpenSpec 3.2/4.1 | i18n-quality-gates | 中日英 key 对齐、CJK/emoji、Windows/Ubuntu 门禁 | M1-KIMI-01/M1-CODEX-01 | Vitest + GitHub Actions 29153829469 | evidence/M1 | M1 | ACCEPTED |
| Z-M3-001 | ADR-002 | payment-evidence | 越权访问被拒绝 | 待新change | TS-M3-01 | evidence/M3 | M3 | BACKLOG |
| Z-M4-001 | ADR-002/005 | resend-notification | 邮件失败不回滚排课 | 待新change | TS-M4A-01 | evidence/M4a | M4a | BACKLOG |
| Z-M4-002 | ADR-005 | resend-notification | 接受不等于送达 | 待新change | TS-M4B-04 | evidence/M4b | M4b | BACKLOG |
| Z-M5-001 | PRD 9.5 | attendance-accounting | 任一写入失败全部回滚 | 待新change | TS-M5-01 | evidence/M5 | M5 | BACKLOG |
| Z-M5-002 | ADR-003 | attendance-accounting | MVP无正式结款入口 | 待新change | TS-M6-02 | evidence/M6 | M5/M6 | BACKLOG |
