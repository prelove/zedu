# Zedu 轻量级教培教务管理系统
## 产品需求文档（PRD）

> **文档版本**：v1.0
> **状态**：评审稿
> **整理日期**：2026-06-09
> **项目代号**：Zedu

---

## 文档说明

本文档是 Zedu 系统的正式产品需求文档（PRD），在以下材料基础上独立整合完成：
- 内部业务讨论稿（V0.1 ~ V0.2）
- ChatGPT 辅助策划稿（V0.1 业务架构讨论）
- 多轮技术选型与部署方案讨论

**整合原则**：吸收两稿优点，独立判断取舍，不做机械合并。本文档可直接用于后续原型设计、ER 图、API 设计、开发任务拆分和部署手册。

---

## 目录

1. 产品定位与目标
2. 用户角色与权限体系
3. 核心业务流程
4. 课程体系设计
5. 功能模块详述
6. 数据模型设计
7. 通知与提醒系统
8. 报表与数据图表
9. 前端设计规范
10. 后端技术架构
11. 部署与发布方案
12. 数据安全与备份
13. 版本演进规划
14. 开发计划
15. 风险与注意事项
16. 附录

---

## 第一章 产品定位与目标

### 1.1 背景

某日语爱好者以兼职形式运营日语学习撮合服务：联络学习者与老师，组织一对一或小班课程，向学生收费，向老师支付课酬，赚取差价。目前使用 Excel + 人工管理，核心痛点是：提醒容易遗漏、续费跟进被动、账务核算繁琐、信息分散难查。

Zedu 由此需求出发，定位为可复制的轻量级教培教务管理系统。

### 1.2 产品名称

```
Zedu
```

命名寓意：Zero-friction Education（零摩擦教育管理）。可展示为"泽度教务"或直接使用英文，各实例可自定义。

### 1.3 产品定位

Zedu 是一套**小而美的轻量级教培教务管理系统**，首期面向日语一对一/小班课程，架构上预留扩展至英语、留学辅导、艺术培训、体育培训、职业证书培训等场景的能力。

**核心设计词**：漂亮、轻量、好部署、账务清楚、提醒可靠、可复制。

### 1.4 当前版本边界

**V1 做的事：**

- 单机构单实例，运营者后台管理
- 学生/老师档案管理
- 课程体系基础维护（领域/方向/等级/能力标签）
- 学生课程报名（含多课程方向）
- 排课与课时管理
- 充值与多币种折算
- 课后确认与流水台账
- 自动邮件提醒（上课/续费/晨报/周报）
- 基础经营报表与 ECharts 图表
- Excel/PDF 导入导出
- All-in-One 单文件发布，多平台支持

**V1 不做的事：**

- 多租户 SaaS / 微服务架构
- 学生端/老师端独立登录
- 小程序（审核周期长，海外适用性低）
- 支付接口对接
- 复杂自动排课算法
- Flutter App / 真正离线同步
- 复杂班级管理 / AI 匹配推荐

### 1.5 目标规模

| 角色 | V1 目标规模 |
|------|------------|
| 学生 | ≤ 1,000 人 |
| 老师 | ≤ 100 人 |
| 运营账号 | 1 Owner + N Operator |
| 并发用户 | ≤ 10（运营者后台）|

### 1.6 多实例部署策略

V1 不做多租户，采用"一客一部署"：

```
zedu.abitcloud.org       → 机构A
friend1.abitcloud.org    → 机构B（复制部署）
custom-domain.com        → 机构C（独立域名 + 定制 Logo）
```

每个实例数据天然隔离，可独立升级、独立备份、独立定制。比多租户在当前阶段更合理，复杂度低，出问题影响范围小。

---

## 第二章 用户角色与权限体系

### 2.1 V1 角色定义

V1 实现两个后台角色。学生和老师 V1 不登录，只被动接收邮件通知。

#### Owner（老板）

```
查看全部经营数据与报表
查看收入、成本、毛利
接收老板周报（每周推送）
管理 Operator 账号
修改系统配置
管理数据备份与恢复
查看完整操作日志
```

#### Operator（教务）

```
管理学生与家长档案
管理老师档案与能力标签
维护课程方向和等级
管理学生课程报名
安排课程（排课）
录入充值记录
执行课后确认
处理换老师
发送或重发通知
查看通知日志
查看基础报表
导入导出数据
```

### 2.2 后续预留角色（V2+）

```
Teacher   老师端：查看课表、确认出勤、填可授时间、查看课酬
Student   学生端：查看余额、查看课表、提交调课申请
Parent    家长端：查看孩子课表与余额（可选）
```

### 2.3 权限实现原则

V1 采用简单 RBAC（Owner 包含 Operator 全部权限），不做细粒度字段级权限。所有写操作记录操作日志，便于追溯。

---

## 第三章 核心业务流程

### 3.1 学生建档与匹配

```
学生咨询/报名
  ↓
Operator 录入学生档案（姓名/联系方式/学习背景）
  ↓
录入家长联系方式（可选）
  ↓
创建学生课程报名项目（选课程方向/当前等级/目标等级）
  ↓
匹配合适老师，创建师生安排
  ↓
进入排课流程
```

### 3.2 老师建档

```
录入老师基础信息（姓名/联系方式）
  ↓
录入教学经历与证书（文本描述）
  ↓
设置可教课程方向与等级（teacher_capability）
  ↓
录入默认课酬（JPY）
  ↓
维护可授时间段（weekday + 时间段）
```

### 3.3 课程报名与多方向学习

**核心设计决策**：一个学生可以同时报名多个课程方向，每个方向可配不同老师。不采用"学生绑定单个老师"的简化模型。

```
王同学
  ├── 课程报名 A：日语 / JLPT / N3→N2 / 老师A
  ├── 课程报名 B：日语 / 会话 / 中级 / 老师B
  └── 课程报名 C：英语 / IELTS（V2 启用）
```

核心关系链：

```
学生 → 课程报名（enrollment）→ 师生安排（assignment）→ 具体课次（lesson）
```

### 3.4 充值与余额流程

```
学生完成线下付款（PayPay/微信/银行转账/现金）
  ↓
Operator 录入充值记录（保存原币种、金额、汇率、折算 JPY）
  ↓
增加对应课程报名项目的余额和课时
  ↓
生成学生账户流水记录
```

### 3.5 排课与课前提醒

```
Operator 创建课次（选学生/课程报名项目/老师/时间/上课链接）
  ↓
课次状态：SCHEDULED
  ↓
定时任务每 10 分钟扫描
  ↓
课前 30 分钟：发送提醒邮件给学生 + 老师（含上课链接）
  ↓
写入通知日志，更新 lesson.remind_sent_at（幂等标记）
```

### 3.6 课后确认与账务（同一事务）

```
课程结束 → Operator 确认出勤
  ↓（同一 SQLite 事务）
  ├── 写 attendance 记录
  ├── 写 student_account_ledger（扣减课时/余额）
  ├── 写 teacher_account_ledger（增加应付款）
  ├── 写 lesson_finance（收/支/毛利快照）
  ├── 更新 enrollment.lesson_balance（缓存）
  ├── 更新 teacher.unpaid_amount（缓存）
  └── 更新 lesson.status = COMPLETED
  ↓
判断余额是否低于预警阈值 → 触发续费提醒
```

### 3.7 换老师流程

```
Operator 在课程报名详情页操作换老师
  ↓
选择新老师（系统检查能力匹配，提示不强制拦截）
  ↓
同一事务：
  ├── 旧 assignment.status = ENDED，记录 end_date + reason
  └── 新建 assignment（新老师，ACTIVE，start_date=今日）
  ↓
enrollment 上的余额和课时不变（不随老师变动）
历史课次记录保留原老师快照
后续新建课次使用新老师
```

---

## 第四章 课程体系设计

### 4.1 设计动机

V1 主要是日语，但系统从设计之初就要支持课程多样性：
- 同一机构可能扩展英语、体育等课程
- 同一老师可能教多个方向，能力有变化
- 学生等级随学习进展变化（N4→N3→N2）
- 体育/艺术类课程可能按年龄、累计课时升级

因此建立四层课程体系：**领域 → 方向 → 等级 → 能力标签**。

### 4.2 四层课程体系

**课程领域（course_domain）**：最顶层业务分类

```
日语 / 英语 / 韩语 / 足球 / 钢琴 / 编程 / 留学辅导 / 职业证书
类型：LANGUAGE / SPORT / ART / ACADEMIC / OTHER
```

**课程方向（course_track）**：某领域下的学习路线，例如日语领域：

```
JLPT 备考 / 日常会话 / 商务日语 / 少儿日语 / 写作强化 / 面试日语
```

**课程等级（course_level）**：某方向下的阶段，例如 JLPT：

```
入门 → N5 → N4 → N3 → N2 → N1
```

附加属性：`min_age / max_age / min_lesson_hours / recommended_lesson_hours`，供体育艺术类课程按年龄/课时升级使用（V1 记录，V2 实现升级规则）。

**能力标签（skill_tag）**：具体能力点

```
语言类：词汇/语法/阅读/听力/口语/写作/综合
考试类：N1真题/IELTS写作/商务敬语
体育类：基础体能/技术动作/战术理解
```

### 4.3 老师能力模型

老师能力是多条记录（teacher_capability），不是单个字段：

```
老师A
  ├── 日语 / JLPT / N5 ← ACTIVE
  ├── 日语 / JLPT / N4 ← ACTIVE
  ├── 日语 / JLPT / N3 ← ACTIVE
  ├── 日语 / 会话 / 中级 ← ACTIVE
  └── 日语 / 会话 / 高级 ← PAUSED（培训中）
```

每条记录包含：有效期、认证状态（未认证/已认证）、能力标签。

**V1 策略**：不做强制拦截，排课时仅软提示。V2 可基于此实现智能推荐。

### 4.4 学生等级变化事件

等级变化必须记录事件，不能只修改字段：

```
王同学 N3 → N2
原因：JLPT N3 合格（2026-12-20）
事件类型：EXAM_PASS
```

事件类型：`ASSESSMENT / EXAM_PASS / HOURS_REACHED / AGE_REACHED / MANUAL`

V1 支持手动录入，V2 可实现规则自动触发。

---

## 第五章 功能模块详述

### 5.1 工作台（Dashboard）

**Owner 视图**

```
今日课程数 / 本月收入 / 本月毛利 / 活跃学生数（指标卡）
月度收入趋势（折线图）
近 30 天新增学生（柱状图）
各课程方向收入占比（饼图）
```

**Operator 视图**

```
今日课程列表（时间/学生/老师/链接，一键发提醒）
待续费学生列表（余额≤阈值，显示联系方式）
待结款老师列表（应付金额汇总）
待确认课次列表（已上课未确认）
失败通知列表（支持重发）
```

### 5.2 学生管理

**学生列表**：搜索、按状态筛选、Excel 导出、快速充值入口。

**学生详情页**：
- 基础信息区：姓名/邮箱/电话/国籍/时区/备注/家长信息
- 学习项目区：每个课程报名卡片（方向/等级/余额/课时/老师），操作：新建/换老师/暂停/结束
- 账务区：余额概览、充值记录、账户流水
- 上课记录区：历史课次列表
- 通知记录区：发给该学生的邮件历史

**Excel 批量导入**：提供模板，字段包含姓名/邮箱/电话/学习方向/当前等级/备注，上传后预览校验，确认后入库。

### 5.3 老师管理

**老师详情页**：
- 基础信息区：姓名/邮箱/电话/简介/证书经历
- 能力标签区：可教课程方向与等级列表（状态/认证状态）
- 可授时间区：周几+时间段，可多条，支持临时停课
- 课酬设置区：默认课酬（JPY/课时）
- 账务区：应付款汇总、结款记录、课时流水
- 带课记录区：历史带过的学生课程项目

### 5.4 课程体系维护

系统设置中管理基础数据：课程领域/方向/等级/能力标签（增删改，启用禁用，可调序）。V1 预置日语完整数据，Operator 可按需调整。

### 5.5 排课管理

**新建课次表单**：

```
选择学生 → 自动列出活跃课程报名项目 → 选择项目
自动代入当前老师（可修改）
选择上课时间（日期+时间，JST 显示）
填写时长（默认 60 分钟）
选择上课方式（微信群/腾讯会议/Zoom/其他）
填写上课链接
备注
```

**日历视图**：月/周视图，手机端自动降级为按日分组列表视图。

**课后确认**：确认学生出勤（是/否）、实际时长、老师备注。不出勤的课次支持 Operator 决定是否扣费。

### 5.6 充值管理

**新建充值**：

```
选择学生 → 选择课程报名项目
选择支付方式（PayPay/微信/银行转账/现金/其他）
输入原始金额和原始币种
输入汇率 → 系统自动计算 JPY 金额
选择课时套餐（预设下拉或手动输入）
录入付款日期 + 备注
```

充值记录不可物理删除，只能标记 `is_void`，保留审计痕迹。

### 5.7 结款管理

**新建结款**：选择老师 + 结款周期，系统自动汇总应付，显示课次明细（可勾选排除），填写实付金额和备注。

### 5.8 通知管理

**提醒规则配置**：课前提醒分钟数（默认 30）、余额预警阈值（默认 3 课时）、晨报时间（默认 08:00 JST）、老板周报时间。

**通知日志**：发送时间/类型/收件人/主题/状态，支持失败重发，支持手动触发。

---

## 第六章 数据模型设计

### 6.1 设计原则

- 所有金额以**整数分（minor unit）**存储（JPY：1円=1，CNY：1分=1）
- Go 层使用 `shopspring/decimal` 处理金额运算
- 所有时间字段存储 **UTC**，时区转换在应用层处理
- 业务枚举值使用 `TEXT`，不依赖数据库 ENUM（保持跨数据库兼容）
- 每张业务表预留 `extra_json TEXT` 字段
- 软删除使用 `deleted_at` 字段（GORM 约定）
- 禁用数据库触发器承载业务逻辑

### 6.2 系统与权限表

```sql
user_account        用户账号（id, username, password_hash, role, status, locked_until...）
system_config       系统配置键值对（key PK, value, updated_by, updated_at）
operation_log       操作日志（operator_id, action, target_type, target_id, detail_json, ip_addr）
backup_log          备份记录（file_name, file_size, trigger_type, status, error_msg）
```

### 6.3 学生与老师

```sql
student             学生（id, name, name_jp, email, phone, nationality, timezone, status, extra_json）
parent              家长（id, student_id, name, email, phone, relationship, is_primary）
teacher             老师（id, name, name_jp, email, phone, bio, default_rate_jpy, status, extra_json）
teacher_availability  可授时间（teacher_id, weekday 0-6, start_time, end_time, effective_from, effective_to）
teacher_capability    老师能力（teacher_id, domain_id, track_id, level_id, skill_tag_codes JSON,
                                status ACTIVE/PAUSED/ENDED, verified 0/1, effective_from, effective_to）
```

### 6.4 课程体系

```sql
course_domain       领域（id, name, code, type LANGUAGE/SPORT/ART/ACADEMIC/OTHER, sort_order, enabled）
course_track        方向（id, domain_id, name, code, sort_order, enabled）
course_level        等级（id, track_id, name, code, sort_order, min_age, max_age,
                         min_lesson_hours, recommended_lesson_hours, enabled）
skill_tag           能力标签（id, domain_id, name, code, sort_order, enabled）
```

### 6.5 学生学习项目

```sql
student_course_enrollment   课程报名（id, student_id, domain_id, track_id,
                              current_level_id, target_level_id,
                              enrollment_type ONE_TO_ONE/GROUP/TRIAL,
                              status ACTIVE/PAUSED/COMPLETED/CANCELLED,
                              charge_per_lesson_jpy,        -- 每课次收费（JPY 整数分）
                              lesson_balance,               -- 剩余课时缓存（REAL）
                              balance_jpy,                  -- 余额缓存（JPY 整数分）
                              started_at, ended_at, extra_json）

student_teacher_assignment  师生安排（id, enrollment_id, student_id, teacher_id,
                              role_type MAIN/SUBSTITUTE/ASSISTANT,
                              rate_jpy,                     -- 该安排下老师课酬（覆盖默认）
                              status ACTIVE/PAUSED/ENDED,
                              start_date, end_date, reason）

student_learning_path       学习路径（id, enrollment_id, student_id,
                              from_level_id, current_level_id, target_level_id,
                              goal_type EXAM/CONVERSATION/BUSINESS/HOBBY/COMPETITION,
                              target_exam_name, target_exam_date,
                              status ACTIVE/COMPLETED/CHANGED/PAUSED）

student_level_event         等级变化事件（id, student_id, enrollment_id, learning_path_id,
                              from_level_id, to_level_id,
                              event_type ASSESSMENT/EXAM_PASS/HOURS_REACHED/AGE_REACHED/MANUAL,
                              event_date, evidence_note, operator_id）
```

### 6.6 班级（预留，V1 数据结构，V2 功能实现）

```sql
class_group         班级（id, name, domain_id, track_id, level_id, main_teacher_id,
                         status ACTIVE/PAUSED/FINISHED,
                         charge_per_lesson_jpy, rate_per_lesson_jpy, max_students）
class_group_member  班级成员（id, class_group_id, student_id, enrollment_id, join_date, leave_date, status）
```

### 6.7 排课与上课

```sql
lesson              课次安排（id, lesson_no, enrollment_id, class_group_id,
                              student_id, teacher_id, domain_id, track_id, level_id,
                              lesson_topic,                 -- 本节课主题
                              scheduled_start_at,           -- UTC
                              scheduled_end_at, duration_min, timezone,
                              meeting_type WECHAT/TENCENT/ZOOM/OTHER,
                              meeting_link,
                              status SCHEDULED/REMINDED/COMPLETED/CANCELLED,
                              remind_sent_at）              -- 幂等标记

attendance          出勤记录（id, lesson_id, enrollment_id, student_id, teacher_id,
                              actual_start_at, actual_end_at,
                              student_attended 0/1, teacher_attended 0/1,
                              lesson_deducted REAL,         -- 实际扣除课时（支持 0.5）
                              charge_jpy,                   -- 实际收费（JPY 整数分）
                              teacher_pay_jpy,              -- 实际课酬（JPY 整数分）
                              teacher_note, operator_note, confirmed_by, confirmed_at）
```

### 6.8 账务体系

```sql
student_payment     充值记录（id, payment_no, student_id, enrollment_id,
                              original_amount,              -- 原始金额（decimal 字符串）
                              original_currency,            -- CNY/USD/JPY...
                              fx_rate_to_jpy,              -- 折算汇率（decimal 字符串）
                              amount_jpy,                   -- 折算后 JPY（整数分）
                              lessons_added REAL,           -- 增加课时
                              package_name,                 -- 套餐名（如"20课时包"）
                              payment_method WECHAT/PAYPAY/BANK/CASH/OTHER,
                              paid_at, operator_id,
                              status CONFIRMED/VOIDED, voided_at, void_reason）

teacher_payout      结款记录（id, payout_no, teacher_id, period_start, period_end,
                              lesson_count REAL, amount_jpy, actual_amount_jpy,
                              payment_method, paid_at, operator_id）

student_account_ledger  学生流水（id, student_id, enrollment_id,
                              biz_type RECHARGE/LESSON_DEDUCT/REFUND/ADJUST/VOID,
                              amount_jpy_delta,             -- 变动（可负）
                              lesson_delta,                 -- 变动课时（可负）
                              balance_jpy_after,            -- 变动后余额快照
                              lesson_balance_after,         -- 变动后课时快照
                              related_payment_id, related_lesson_id, operator_id）

teacher_account_ledger  老师流水（id, teacher_id,
                              biz_type LESSON_PAYABLE/PAYOUT/ADJUST,
                              amount_jpy_delta, unpaid_amount_after,
                              related_lesson_id, related_payout_id, operator_id）

lesson_finance      单课财务快照（id, lesson_id, enrollment_id, student_id, teacher_id,
                              charge_jpy, teacher_pay_jpy,
                              gross_profit_jpy）            -- 毛利 = charge - teacher_pay

fx_rate_snapshot    汇率快照（id, from_currency, to_currency, rate, source MANUAL, recorded_at）
```

### 6.9 通知体系

```sql
notification_template  通知模板（id, code 唯一, type EMAIL, language ZH/JA/EN/BILINGUAL,
                                 subject_tpl, body_tpl HTML, enabled, updated_by）

notification_log      通知日志（id, template_code, type, lesson_id,
                                recipient_id, recipient_type STUDENT/TEACHER/OPERATOR,
                                recipient_email, recipient_name,
                                subject, body_preview,
                                status PENDING/SENT/FAILED/CANCELLED,
                                error_msg, retry_count, sent_at）
```

**合计：29 张表**（含预留班级表）

---

## 第七章 通知与提醒系统

### 7.1 通知渠道

**V1 仅支持邮件**：
- 主通道：Resend API（免费 3,000 封/月，无需配置 SMTP）
- 备用通道：SMTP（Fallback，Resend 失败时自动切换）

通知发送器接口抽象，便于后续扩展：

```go
type NotificationSender interface {
    Send(ctx context.Context, msg *Message) error
}
// V1 实现：ResendSender / SmtpSender
// V2+ 预留：LineSender / WechatSender / SmsSender
```

### 7.2 六类定时任务

**任务 1：课前提醒**（每 10 分钟轮询）

查询条件：`remind_sent_at IS NULL AND scheduled_start_at BETWEEN NOW()+20min AND NOW()+40min AND status=SCHEDULED`

操作：发提醒邮件给学生（含上课链接）+ 老师（含学生信息）→ 更新 `remind_sent_at`（幂等保证只发一次）

**任务 2：教务晨报**（每天 08:00 JST，发给 Operator）

内容：今日课程列表 / 待确认课次 / 余额不足学生 / 待结款老师 / 失败通知列表

**任务 3：余额预警扫描**（每天 20:00 JST，发给 Operator）

触发条件（满足其一）：
- `lesson_balance ≤ 配置阈值（默认 3）`
- `balance_jpy < 下一节课 charge_per_lesson_jpy`
- `未来 7 天已排课次数 > lesson_balance`

*课后确认时也实时触发一次余额判断，不等待每日扫描。*

**任务 4：老板周报**（每周一 08:00 JST，发给 Owner）

内容：上周收入/课酬/毛利 / 完成课次数 / 新增/流失学生 / 各课程方向分布 / 本月累计对比

**任务 5：通知重试**（每 30 分钟）

查询 `status=FAILED AND retry_count < 3`，重新发送，最多重试 3 次。

**任务 6：自动关闭过期课次**（每天 03:00 JST）

将 `scheduled_end_at < NOW()-4h` 且状态为 SCHEDULED/REMINDED 的课次自动置为 COMPLETED（仅状态变更，不触发账务）。4 小时缓冲避免误关进行中的课次。

### 7.3 必须实现的邮件模板

| 模板代码 | 发送对象 | 场景 |
|---------|---------|------|
| LESSON_REMINDER_STUDENT | 学生 | 课前提醒（含上课链接）|
| LESSON_REMINDER_TEACHER | 老师 | 课前提醒（含学生信息）|
| BALANCE_ALERT_OPERATOR | Operator | 余额不足预警汇总 |
| MORNING_REPORT_OPERATOR | Operator | 每日晨报 |
| OWNER_WEEKLY_REPORT | Owner | 老板周报 |

邮件中手机号脱敏（`138****1234`）。模板存储在数据库，支持后台编辑，变量使用 `{{.VariableName}}` 占位符。

---

## 第八章 报表与数据图表

### 8.1 工作台 ECharts 图表

| 图表 | 类型 | 内容 |
|-----|------|------|
| 月度收入趋势 | 折线图 | 近 12 个月收入/课酬/毛利对比 |
| 本月收支构成 | 柱状图 | 收入/课酬/毛利分周展示 |
| 课程方向分布 | 饼图 | 各方向课时占比 |
| 学生增长曲线 | 面积图 | 近 6 个月累计活跃学生数 |
| 老师带课分布 | 横向柱状图 | 各老师本月课时数 |

### 8.2 导出报表

| 报表 | 格式 | 内容 |
|-----|------|------|
| 月度收支报表 | Excel / PDF | 收入/课酬/毛利/差价按月汇总 |
| 学生账单 | PDF | 单个学生充值与消费明细 |
| 老师结款单 | PDF | 单个老师课时与应付明细 |
| 学生列表 | Excel | 全字段，支持筛选 |
| 老师列表 | Excel | 全字段 |
| 课次记录 | Excel | 指定时间段内全部课次 |

图表截图使用 ECharts 内置导出（PNG）。

---

## 第九章 前端设计规范

### 9.1 技术选型

```
框架：Vue 3 + Vite + TypeScript
UI 组件库：Naive UI（Soybean Admin 内置）
Admin 模板：Soybean Admin（Vue3 + Naive UI 版本）
图表：ECharts 5（Soybean 内置集成）
状态管理：Pinia
CSS：UnoCSS
HTTP 客户端：axios
```

### 9.2 选择 Soybean Admin 的理由

| 维度 | Soybean Admin | RuoYi-Vue3 |
|------|--------------|------------|
| 视觉风格 | 现代、清新、卡片化 | 传统后台风格 |
| 主题系统 | 动态主题，实时切换主色/圆角 | 固定主题 |
| 移动端适配 | 较好，响应式布局 | 一般 |
| ECharts | 内置完整支持 | 需额外配置 |
| TypeScript | 全量支持 | 部分支持 |
| 构建速度 | Vite 8，极快 | 相对慢 |
| 后端绑定 | 不绑定，纯前端 | 强绑若依后端 |

备选：Art Design Pro（Element Plus，暗色模式精致）

### 9.3 手机端策略

V1 不单独开发移动端项目，在 Soybean Admin 内提供移动优先页面：

```
/mobile/today      今日课程（卡片列表 + 快速确认）
/mobile/confirm    待确认课次
/mobile/recharge   快速充值录入
/mobile/alerts     待续费学生列表
```

手机端原则：少表格、多卡片、大按钮、少字段、底部快捷导航。

V1.5 可视情况独立拆出 `/m` 移动端（TDesign Mobile Vue / Varlet / NutUI）。

---

## 第十章 后端技术架构

### 10.1 技术栈

| 层次 | 选型 | 说明 |
|------|------|------|
| 语言 | Go 1.22+ | 单二进制，~30MB 内存，跨平台编译 |
| Web 框架 | Gin | 主流，中间件生态完善 |
| ORM | GORM | 支持 SQLite/MySQL/PostgreSQL，迁移无缝 |
| 数据库 | SQLite（`modernc.org/sqlite`）| 纯 Go 驱动，无 CGO，跨平台 |
| 数据库迁移 | goose | SQL 文件版本管理，按 dialect 分目录 |
| 认证 | JWT + Refresh Token | Access 60 分钟，Refresh 14 天 |
| 金额运算 | shopspring/decimal | 精确 decimal，避免浮点 |
| 定时任务 | robfig/cron v3 | 标准 Cron 表达式 |
| 邮件主通道 | Resend Go SDK | 零配置，API Key 驱动 |
| 邮件备用 | net/smtp | 标准库 Fallback |
| Excel | excelize | Go 生态最成熟的 Excel 库 |
| PDF | gofpdf | 报表/账单 PDF 生成 |
| 静态嵌入 | go:embed | 前端 dist 打包进二进制 |
| 日志 | zerolog | 结构化日志，性能优秀 |
| 配置 | viper | YAML + 环境变量覆盖 |
| 备份（可选）| Litestream | SQLite 流式备份到 S3/R2 |
| 桌面（可选）| Wails v2 | Go + Vue 桌面应用，系统托盘 |

### 10.2 工程目录结构

```
zedu/
├── backend/
│   ├── cmd/zedu-server/main.go
│   ├── internal/
│   │   ├── app/          应用初始化、路由注册、中间件
│   │   ├── auth/         JWT 认证、登录、Refresh Token
│   │   ├── user/         用户账号管理
│   │   ├── student/      学生档案
│   │   ├── parent/       家长信息
│   │   ├── teacher/      老师档案与能力
│   │   ├── course/       课程体系
│   │   ├── enrollment/   课程报名与师生安排
│   │   ├── lesson/       排课与课后确认
│   │   ├── finance/      账务（充值/结款/流水）
│   │   ├── notification/ 通知模板与发送
│   │   ├── report/       报表与图表数据
│   │   ├── system/       系统配置
│   │   ├── job/          定时任务（6 类）
│   │   ├── audit/        操作日志
│   │   └── backup/       数据备份
│   ├── pkg/
│   │   ├── response/     统一响应格式
│   │   ├── pagination/   分页工具
│   │   ├── validator/    入参校验
│   │   ├── money/        金额工具（decimal 封装）
│   │   ├── datetime/     时区转换工具
│   │   ├── crypto/       密码哈希、JWT
│   │   └── errors/       业务错误码
│   ├── migrations/
│   │   ├── sqlite/
│   │   ├── mysql/        （V2 备用）
│   │   └── postgres/     （V3 备用）
│   ├── web/admin-dist/   前端构建产物（go:embed 目标）
│   ├── config/config.example.yaml
│   ├── go.mod
│   └── go.sum
├── frontend/admin/
│   ├── src/
│   │   ├── api/          后端接口调用
│   │   ├── views/        页面组件（dashboard/student/teacher/course/enrollment/lesson/finance/notification/report/system/mobile）
│   │   ├── router/
│   │   ├── store/        Pinia 状态
│   │   ├── components/   公共组件
│   │   └── utils/
│   ├── package.json
│   └── vite.config.ts
├── deploy/
│   ├── zedu.service           systemd 服务配置
│   ├── zedu-service.xml       WinSW 配置
│   ├── nginx.conf             反向代理配置
│   ├── litestream.yml         备份配置示例
│   ├── install.sh             Linux 一键安装
│   └── install-service.bat   Windows 服务安装
├── scripts/
│   ├── build.sh               多平台构建
│   ├── build.ps1              Windows 构建
│   └── release.sh             GitHub Release 发布
├── docs/
│   ├── prd.md / api.md / database.md / deployment.md
├── Makefile
└── README.md
```

### 10.3 模块代码规范

```
student/
├── model.go       GORM Model（对应数据库表）
├── dto.go         API 入参/出参 DTO（不直接暴露 Model）
├── handler.go     HTTP Handler（只处理请求/响应）
├── service.go     业务逻辑与事务（核心规则在此）
├── repository.go  数据库访问（GORM 查询封装）
├── routes.go      路由注册
└── errors.go      模块级业务错误码
```

约束：业务逻辑不写在 handler；数据库操作不散落在 service 外部；Model 不直接作为 API 响应。

### 10.4 API 规范

**基础前缀**：`/api/v1`

**RESTful 命名**（资源名复数）：

```
GET    /api/v1/students
POST   /api/v1/students
GET    /api/v1/students/{id}
PUT    /api/v1/students/{id}
DELETE /api/v1/students/{id}
```

**业务动作**：

```
POST /api/v1/lessons/{id}/confirm
POST /api/v1/lessons/{id}/cancel
POST /api/v1/students/{id}/payments
POST /api/v1/teachers/{id}/payouts
POST /api/v1/notifications/{id}/resend
POST /api/v1/enrollments/{id}/assignments/change-teacher
```

**API 分组**：`/api/v1/auth` / `users` / `students` / `parents` / `teachers` / `courses` / `enrollments` / `lessons` / `finance` / `reports` / `notifications` / `system` / `backup`；`/healthz`（无版本前缀）

**统一响应格式**：

```json
{ "code": 0, "message": "success", "data": {}, "traceId": "20260609-xxxx" }
{ "code": 40001, "message": "余额不足", "traceId": "20260609-xxxx" }
```

**版本策略**：非破坏性变更在 v1 追加；破坏性变更启用 `/api/v2`。

### 10.5 SQLite 关键配置

启动时必须执行：

```sql
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;
```

跨数据库兼容原则：避免 SQLite 独有语法；避免数据库触发器承载业务；枚举用 TEXT；时间用 DATETIME 字符串（UTC）；repository 层隔离数据库访问；migration 按 dialect 分目录。

---

## 第十一章 部署与发布方案

### 11.1 All-in-One 设计

核心理念：**一个压缩包解决所有问题**。

```
go:embed 将前端 dist 产物打包进 Go 二进制
modernc.org/sqlite 纯 Go SQLite 驱动，无 CGO
内嵌 HTTP 服务器 + 静态资源服务
内嵌定时任务调度器
首次运行自动创建数据库并执行 migration
```

**发布包结构**：

```
zedu/
├── zedu-server(.exe)    主程序（含前端、API、定时任务）
├── config.yaml          配置文件（首次运行自动生成模板）
└── data/zedu.db         数据库（首次运行自动创建）
```

### 11.2 四种部署模式

**模式 A：Windows 双击运行（最简单）**
双击 `zedu-server.exe`，浏览器访问 `http://localhost:8080`，关窗即停。适合个人本地使用。

**模式 B：Windows Service（长期运行）**
使用 WinSW（微软维护，单 exe + xml，无需 .NET）：

```xml
<service>
  <id>zedu</id><name>Zedu 教务管理系统</name>
  <executable>%BASE%\zedu-server.exe</executable>
  <onfailure action="restart" delay="10 sec"/>
  <startmode>Automatic</startmode>
</service>
```

```bat
winsw install zedu-service.xml && winsw start zedu-service.xml
```

**模式 C：Linux / AWS 云端（推荐生产）**
AWS t2.micro（1 核 1G RAM）完全够用，内存余量充足。

```bash
systemctl enable zedu && systemctl start zedu
```

配合 Nginx 反向代理 + Let's Encrypt SSL 提供 HTTPS 访问。

**模式 D：Wails 桌面应用（可选，对外分发）**
Wails v2 打包为真正桌面应用（系统 WebView，非 Electron），安装包约 20MB，系统托盘图标。
Wails 优于 Tauri：后端用 Go，与本项目一致，零额外学习成本。

### 11.3 跨平台编译

```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o dist/zedu_windows_amd64.exe .
GOOS=linux   GOARCH=amd64 go build -ldflags="-w -s" -o dist/zedu_linux_amd64 .
GOOS=darwin  GOARCH=arm64 go build -ldflags="-w -s" -o dist/zedu_darwin_arm64 .
```

`git tag v1.x` 触发 GitHub Actions 自动构建，产物发布到 GitHub Releases。

### 11.4 配置文件

```yaml
server:
  host: 0.0.0.0
  port: 8080
  public_url: https://zedu.abitcloud.org

app:
  name: Zedu
  timezone: Asia/Tokyo
  default_currency: JPY

database:
  driver: sqlite           # sqlite / mysql / postgres
  dsn: ./data/zedu.db

auth:
  jwt_secret: change-me-in-production
  access_token_minutes: 60
  refresh_token_days: 14

mail:
  primary: resend          # resend / smtp
  resend:
    api_key: re_xxxxxxxxxx
    from_email: noreply@abitcloud.org
    from_name: Zedu 教务
  smtp:                    # Fallback
    host: smtp.gmail.com
    port: 587
    username: ""
    password: ""

lesson:
  reminder_minutes: 30
  balance_alert_lessons: 3

backup:
  auto_enabled: true
  cron: "0 0 2 * * *"
  path: ./backup
  retention_days: 30
  litestream_enabled: false
  s3_bucket: ""
  s3_access_key: ""
  s3_secret_key: ""
  s3_endpoint: ""          # Cloudflare R2 endpoint
```

---

## 第十二章 数据安全与备份

### 12.1 安全措施

| 项目 | 实现方式 |
|------|---------|
| 密码存储 | bcrypt 哈希，cost=12 |
| 登录防暴力 | 连续失败 5 次，锁定 15 分钟 |
| Token 管理 | JWT（60 分钟）+ Refresh Token（14 天）|
| 操作审计 | 所有写操作写入 operation_log |
| 隐私保护 | 手机号脱敏（138****1234）|
| 配置安全 | JWT Secret 支持从环境变量注入 |
| HTTPS | 生产环境必须启用，Let's Encrypt 免费证书 |

### 12.2 数据合规（日本 APPI）

- 学生/老师联系方式不在邮件正文完整暴露
- 系统 V1 不对公众开放
- V3 增加数据删除/导出功能，支持 APPI 数据权利请求

### 12.3 备份策略

| 方式 | 实现 | 适用场景 |
|------|------|---------|
| 手动备份 | 系统设置页一键下载 `.zip`（db + config）| 随时按需 |
| 自动本地备份 | 每天 02:00，`VACUUM INTO` 方式，保留 30 天 | 本地运行 |
| Litestream 流式备份 | 实时同步到 S3/Cloudflare R2，每 5 秒一次 | 云端强烈推荐 |

**重要**：不能直接 `cp` 正在写入的 db 文件，必须使用 `VACUUM INTO` 或 SQLite Backup API。

Cloudflare R2 免费 10GB，对本系统足够多年使用。

---

## 第十三章 版本演进规划

### 13.1 V1（当前目标）

单机构后台管理，Owner + Operator，全功能账务与提醒，All-in-One 单文件发布。

### 13.2 V1.5（功能补完）

独立移动端 `/m`（TDesign Mobile Vue / Varlet）/ PWA / Wails 桌面版 / 更完整报表 / Excel 历史数据导入

### 13.3 V2（多端扩展）

```
老师端：查看课表/确认出勤/填可授时间/查看课酬
学生端：查余额/查课表/调课申请/续费申请
调课申请审批流程
老师能力智能提示（排课时自动过滤不合适的老师）
不同课程方向独立定价
LINE / WeChat 通知通道（日本市场 LINE 优先）
多人 Operator 协作
```

### 13.4 V3（平台化）

```
Flutter App（iOS + Android）
PayPay / Stripe 支付接口
多实例管理控制台
对外招生门户 + SEO
AI 老师匹配推荐（基于能力标签）
AI 学习报告生成
课程成长路径可视化
等级升级规则自动执行
```

### 13.5 扩展时机参考

| 触发条件 | 建议扩展 | 版本 |
|----------|---------|------|
| 学生 > 200，续费跟进成本高 | 学生自助端 | V1.5/V2 |
| 老师 > 30，排课沟通成本高 | 老师自助端 | V2 |
| 手机操作频率高 | 独立移动端 | V1.5 |
| 课程类型扩展 | 多领域（架构已支持）| V1/V2 |
| 月流水达规模 | 支付接口 | V3 |
| 想对外招生 | 门户 + SEO | V3 |
| 多机构管理 | 多实例控制台 | V3 |

---

## 第十四章 开发计划

### 14.1 Sprint 规划

| Sprint | 周期 | 核心内容 | 交付目标 |
|--------|------|---------|---------|
| S0 | 1 周 | 工程脚手架 + SQLite + Soybean Admin + 登录认证 | 最小可运行框架 |
| S1 | 1 周 | 课程体系维护 + 学生/老师档案 CRUD + 课程报名 | 数据录入可用 |
| S2 | 1 周 | 排课管理 + 日历视图 + 课后确认 + 账务事务 | 核心业务流跑通 |
| S3 | 1 周 | 充值录入 + 结款 + 流水台账 + 多币种折算 | 完整账本逻辑 |
| S4 | 1 周 | Resend 邮件集成 + 六类定时任务 + 通知日志 | 自动提醒上线 |
| S5 | 1 周 | 工作台仪表盘（ECharts）+ 报表导出（Excel/PDF）| 系统完整可用 |
| S6 | 1 周 | Excel 导入 + 数据备份 + All-in-One 打包 + 多平台编译 | 可发布版本 |
| S7 | 按需 | 测试 + Bug 修复 + 数据迁移 + 文档 + 正式上线 | V1 正式发布 |

总计约 **7~8 周**，以 AI 辅助开发（Claude Code / Cursor / Windsurf）为主力，可压缩至 4~5 周。

### 14.2 AI 辅助开发策略

1. 给定 ER 图 + 表结构 → AI 生成 GORM Model + Gin Handler + Service 骨架
2. 在骨架上填充业务逻辑（差价计算/课时扣减/幂等控制）
3. 参考 Soybean Admin 示例 → AI 生成 Vue 页面组件，前后端联调
4. 邮件模板 → AI 生成 HTML，Resend 控制台预览效果

### 14.3 关键开发注意事项

**账务事务**：课后确认的 6 个写操作必须同一事务，任意失败全部回滚。充值/结款只做软删除。

**幂等设计**：`remind_sent_at IS NULL` 作为提醒幂等门控；充值/结款接口支持幂等 Key（payment_no/payout_no）。

**时区处理**：数据库存 UTC，Go 层按 `config.app.timezone` 转换；定时任务触发时间基于系统时区；邮件模板可选双时区显示。

**金额处理**：Go 层全程 `shopspring/decimal`，最终转整数分写入数据库；报表层转回 decimal 展示。

**SQLite 并发**：启动执行三条 PRAGMA；写操作串行化，避免 `SQLITE_BUSY`。

---

## 第十五章 风险与注意事项

### 15.1 业务层面

| 事项 | 说明 |
|------|------|
| 数据迁移 | Excel 提供标准模板，S6 实现批量导入；迁移前须人工核对余额 |
| 邮件接受度 | 提醒邮件要简洁，主题让收件人一眼识别；Resend 正式域名有助于送达率 |
| 老师参与意愿 | V1 老师只收邮件，门槛最低；提醒频率需合理，避免被屏蔽 |
| 学生隐私（APPI）| 邮件不完整暴露手机号；数据不对外开放；V3 增加数据删除功能 |
| 付款凭证 | 系统充值记录为内部台账，非法律凭证；建议另行留存转账截图 |
| 汇率混用 | 汇率手动维护，不做自动抓取（避免依赖外部服务）|
| 业务连续性 | 本地运行有宕机风险，推荐云端部署 + Litestream 备份 |

### 15.2 技术层面

| 事项 | 说明 |
|------|------|
| SQLite 扩展边界 | 1,000 学生无压力；> 5,000 且并发增加时，GORM 改驱动迁移到 PostgreSQL |
| Resend 额度 | 估算约 1,830 封/月，在免费 3,000 封范围内 |
| 定时任务精度 | 课前提醒在课前 20~40 分钟内发出，可接受 |
| 前端 build 嵌入 | Makefile 中 `go generate` 触发前端构建，再 `go build` 打包，确保发布一致性 |
| 跨数据库迁移 | goose 按 dialect 分目录；GORM 只换 driver 和 dsn；业务代码不变 |

---

## 附录

### 附录 A：后端核心依赖

```
github.com/gin-gonic/gin              Web 框架
gorm.io/gorm                          ORM
modernc.org/sqlite                    SQLite 驱动（无 CGO）
github.com/pressly/goose/v3           数据库迁移
github.com/robfig/cron/v3             定时任务
github.com/resend/resend-go/v2        邮件（Resend）
github.com/shopspring/decimal         精确金额运算
github.com/golang-jwt/jwt/v5          JWT
github.com/spf13/viper                配置管理
github.com/rs/zerolog                 结构化日志
github.com/xuri/excelize/v2           Excel 读写
github.com/jung-kurt/gofpdf           PDF 生成
```

### 附录 B：前端 Admin 模板颜值对比

| 名称 | Stars | 组件库 | 颜值 | 移动端 | 图表 | 推荐 |
|------|-------|--------|------|--------|------|------|
| **Soybean Admin** | 14k+ | Naive UI | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ECharts 内置 | ✅ 首选 |
| Art Design Pro | 5k+ | Element Plus | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ECharts 内置 | 备选 |
| Vue Pure Admin | 15k+ | Element Plus | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ECharts 内置 | 备选 |
| RuoYi-Vue3 | 6k+ | Element Plus | ⭐⭐⭐ | ⭐⭐⭐ | ECharts | 不推荐 |

### 附录 C：部署资源对比

| 方式 | 内存 | 启动 | 门槛 | 适用场景 |
|------|------|------|------|---------|
| Go 单二进制（直接运行）| ~30MB | <1s | 极低 | 个人本地 |
| Go + WinSW 服务 | ~30MB | 开机自启 | 低 | Windows 长期运行 |
| Go + systemd（Linux）| ~30MB | 开机自启 | 低 | AWS/VPS |
| Wails 桌面应用 | ~50MB | <2s | 中 | 对外分发 |
| Java/Spring Boot | ~512MB | 15~30s | 中 | 不推荐 |

### 附录 D：数据库表清单（29 张）

```
系统与权限（4）：user_account / system_config / operation_log / backup_log
学生与老师（5）：student / parent / teacher / teacher_availability / teacher_capability
课程体系（4）： course_domain / course_track / course_level / skill_tag
学习项目（4）： student_course_enrollment / student_teacher_assignment /
               student_learning_path / student_level_event
班级预留（2）： class_group / class_group_member
排课上课（2）： lesson / attendance
账务体系（6）： student_payment / teacher_payout / student_account_ledger /
               teacher_account_ledger / lesson_finance / fx_rate_snapshot
通知体系（2）： notification_template / notification_log
```

### 附录 E：下一步行动

1. **业务对齐**：与运营者确认课程方向/等级配置、课时套餐定价、充值币种习惯、邮件语言偏好、部署偏好
2. **ER 图设计**：基于附录 D，绘制完整 ER 图，重点评审账务流水外键关系和索引设计
3. **API 设计**：编写 `/api/v1` 接口清单，重点设计课后确认和充值的事务边界
4. **UI 原型**：工作台 Dashboard + 排课页面 + 移动端 Today 页低保真原型
5. **环境准备**：注册 Resend 账号（验证发件域名）/ 开通 AWS Free Tier 或准备 VPS / 开发环境（Go 1.22+, Node 20+）
6. **Sprint 0 启动**：搭建工程骨架，跑通 Go + SQLite + Soybean Admin + JWT 登录的最小可运行系统

---

*Zedu PRD v1.0 · 2026-06-09*
*文档版本历史：V0.1 初始业务背景 → V0.2 更新技术选型 → V1.0 整合完整数据模型、API 规范、工程结构*
