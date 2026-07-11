# Zedu 轻量级教培教务管理系统
## 产品需求文档（PRD）— 完整实装版

> **文档版本**：v2.0（实装基准版）
> **状态**：可作为开发基准
> **整理日期**：2026-07-03
> **项目代号**：Zedu（Zero-friction Education）
> **上一版本**：v1.0（业务与技术架构讨论稿）

---

## 文档说明

本文档在 v1.0 基础上进一步细化，目标是**可以直接作为原型设计、ER 建库、API 开发、前端开发、测试用例编写的基准文档**，而不再需要额外的中间设计文档。

相比 v1.0，本版新增或大幅扩充的内容：

- 完整可执行的数据库 DDL（SQLite 版本，含索引、约束）
- 全量 API 接口清单（方法/路径/入参/出参/错误码，逐个列出）
- 业务错误码总表
- 字段级校验规则总表
- 页面级 UI 规格说明（逐页面列出字段、按钮、状态、交互）
- 通知邮件模板原文（中/日双语示例）
- 状态机图（课次状态、报名状态、充值状态等）
- 非功能性需求（性能、可用性、国际化、可访问性）
- 测试用例检查清单
- 更细粒度的开发任务拆分（可直接建 Issue/Ticket）
- 术语表
- 风险登记表（含负责人与应对措施占位）

**使用建议**：本文档章节较多，建议按角色分工阅读——后端开发重点看第六、七、十章；前端开发重点看第五、九章；产品/运营重点看第一至四章、第八章；测试重点看第十三章；部署运维重点看第十一、十二章。

---

## 目录

**第一部分：产品与业务**
1. 产品定位与目标
2. 用户角色与权限体系
3. 核心业务流程（含异常分支）
4. 课程体系设计
5. 功能模块与页面级 UI 规格

**第二部分：数据与接口**
6. 数据模型设计（完整 DDL）
7. API 接口规范（全量清单）
8. 业务规则与字段校验总表
9. 状态机设计

**第三部分：系统与工程**
10. 后端技术架构
11. 前端技术架构
12. 部署与发布方案
13. 数据安全与备份
14. 通知与提醒系统（含模板原文）
15. 非功能性需求

**第四部分：项目管理**
16. 开发计划与任务拆分
17. 测试计划与用例清单
18. 风险登记表
19. 版本演进规划

**附录**
A. 术语表
B. 技术选型速查
C. 配置文件完整示例
D. 下一步行动清单

---

# 第一部分：产品与业务

## 第一章 产品定位与目标

### 1.1 背景

某日语教育从业者以兼职形式运营日语学习撮合服务：联络学习者（学生）与老师，组织一对一或小班课程，向学生收费，向老师支付课酬，赚取差价作为运营收益。目前使用 Excel + 人工管理，核心痛点：

- 上课提醒容易遗漏（无自动化机制，全靠人工记忆）
- 续费跟进被动（余额不足无预警，靠定期翻表格）
- 费用核算繁琐（充值、扣课、课酬、差价手工计算，易出错）
- 调课沟通散乱（时间变动靠逐一手动通知）
- 信息分散（学生档案、上课记录、收支数据分散在多处）

Zedu 由此需求出发，定位为**可复制部署的轻量级教培教务管理系统**。

### 1.2 产品名称与品牌

```
系统名称：Zedu
命名含义：Zero-friction Education（零摩擦教育管理）
中文展示名：可自定义（如"泽度教务"），各部署实例可独立设置
```

### 1.3 产品定位声明

> Zedu 是一套小而美的轻量级教培教务管理系统，追求"漂亮、轻量、好部署、账务清楚、提醒可靠、可复制"，而非重型 SaaS 平台或复杂 ERP。首期面向日语一对一/小班课程管理场景，架构预留扩展至其他语言、体育、艺术、职业培训等场景的能力。

### 1.4 V1 范围边界（Scope）

**明确包含（In Scope）：**

| # | 能力 | 说明 |
|---|------|------|
| 1 | 运营者后台管理 | Owner + Operator 双角色，PC/Pad/Mobile 自适应 |
| 2 | 学生档案管理 | 含家长信息，支持 Excel 批量导入 |
| 3 | 老师档案管理 | 含能力标签、可授时间、课酬设置 |
| 4 | 课程体系维护 | 领域/方向/等级/能力标签四层结构 |
| 5 | 学生课程报名 | 支持一个学生多个课程方向并行 |
| 6 | 排课管理 | 日历视图 + 列表视图，多种上课方式 |
| 7 | 课后确认 | 出勤确认，触发账务事务 |
| 8 | 充值管理 | 多币种，自动折算 JPY |
| 9 | 结款管理 | 老师课酬结算 |
| 10 | 账务流水 | 学生/老师双向流水，单课财务快照 |
| 11 | 自动提醒 | 邮件通道，六类定时任务 |
| 12 | 数据图表 | ECharts 工作台图表 |
| 13 | 报表导出 | Excel/PDF |
| 14 | 数据导入 | Excel 批量导入学生/老师 |
| 15 | 数据备份 | 手动+自动+云端流式备份 |
| 16 | All-in-One 部署 | 单文件发布，四种部署模式 |

**明确不包含（Out of Scope，V1 不做）：**

```
多租户 SaaS 架构
微服务架构拆分
学生端 / 老师端 / 家长端独立登录系统
微信小程序（审核周期长，海外适用性存疑）
在线支付接口对接（PayPay/Stripe/微信支付 API）
复杂自动排课算法（AI 匹配、冲突自动求解）
Flutter / React Native 原生 App
真正的离线数据同步
复杂多级班级管理（分班、排位、竞赛管理）
课程内容管理（教材、题库、作业系统）
在线视频会议内嵌（V1 仅提供第三方会议链接）
```

### 1.5 目标用户与规模

| 维度 | 目标值 |
|------|--------|
| 目标用户画像 | 兼职/小型教培运营者，管理 1 人或极小团队 |
| 学生规模 | ≤ 1,000 人（活跃 + 历史累计）|
| 老师规模 | ≤ 100 人 |
| 运营账号数 | 1 Owner + 2~5 Operator（预留）|
| 并发在线用户 | ≤ 10（纯后台系统，无面向公众访问）|
| 日均课次数 | 预估 10~50 节 |
| 数据保留期 | 无限期（历史记录不物理删除）|

### 1.6 部署与多实例策略

V1 不做多租户共享架构，采用**"一客一部署"（Single-tenant-per-instance）**模式：

```
zedu.abitcloud.org       → 机构 A（当前主实例）
friend1.abitcloud.org    → 机构 B（复制部署，独立数据库）
custom-domain.com        → 机构 C（独立域名 + 定制 Logo/主题色）
```

**选择理由**（相对于多租户 SaaS）：

| 维度 | 一客一部署 | 多租户 SaaS |
|------|-----------|------------|
| 数据隔离 | 天然隔离（物理隔离）| 需要严格的租户过滤逻辑，出错风险高 |
| 开发复杂度 | 低（无需 tenant_id 贯穿所有表）| 高 |
| 定制能力 | 高（可独立改 Logo/主题/字段）| 低（需做成通用配置项）|
| 故障影响范围 | 单实例，互不影响 | 一处故障可能影响所有租户 |
| 运维成本 | 随实例数线性增长 | 集中运维，边际成本低 |
| 适用规模 | 个位数到十位数客户 | 成百上千客户 |

在当前及可预见阶段（个位数到十位数部署实例），一客一部署的低复杂度收益远大于多租户的运维集约化收益，故 V1 采用此策略。V3 阶段若部署实例显著增多，可评估"多实例管理控制台"（集中监控 + 一键部署新实例，但数据仍物理隔离）。

---

## 第二章 用户角色与权限体系

### 2.1 角色总览

| 角色代码 | 角色名称 | V1 是否登录 | 说明 |
|---------|---------|:---:|------|
| `OWNER` | 老板 | ✅ | 机构经营者，最高权限 |
| `OPERATOR` | 教务 | ✅ | 日常运营操作人员 |
| `TEACHER` | 老师 | ❌（V1）| 仅接收邮件通知，V2 开放登录 |
| `STUDENT` | 学生 | ❌（V1）| 仅接收邮件通知，V2 开放登录 |
| `PARENT` | 家长 | ❌（V1）| 仅接收邮件通知（可选），V2 开放登录 |

### 2.2 Owner 权限详情

```
【数据查看】
  查看全部学生、老师、课程、账务数据（无范围限制）
  查看经营报表（收入/成本/毛利/趋势）
  接收老板周报（自动邮件推送）

【系统管理】
  管理 Operator 账号（新建/禁用/重置密码）
  修改系统配置（提醒规则、课时套餐、SMTP/Resend 配置等）
  管理数据备份（手动备份、查看备份历史、恢复数据）
  查看完整操作日志（含 Operator 的所有操作记录）

【业务操作】
  拥有 Operator 的全部业务操作权限
```

### 2.3 Operator 权限详情

```
【学生与老师】
  新建/编辑/查看学生档案（不可删除，仅可置为"已结束"状态）
  新建/编辑/查看老师档案（不可删除，仅可置为"已结束"状态）
  管理老师能力标签与可授时间

【课程与排课】
  维护课程体系基础数据（领域/方向/等级/标签）
  管理学生课程报名（新建/暂停/结束/换老师）
  创建/编辑/取消课次
  执行课后确认

【账务】
  录入充值记录（不可删除，仅可作废）
  录入结款记录
  查看账务流水（不可编辑历史流水，流水为系统自动生成）

【通知】
  查看通知日志
  手动触发/重发通知
  编辑通知模板内容（不可禁用系统级模板）

【报表】
  查看基础报表（非经营层面的敏感汇总，如毛利总额，V1 阶段 Operator 也可见，因运营规模小、通常为信任关系；如需限制可在系统配置中开启"隐藏毛利"选项）
  导出 Excel/PDF

【数据】
  Excel 批量导入学生/老师
  查看备份记录（不可触发恢复操作，恢复为 Owner 专属）
```

### 2.4 权限设计原则

- V1 采用**简单二级 RBAC**：Owner ⊇ Operator，不做字段级/记录级细粒度权限
- 所有写操作无条件记录 `operation_log`，用于事后审计，弥补权限粒度较粗的风险
- 关键操作（作废充值、删除数据）二次确认弹窗，非纯前端校验，后端同样二次校验
- V2 可扩展：多 Operator 场景下按"负责学生范围"做数据权限过滤（预留 `operator_scope` 表结构，V1 不实现）

### 2.5 老师/学生/家长的 V1 交互方式

V1 中 Teacher / Student / Parent **没有账号、不能登录**，仅作为通知接收方存在于系统中（作为数据记录，而非系统用户）。所有交互通过邮件单向触达：

```
系统 → 邮件 → 学生/老师/家长（仅接收，无法通过系统回复或操作）
```

若学生/老师需要变更信息、请假、调课，仍通过线下渠道（微信、电话）联系 Operator，由 Operator 在系统中代为操作。这是 V1 刻意保持的"轻"设计——避免过早引入登录体系增加的复杂度和维护成本。

---

## 第三章 核心业务流程（含异常分支）

本章在 v1.0 基础上，为每个核心流程补充**异常处理分支**，确保开发时不遗漏边界情况。

### 3.1 学生建档与课程报名流程

**主流程：**

```
学生咨询/报名
  ↓
Operator 打开"新建学生"页面
  ↓
录入基础信息（姓名*/邮箱*/电话/国籍/时区/备注）
  ↓
（可选）录入家长联系方式（可录入多个，标记主联系人）
  ↓
保存学生档案 → student.status = ACTIVE
  ↓
在学生详情页点击"新建课程报名"
  ↓
选择课程领域 → 课程方向 → 当前等级 → 目标等级
  ↓
选择报名类型（一对一/小班/试听）
  ↓
设置每课次收费（可先留空，充值时再确定）
  ↓
保存 → student_course_enrollment.status = ACTIVE
  ↓
（可选，也可稍后进行）匹配老师 → 创建 student_teacher_assignment
```

**异常分支：**

| 异常情况 | 处理方式 |
|---------|---------|
| 邮箱格式非法 | 前端 + 后端双重校验，阻止保存，提示具体错误字段 |
| 邮箱与已有学生重复 | 提示"该邮箱已存在学生：XXX"，可选择"仍然新建"（同一邮箱可能是双胞胎等情况）或"跳转至已有学生" |
| 学生暂无邮箱 | 邮箱设为非必填，但若无邮箱则该学生无法接收邮件提醒，前端明确提示"未填写邮箱，将无法接收自动提醒" |
| 课程方向下暂无可选老师 | 允许先创建报名项目、暂不指定老师（师生安排后补），排课时若无老师则不允许创建课次 |
| 学生中途要求"试听转正式" | enrollment_type 从 TRIAL 改为 ONE_TO_ONE，历史课次不受影响 |

### 3.2 老师建档流程

**主流程：**

```
录入老师基础信息（姓名*/邮箱*/电话）
  ↓
录入简介与教学经历（自由文本）
  ↓
新增能力记录（可多条）：领域 + 方向 + 等级 + 能力标签
  ↓
设置默认课酬（JPY/课次）
  ↓
维护可授时间（可多条：周几 + 起止时间）
  ↓
保存 → teacher.status = ACTIVE
```

**异常分支：**

| 异常情况 | 处理方式 |
|---------|---------|
| 老师暂无固定可授时间（时间灵活）| 可授时间设为选填，留空表示"时间灵活，需单独确认" |
| 老师能力记录冲突（同一方向同一等级重复添加）| 后端校验唯一性（teacher_id + track_id + level_id 组合），提示"该能力已存在，请编辑而非新增" |
| 老师暂停接单但仍在系统中 | status = PAUSED，排课时该老师不出现在候选列表，但历史数据保留 |

### 3.3 学生课程报名与师生匹配（多对多模型）

**核心设计**（沿用 v1.0 决策）：学生可同时报名多个课程方向，每个方向可配不同老师，不采用"学生绑定单一老师"的简化模型。

**换老师流程主流程：**

```
Operator 在课程报名详情页点击"更换老师"
  ↓
弹窗选择新老师，系统显示新老师的能力标签供参考（软提示，不强制拦截）
  ↓
填写更换原因（选填，如"老师请假"/"学生要求更换"/"能力不匹配"）
  ↓
确认更换 → 后端事务：
  ├── 旧 assignment.status = ENDED, end_date = 今日
  └── 新 assignment.status = ACTIVE, start_date = 今日
  ↓
enrollment.lesson_balance / balance_jpy 不变（余额随学生走，不随老师走）
  ↓
未来（scheduled_start_at > 今日）状态为 SCHEDULED 的课次：
  提示 Operator 是否批量将老师字段更新为新老师（默认询问，允许逐条处理）
  ↓
历史（已完成/已过去）课次不受影响，保留原老师记录
```

**异常分支：**

| 异常情况 | 处理方式 |
|---------|---------|
| 换老师时存在未来已排课次 | 弹窗提示"存在 N 个未来课次仍使用旧老师，是否批量更新？" |
| 新老师与旧老师是同一人（误操作）| 前端禁用当前老师作为"新老师"选项 |
| 换老师后需要撤销 | 支持"撤销更换"（本质是再次换回原老师，生成新的 assignment 记录，历史保留每一次变更）|

### 3.4 充值与余额流程

**主流程：**

```
学生/家长线下付款（PayPay/微信/银行转账/现金）
  ↓
Operator 打开学生详情页 → "新建充值"
  ↓
选择课程报名项目（若该学生只有一个 ACTIVE 项目，自动选中）
  ↓
选择支付方式
  ↓
输入原始金额 + 原始币种
  ↓
若币种非 JPY → 输入当时汇率 → 系统自动计算 amount_jpy = original_amount × fx_rate_to_jpy
  ↓
选择课时套餐（预设下拉，如"20课时包"）或手动输入课时数
  ↓
录入付款日期，备注
  ↓
提交 → 后端事务：
  ├── 写入 student_payment（status=CONFIRMED）
  ├── 写入 student_account_ledger（biz_type=RECHARGE，正向变动）
  └── 更新 enrollment.balance_jpy += amount_jpy, lesson_balance += lessons_added
  ↓
若该学生此前处于"余额预警"名单，重新计算后若已脱离预警范围，从下次晨报名单中自动移除（无需手动处理，晨报每次查询实时数据）
```

**异常分支：**

| 异常情况 | 处理方式 |
|---------|---------|
| 充值录入后发现金额录错 | 不允许直接编辑/删除该 payment 记录；须"作废"该记录（status=VOIDED，同时生成一条反向 ledger 冲正），再重新录入正确记录 |
| 学生有多个报名项目，充值应分配到哪个 | 强制要求选择归属项目，不支持"充值到学生账户不指定项目"（避免账目混乱）；若确有通用余额需求，V2 再评估是否支持"学生级别公共余额池" |
| 汇率获取困难（Operator 不知道当天汇率）| 系统设置中维护"参考汇率"，充值表单自动带入参考值，Operator 可手动覆盖 |
| 学生要求部分退款 | biz_type=REFUND，允许负向金额调整，需在备注中说明原因，Owner 权限可见退款明细 |

### 3.5 排课与课前提醒流程

**主流程：**

```
Operator 打开"新建课次"
  ↓
选择学生 → 自动列出该学生 ACTIVE 状态的课程报名项目 → 选择项目
  ↓
系统自动代入该项目当前 ACTIVE 的 assignment 对应老师（可手动改选其他老师）
  ↓
选择上课日期时间（前端按 Asia/Tokyo 展示，后端存 UTC）
  ↓
填写时长（默认 60 分钟，可调整）
  ↓
选择上课方式（微信群/腾讯会议/Zoom/其他）+ 填写链接
  ↓
（可选）填写本节课主题
  ↓
保存 → lesson.status = SCHEDULED
  ↓
── 系统自动流程（无需人工介入）──
定时任务每 10 分钟扫描一次
  ↓
课前 30 分钟（可配置）：
  发送提醒邮件给学生（若有邮箱）
  发送提醒邮件给老师（若有邮箱）
  ↓
更新 lesson.remind_sent_at = NOW()，lesson.status = REMINDED
写入 notification_log（学生一条 + 老师一条）
```

**异常分支：**

| 异常情况 | 处理方式 |
|---------|---------|
| 排课时该学生余额不足以支付本课次 | 前端警示（黄色提示条："当前余额不足，剩余课时 X，本次排课后将变为负数"），但不强制拦截创建（允许先排课、后续再处理续费，因为很多情况下是"先上课后续费"的信任关系）|
| 排课时老师能力不匹配当前等级 | 软提示，不拦截（同 3.3 节设计）|
| 排课时该学生/老师在同一时段已有其他课次（时间冲突）| 提示"该学生/老师在此时段已有课程安排：[课次详情]"，Operator 可选择"仍然创建"（可能是合理的连续课或例外情况）|
| 学生/老师无邮箱，无法接收提醒 | 创建课次时提示，但不阻止创建；该课次的提醒任务扫描时会跳过无邮箱的一方，仅通知有邮箱的一方 |
| 上课时间设置为过去时间 | 允许（用于补录历史课次），但状态直接为下拉可选 COMPLETED，不会触发提醒 |
| 课前提醒发送失败（邮件服务异常）| 写入 notification_log(status=FAILED)，通知重试任务（每 30 分钟）自动重试最多 3 次；若仍失败，晨报中会列出"失败通知"提醒 Operator 人工介入 |

### 3.6 课后确认与账务流程

**主流程（核心事务）：**

```
课程结束后（或次日）
  ↓
Operator 打开课次详情页 → "确认出勤"
  ↓
填写：学生是否出勤 / 老师是否出勤 / 实际时长 / 老师备注 / 是否扣费
  ↓
提交 → 后端在单一数据库事务内完成：
  ┌─ 写入 attendance 记录
  ├─ 若扣费：
  │    写入 student_account_ledger（biz_type=LESSON_DEDUCT，负向变动）
  │    更新 enrollment.lesson_balance -= 扣除课时
  │    更新 enrollment.balance_jpy -= charge_jpy
  ├─ 若老师应得课酬（默认出勤即应得，即使学生缺勤，视机构政策）：
  │    写入 teacher_account_ledger（biz_type=LESSON_PAYABLE，正向变动）
  │    更新 teacher.unpaid_amount += teacher_pay_jpy
  ├─ 写入 lesson_finance（收费/课酬/毛利快照）
  └─ 更新 lesson.status = COMPLETED
  ↓
事务提交后（事务外异步判断，不阻塞主流程）：
  判断 enrollment 是否触发余额预警条件
  ↓（若触发）
  当日 20:00 定时任务扫描时会捕获到，纳入余额预警邮件
  （亦可选择：课后确认后若余额已为 0 或负数，立即触发一条即时提醒邮件，不等到 20:00）
```

**异常分支：**

| 异常情况 | 处理方式 |
|---------|---------|
| 学生请假未上课，是否扣费 | Operator 在确认表单中勾选"是否扣费"，默认值可在系统设置中配置（如"提前 24 小时请假不扣费，否则扣费"由人工判断，V1 不做自动规则），若不扣费则不生成 LESSON_DEDUCT 流水，但如老师已到场仍可选择支付老师课酬（学生爽约老师仍应得课酬，是常见行业规则）|
| 老师缺席（老师原因未上课）| 不扣学生课时，不产生老师应付款，Operator 备注原因，可能需要另行安排补课（新建一条新课次）|
| 确认出勤时系统崩溃/网络中断 | 依赖数据库事务的原子性，未完全提交的操作自动回滚，前端提示"提交失败，请重试"，不会出现"扣了课时但没记老师课酬"的半成功状态 |
| 误操作确认了错误的课次 | V1 不支持"撤销课后确认"（避免账务被随意回滚造成混乱）；如确有错误，走"人工调整"流程：Operator 创建一条 biz_type=ADJUST 的流水手动修正，并在备注中详细说明原因，操作日志留痕 |
| 忘记确认出勤，课次一直挂起 | 定时任务 6（自动关闭过期课次）在课次结束 4 小时后自动将状态置为 COMPLETED，但**不会自动触发账务扣减**（避免无人工确认的情况下擅自扣费），Operator 仍需后续手动补充确认账务，工作台"待确认课次"列表会持续提醒 |

### 3.7 结款流程

**主流程：**

```
Operator 打开老师详情页 → "新建结款"
  ↓
选择结算周期（开始日期 - 结束日期）
  ↓
系统自动查询该周期内 teacher_account_ledger 中 biz_type=LESSON_PAYABLE 且未被结算的记录
  ↓
列表展示明细（课次/学生/日期/应付金额），Operator 可勾选排除个别记录（特殊情况）
  ↓
系统汇总应付总额，Operator 确认或调整实付金额（如有折扣/补贴）
  ↓
提交 → 后端事务：
  ├─ 写入 teacher_payout
  ├─ 写入 teacher_account_ledger（biz_type=PAYOUT，负向变动，冲抵应付）
  └─ 更新 teacher.unpaid_amount -= 实付金额对应的应付部分
  ↓
被结算的 LESSON_PAYABLE 记录标记 payout_id 关联，避免重复结算
```

**异常分支：**

| 异常情况 | 处理方式 |
|---------|---------|
| 结款周期内应付金额为 0 | 提示"该周期内无可结算课次"，不生成空结款单 |
| 部分结算（先结一部分，剩余下次结）| 支持，Operator 可在明细列表中只勾选部分记录进行本次结算 |
| 结款后发现某课次不应计入 | 生成一条 ADJUST 流水冲正，说明原因，不修改历史 payout 记录 |

---

## 第四章 课程体系设计

（本章内容与 v1.0 一致，此处保留核心要点，供后续章节引用，详见 v1.0 文档第四章获取完整背景说明）

### 4.1 四层课程体系

```
课程领域（course_domain）   日语 / 英语 / 韩语 / 足球 / 钢琴 ...
    └─ 课程方向（course_track）   JLPT备考 / 日常会话 / 商务日语 ...
         └─ 课程等级（course_level）   N5 / N4 / N3 / N2 / N1 ...
              └─ 能力标签（skill_tag）  词汇 / 语法 / 阅读 / 听力 / 口语 ...
```

### 4.2 V1 初始化数据（日语领域）

系统首次部署时，通过 migration 或初始化脚本预置以下数据，Operator 可在系统设置中增删改：

**课程领域**：`日语（LANGUAGE）`

**课程方向**（domain=日语）：
```
JLPT 备考 / 日常会话 / 商务日语 / 少儿日语 / 写作强化 / 面试日语
```

**课程等级**（track=JLPT备考）：
```
入门 → N5 → N4 → N3 → N2 → N1
```

**课程等级**（track=日常会话）：
```
初级 → 中级 → 高级
```

**能力标签**（domain=日语）：
```
词汇 / 语法 / 阅读 / 听力 / 口语 / 写作 / 综合 / 面试技巧 / 商务敬语
```

### 4.3 老师能力与学生等级变化

详见 v1.0 文档第四章第 4.3、4.4 节，设计不变：老师能力为多条 `teacher_capability` 记录（非单一字段），排课时软提示不强制拦截；学生等级变化通过 `student_level_event` 记录事件而非直接改字段，保留完整变化历史。

---

## 第五章 功能模块与页面级 UI 规格

本章将每个页面拆解到字段、按钮、交互状态级别，作为前端开发的直接依据。

### 5.1 全局导航结构

```
顶部栏：Logo + 系统名 | 通知铃铛（未读数角标）| 用户头像下拉
左侧栏：工作台 / 学员管理(学生列表+充值记录) / 教师管理(老师列表+结款记录) /
       课程体系 / 课程管理(排课列表+日历视图) /
       财务管理(收支概览+学生账单+老师账单+报表导出) / 数据图表 /
       通知管理(提醒规则+通知日志) / 系统设置
```

移动端（<768px）：左侧栏收起为底部 Tab Bar：`今日` `确认` `学生` `更多`。

### 5.2 工作台（Dashboard）— `/dashboard`

1. 问候语+日期条
2. 指标卡区（4张，Owner视图：今日课程数/本月收入/本月毛利/活跃学生数；Operator视图：今日课程数/待续费学生数/待结款老师数/待确认课次数），每卡含环比箭头，点击跳转
3. 今日课程列表卡片（时间/学生/老师/上课方式/状态/操作，含[查看链接][手动发提醒][确认出勤]）
4. 待续费学生卡片（最多5条+查看全部，剩余课时≤1红色高亮）
5. 待结款老师卡片
6. 图表区（月度收入趋势折线图/课程方向占比饼图/学生增长趋势面积图，Owner默认展开，Operator默认折叠）
7. 失败通知提醒条（存在FAILED通知时置顶显示）

### 5.3 学生列表 — `/students`

顶部：[+新建学生][导入Excel][导出Excel] + 搜索框（姓名/邮箱/电话）+ 筛选（状态/课程方向）
表格列：姓名/邮箱/电话/学习方向(标签)/总余额/总剩余课时/状态/操作(查看/编辑/快速充值)
分页每页20条；空状态显示引导插图+[新建第一个学生]

### 5.4 学生详情页 — `/students/:id`

**左侧基础信息卡**：姓名/日文名/邮箱/电话/国籍/时区/状态(可切换)/备注 + 家长信息子卡（可多条，含主联系人标记）

**右侧 Tab 页签**：
- Tab1「学习项目」：课程报名卡片列表（方向+等级箭头/当前老师/剩余课时/余额/状态/[查看课次][换老师][暂停恢复][编辑]）+ [新建课程报名]
- Tab2「账务」：充值记录表格 + 账户流水表格 + [新建充值][导出账单PDF]
- Tab3「上课记录」：历史课次表格，按项目/日期筛选
- Tab4「通知记录」：发送记录表格，失败可[重发]

### 5.5 老师列表/详情页 — `/teachers`、`/teachers/:id`

结构对称学生模块。列表额外列：可教方向/在带学生数/累计应付未结款(红色高亮>0)。
详情页Tab：能力与时间（能力标签列表+可授时间周视图）/ 账务（应付概览+结款记录+课时流水）/ 带课记录 / 通知记录

### 5.6 课程体系维护 — `/courses`

三栏联动：左-课程领域列表（增删改拖拽排序）/ 中-选中领域后的方向列表 / 右-选中方向后的等级列表+该领域能力标签管理
新建/编辑字段：名称/代码(自动建议可编辑)/排序/启用开关；等级额外含最小最大年龄、最小/建议课时数（选填）

### 5.7 排课管理 — `/lessons`、`/lessons/calendar`

**列表视图**：[+新建课次] + 筛选（日期范围/学生/老师/状态多选）
表格列：日期时间/学生/老师/课程方向-等级/上课方式/状态(颜色标签)/操作
状态颜色：SCHEDULED蓝/REMINDED橙/COMPLETED绿/CANCELLED灰
操作：[查看][编辑](仅SCHEDULED)[取消](SCHEDULED/REMINDED)[确认出勤](时间已过未确认时高亮)

**日历视图**：月/周视图切换，课次色块展示，点击弹详情浮层；移动端自动降级为按日分组列表

**新建/编辑课次表单**：学生*→课程报名项目(联动，仅ACTIVE)* / 老师*(自动带入当前assignment，可改选) / 上课日期*+开始时间*+时长(默认60分钟)* / 上课方式*(微信群/腾讯会议/Zoom/其他) / 上课链接 / 本节课主题(选填) / 备注(选填)
提交时检测时间冲突或余额不足→黄色警示但不阻止提交

**课后确认表单**：课次信息只读展示 + 学生是否出勤(默认是) + 老师是否出勤(默认是) + 实际时长(默认带入计划时长) + 是否扣课时(联动学生出勤状态) + 老师备注 + 运营备注 + [确认提交]（不可撤销，二次确认弹窗）

### 5.8 充值记录页 — `/finance/payments`

[+新建充值] + 筛选（学生/日期范围/支付方式/状态）
表格列：单号/日期/学生/项目/原始金额币种/折算JPY/课时/方式/状态/操作
VOIDED状态整行置灰，仅显示[查看]；CONFIRMED状态显示[查看][作废](二次确认+填写原因)

### 5.9 结款记录页 — `/finance/payouts`

结构对称充值记录页，字段替换为：老师/结算周期/课时数/应付/实付

### 5.10 财务报表 — `/finance/report`

日期范围选择器（预设：本月/上月/本季度/本年/自定义）
汇总卡片：总收入/总课酬支出/总毛利/毛利率
明细表格：按月/周/日汇总（可切换粒度）+ [导出Excel][导出PDF]

### 5.11 数据图表页 — `/reports`

图表1学生增长曲线（折线，可切换时间跨度）/ 图表2月度课时完成率（完成/取消/总计堆叠柱状图）/
图表3收入差价走势（收入线+课酬线+毛利线三线折线图）/ 图表4各老师带课分布（横向柱状图）/
图表5各课程方向收入占比（环形图）；每图表右上角[导出PNG]

### 5.12 通知管理 — `/notifications`

Tab1「提醒规则」：课前提醒分钟数(默认30) / 余额预警阈值(默认3课时) / 晨报发送时间(默认08:00) /
老板周报发送日+时间(默认周一08:00) / 通知语言(中文/日文/双语) + [保存配置]

Tab2「通知日志」：筛选（类型/状态/日期范围/收件人搜索）
表格列：发送时间/类型/收件人/收件人类型/主题/状态/操作；失败行显示[重发]+失败原因悬浮提示

### 5.13 系统设置 — `/settings`

Tab「基本信息」：机构名称/Logo上传/联系方式/系统时区/默认币种
Tab「邮件配置」：Resend API Key(脱敏显示)/发件人名称/发件邮箱/SMTP备用配置
Tab「课时套餐」：套餐管理列表（名称/金额/课时数/启用状态），增删改
Tab「数据备份」(仅Owner)：自动备份开关+Cron配置 / Litestream云备份配置(S3/R2参数) /
备份历史列表+[立即备份][下载][恢复](需输入确认文字二次确认)
Tab「账号管理」(仅Owner)：Operator账号列表+[新建账号][禁用][重置密码]
Tab「操作日志」：全量操作记录，筛选+搜索

### 5.14 移动端专属页面

**`/mobile/today`**：日期切换 + 课程卡片列表（时间段大字号/学生老师姓名/上课方式图标+[复制链接]/状态标签/已过时间显示[确认出勤]大按钮）+ 底部Tab Bar

**`/mobile/confirm`**：卡片列表仅展示待确认课次，每卡内嵌简化课后确认表单（出勤开关+提交按钮）

**`/mobile/recharge`**：学生搜索大输入框 + 极简充值表单（金额/币种/课时/方式）+ [提交]大按钮

**`/mobile/alerts`**：卡片列表（学生姓名+剩余课时大号红字+联系方式）+ [一键拨号][复制微信号][去充值]快捷按钮

---

# 第二部分：数据与接口

## 第六章 数据模型设计（完整 DDL）

### 6.1 设计原则

- 金额一律以**整数分（minor unit）**存储：JPY 用 `INTEGER`（1円=1），涉及 CNY 等两位小数币种时统一换算为 JPY 整数分入库，原始金额单独用 `TEXT`（decimal 字符串）保存
- 时间统一 `DATETIME` 存 UTC（ISO8601 字符串或 SQLite 内建 datetime），应用层转时区
- 主键统一 `INTEGER PRIMARY KEY AUTOINCREMENT`
- 枚举字段用 `TEXT` + 应用层校验，不使用数据库 ENUM
- 每张业务表含 `created_at`，多数含 `updated_at`；支持软删除的表含 `deleted_at`
- 外键统一显式声明，启动时 `PRAGMA foreign_keys=ON`
- 预留 `extra_json TEXT` 字段用于未来扩展属性，避免频繁改表结构

### 6.2 完整 DDL（SQLite 方言）

```sql
-- ========== 系统与权限 ==========

CREATE TABLE user_account (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK(role IN ('OWNER','OPERATOR')),
  display_name TEXT NOT NULL,
  email TEXT,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','DISABLED')),
  last_login_at DATETIME,
  login_fail_count INTEGER NOT NULL DEFAULT 0,
  locked_until DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);

CREATE TABLE system_config (
  config_key TEXT PRIMARY KEY,
  config_value TEXT NOT NULL,
  description TEXT,
  updated_by INTEGER REFERENCES user_account(id),
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE operation_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  operator_id INTEGER REFERENCES user_account(id),
  operator_name TEXT,
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id INTEGER,
  detail_json TEXT,
  ip_addr TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_oplog_target ON operation_log(target_type, target_id);
CREATE INDEX idx_oplog_created ON operation_log(created_at);

CREATE TABLE backup_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_name TEXT NOT NULL,
  file_size INTEGER,
  trigger_type TEXT NOT NULL CHECK(trigger_type IN ('AUTO','MANUAL')),
  status TEXT NOT NULL CHECK(status IN ('SUCCESS','FAILED')),
  error_msg TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ========== 学生与老师 ==========

CREATE TABLE student (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  name_jp TEXT,
  email TEXT,
  phone TEXT,
  nationality TEXT,
  timezone TEXT NOT NULL DEFAULT 'Asia/Tokyo',
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','ENDED')),
  note TEXT,
  extra_json TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX idx_student_email ON student(email);
CREATE INDEX idx_student_status ON student(status);
CREATE INDEX idx_student_name ON student(name);

CREATE TABLE parent (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  student_id INTEGER NOT NULL REFERENCES student(id),
  name TEXT NOT NULL,
  email TEXT,
  phone TEXT,
  relationship TEXT CHECK(relationship IN ('FATHER','MOTHER','OTHER')),
  is_primary INTEGER NOT NULL DEFAULT 0,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_parent_student ON parent(student_id);

CREATE TABLE teacher (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  name_jp TEXT,
  email TEXT,
  phone TEXT,
  bio TEXT,
  default_rate_jpy INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','ENDED')),
  note TEXT,
  extra_json TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX idx_teacher_email ON teacher(email);
CREATE INDEX idx_teacher_status ON teacher(status);

CREATE TABLE teacher_availability (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  weekday INTEGER NOT NULL CHECK(weekday BETWEEN 0 AND 6),
  start_time TEXT NOT NULL,
  end_time TEXT NOT NULL,
  effective_from DATE,
  effective_to DATE,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_avail_teacher ON teacher_availability(teacher_id);

CREATE TABLE teacher_capability (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  domain_id INTEGER NOT NULL REFERENCES course_domain(id),
  track_id INTEGER REFERENCES course_track(id),
  level_id INTEGER REFERENCES course_level(id),
  skill_tag_codes TEXT,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','ENDED')),
  verified INTEGER NOT NULL DEFAULT 0,
  effective_from DATE,
  effective_to DATE,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(teacher_id, track_id, level_id)
);
CREATE INDEX idx_cap_teacher ON teacher_capability(teacher_id);

-- ========== 课程体系 ==========

CREATE TABLE course_domain (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  code TEXT NOT NULL UNIQUE,
  type TEXT NOT NULL CHECK(type IN ('LANGUAGE','SPORT','ART','ACADEMIC','OTHER')),
  sort_order INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE course_track (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  domain_id INTEGER NOT NULL REFERENCES course_domain(id),
  name TEXT NOT NULL,
  code TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(domain_id, code)
);
CREATE INDEX idx_track_domain ON course_track(domain_id);

CREATE TABLE course_level (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  track_id INTEGER NOT NULL REFERENCES course_track(id),
  name TEXT NOT NULL,
  code TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  min_age INTEGER,
  max_age INTEGER,
  min_lesson_hours REAL,
  recommended_lesson_hours REAL,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(track_id, code)
);
CREATE INDEX idx_level_track ON course_level(track_id);

CREATE TABLE skill_tag (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  domain_id INTEGER NOT NULL REFERENCES course_domain(id),
  name TEXT NOT NULL,
  code TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(domain_id, code)
);

-- ========== 学生学习项目 ==========

CREATE TABLE student_course_enrollment (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  student_id INTEGER NOT NULL REFERENCES student(id),
  domain_id INTEGER NOT NULL REFERENCES course_domain(id),
  track_id INTEGER NOT NULL REFERENCES course_track(id),
  current_level_id INTEGER REFERENCES course_level(id),
  target_level_id INTEGER REFERENCES course_level(id),
  enrollment_type TEXT NOT NULL DEFAULT 'ONE_TO_ONE' CHECK(enrollment_type IN ('ONE_TO_ONE','GROUP','TRIAL')),
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','COMPLETED','CANCELLED')),
  charge_per_lesson_jpy INTEGER NOT NULL DEFAULT 0,
  lesson_balance REAL NOT NULL DEFAULT 0,
  balance_jpy INTEGER NOT NULL DEFAULT 0,
  started_at DATE,
  ended_at DATE,
  note TEXT,
  extra_json TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX idx_enrollment_student ON student_course_enrollment(student_id);
CREATE INDEX idx_enrollment_status ON student_course_enrollment(status);

CREATE TABLE student_teacher_assignment (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  student_id INTEGER NOT NULL REFERENCES student(id),
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  role_type TEXT NOT NULL DEFAULT 'MAIN' CHECK(role_type IN ('MAIN','SUBSTITUTE','ASSISTANT')),
  rate_jpy INTEGER,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','ENDED')),
  start_date DATE NOT NULL,
  end_date DATE,
  reason TEXT,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_assign_enrollment ON student_teacher_assignment(enrollment_id);
CREATE INDEX idx_assign_teacher ON student_teacher_assignment(teacher_id);
CREATE INDEX idx_assign_status ON student_teacher_assignment(status);

CREATE TABLE student_learning_path (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  student_id INTEGER NOT NULL REFERENCES student(id),
  from_level_id INTEGER REFERENCES course_level(id),
  current_level_id INTEGER REFERENCES course_level(id),
  target_level_id INTEGER REFERENCES course_level(id),
  goal_type TEXT CHECK(goal_type IN ('EXAM','CONVERSATION','BUSINESS','HOBBY','COMPETITION')),
  target_exam_name TEXT,
  target_exam_date DATE,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','COMPLETED','CHANGED','PAUSED')),
  started_at DATE,
  ended_at DATE,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_path_enrollment ON student_learning_path(enrollment_id);

CREATE TABLE student_level_event (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  student_id INTEGER NOT NULL REFERENCES student(id),
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  learning_path_id INTEGER REFERENCES student_learning_path(id),
  from_level_id INTEGER REFERENCES course_level(id),
  to_level_id INTEGER NOT NULL REFERENCES course_level(id),
  event_type TEXT NOT NULL CHECK(event_type IN ('ASSESSMENT','EXAM_PASS','HOURS_REACHED','AGE_REACHED','MANUAL')),
  event_date DATE NOT NULL,
  evidence_note TEXT,
  operator_id INTEGER REFERENCES user_account(id),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_levelevent_student ON student_level_event(student_id);

-- ========== 班级（V1 结构预留） ==========

CREATE TABLE class_group (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  domain_id INTEGER NOT NULL REFERENCES course_domain(id),
  track_id INTEGER REFERENCES course_track(id),
  level_id INTEGER REFERENCES course_level(id),
  main_teacher_id INTEGER REFERENCES teacher(id),
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','FINISHED')),
  charge_per_lesson_jpy INTEGER NOT NULL DEFAULT 0,
  rate_per_lesson_jpy INTEGER NOT NULL DEFAULT 0,
  max_students INTEGER,
  started_at DATE,
  ended_at DATE,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE class_group_member (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  class_group_id INTEGER NOT NULL REFERENCES class_group(id),
  student_id INTEGER NOT NULL REFERENCES student(id),
  enrollment_id INTEGER REFERENCES student_course_enrollment(id),
  join_date DATE NOT NULL,
  leave_date DATE,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','LEFT')),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ========== 排课与上课 ==========

CREATE TABLE lesson (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  lesson_no TEXT NOT NULL UNIQUE,
  enrollment_id INTEGER REFERENCES student_course_enrollment(id),
  class_group_id INTEGER REFERENCES class_group(id),
  student_id INTEGER NOT NULL REFERENCES student(id),
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  domain_id INTEGER REFERENCES course_domain(id),
  track_id INTEGER REFERENCES course_track(id),
  level_id INTEGER REFERENCES course_level(id),
  lesson_topic TEXT,
  scheduled_start_at DATETIME NOT NULL,
  scheduled_end_at DATETIME NOT NULL,
  duration_min INTEGER NOT NULL DEFAULT 60,
  timezone TEXT NOT NULL DEFAULT 'Asia/Tokyo',
  meeting_type TEXT NOT NULL CHECK(meeting_type IN ('WECHAT','TENCENT','ZOOM','OTHER')),
  meeting_link TEXT,
  status TEXT NOT NULL DEFAULT 'SCHEDULED' CHECK(status IN ('SCHEDULED','REMINDED','COMPLETED','CANCELLED')),
  remind_sent_at DATETIME,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_lesson_student ON lesson(student_id);
CREATE INDEX idx_lesson_teacher ON lesson(teacher_id);
CREATE INDEX idx_lesson_scheduled ON lesson(scheduled_start_at);
CREATE INDEX idx_lesson_status ON lesson(status);
CREATE INDEX idx_lesson_remind ON lesson(remind_sent_at, status);

CREATE TABLE attendance (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  lesson_id INTEGER NOT NULL REFERENCES lesson(id),
  enrollment_id INTEGER REFERENCES student_course_enrollment(id),
  student_id INTEGER NOT NULL REFERENCES student(id),
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  actual_start_at DATETIME,
  actual_end_at DATETIME,
  student_attended INTEGER NOT NULL DEFAULT 1,
  teacher_attended INTEGER NOT NULL DEFAULT 1,
  lesson_deducted REAL NOT NULL DEFAULT 0,
  charge_jpy INTEGER NOT NULL DEFAULT 0,
  teacher_pay_jpy INTEGER NOT NULL DEFAULT 0,
  teacher_note TEXT,
  operator_note TEXT,
  confirmed_by INTEGER REFERENCES user_account(id),
  confirmed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_attendance_lesson ON attendance(lesson_id);

-- ========== 账务体系 ==========

CREATE TABLE student_payment (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  payment_no TEXT NOT NULL UNIQUE,
  student_id INTEGER NOT NULL REFERENCES student(id),
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  original_amount TEXT NOT NULL,
  original_currency TEXT NOT NULL DEFAULT 'JPY',
  fx_rate_to_jpy TEXT NOT NULL DEFAULT '1',
  amount_jpy INTEGER NOT NULL,
  lessons_added REAL NOT NULL DEFAULT 0,
  package_name TEXT,
  payment_method TEXT NOT NULL CHECK(payment_method IN ('WECHAT','PAYPAY','BANK','CASH','OTHER')),
  paid_at DATETIME NOT NULL,
  operator_id INTEGER REFERENCES user_account(id),
  status TEXT NOT NULL DEFAULT 'CONFIRMED' CHECK(status IN ('CONFIRMED','VOIDED')),
  voided_at DATETIME,
  void_reason TEXT,
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_payment_student ON student_payment(student_id);
CREATE INDEX idx_payment_enrollment ON student_payment(enrollment_id);
CREATE INDEX idx_payment_status ON student_payment(status);

CREATE TABLE teacher_payout (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  payout_no TEXT NOT NULL UNIQUE,
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  period_start DATE NOT NULL,
  period_end DATE NOT NULL,
  lesson_count REAL NOT NULL DEFAULT 0,
  amount_jpy INTEGER NOT NULL DEFAULT 0,
  actual_amount_jpy INTEGER NOT NULL DEFAULT 0,
  payment_method TEXT,
  paid_at DATETIME,
  operator_id INTEGER REFERENCES user_account(id),
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_payout_teacher ON teacher_payout(teacher_id);

CREATE TABLE student_account_ledger (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  student_id INTEGER NOT NULL REFERENCES student(id),
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  biz_type TEXT NOT NULL CHECK(biz_type IN ('RECHARGE','LESSON_DEDUCT','REFUND','ADJUST','VOID')),
  amount_jpy_delta INTEGER NOT NULL,
  lesson_delta REAL NOT NULL DEFAULT 0,
  balance_jpy_after INTEGER NOT NULL,
  lesson_balance_after REAL NOT NULL,
  related_payment_id INTEGER REFERENCES student_payment(id),
  related_lesson_id INTEGER REFERENCES lesson(id),
  operator_id INTEGER REFERENCES user_account(id),
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_ledger_student ON student_account_ledger(student_id);
CREATE INDEX idx_ledger_enrollment ON student_account_ledger(enrollment_id);
CREATE INDEX idx_ledger_created ON student_account_ledger(created_at);

CREATE TABLE teacher_account_ledger (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  biz_type TEXT NOT NULL CHECK(biz_type IN ('LESSON_PAYABLE','PAYOUT','ADJUST')),
  amount_jpy_delta INTEGER NOT NULL,
  unpaid_amount_after INTEGER NOT NULL,
  related_lesson_id INTEGER REFERENCES lesson(id),
  related_payout_id INTEGER REFERENCES teacher_payout(id),
  operator_id INTEGER REFERENCES user_account(id),
  note TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_tledger_teacher ON teacher_account_ledger(teacher_id);

CREATE TABLE lesson_finance (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  lesson_id INTEGER NOT NULL UNIQUE REFERENCES lesson(id),
  enrollment_id INTEGER REFERENCES student_course_enrollment(id),
  student_id INTEGER NOT NULL REFERENCES student(id),
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  charge_jpy INTEGER NOT NULL DEFAULT 0,
  teacher_pay_jpy INTEGER NOT NULL DEFAULT 0,
  gross_profit_jpy INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_finance_created ON lesson_finance(created_at);

CREATE TABLE fx_rate_snapshot (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  from_currency TEXT NOT NULL,
  to_currency TEXT NOT NULL DEFAULT 'JPY',
  rate TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT 'MANUAL',
  recorded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  operator_id INTEGER REFERENCES user_account(id)
);

-- ========== 通知体系 ==========

CREATE TABLE notification_template (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  code TEXT NOT NULL UNIQUE,
  type TEXT NOT NULL DEFAULT 'EMAIL',
  language TEXT NOT NULL DEFAULT 'ZH' CHECK(language IN ('ZH','JA','EN','BILINGUAL')),
  subject_tpl TEXT NOT NULL,
  body_tpl TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  updated_by INTEGER REFERENCES user_account(id),
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE notification_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  template_code TEXT,
  type TEXT NOT NULL DEFAULT 'EMAIL',
  lesson_id INTEGER REFERENCES lesson(id),
  recipient_id INTEGER,
  recipient_type TEXT CHECK(recipient_type IN ('STUDENT','TEACHER','OPERATOR')),
  recipient_email TEXT,
  recipient_name TEXT,
  subject TEXT,
  body_preview TEXT,
  status TEXT NOT NULL DEFAULT 'PENDING' CHECK(status IN ('PENDING','SENT','FAILED','CANCELLED')),
  error_msg TEXT,
  retry_count INTEGER NOT NULL DEFAULT 0,
  sent_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_notiflog_status ON notification_log(status);
CREATE INDEX idx_notiflog_lesson ON notification_log(lesson_id);
CREATE INDEX idx_notiflog_created ON notification_log(created_at);
```

### 6.3 启动时必须执行的 PRAGMA

```sql
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;
PRAGMA synchronous=NORMAL;
```

### 6.4 ER 关系图（文字版）

```
student 1─N parent
student 1─N student_course_enrollment
student_course_enrollment 1─N student_teacher_assignment ─N─1 teacher
student_course_enrollment 1─N student_learning_path
student_course_enrollment 1─N student_level_event
student_course_enrollment 1─N lesson
teacher 1─N teacher_availability
teacher 1─N teacher_capability
course_domain 1─N course_track 1─N course_level
course_domain 1─N skill_tag
lesson 1─1 attendance
lesson 1─1 lesson_finance
student_course_enrollment 1─N student_payment
student 1─N student_account_ledger
teacher 1─N teacher_account_ledger
teacher 1─N teacher_payout
lesson 1─N notification_log
```


## 第七章 API 接口规范（全量清单）

### 7.1 通用约定

```
Base URL: /api/v1
认证方式: Authorization: Bearer <JWT access_token>
Content-Type: application/json（除文件上传用 multipart/form-data）
时间格式: ISO8601 UTC，如 2026-07-03T09:00:00Z
分页参数: page（默认1）、pageSize（默认20，最大100）
排序参数: sortBy、sortOrder（asc/desc）
```

**统一响应包裹**：

```json
{ "code": 0, "message": "success", "data": {}, "traceId": "..." }
```

### 7.2 认证模块 `/api/v1/auth`

| 方法 | 路径 | 说明 | 请求体 | 成功响应 data |
|------|------|------|--------|--------------|
| POST | /auth/login | 登录 | `{username, password}` | `{accessToken, refreshToken, user:{id,username,role,displayName}}` |
| POST | /auth/refresh | 刷新Token | `{refreshToken}` | `{accessToken, refreshToken}` |
| POST | /auth/logout | 登出 | `{refreshToken}` | `{}` |
| GET  | /auth/me | 当前用户信息 | - | `{id,username,role,displayName,email}` |
| POST | /auth/change-password | 修改密码 | `{oldPassword,newPassword}` | `{}` |

### 7.3 用户账号 `/api/v1/users`（仅 Owner）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /users | 列表（分页） |
| POST | /users | 新建 Operator 账号 |
| PUT | /users/{id} | 编辑（displayName/email） |
| POST | /users/{id}/disable | 禁用账号 |
| POST | /users/{id}/enable | 启用账号 |
| POST | /users/{id}/reset-password | 重置密码（生成随机密码并返回一次）|

### 7.4 学生 `/api/v1/students`

| 方法 | 路径 | 说明 | 关键参数/请求体 |
|------|------|------|----------------|
| GET | /students | 列表（分页/搜索/筛选） | `?keyword=&status=&trackId=&page=&pageSize=` |
| POST | /students | 新建 | `{name,nameJp,email,phone,nationality,timezone,note}` |
| GET | /students/{id} | 详情（含enrollments概要）| - |
| PUT | /students/{id} | 编辑基础信息 | 同新建字段 |
| POST | /students/{id}/status | 变更状态 | `{status}` |
| GET | /students/{id}/enrollments | 该学生课程报名列表 | - |
| GET | /students/{id}/lessons | 该学生课次记录 | `?enrollmentId=&from=&to=` |
| GET | /students/{id}/ledger | 该学生账户流水 | `?enrollmentId=&from=&to=` |
| GET | /students/{id}/notifications | 该学生通知记录 | - |
| POST | /students/import | Excel 批量导入 | multipart file |
| GET | /students/export | Excel 导出 | `?status=&trackId=`（返回文件流）|
| GET | /students/import-template | 下载导入模板 | - |

### 7.5 家长 `/api/v1/parents`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /students/{studentId}/parents | 家长列表 |
| POST | /students/{studentId}/parents | 新建家长联系方式 |
| PUT | /parents/{id} | 编辑 |
| DELETE | /parents/{id} | 删除 |
| POST | /parents/{id}/set-primary | 设为主联系人 |

### 7.6 老师 `/api/v1/teachers`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /teachers | 列表（分页/搜索/筛选可教方向）|
| POST | /teachers | 新建 |
| GET | /teachers/{id} | 详情 |
| PUT | /teachers/{id} | 编辑基础信息 |
| POST | /teachers/{id}/status | 变更状态 |
| GET | /teachers/{id}/capabilities | 能力列表 |
| POST | /teachers/{id}/capabilities | 新建能力记录 |
| PUT | /teacher-capabilities/{id} | 编辑能力记录 |
| DELETE | /teacher-capabilities/{id} | 删除能力记录 |
| GET | /teachers/{id}/availability | 可授时间列表 |
| POST | /teachers/{id}/availability | 新建可授时间 |
| PUT | /teacher-availability/{id} | 编辑 |
| DELETE | /teacher-availability/{id} | 删除 |
| GET | /teachers/{id}/students | 带课学生列表 |
| GET | /teachers/{id}/ledger | 账务流水 |
| POST | /teachers/import | Excel 批量导入 |
| GET | /teachers/export | Excel 导出 |

### 7.7 课程体系 `/api/v1/courses`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /courses/domains | 领域列表 |
| POST | /courses/domains | 新建领域 |
| PUT | /courses/domains/{id} | 编辑 |
| POST | /courses/domains/{id}/reorder | 调整排序 |
| GET | /courses/domains/{id}/tracks | 该领域下方向列表 |
| POST | /courses/tracks | 新建方向 |
| PUT | /courses/tracks/{id} | 编辑 |
| GET | /courses/tracks/{id}/levels | 该方向下等级列表 |
| POST | /courses/levels | 新建等级 |
| PUT | /courses/levels/{id} | 编辑 |
| GET | /courses/domains/{id}/skill-tags | 该领域能力标签列表 |
| POST | /courses/skill-tags | 新建标签 |
| PUT | /courses/skill-tags/{id} | 编辑 |

### 7.8 课程报名与师生安排 `/api/v1/enrollments`

| 方法 | 路径 | 说明 | 请求体要点 |
|------|------|------|-----------|
| GET | /enrollments | 列表（可按学生/状态筛选） | - |
| POST | /enrollments | 新建课程报名 | `{studentId,domainId,trackId,currentLevelId,targetLevelId,enrollmentType,chargePerLessonJpy}` |
| GET | /enrollments/{id} | 详情 | - |
| PUT | /enrollments/{id} | 编辑 | - |
| POST | /enrollments/{id}/status | 变更状态（暂停/结束/恢复）| `{status,reason}` |
| GET | /enrollments/{id}/assignments | 师生安排历史 | - |
| POST | /enrollments/{id}/assignments/change-teacher | 更换老师 | `{newTeacherId,reason,updateFutureLessons:bool}` |
| GET | /enrollments/{id}/learning-paths | 学习路径历史 | - |
| POST | /enrollments/{id}/learning-paths | 新建学习路径 | - |
| GET | /enrollments/{id}/level-events | 等级变化事件 | - |
| POST | /enrollments/{id}/level-events | 新建等级变化事件 | `{toLevelId,eventType,eventDate,evidenceNote}` |

### 7.9 排课与上课 `/api/v1/lessons`

| 方法 | 路径 | 说明 | 请求体要点 |
|------|------|------|-----------|
| GET | /lessons | 列表（分页/筛选：日期范围/学生/老师/状态）| - |
| GET | /lessons/calendar | 日历视图数据 | `?from=&to=&teacherId=&studentId=` |
| POST | /lessons | 新建课次 | `{enrollmentId,teacherId,scheduledStartAt,durationMin,meetingType,meetingLink,lessonTopic,note}` |
| GET | /lessons/{id} | 详情 | - |
| PUT | /lessons/{id} | 编辑（仅SCHEDULED可编辑）| 同新建字段 |
| POST | /lessons/{id}/cancel | 取消课次 | `{reason}` |
| POST | /lessons/{id}/confirm | 课后确认 | `{studentAttended,teacherAttended,actualDurationMin,deductLesson:bool,teacherNote,operatorNote}` |
| POST | /lessons/{id}/remind | 手动触发提醒邮件 | - |
| GET | /lessons/{id}/conflicts | 时间冲突检测 | - |

### 7.10 账务 `/api/v1/finance`

| 方法 | 路径 | 说明 | 请求体要点 |
|------|------|------|-----------|
| GET | /finance/payments | 充值记录列表 | `?studentId=&from=&to=&status=` |
| POST | /finance/payments | 新建充值 | `{studentId,enrollmentId,originalAmount,originalCurrency,fxRateToJpy,lessonsAdded,packageName,paymentMethod,paidAt,note}` |
| GET | /finance/payments/{id} | 详情 | - |
| POST | /finance/payments/{id}/void | 作废 | `{reason}` |
| GET | /finance/payouts | 结款记录列表 | `?teacherId=&from=&to=` |
| POST | /finance/payouts/preview | 预览待结算明细 | `{teacherId,periodStart,periodEnd}` |
| POST | /finance/payouts | 提交结款 | `{teacherId,periodStart,periodEnd,excludeLessonIds[],actualAmountJpy,paymentMethod,note}` |
| GET | /finance/ledger/student/{studentId} | 学生流水 | - |
| GET | /finance/ledger/teacher/{teacherId} | 老师流水 | - |
| POST | /finance/ledger/adjust | 人工调整流水（ADJUST类型）| `{targetType,targetId,amountJpyDelta,reason}` |
| GET | /finance/report | 财务汇总报表 | `?from=&to=&groupBy=month|week|day` |
| GET | /finance/report/export | 导出报表 | `?format=excel|pdf&from=&to=` |
| GET | /finance/packages | 课时套餐列表 | - |
| POST | /finance/packages | 新建套餐 | `{name,amountJpy,lessons,enabled}` |
| PUT | /finance/packages/{id} | 编辑套餐 | - |

### 7.11 报表 `/api/v1/reports`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /reports/dashboard | 工作台聚合数据（指标卡+今日课程+待办列表）|
| GET | /reports/revenue-trend | 月度收入趋势 |
| GET | /reports/student-growth | 学生增长曲线 |
| GET | /reports/track-distribution | 课程方向分布 |
| GET | /reports/teacher-workload | 老师带课分布 |
| GET | /reports/completion-rate | 课时完成率 |

### 7.12 通知 `/api/v1/notifications`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /notifications/logs | 通知日志列表（分页/筛选）|
| POST | /notifications/logs/{id}/resend | 重发 |
| GET | /notifications/templates | 模板列表 |
| PUT | /notifications/templates/{code} | 编辑模板内容 |
| GET | /notifications/rules | 获取提醒规则配置 |
| PUT | /notifications/rules | 更新提醒规则配置 |

### 7.13 系统 `/api/v1/system`

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /system/config | 获取系统配置（基本信息/邮件/币种等）|
| PUT | /system/config | 更新系统配置 |
| GET | /system/operation-logs | 操作日志列表 |
| GET | /system/fx-rates | 参考汇率列表 |
| PUT | /system/fx-rates | 更新参考汇率 |

### 7.14 备份 `/api/v1/backup`（仅 Owner）

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /backup/logs | 备份历史列表 |
| POST | /backup/trigger | 立即手动备份 |
| GET | /backup/{id}/download | 下载备份文件 |
| POST | /backup/{id}/restore | 恢复数据（需二次确认参数 `{confirmText:"CONFIRM"}`）|

### 7.15 健康检查

```
GET /healthz  →  { "status": "ok", "version": "1.0.0", "uptime": 12345 }
```

---

## 第八章 业务规则与字段校验总表

### 8.1 字段校验规则速查表

| 字段 | 规则 |
|------|------|
| student.name | 必填，1~50字符 |
| student.email | 选填，标准邮箱格式，若填写则系统内唯一性提示（非强制唯一）|
| student.phone | 选填，允许国际格式（+86/+81前缀）|
| student.timezone | 必填，IANA时区字符串，默认 Asia/Tokyo |
| teacher.name | 必填，1~50字符 |
| teacher.default_rate_jpy | 必填，≥0 整数 |
| enrollment.charge_per_lesson_jpy | 必填，≥0 整数 |
| lesson.scheduled_start_at | 必填，须为合法日期时间 |
| lesson.duration_min | 必填，10~480 之间整数 |
| lesson.meeting_link | 选填，若 meeting_type≠OTHER 建议填写但不强制 |
| student_payment.original_amount | 必填，> 0，decimal 字符串，最多2位小数 |
| student_payment.fx_rate_to_jpy | 必填，> 0，decimal 字符串 |
| student_payment.lessons_added | 必填，> 0 |
| attendance.actual_duration_min | 选填，若填写须 ≤ lesson.duration_min × 2（防止异常录入）|
| user_account.password | 新建/重置时 ≥ 8 位，须含字母+数字 |
| course_domain.code / course_track.code / course_level.code | 必填，同层级内唯一，仅允许字母数字下划线 |

### 8.2 业务规则清单（易被忽略的边界规则）

```
R1  学生 status=ENDED 后不可再新建课次或课程报名，但历史数据只读可查
R2  老师 status=ENDED 后不可再被指派为新课次的老师，历史数据只读可查
R3  enrollment.status=CANCELLED/COMPLETED 后不可再排课
R4  同一 enrollment 下同一时间只能有一个 status=ACTIVE 的 student_teacher_assignment（同一role_type=MAIN）
R5  lesson.status=COMPLETED 后不可再编辑课次基础信息（时间/老师等），仅可编辑 note
R6  attendance 一旦创建（课后确认提交后）不可删除、不可修改（如需修正走 ADJUST 流水）
R7  student_payment.status=VOIDED 的记录不计入任何余额和报表统计
R8  charge_per_lesson_jpy 可在 enrollment 编辑时修改，但不影响历史 lesson_finance 快照（快照落盘时的值为准）
R9  teacher_capability 的 (teacher_id, track_id, level_id) 组合唯一，重复添加需走编辑
R10 通知重试最多 3 次，超过后需人工手动重发
R11 一个 lesson 只能对应一条 attendance 记录（1对1唯一约束）
R12 备份恢复操作会完全覆盖当前数据库，操作前系统自动先对当前状态做一次快照备份
R13 Excel 导入时，若邮箱已存在则整行跳过并在导入报告中标注，不做自动合并
R14 汇率 fx_rate_to_jpy 若原币种=JPY，固定为 "1"，前端禁用输入
R15 结款预览时，已被其他 payout 关联的 LESSON_PAYABLE 记录不会重复出现在下次预览中
```

### 8.3 业务错误码总表

| 错误码 | HTTP状态 | 说明 |
|--------|---------|------|
| 0 | 200 | 成功 |
| 40001 | 400 | 参数校验失败（详见 message 具体字段）|
| 40002 | 400 | 学生余额不足 |
| 40003 | 400 | 时间冲突（学生或老师在该时段已有课次）|
| 40101 | 401 | 未登录或 Token 过期 |
| 40102 | 401 | 用户名或密码错误 |
| 40103 | 401 | 账号已被锁定 |
| 40301 | 403 | 权限不足（Operator 尝试访问 Owner 专属接口）|
| 40401 | 404 | 资源不存在 |
| 40901 | 409 | 数据冲突（如邮箱重复、能力记录重复）|
| 42201 | 422 | 状态不允许该操作（如尝试编辑已COMPLETED的课次）|
| 50001 | 500 | 服务器内部错误 |
| 50002 | 500 | 数据库事务失败 |
| 50301 | 503 | 邮件服务暂时不可用 |

---

## 第九章 状态机设计

### 9.1 课次状态机（lesson.status）

```
SCHEDULED ──(课前提醒任务触发)──▶ REMINDED
SCHEDULED ──(Operator取消)──▶ CANCELLED
REMINDED  ──(Operator取消)──▶ CANCELLED
SCHEDULED ──(自动关闭任务，超时未确认)──▶ COMPLETED（仅状态变更，无账务）
REMINDED  ──(自动关闭任务，超时未确认)──▶ COMPLETED（仅状态变更，无账务）
SCHEDULED ──(Operator课后确认)──▶ COMPLETED（含账务事务）
REMINDED  ──(Operator课后确认)──▶ COMPLETED（含账务事务）
COMPLETED, CANCELLED 为终态，不可再变更
```

### 9.2 课程报名状态机（enrollment.status）

```
ACTIVE ──(Operator暂停)──▶ PAUSED
PAUSED ──(Operator恢复)──▶ ACTIVE
ACTIVE ──(达成学习目标)──▶ COMPLETED
ACTIVE/PAUSED ──(学生放弃/终止)──▶ CANCELLED
COMPLETED, CANCELLED 为终态
```

### 9.3 充值记录状态机（student_payment.status）

```
CONFIRMED ──(Operator作废，需填写原因)──▶ VOIDED
VOIDED 为终态，不可恢复为 CONFIRMED（如需恢复须重新新建一条记录）
```

### 9.4 师生安排状态机（student_teacher_assignment.status）

```
ACTIVE ──(换老师操作，原assignment)──▶ ENDED
ACTIVE ──(临时暂停，如老师请假)──▶ PAUSED
PAUSED ──(恢复)──▶ ACTIVE
ENDED 为终态
```

### 9.5 通知状态机（notification_log.status）

```
PENDING ──(发送成功)──▶ SENT
PENDING ──(发送失败)──▶ FAILED
FAILED ──(重试成功)──▶ SENT
FAILED ──(重试仍失败，达到最大次数)──▶ FAILED（终态，需人工处理）
PENDING/FAILED ──(课次被取消)──▶ CANCELLED
```


---

# 第三部分：系统与工程

## 第十章 后端技术架构

### 10.1 技术栈总览

| 层次 | 选型 | 版本参考 |
|------|------|---------|
| 语言 | Go | 1.22+ |
| Web 框架 | Gin | v1.10+ |
| ORM | GORM | v1.25+ |
| 数据库 | SQLite（modernc.org/sqlite）| 纯Go驱动，无CGO |
| 迁移工具 | goose | v3+ |
| JWT | golang-jwt/jwt | v5 |
| 金额运算 | shopspring/decimal | 最新稳定版 |
| 定时任务 | robfig/cron | v3 |
| 邮件 | resend-go | v2 |
| Excel | excelize | v2 |
| PDF | gofpdf | 最新稳定版 |
| 日志 | zerolog | 最新稳定版 |
| 配置 | viper | 最新稳定版 |

### 10.2 工程目录结构（与 v1.0 一致，此处不重复列出，详见 v1.0 文档第十章）

### 10.3 关键代码规范补充

**统一响应封装（pkg/response）：**

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    TraceID string      `json:"traceId"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(200, Response{Code: 0, Message: "success", Data: data, TraceID: getTraceID(c)})
}

func Error(c *gin.Context, httpStatus, bizCode int, message string) {
    c.JSON(httpStatus, Response{Code: bizCode, Message: message, TraceID: getTraceID(c)})
}
```

**课后确认事务示例（service 层核心逻辑）：**

```go
func (s *LessonService) ConfirmAttendance(ctx context.Context, lessonID int64, req ConfirmRequest) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        var lesson Lesson
        if err := tx.First(&lesson, lessonID).Error; err != nil {
            return err
        }
        if lesson.Status == "COMPLETED" || lesson.Status == "CANCELLED" {
            return ErrInvalidState
        }

        attendance := Attendance{
            LessonID: lessonID, StudentAttended: req.StudentAttended,
            TeacherAttended: req.TeacherAttended, /* ... */
        }
        if err := tx.Create(&attendance).Error; err != nil { return err }

        if req.DeductLesson {
            if err := s.deductStudentBalance(tx, lesson, attendance); err != nil { return err }
        }
        if req.TeacherAttended {
            if err := s.addTeacherPayable(tx, lesson, attendance); err != nil { return err }
        }
        if err := s.writeLessonFinance(tx, lesson, attendance); err != nil { return err }

        lesson.Status = "COMPLETED"
        if err := tx.Save(&lesson).Error; err != nil { return err }

        return nil // 事务提交，任一步失败自动回滚
    })
}
```

### 10.4 定时任务实现要点

```go
func RegisterJobs(c *cron.Cron, svc *Services) {
    c.AddFunc("*/10 * * * *", svc.Job.LessonReminder)         // 每10分钟
    c.AddFunc("0 8 * * *", svc.Job.MorningReport)              // 每天08:00
    c.AddFunc("0 20 * * *", svc.Job.BalanceAlert)               // 每天20:00
    c.AddFunc("0 8 * * 1", svc.Job.OwnerWeeklyReport)           // 每周一08:00
    c.AddFunc("*/30 * * * *", svc.Job.NotificationRetry)        // 每30分钟
    c.AddFunc("0 3 * * *", svc.Job.AutoCloseLessons)            // 每天03:00
    c.AddFunc("0 2 * * *", svc.Job.AutoBackup)                  // 每天02:00
}
```

所有任务须包裹 panic-recover，单个任务异常不应影响其他任务和主服务运行；任务执行日志写入应用日志（非 notification_log），便于运维排查。

---

## 第十一章 前端技术架构

（技术栈与选型理由详见 v1.0 文档第九章，此处补充实施细节）

### 11.1 请求层封装

```typescript
// api/request.ts
import axios from 'axios';

const request = axios.create({ baseURL: '/api/v1', timeout: 15000 });

request.interceptors.request.use(config => {
  const token = useAuthStore().accessToken;
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

request.interceptors.response.use(
  res => {
    if (res.data.code !== 0) {
      // 统一错误提示（Naive UI message）
      window.$message?.error(res.data.message);
      return Promise.reject(res.data);
    }
    return res.data.data;
  },
  async err => {
    if (err.response?.status === 401) {
      // 尝试刷新 token，失败则跳转登录页
    }
    return Promise.reject(err);
  }
);
```

### 11.2 状态管理（Pinia Store 划分）

```
store/
  auth.ts       登录态、用户信息、Token
  app.ts        全局配置（系统信息、币种、时区）
  dictionary.ts 课程体系字典缓存（领域/方向/等级/标签，减少重复请求）
```

### 11.3 关键组件复用清单

```
<StudentSelector />       学生选择器（下拉搜索）
<TeacherSelector />       老师选择器
<EnrollmentSelector />    课程报名项目选择器（联动学生）
<CourseTrackCascader />   课程领域/方向/等级级联选择器
<MoneyInput />            金额输入框（自动格式化千分位）
<DateTimePicker />        日期时间选择器（统一时区处理）
<StatusTag />             状态标签（自动映射颜色）
<ConfirmDialog />         二次确认弹窗（危险操作统一样式）
```

---

## 第十二章 部署与发布方案

（与 v1.0 第十一章一致：All-in-One 单文件发布、四种部署模式 A/B/C/D、跨平台编译、config.yaml 完整示例，详见 v1.0 文档，此处不重复。）

### 12.1 发布检查清单（Release Checklist）

```
□ 前端 npm run build 无报错，产物体积检查（<5MB gzip）
□ go build 三平台交叉编译成功
□ 首次启动自动建库测试（删除 data.db 后启动，验证 migration 正常执行）
□ 默认管理员账号首次启动自动生成，密码随机生成并打印到控制台/日志
□ 健康检查接口 /healthz 返回正常
□ 定时任务在测试环境实际触发验证（可临时改短 Cron 间隔测试）
□ 备份/恢复完整流程验证
□ 版本号写入 /healthz 返回体，便于确认部署版本
```

---

## 第十三章 数据安全与备份

（与 v1.0 第十二章一致：安全措施表、APPI 合规要点、三种备份方式、Litestream 配置，详见 v1.0 文档。）

### 13.1 补充：密码找回机制

V1 无自助找回密码功能（系统不对外开放注册），Operator 忘记密码时由 Owner 在"账号管理"页面执行"重置密码"，系统生成随机密码，Owner 通过线下渠道（微信/当面）告知对方，首次登录强制要求修改密码。

### 13.2 补充：默认管理员初始化

系统首次启动（检测到 `user_account` 表为空）时，自动创建一个 Owner 账号：

```
username: admin
password: 随机生成16位强密码，仅在首次启动时打印到控制台和 logs/init.log
```

首次登录后系统强制弹出修改密码提示，不可跳过。

---

## 第十四章 通知与提醒系统（含模板原文）

### 14.1 六类定时任务（与 v1.0 第七章一致，此处补充完整邮件模板原文）

### 14.2 邮件模板原文

**模板 1：LESSON_REMINDER_STUDENT（学生课前提醒）**

```
主题：【上课提醒】{{.LessonDate}} {{.LessonTime}} 您的日语课即将开始

正文：
{{.StudentName}} 同学，您好：

您与 {{.TeacherName}} 老师的课程即将开始：

  课程：{{.TrackName}} · {{.LevelName}}
  时间：{{.LessonDate}} {{.LessonTime}}（{{.Duration}}分钟）
  方式：{{.MeetingType}}
  链接：{{.MeetingLink}}

请提前 5 分钟进入等待老师。如需调整时间，请联系教务老师。

祝学习愉快！
{{.OrgName}} 教务组
```

**模板 2：LESSON_REMINDER_TEACHER（老师课前提醒）**

```
主题：【上课提醒】{{.LessonDate}} {{.LessonTime}} 与 {{.StudentName}} 同学的课程

正文：
{{.TeacherName}} 老师，您好：

您与 {{.StudentName}} 同学的课程即将开始：

  课程：{{.TrackName}} · {{.LevelName}}
  时间：{{.LessonDate}} {{.LessonTime}}（{{.Duration}}分钟）
  方式：{{.MeetingType}}
  链接：{{.MeetingLink}}

如学生本节课学习目标或注意事项，请查看备注：{{.LessonNote}}

辛苦了！
{{.OrgName}} 教务组
```

**模板 3：BALANCE_ALERT_OPERATOR（余额不足预警，发给运营者）**

```
主题：【余额预警】{{.Count}} 位学生课时余额不足，请及时跟进续费

正文：
以下学生课时余额已低于预警线，请及时联系续费：

{{range .Students}}
  · {{.Name}}（{{.TrackName}}）剩余 {{.LessonBalance}} 课时 | 联系方式：{{.Contact}}
{{end}}

请登录系统查看详情并处理。
```

**模板 4：MORNING_REPORT_OPERATOR（教务晨报）**

```
主题：【教务晨报】{{.Date}} 今日待办摘要

正文：
早上好！以下是今日教务摘要：

■ 今日课程（共 {{.TodayLessonCount}} 节）
{{range .TodayLessons}}
  {{.Time}} {{.StudentName}} × {{.TeacherName}}（{{.TrackName}}）
{{end}}

■ 待确认课次（{{.PendingConfirmCount}} 个）
{{range .PendingConfirms}}
  {{.Date}} {{.StudentName}} × {{.TeacherName}}
{{end}}

■ 待续费学生（{{.BalanceAlertCount}} 人）
{{range .AlertStudents}}
  {{.Name}} 剩余 {{.LessonBalance}} 课时
{{end}}

■ 待结款老师（{{.PendingPayoutCount}} 人，合计 ¥{{.TotalUnpaid}}）

■ 发送失败通知（{{.FailedNotificationCount}} 条，请检查）

祝工作顺利！
```

**模板 5：OWNER_WEEKLY_REPORT（老板周报）**

```
主题：【经营周报】{{.WeekRange}} 数据摘要

正文：
本周经营数据摘要：

  总收入：¥{{.TotalRevenue}}
  总课酬支出：¥{{.TotalPayable}}
  毛利：¥{{.GrossProfit}}（毛利率 {{.ProfitMargin}}%）
  完成课次：{{.CompletedLessons}} 节 | 取消课次：{{.CancelledLessons}} 节
  新增学生：{{.NewStudents}} 人

各课程方向课时分布：
{{range .TrackDistribution}}
  {{.TrackName}}：{{.LessonCount}} 课时（占比 {{.Percentage}}%）
{{end}}

本月累计对比上月：收入 {{.MonthRevenueChange}}，毛利 {{.MonthProfitChange}}
```

### 14.3 通知发送失败处理流程

```
发送失败 → notification_log.status=FAILED，记录 error_msg
  ↓
每30分钟重试任务扫描 retry_count < 3 的 FAILED 记录
  ↓
重试成功 → status=SENT
重试失败达3次 → 保持 FAILED，等待人工处理（晨报会列出）
  ↓
Operator 在通知日志页面手动点击【重发】→ retry_count 重置为0，重新进入重试队列
```

---

## 第十五章 非功能性需求

### 15.1 性能要求

| 指标 | 目标值 |
|------|--------|
| 页面首屏加载（工作台）| < 2秒（本地/同城部署）|
| API 响应时间（P95）| < 500ms（不含邮件发送等异步操作）|
| 列表分页查询（1000条内）| < 300ms |
| 并发用户支持 | ≥ 10（V1场景下绰绰有余）|
| 数据库单表数据量预估上限 | lesson/attendance 表 5万+ 条不影响性能（有索引）|

### 15.2 可用性要求

```
目标可用性：单实例部署，无高可用集群要求（V1不做）
计划外停机容忍度：可接受，运营者非7×24小时依赖场景
故障恢复：备份可在30分钟内完成恢复（人工介入）
```

### 15.3 国际化与本地化

```
V1 界面语言：中文（简体）为主
通知邮件语言：可配置中文/日文/双语（系统设置中切换）
时区：系统级配置默认 Asia/Tokyo，学生/老师可各自设置时区（V2 考虑按收件人时区分别显示邮件时间）
金额显示：统一 JPY 为主显示币种，千分位分隔符，如 ¥12,000
日期格式：中文界面用 2026-07-03，邮件正文可用更口语化格式如"7月3日（周五）"
```

### 15.4 可访问性（Accessibility）

```
V1 基础要求：
  颜色对比度符合基本可读性（不做 WCAG AA 完整认证）
  表单错误提示明确到具体字段，不仅依赖颜色区分
  移动端触摸区域 ≥ 44×44px（符合移动端可用性基本标准）
```

### 15.5 浏览器兼容性

```
支持：Chrome/Edge/Safari 最新两个大版本，移动端 Safari (iOS) / Chrome (Android)
不刻意支持：IE（已停止维护，不做兼容）
```

### 15.6 数据规模压测建议（上线前）

```
建议在 Sprint 7 测试阶段，用脚本批量生成模拟数据（1000学生/100老师/1万课次/2万流水记录）验证：
  列表分页查询响应时间
  工作台聚合查询响应时间（涉及多表 JOIN 或子查询）
  Excel 导出大数据量时的内存占用与耗时
```


---

# 第四部分：项目管理

## 第十六章 开发计划与任务拆分

### 16.1 Sprint 总览

| Sprint | 周期 | 主题 |
|--------|------|------|
| S0 | 1周 | 工程脚手架 |
| S1 | 1周 | 基础档案与课程体系 |
| S2 | 1周 | 排课与课后确认 |
| S3 | 1周 | 账务体系 |
| S4 | 1周 | 通知与定时任务 |
| S5 | 1周 | 报表与图表 |
| S6 | 1周 | 导入导出与打包发布 |
| S7 | 1~2周 | 测试与上线 |

总计 8~9 周，AI 辅助开发可压缩至 4~5 周。

### 16.2 详细任务拆分（可直接建 Issue）

**Sprint 0：工程脚手架**
```
[ ] 初始化 Go 项目结构（backend/），配置 go.mod
[ ] 初始化 Soybean Admin 前端项目（frontend/admin/）
[ ] 配置 SQLite 连接 + GORM + goose migration 框架
[ ] 编写 6.2 节完整 DDL 的 migration 文件
[ ] 实现 JWT 登录/刷新/登出接口
[ ] 实现默认管理员账号首次启动初始化逻辑
[ ] 前端登录页 + 路由守卫 + Pinia auth store
[ ] 配置 go:embed 嵌入前端产物，验证单二进制可运行
[ ] 配置 viper 读取 config.yaml
[ ] 搭建 zerolog 日志框架
[ ] 编写 Makefile（build/dev/test 命令）
[ ] GitHub Actions CI 骨架（lint + build 验证）
```

**Sprint 1：基础档案与课程体系**
```
[ ] 课程体系四层 CRUD API（domain/track/level/skill_tag）
[ ] 课程体系初始化数据脚本（日语领域预置数据）
[ ] 学生 CRUD API + 家长 CRUD API
[ ] 老师 CRUD API + 能力 CRUD API + 可授时间 CRUD API
[ ] 课程报名 enrollment CRUD API
[ ] 师生安排 assignment 创建/换老师 API
[ ] 前端：课程体系维护页面（三栏联动）
[ ] 前端：学生列表 + 详情页（基础信息 + 学习项目Tab）
[ ] 前端：老师列表 + 详情页（能力与时间Tab）
[ ] 字段校验规则前后端双重实现（第八章8.1）
```

**Sprint 2：排课与课后确认**
```
[ ] lesson CRUD API + 时间冲突检测逻辑
[ ] 课次列表 + 日历视图数据接口
[ ] attendance 课后确认事务实现（10.3节示例）
[ ] lesson_finance / ledger 联动写入逻辑
[ ] 前端：排课列表页 + 新建/编辑课次弹窗
[ ] 前端：日历视图（月/周切换）
[ ] 前端：课后确认表单
[ ] 移动端：/mobile/today、/mobile/confirm 页面
[ ] 单元测试：课后确认事务的原子性测试（模拟中途失败场景）
```

**Sprint 3：账务体系**
```
[ ] student_payment CRUD API + 作废逻辑
[ ] teacher_payout 预览 + 提交 API
[ ] 账户流水查询 API（学生/老师）
[ ] 汇率快照与课时套餐管理 API
[ ] 前端：充值记录页 + 新建充值表单
[ ] 前端：结款记录页 + 预览结款明细
[ ] 前端：学生详情页账务Tab完善
[ ] 前端：老师详情页账务Tab完善
[ ] 移动端：/mobile/recharge 页面
[ ] money 工具包（decimal封装）单元测试
```

**Sprint 4：通知与定时任务**
```
[ ] NotificationSender 接口 + ResendSender/SmtpSender 实现
[ ] notification_template 数据初始化（5个模板，14.2节原文）
[ ] 六类定时任务实现（10.4节）
[ ] 通知日志 API + 重发逻辑
[ ] 前端：通知管理页（提醒规则 + 通知日志）
[ ] 移动端：/mobile/alerts 页面
[ ] 邮件模板变量渲染引擎（Go text/template）
[ ] 定时任务集成测试（临时改短Cron间隔验证触发）
```

**Sprint 5：报表与图表**
```
[ ] 工作台聚合数据 API（/reports/dashboard）
[ ] 各类图表数据 API（收入趋势/学生增长/方向分布/老师工作量）
[ ] 财务报表 API + Excel/PDF 导出
[ ] 前端：工作台页面完整实现（含ECharts图表）
[ ] 前端：数据图表页
[ ] 前端：财务报表页
```

**Sprint 6：导入导出与打包发布**
```
[ ] Excel 学生/老师批量导入 API（含校验+导入报告）
[ ] Excel 导出（学生/老师/课次列表）
[ ] 数据备份 API（手动备份/VACUUM INTO实现）
[ ] Litestream 集成（可选配置）
[ ] 三平台交叉编译脚本
[ ] Windows WinSW 服务配置 + install-service.bat
[ ] Linux systemd install.sh
[ ] 备份恢复完整流程实现（含二次确认）
[ ] README 编写（含快速开始指南）
```

**Sprint 7：测试与上线**
```
[ ] 按第十七章测试用例清单逐项执行
[ ] 批量模拟数据压测（1000学生/100老师/1万课次）
[ ] 安全测试（登录锁定/密码强度/越权访问）
[ ] 真实数据迁移演练（Excel导入现有数据）
[ ] 用户验收测试（运营者实际操作走查）
[ ] Bug修复与迭代
[ ] 正式环境部署 + 首次数据录入
[ ] 编写操作手册（供Operator日常使用参考）
```

---

## 第十七章 测试计划与用例清单

### 17.1 测试范围与策略

```
单元测试：核心业务逻辑（money计算、状态机流转、事务原子性）目标覆盖率 ≥ 60%
集成测试：API 端到端测试，覆盖主要业务流程
手动测试：UI交互、跨浏览器、移动端适配
压力测试：大数据量下的查询性能
```

### 17.2 核心测试用例清单

**账务事务测试（最高优先级）**

```
TC-01 课后确认后，学生余额正确扣减，老师应付正确增加
TC-02 课后确认事务中途模拟数据库异常，验证全部回滚（无部分写入）
TC-03 学生请假不扣课时，验证 ledger 无 LESSON_DEDUCT 记录
TC-04 老师缺席不产生应付款，验证 teacher_account_ledger 无对应记录
TC-05 充值作废后，验证余额正确冲正，且 ledger 生成 VOID 记录
TC-06 同一学生多个enrollment并行充值，验证互不干扰
TC-07 多币种充值，验证汇率折算金额正确（含小数精度边界测试，如0.1+0.2类问题）
```

**幂等性测试**

```
TC-08 课前提醒任务重复执行（模拟任务重叠触发），验证同一课次不会收到两次提醒
TC-09 充值接口重复提交相同请求（网络重试场景），验证是否需要幂等Key机制生效
```

**状态机测试**

```
TC-10 尝试编辑已COMPLETED的课次，验证返回42201错误
TC-11 尝试对VOIDED的充值记录再次作废，验证被拒绝
TC-12 换老师后，验证旧assignment正确ENDED，新assignment正确ACTIVE，且同一时刻只有一个ACTIVE
```

**权限测试**

```
TC-13 Operator 访问 /users 等 Owner 专属接口，验证返回40301
TC-14 未登录访问任意 /api/v1/* 接口（除/auth/login），验证返回40101
TC-15 Token 过期后自动刷新流程验证
```

**边界与异常输入测试**

```
TC-16 学生邮箱为空时，验证提醒邮件任务正确跳过该学生（不报错、不崩溃）
TC-17 排课时间设置为过去时间，验证不触发提醒任务
TC-18 Excel导入含重复邮箱数据，验证正确跳过并生成导入报告
TC-19 金额输入负数/超大数值，验证前后端校验拦截
TC-20 时区边界测试：日本时间23:30创建课次，验证UTC存储和显示换算正确
```

**定时任务测试**

```
TC-21 课前提醒：验证仅在设定时间窗口内（如课前20~40分钟）触发
TC-22 自动关闭过期课次：验证4小时缓冲期正确生效，不误关进行中课次
TC-23 通知重试：验证最多重试3次后停止，且重试间隔符合预期
```

**数据备份与恢复测试**

```
TC-24 手动触发备份，验证生成的备份文件可正常恢复
TC-25 备份过程中模拟并发写入，验证 VACUUM INTO 不影响正在进行的事务
TC-26 恢复操作前自动生成当前状态快照，验证快照存在
```

**跨平台部署测试**

```
TC-27 Windows双击运行，验证浏览器可正常访问
TC-28 Linux systemd 服务安装，验证开机自启
TC-29 三平台编译产物均可独立运行（无需额外依赖库）
```

---

## 第十八章 风险登记表

| 编号 | 风险描述 | 影响 | 可能性 | 应对措施 | 负责人 |
|------|---------|------|--------|---------|--------|
| RISK-01 | SQLite 并发写入锁冲突导致操作失败 | 中 | 低 | WAL模式+busy_timeout+应用层写串行化 | 后端开发 |
| RISK-02 | Resend 邮件送达率不达预期（进垃圾箱）| 中 | 中 | 验证发件域名SPF/DKIM，监控失败率 | 待定 |
| RISK-03 | 账务事务设计遗漏边界情况导致账目错误 | 高 | 中 | 按17.2节测试用例充分测试+人工调整流水兜底 | 后端开发 |
| RISK-04 | 时区处理错误导致提醒时间不准 | 中 | 中 | 统一UTC存储+集成测试专项验证 | 后端开发 |
| RISK-05 | Excel导入格式不规范导致导入失败或脏数据 | 低 | 高 | 提供标准模板+详细导入报告+失败行不影响成功行 | 前后端开发 |
| RISK-06 | 运营者遗忘做课后确认，账务滞后 | 中 | 高 | 工作台"待确认"醒目提示+晨报持续提醒 | 产品/运营 |
| RISK-07 | 备份未生效导致数据丢失 | 高 | 低 | 首次部署后立即验证备份恢复流程，定期演练 | 运维 |
| RISK-08 | Go/前端依赖库版本升级导致兼容性问题 | 低 | 中 | 锁定版本号(go.sum/package-lock.json)，谨慎升级 | 后端/前端开发 |
| RISK-09 | 单机部署无高可用，服务器故障造成服务中断 | 中 | 低 | 云端部署+Litestream实时备份，接受一定停机容忍度 | 运维 |
| RISK-10 | 需求理解偏差导致开发返工 | 中 | 中 | Sprint 0启动前与运营者再次确认附录D行动清单中的细节 | 产品 |

---

## 第十九章 版本演进规划

（与 v1.0 第十三章一致：V1.5/V2/V3 规划详见 v1.0 文档，此处仅列版本节奏总览）

```
V1  （本文档范围）  单机构后台管理，全功能账务与提醒，All-in-One发布
V1.5（功能补完）    独立移动端/PWA/Wails桌面版/历史数据导入
V2  （多端扩展）    老师端/学生端/调课审批/智能能力提示/多Operator协作
V3  （平台化）      Flutter App/支付接口/多实例管理控制台/AI匹配推荐
```

---

# 附录

## 附录 A：术语表

| 术语 | 英文/代码 | 说明 |
|------|-----------|------|
| 课程领域 | course_domain | 最顶层课程分类，如"日语" |
| 课程方向 | course_track | 领域下的学习路线，如"JLPT备考" |
| 课程等级 | course_level | 方向下的阶段，如"N3" |
| 课程报名 | enrollment | 学生对某课程方向的报名记录，含余额和课时 |
| 师生安排 | assignment | enrollment下具体由哪位老师授课的记录 |
| 课次 | lesson | 一次具体的上课安排 |
| 出勤记录 | attendance | 课后确认产生的记录，触发账务 |
| 账户流水 | ledger | 余额/应付款每次变动的流水记录 |
| 单课财务 | lesson_finance | 每节课的收费/课酬/毛利快照 |
| 幂等 | Idempotent | 同一操作重复执行结果一致，不产生副作用 |
| Minor Unit | 最小货币单位 | 如JPY的"円"、CNY的"分" |
| RBAC | Role-Based Access Control | 基于角色的权限控制 |
| APPI | 个人情报保护法 | 日本个人信息保护法规 |
| WAL | Write-Ahead Logging | SQLite的一种日志模式，提升并发性能 |

## 附录 B：技术选型速查（汇总自 v1.0）

详见 v1.0 文档附录 A/B/C：后端核心依赖清单、前端Admin模板颜值对比、部署资源对比。

## 附录 C：配置文件完整示例

（与 v1.0 文档第十一章 11.4 节一致，此处不重复）

## 附录 D：下一步行动清单

```
□ 业务对齐会议：确认课程方向/等级初始配置清单（日语具体开几个方向几个等级）
□ 业务对齐会议：确认课时套餐定价档次（至少3档：入门/标准/优惠）
□ 业务对齐会议：确认充值币种使用习惯占比（JPY独占 or CNY/JPY混用）
□ 业务对齐会议：确认邮件语言偏好（中文/日文/双语，是否按学生国籍差异化）
□ 业务对齐会议：确认部署偏好（优先本地Windows试运行 or 直接云端部署）
□ 技术准备：注册 Resend 账号，完成发件域名 SPF/DKIM 验证
□ 技术准备：开通 AWS Free Tier 账号或准备替代 VPS
□ 技术准备：搭建开发环境（Go 1.22+, Node 20+, VS Code + Go + Volar插件）
□ 技术准备：初始化 Git 仓库，配置 GitHub Actions
□ 启动 Sprint 0，目标两周内产出可登录、可运行的最小系统
```

---

*Zedu PRD v2.0（完整实装版）· 整理日期 2026-07-03*

*本文档整合了全部历史讨论（v0.1业务讨论稿、v0.2部署方案讨论、ChatGPT策划稿、v1.0正式PRD），新增完整DDL、全量API清单、页面级UI规格、状态机设计、通知模板原文、测试用例清单、风险登记表和细粒度任务拆分，可直接作为产品原型设计、数据库建库、前后端开发、测试用例编写的基准依据。*

*文档版本历史：*
- *v0.1/v0.2 讨论稿：业务背景、需求概要、初步技术选型*
- *v1.0 正式PRD：完整数据模型、API规范纲要、工程结构*
- *v2.0 完整实装版（本版）：细化至字段级、接口级、页面级、测试用例级，可直接进入开发阶段*
