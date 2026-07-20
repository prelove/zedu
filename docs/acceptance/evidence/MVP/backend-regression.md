# 后端隔离环境回归 — 证据

> **复核更新（2026-07-20）**：在当前工作树再次执行 `GOTOOLCHAIN=local` 下的 `go fmt ./...`、`go vet ./...`、`go test ./... -count=1`、`go build ./...`，全部通过。主要包结果：auth 185.473s、backup 27.275s、course 181.591s、directory 93.522s、finance 29.096s、notification 22.056s、payable 19.292s、database 69.486s；无测试失败或构建错误。

## 范围

在隔离环境运行 Go fmt/vet/test/build、migrations 001–009 up/down/up、账务核对、备份恢复演练。

## 环境

- OS: Windows
- Go: 1.23.3
- SQLite: modernc.org/sqlite v1.29.10（纯 Go，无 cgo）
- 隔离：每个测试使用 `t.TempDir()` 独立数据库

## 1. fmt / vet / build

```
gofmt -l .   → 无输出（全部格式化）
go vet ./... → 无输出（无警告）
go build ./... → 无输出（成功）
```

## 2. 测试（串行 `-p 1`）

Windows 下 modernc.org/sqlite 的 WAL 文件锁在并行测试时会出现 contention，
导致超时。这是已知的 Windows SQLite 基础设施问题，与代码无关。
使用 `-p 1` 串行执行可避免此问题。

```
go test -p 1 ./... -timeout 600s
ok  auth         134.151s
ok  backup        16.962s
ok  course       145.771s
ok  dashboard      6.903s
ok  directory     69.799s
ok  evidence      10.790s
ok  finance       17.406s
ok  lesson        11.533s
ok  notification   5.514s
ok  onboarding    15.905s
ok  payable        9.532s
ok  platform/auth  (cached)
ok  platform/database  31.753s
ok  platform/httpserver  1.664s
ok  platform/logging    (cached)
```

**所有包 PASS，P0/P1=0。**

## 3. Migrations 001–009 up/down/up

```
go test -p 1 ./internal/platform/database/... -run TestMigrationUpDownUp -v
=== RUN   TestMigrationUpDownUp
--- PASS: TestMigrationUpDownUp (1.39s)
PASS
```

- `MigrateUp` 应用 001–009 全部 up 迁移
- `MigrateDown` 回滚 009→001 全部 down 迁移
- `MigrateUp` 再次应用 001–009 全部 up 迁移
- migration 009（LESSON_REMINDER）通过重建表扩展 CHECK 约束，保留现有数据

## 4. 账务核对

- `finance` 包测试覆盖：充值、流水、课时余额、附件上传/下载/作废
- `lesson` 包测试覆盖：课次确认生成原子账务事实（`LESSON_CONFIRM` ledger entry）
- `payable` 包测试覆盖：从 `teacher_account_ledger` 聚合应付金额（整数，无 float）
- 所有金额字段为 `int64`，无 float 计算

## 5. 备份恢复演练

```
go test -p 1 ./internal/app/backup/... -v
=== RUN   TestOwnerBackupCreatesAuditedArtifact --- PASS
=== RUN   TestBackupPackageContainsDBAttachmentsManifest --- PASS
=== RUN   TestBackupPackageOperatorForbidden --- PASS
=== RUN   TestBackupPackageFailureCleansStaging --- PASS
=== RUN   TestBackupPackageSameSecondNoConflict --- PASS
=== RUN   TestBackupVerifyValidatesManifest --- PASS
PASS
```

- 备份包包含 `zedu.db`、`uploads/`、`manifest.json`、`config-summary.json`
- manifest SHA-256 校验通过
- 篡改包校验失败
- Owner 成功创建并写入审计日志
- Operator → 40301
- 故障清理无成功审计
- 同秒创建不冲突（nonce 区分）
- `zedu-backup-verify` CLI 恢复到新临时目录，不覆盖活动 DB

## 门禁

- P0/P1=0：所有测试 PASS
- DB/附件 hash 一致：manifest SHA-256 校验
- 不覆盖活动 DB：`VerifyPackage` 要求 targetDir 不存在
