# 旧 OpenSpec 迁移映射

| 旧change | 新能力 | PRD章节 | 处置 |
|---|---|---|---|
| 001 | engineering-foundation/schema-baseline | 12,16-20,23 | SPLIT |
| 002 | auth-init | 3,5.2,7.1,13,15 | MIGRATE |
| 003 | people-directory/course-catalog | 3,5,10,14 | SPLIT |
| 004 | enrollment-assignment | 6,9.2-9.3,15 | MIGRATE |
| 005 | lesson-scheduling | 8,9.4,10,15 | MIGRATE |
| 006 | attendance-accounting | 8,9.5,14,15,24.3 | REWRITE |
| 007 | payment-ledger/payment-evidence | 7,9.6,10,14,15 | SPLIT |
| 008 | dashboard/verified-backup | 10,20,23.2 | SPLIT |
| 009 | lesson-enhancement/v1-payout | 9.3-9.4,9.7,15 | SPLIT_DEFER |
| 010 | resend-notification | 1.2,3,9.4,21,24 | REWRITE |
| 011 | payment-evidence/v1-dictionary | 7.3,8.3,10,14,20 | SPLIT |
| 012 | v1-reporting/v1-data-io | 10,12,13,23.3 | DEFER |
| 013 | v1-mobile-packaging | 11,18-20,22 | DEFER |
| 014 | release-gates/v1-migration | 24-26 | SPLIT_MANUAL_GATE |

细粒度 Requirement/Scenario/Test 映射将在各新 change 创建时进入 `requirements-matrix.md`，CI 以 orphan=0 为门禁。
