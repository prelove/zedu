# Zedu MVP 交付 PRD（执行基准）

## 1. 目标

在完整路线图持续可追踪的前提下，先交付并验收可运营 MVP，再进入 V1。唯一业务事实源为 `docs/2_prd/Zedu-PRD-Final-v3.1.md`；本文件记录经用户批准的阶段性范围修订。

## 2. 已批准范围修订

- MVP 新增：Resend 邮件通知、付款凭证上传。
- MVP 不含：正式老师结款。MVP 的 UI、路由、API、权限绕过入口和隐藏开关均不得提供可执行结款能力。
- “7天 MVP”是进度目标而非质量承诺；不得跳过财务、安全、授权、恢复或验收门禁换取日期。

## 3. MVP 能力闭环

`初始化/登录 → 人员与课程 → 报名与师生安排 → 充值与凭证 → 排课 → 邮件提醒 → 课后确认 → 学生流水/老师应付 → 工作台 → 备份恢复`

### M0 治理与规格迁移

- OpenSpec 固定至官方 1.6.x，按实际 `--help` 编写命令并在 CI 严格校验。
- Superpowers 固定官方版本及各 AI harness 的可复现安装方式。
- 冻结旧 001-014 为只读迁移输入，保存 SHA-256 清单。
- 建立 `旧change → 新能力 → PRD → Requirement → Scenario → Task → Test → Evidence` 零孤儿矩阵。
- 完成 Claude 文档评审、编码/测试/验收/安全/i18n/AI协作规范和总路线图。

### M1 工程与质量基线

- Go/Gin/GORM/SQLite 与 Vue/Vite/TypeScript 工程。
- migration up/down/up、日语模板、CI、结构化日志、i18n 基础设施。
- UTF-8、LF、Windows 10 日文环境及中文/日文路径验证。

### M2 认证、初始化、人员课程与报名安排

- Owner/Operator 登录、锁定、刷新轮换、强制改密、RBAC、审计。
- 学生/家长/老师、四层课程、报名、师生安排。
- 暂停老师不进入新安排候选，但历史引用可见。

### M3 充值与付款凭证

- 充值、作废冲正、学生余额/流水、凭证上传/查看/下载/替换/删除。
- 对象级授权与 IDOR 防护；扩展名、Content-Type、magic bytes 一致性校验。
- 限制大小/数量；随机存储名；非 Web 根目录；安全下载响应。
- 临时文件、fsync、原子 rename、DB 状态与显式补偿；孤儿扫描。
- 备份 manifest 包含相对路径、大小、SHA-256，DB 与文件双向核对。

### M4a 排课

- 创建、修改、取消、冲突提示、时区和状态机。
- 排课事务仅写核心数据及同库通知待办；不在事务中调用 Resend。
- 邮件不可用不得回滚已成功排课。

### M4b Resend 通知

- MVP：课次创建/变更/取消和批准的课前提醒、模板、通知日志、重试、人工重放、失败可见。
- 不含：SMTP fallback、晨报、周报、结款通知、通用消息总线。
- `notification_log` 每个 recipient 一行。
- 状态采用 `PENDING/SENDING/PROVIDER_ACCEPTED/FAILED/DEAD/SUPERSEDED`。
- `PROVIDER_ACCEPTED` 只表示 Resend API 接受，不表示实际送达；MVP 不接 webhook 时送达状态未知。
- DB UNIQUE 幂等约束至少覆盖事件类型、课次、排期版本、recipient、模板及 locale。
- 改期/取消在事务内将旧 PENDING 记录标记 SUPERSEDED。
- SQLite 条件 UPDATE + 事务原子领取 lease；MVP 单 worker，双实例竞争测试只允许一个成功。
- 429/5xx/超时退避；永久错误和达到上限进入 DEAD；管理员重放必须有理由、审计及新代次。
- `RESEND_API_KEY` 仅来自环境/secret store，禁止进入 DB、UI、日志、错误响应、导出和备份；dev/test/prod 隔离。

### M5 课后确认与账务闭环

- 出勤确认、学生扣减/流水、老师应付、课次状态、幂等、并发、回滚、核账。
- 老师应付只是已发生义务，不是结款。
- 多表写入同一 SQLite 事务；金额禁止 float；历史事实不可覆盖。

### M6 工作台、恢复与 MVP 验收

- 今日课次、待确认、余额预警、通知失败。
- DB+uploads 完整备份；在临时目录/临时 DB 恢复并校验后原子切换，失败保持当前状态。
- 完整 E2E、Win10 JP、三语、操作手册、恢复手册和 Product Owner 签字。

## 4. 国际化范围

- locale：`zh-CN`、`ja-JP`、`en-US`。
- MVP 核心页面、校验错误、邮件主题/正文、日期时间、金额、文件名与下载必须覆盖三语。
- API 使用稳定错误码，UI 本地化；缺 key 在 CI 失败，生产 fallback 不得显示空字符串。
- 时间存 UTC，显示默认 Asia/Tokyo。

## 5. 决策门禁

代理可自主决定技术与文档治理；产品范围、财务语义、角色权限、隐私、持续费用、生产部署和真实迁移必须人工确认。

所有暂定运营参数进入 Decision Register，记录 owner、依据、`provisional` 状态、最晚确认里程碑和影响。上线前未决项必须归零或由 Product Owner 明确接受默认值。

## 6. V1 与后续

- V1：正式老师结款、完整通知自动化、配置字典、报表/数据IO、移动/打包、真实迁移及并行运行。
- V1.5/V2：按正式 PRD 的触发条件管理，不在 MVP 创建空壳。
