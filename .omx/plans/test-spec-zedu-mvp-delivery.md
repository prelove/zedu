# Zedu MVP 测试与验收规格

## 1. 通用完成门禁

每个里程碑必须保存到 `docs/acceptance/evidence/<milestone>/`：命令、工具版本、时间、commit SHA、输出摘要、失败与处置。`commit`、截图或口头声明不能单独证明完成。

通用命令按项目生成后的真实脚本为准，至少覆盖：

```powershell
openspec validate --all --strict
git diff --check
go vet ./...
go test ./... -race -count=1
npm run lint
npm run typecheck
npm run test:unit
npm run test:e2e
npm run build
```

## 2. 分层测试

### Unit

- 金额、课时、状态机、错误码、RBAC、上传校验、i18n fallback、通知幂等/退避/lease、账务规则。
- 新增代码行覆盖率 ≥80%；财务、认证、RBAC、通知状态机和上传安全关键分支 100%。

### Integration

- 真实 SQLite 临时库与 migration、fake Resend HTTP server、临时文件系统。
- up/down/up、外键/唯一键/PRAGMA、事务回滚、并发幂等、通知原子待办、lease 竞争、DB/文件补偿、备份恢复、IDOR。
- 核账差异 0、重复通知 0、测试清理后孤儿文件 0。

### E2E

- Playwright 为默认浏览器工具；computer-use 用于 Win10 原生安装、文件选择器、下载等场景。
- MVP 主链、三语核心旅程、断网/重复提交、越权凭证、通知失败/重放、恢复及无结款入口通过率 100%。

### Observability

- request/correlation ID、actor、error code、通知 event/recipient/provider ID、队列积压、lease 恢复、上传失败/孤儿、事务回滚、备份状态、核账差异。
- `DEAD > 0`、最老 PENDING >15分钟、过期 lease 持续5分钟、孤儿持续一个扫描周期、备份超过24小时均告警。
- 核账差异非0或日志泄露 secret 为 P0；完整邮箱/密码/凭证内容进入日志为 P1。

## 3. 核心 Given/When/Then

### TS-M3-01 越权凭证

Given Operator A 无权访问学生 S002；When 猜测并请求 S002 的凭证；Then 返回批准的 403/404，不泄露实体/文件路径，并记录拒绝审计。

### TS-M3-02 伪装文件

Given 文件名和 Content-Type 为 PDF、magic bytes 为 PE；When 上传；Then 返回 42201，无 READY 记录，临时文件被清理。

### TS-M3-03 DB/文件半成功

Given 文件已 rename 且 DB commit 被故障注入失败；When 上传；Then 文件被补偿删除或登记为可恢复孤儿，不能下载，扫描后孤儿归零。

### TS-M4A-01 邮件不可用不回滚排课

Given Resend worker 停止；When 创建课次 L001；Then lesson 为 SCHEDULED、同库待办存在、API 成功。

### TS-M4B-01 收件人部分成功

Given 学生 provider 接受、老师返回503；When worker 处理；Then 两个 recipient 行分别为 PROVIDER_ACCEPTED 与 FAILED，课次状态不变化。

### TS-M4B-02 双 worker 竞争

Given N001 为 PENDING；When W1/W2 同时 claim；Then 条件更新只有一个影响行数为1，provider 仅收到一次请求。

### TS-M4B-03 改期废止旧提醒

Given 课次版本1存在 PENDING；When 改期生成版本2；Then 版本1为 SUPERSEDED，只有版本2可发送。

### TS-M4B-04 供应商接受不等于送达

Given Resend 返回 message ID 且无 webhook；When worker 成功；Then UI 显示“供应商已接受/送达未知”，不得显示“已送达”。

### TS-M5-01 事务回滚

Given 余额10课时、应扣1、老师应付3000JPY且应付写入故障；When 确认；Then lesson/余额/学生流水/老师应付全部保持原状。

### TS-M5-02 并发确认

Given 未确认课次；When 两请求并发确认；Then 仅一个成功，只扣1课时、只生成一条学生流水和老师应付。

### TS-M6-01 原子恢复

Given 当前可用数据及备份 manifest；When 在临时位置恢复并发现 checksum 不符；Then 原系统不切换且仍可用。校验通过时才原子切换。

### TS-M6-02 无结款入口

Given MVP 构建；When 扫描菜单、路由、OpenAPI、API 直链及 feature flags；Then 可执行结款入口数量为0，绕过尝试失败。

## 4. 三语与 Win10 JP 矩阵

登录/初始化、人员课程、报名排课、充值凭证、出勤流水、工作台、API/UI错误、邮件、日期金额、文件名下载均在 `zh-CN/ja-JP/en-US` 验证；Win10 JP 验证中文/日文/全角/空格路径、UTF-8 文档、Excel/CSV策略和邮件预览。

## 5. MVP Go/No-Go

- OpenSpec strict、lint、typecheck、unit/integration/E2E/build 全部退出码0。
- P0/P1=0；P2 有 owner、期限和不阻断理由。
- 财务核账差异0；DB/凭证恢复 checksum 100%一致。
- 三语矩阵100%；真实 Resend 仅批准测试收件箱 smoke 通过，自动测试使用 fake provider。
- 暂定运营参数已确认或明确接受默认值。
- Product Owner 完整演练并签字。

