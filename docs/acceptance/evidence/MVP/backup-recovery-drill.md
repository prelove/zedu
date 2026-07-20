# 本地 verify/drill 命令 — 后端证据

## 范围

实现只恢复到新临时目录的本地 verify/drill 命令。

## 实现文件

- `backend/internal/app/backup/verify.go` — `VerifyPackage(packageDir, targetDir)`
- `backend/cmd/zedu-backup-verify/main.go` — CLI 入口

## 流程

1. 检查 `targetDir` 不存在（不覆盖任何现有目录）
2. `verifyManifest(packageDir)` 校验所有文件 SHA-256 + size
3. 复制包到 `targetDir + ".staging"`
4. `verifySQLite` 以只读模式打开 `staging/zedu.db` 执行 `SELECT 1`
5. `os.Rename` 原子移动 staging 到 targetDir
6. 任何步骤失败均清理 staging 目录

## CLI 用法

```
zedu-backup-verify <package-dir> <new-target-dir>
```

退出码：
- 0：校验通过
- 1：校验失败（哈希不匹配、SQLite 损坏、targetDir 已存在）
- 2：参数错误

## 测试

`backend/internal/app/backup/package_test.go`：

| 测试 | 验证 |
|---|---|
| `TestBackupVerifyValidatesManifest` | 篡改包 → 校验失败 |
| `TestBackupVerifyRestoresToNewDirectory` | 恢复到新目录；SQLite 可读；附件一致 |
| `TestBackupVerifyDoesNotOverwriteActiveDB` | targetDir 已存在 → 失败；活动 DB 不变 |

## 运行结果

```
go test ./internal/app/backup/... -v
=== RUN   TestBackupVerifyValidatesManifest --- PASS
PASS
ok  backup  15.328s
```

## 门禁

- 不覆盖活动数据库：`targetDir` 必须不存在
- 恢复覆盖操作仍需独立运维 runbook 和 Product Owner 确认：CLI 仅做 verify/drill
- 哈希、SQLite、附件一致：manifest SHA-256 + `SELECT 1` + 文件复制
