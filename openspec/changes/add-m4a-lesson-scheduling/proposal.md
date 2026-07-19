## Why

M3 已补齐充值、余额与付款凭证，MVP 主链路目前只缺“把已建档、已报名、已安排老师的学生真正排进课表”这一步。若没有基础排课，运营者无法从报名进入上课前准备，后续通知、课后确认与经营核对也缺少统一的课次事实源。

现在进入 M4a 的原因是：M2/M3 的身份、主数据与账务边界已冻结，可以在不引入通知、副作用解耦、老师结款等更高复杂度能力的前提下，先交付最小可验证的排课闭环，并为 M4b/M5 提供稳定契约。

## What Changes

- 新增基础排课 capability：允许 Owner/Operator 基于有效 enrollment 和 ACTIVE assignment 创建、查看、编辑、取消基础课次。
- 定义 lesson 的最小业务字段与状态机，包括 `SCHEDULED`、`COMPLETED`、`CANCELLED`，并冻结创建/更新时的校验规则。
- 明确课次时间以 UTC 存储、按提交 timezone 解释输入时间，确保 Windows/SQLite 环境下行为一致。
- 规定排课与财务/通知的边界：排课不触发通知发送，不写账务流水，不生成老师结款，不修改 attendance/accounting。
- 为后续 M4b/M5 预留稳定字段和状态约束，但不提前实现冲突检测增强、通知 outbox、自动完课、课后确认或任何报表能力。

## Capabilities

### New Capabilities
- `lesson-scheduling`: 基础课次 CRUD、状态守卫、时区处理和最小列表查询。

### Modified Capabilities
- 无

## Impact

- Affected specs: 新增 `lesson-scheduling`
- Affected backend: `backend/internal/lesson/**`、路由注册、SQLite migration、与 enrollment/assignment 的只读校验接缝
- Affected frontend: `frontend/src/features/lesson/**`、路由/菜单、三语文案、列表与表单校验
- Affected APIs: 新增 `/api/v1/lessons` 及明细/更新/取消相关接口
- Dependencies: 依赖 M2 已交付的认证/RBAC、enrollment、assignment；依赖 M3 冻结的 base currency/余额只读提示语义，但不得写入账务
- Risks:
  - PRD 第 9.4 节与 legacy 005 对 meeting link、duration、timezone 有明确约束，若字段冻结不完整会阻塞后续 M4b/M5
  - 若把通知、冲突检测增强或老师结款带入本 change，会破坏 MVP 边界并扩大测试面
  - 时区与 Windows 本地环境处理错误会直接影响后续提醒窗口和课后确认
- PRD refs: 8、9.4、10.7、13.8、15、24.3
- Non-Goals:
  - 不实现 Resend 通知、outbox-lite、送达状态或任何消息模板
  - 不实现老师结款、老师应付、经营报表、备份恢复或移动端 today 流程
  - 不实现冲突检测增强、批量换老师联动、自动排课或 AI 推荐
  - 不实现 attendance、课消、余额扣减、suggested_* 结算字段或任何账务副作用
