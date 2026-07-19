## Context

M4a 位于 M2 主数据与 M3 财务闭环之后、M4b 通知之前。当前系统已经具备：

- Owner/Operator 认证与 RBAC
- 学生、课程、报名项目与老师安排主数据
- 充值、余额、付款凭证等财务事实

缺口是：系统仍没有一个稳定的 `lesson` 事实源，运营人员无法把有效 enrollment 与 ACTIVE assignment 组合成可追踪课次。PRD 9.4、10.7、13.8 和 legacy change 005 已给出基础排课约束，但旧文档把后续通知/增强场景混杂在一起，需要在 M4a 重新收敛。

约束：

- 仅交付基础排课，不把通知、冲突检测增强、老师结款、attendance/accounting 带入本 change
- 兼容 Windows 10 日文环境与 SQLite/modernc 运行时
- 需要给 M4b 通知和 M5 课后确认提供稳定 lesson 契约

## Goals / Non-Goals

**Goals:**

- 定义 `lesson` 的最小数据模型、状态机和 API 契约
- 允许 Owner/Operator 创建、列表、查看、编辑、取消课次
- 保证创建与变更时对 enrollment/assignment/status/timezone 的校验一致且可测试
- 明确排课不产生通知与账务副作用
- 为后续 attendance、notification 预留只读/扩展接缝

**Non-Goals:**

- 不实现 Resend outbox、提醒窗口、送达状态、消息模板
- 不实现时间冲突检测、软警告条、批量换老师联动或自动排课
- 不实现课后确认、扣费、suggested_*、老师应付或报表
- 不实现移动端 today、日历高级视图或 dashboard 聚合

## Decisions

### 1. lesson 作为独立聚合，创建时冻结最小排课事实

`lesson` 单独建表，持有：

- `id`, `lesson_no`
- `enrollment_id`, `assignment_id`, `teacher_id`, `student_id`
- `scheduled_start_at`, `scheduled_end_at`, `duration_min`, `timezone`
- `meeting_type`, `meeting_link`, `lesson_topic`, `note`
- `status`, `cancel_reason`
- 审计字段

创建时从 enrollment/assignment 读取并冻结 `teacher_id` / `student_id`，避免后续 assignment 变化影响历史 lesson 事实。

Rejected:

- 仅保存 `assignment_id`，查询时再 join 当前 teacher/student | 会让历史课次随 assignment 替换而漂移
- 在 M4a 同时引入 attendance 字段 | 会把 M5 范围提前带入

### 2. 创建/更新/取消统一在单事务内完成

每次写入使用单数据库事务，顺序为：

1. 校验调用者角色
2. 读取 enrollment、assignment 当前状态
3. 校验业务规则
4. 生成/更新 lesson
5. 写 operation/audit 日志

本 change 不调用通知、账务或外部系统，因此失败补偿以事务回滚为主，不需要 outbox。

Rejected:

- 先查后写、无事务 | 状态竞争下可能把已结束 assignment 或终态 enrollment 写成课次
- 创建成功后异步补写审计 | 审计会与业务事实脱节

### 3. 课次状态机先收敛为三态

M4a 只允许：

- `SCHEDULED`：创建后的默认状态，可编辑、可取消
- `COMPLETED`：仅允许由更新接口显式标记，为 M5 过渡接缝
- `CANCELLED`：由取消接口进入，不可再编辑

M4a 不引入 `REMINDED`、自动完课或缺席分类状态。

Rejected:

- 提前引入 `REMINDED` | 依赖 M4b outbox，与本里程碑耦合
- 只做创建/删除，不做取消状态 | 会破坏未来通知和审计链路

### 4. 时间输入按本地 timezone 解释，库存储统一 UTC

接口接收业务本地时间与 `timezone`，服务层在写库前统一换算为 UTC，返回时同时保留 `timezone` 与 UTC 时间戳。这样可以保证：

- Windows/SQLite 一致
- 后续通知窗口计算有稳定基准
- 日本时区等场景可做确定性测试

Rejected:

- 直接存本地时间字符串 | 后续提醒窗口、跨环境测试会失真
- 强制只支持 Asia/Tokyo | 与产品三语/多地区定位不符

### 5. meeting_link 与输入约束在 M4a 冻结

校验规则在创建和更新时一致执行：

- `duration_min` MUST 在批准范围内
- `meeting_type=WECHAT` 时 `meeting_link` MUST 为合法 URL
- enrollment 为 `COMPLETED/CANCELLED` 时 MUST 拒绝创建
- assignment 非 ACTIVE 或已结束时 MUST 拒绝创建

这样可以提前锁定前后端错误码与表单行为。

Rejected:

- 依赖前端校验，后端放行 | 不能满足真实 HTTP focused suite
- 现在引入复杂会议类型词典 | 超出 M4a 范围

### 6. 安全与可观测性优先采用现有模式

- 仅 `OWNER` / `OPERATOR` 可访问 lesson 写接口
- 所有写操作进入现有审计/operation log
- 日志包含 lesson id / lesson_no / enrollment id / actor id，不记录敏感外链明文以外的额外隐私

Rejected:

- 为 lesson 单独新增权限系统 | 与 M2 RBAC 重复
- 不记录取消原因 | 不利于后续通知与经营回溯

## Risks / Trade-offs

- [状态竞争] enrollment 或 assignment 在并发下刚好变为终态 → 使用单事务重读并以数据库当前状态为准
- [时区误差] 本地时间转 UTC 处理不一致 → 增加 Asia/Tokyo 确定性测试，统一服务层转换
- [范围蔓延] 冲突检测/通知/attendance 被顺手带入 → proposal/spec/tasks 明确 Non-Goals，验收加负向断言
- [历史漂移] assignment 更换老师后旧课次被“跟着变” → 创建时冻结 teacher_id/student_id
- [可用性取舍] M4a 不做冲突检测会保留人工误排风险 → 接受该风险，后续在增强 change 处理，不阻塞 MVP 主链路

