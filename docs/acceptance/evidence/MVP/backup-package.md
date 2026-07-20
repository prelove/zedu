# 可携带备份包 — 后端证据

## 范围

实现 staging→SQLite/上传附件/配置摘要→manifest SHA-256→原子发布的 Owner 备份包。

## 实现文件

- `backend/internal/app/backup/package.go` — `CreatePackage`、manifest 生成与校验
- `backend/internal/app/backup/verify.go` — `VerifyPackage`（恢复到新临时目录）
- `backend/internal/app/backup/handler.go` — `POST /system/backups`（Owner only）
- `backend/cmd/zedu-backup-verify/main.go` — 独立 verify CLI

## 流程

1. 在 `backupDir/.tmp/` 下创建私有 staging 目录（随机 nonce 命名）
2. `VACUUM INTO` 将 SQLite 快照写入 `staging/zedu.db`
3. 复制 `dataRoot/uploads/` 到 `staging/uploads/`（保留相对路径）
4. 写 `config-summary.json`（`formatVersion`、`database`、`attachments`，无 secrets）
5. 遍历 staging 生成 `manifest.json`（每文件 SHA-256 + size）
6. `verifyManifest` 重新哈希校验
7. `os.Rename` 原子发布到 `backupDir/<name>`
8. 失败时 `defer` 清理 staging 和 published 目录

## 命名

`zedu-<UTC ISO 8601>-<6 字节 hex nonce>`，同秒不冲突。

## 测试

`backend/internal/app/backup/package_test.go`、`handler_test.go`：

| 测试 | 验证 |
|---|---|
| `TestBackupPackageContainsDBAttachmentsManifest` | 包含 `zedu.db`、`uploads/`、`manifest.json`；manifest SHA-256 匹配 |
| `TestOwnerBackupCreatesAuditedArtifact` | Owner 成功创建并写入审计日志 |
| `TestBackupPackageOperatorForbidden` | Operator → 40301 |
| `TestBackupPackageFailureCleansStaging` | 故障时 staging 和 published 目录被清理，无成功审计 |
| `TestBackupPackageSameSecondNoConflict` | 同秒创建两个包不冲突（nonce 区分） |
| `TestBackupVerifyValidatesManifest` | 篡改包校验失败 |

## 运行结果

```
go test ./internal/app/backup/... -v
=== RUN   TestOwnerBackupCreatesAuditedArtifact --- PASS
=== RUN   TestBackupPackageContainsDBAttachmentsManifest --- PASS
=== RUN   TestBackupPackageOperatorForbidden --- PASS
=== RUN   TestBackupPackageFailureCleansStaging --- PASS
=== RUN   TestBackupPackageSameSecondNoConflict --- PASS
=== RUN   TestBackupVerifyValidatesManifest --- PASS
PASS
ok  backup  15.328s
```

## 门禁

- 不暴露 HTTP restore：仅 `POST /system/backups` 创建端点
- 无 secrets：`config-summary.json` 仅含 `formatVersion`、`database`、`attachments`
- 原子发布：`os.Rename` 保证发布原子性
