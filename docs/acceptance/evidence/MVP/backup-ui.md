# Owner 备份 UI — 前端证据

## 范围

更新 Owner 备份 UI 展示包名/错误态（三语）。

## 实现文件

- `frontend/src/features/dashboard/DashboardView.vue` — 备份创建按钮、包名展示、错误态
- `frontend/src/i18n/locales/{zh-CN,ja-JP,en-US}.ts` — `dashboard.createBackup`、`backupCreating`、`backupCreated`、`backupError`

## i18n 键

| 键 | zh-CN | ja-JP | en-US |
|---|---|---|---|
| `dashboard.createBackup` | 创建备份 | バックアップを作成 | Create backup |
| `dashboard.backupCreating` | 备份中… | バックアップ中… | Creating backup… |
| `dashboard.backupCreated` | 备份文件：{file} | バックアップ：{file} | Backup file: {file} |
| `dashboard.backupError` | 备份创建失败，请稍后再试。 | バックアップ作成に失敗しました。後でもう一度お試しください。 | Backup creation failed. Please try again later. |

## 行为

- Owner 登录后在工作台看到"创建备份"按钮
- 点击后按钮禁用并显示"备份中…"
- 成功后显示"备份文件：{file}"
- 失败后显示错误提示（无敏感信息）
- Operator 不显示按钮（后端 40301 + 前端 RBAC）

## 测试

- `tests/i18n.test.ts`、`tests/i18n-m2.test.ts`、`tests/i18n-m2-kimi02.test.ts` — 三语 key parity PASS
- `tests/negative-scope.test.ts` — 无 `/restore` 路由

## 门禁

- Owner 成功：按钮可见且可点击
- Operator 无入口：前端 RBAC + 后端 40301
- 无 restore 控件：仅创建按钮
- 三语：所有键在 zh-CN/ja-JP/en-US 均存在
