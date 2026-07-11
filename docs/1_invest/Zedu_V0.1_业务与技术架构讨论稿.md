# Zedu 轻量级教培教务管理系统 — 业务与技术架构讨论稿

> 版本：V0.1 讨论稿  
> 整理日期：2026-06-09  
> 定位：小而美、单机构、前后端一体化、轻量部署、可复制定制  
> 当前阶段：内部讨论稿，后续可扩展为产品愿景、PRD、ER 设计、API 设计与开发任务书

---

## 0. 本稿定位

本文档用于沉淀当前关于 **Zedu 轻量级教培教务管理系统** 的阶段性讨论结论。

本稿不是最终 PRD，也不是详细设计说明书，而是一份用于后续继续深化的业务与技术架构讨论稿。后续可以在此基础上继续形成：

- 产品愿景文档
- 正式 PRD
- 原型设计说明
- ER 数据库设计
- API 设计规范
- 前后端开发任务拆分
- 部署运维手册

---

## 1. 产品定位

### 1.1 系统名称

暂定名称：

```text
Zedu
```

### 1.2 产品定位

Zedu 是一套 **小而美的轻量级教培教务管理系统**，首期面向日语一对一 / 小班课程管理场景，后续可扩展到英语、留学辅导、艺术培训、体育培训、职业证书培训等轻量教培业务。

### 1.3 核心目标

系统目标不是做重型 SaaS 平台，也不是做复杂 ERP，而是优先满足小型教培运营者的真实高频需求：

```text
漂亮
轻量
好部署
手机可用
账务清楚
提醒可靠
可复制部署
可定制界面
可快速上线
```

### 1.4 当前定位边界

当前版本明确：

```text
不做多租户
不做重型 SaaS
不做微服务
不做复杂权限平台
不做 Flutter 全覆盖
不做真正离线同步
不直接二开大框架
```

采用：

```text
单机构
单实例
单数据库
前后端一体化发布
一个可执行文件 + 一个数据库文件 + 一个可选配置文件
```

---

## 2. 部署与使用场景

### 2.1 主部署方式

主系统部署在：

```text
https://zedu.abitcloud.org
```

访问方式：

```text
PC 浏览器
iPad / 平板浏览器
手机浏览器
后续可包装 Electron 桌面壳
后续如有 App 需求，可复用 REST API
```

### 2.2 多客户使用方式

当前不做多租户。如果后续其他朋友或教培机构要使用，采用“一客一部署”方式：

```text
zedu.abitcloud.org       → 当前实例
friend1.abitcloud.org    → 复制部署一套
friend2.abitcloud.org    → 复制部署一套
custom-domain.com        → 独立部署 + 定制 Logo / UI
```

优点：

```text
复杂度低
数据天然隔离
可定制 UI
可独立升级
可独立备份
出现问题影响范围小
```

---

## 3. 开源框架借鉴原则

本项目不直接 fork 或二开某个大型后台框架，而是借鉴其成熟工程经验，自建干净的轻量项目。

### 3.1 借鉴对象

| 参考对象 | 借鉴内容 | 不采用内容 |
|---|---|---|
| Gin-Vue-Admin | Go 中后台工程组织、JWT、权限、菜单、代码生成思路 | 不直接采用其前端，不二开全家桶 |
| RuoYi-Go | Gin + GORM 后台模块划分、后台管理思想 | 不复刻若依传统 UI 与重后台风格 |
| PocketBase | 单文件后端、SQLite、轻量发布理念 | 不采用其模型作为业务核心，不受限于其数据模型 |
| go-admin | 后台权限、日志、系统初始化、迁移思路 | 不采用其传统后台体系 |

### 3.2 核心原则

```text
借鉴工程经验
不绑定大框架
不被脚手架反向绑架
业务逻辑自己掌控
前端采用更漂亮的 Soybean Admin
后端采用小型 Go 服务组合
```

---

## 4. 总体业务架构

### 4.1 角色体系

V1 主要实现：

```text
Owner / 老板
Operator / 教务
```

后续预留：

```text
Teacher / 老师
Student / 学生
Parent / 家长
```

### 4.2 角色职责

#### Owner / 老板

```text
查看经营概况
查看收入、成本、毛利
查看待续费学生
查看待结款老师
查看课程与老师贡献
查看经营摘要
管理系统基础配置
```

#### Operator / 教务

```text
管理学生档案
管理老师档案
维护课程方向和等级
安排课程
录入充值
课后确认
处理扣课
生成老师应付
发送或重发通知
维护通知模板
查看操作与通知日志
```

#### Teacher / 老师，后续

V1 不登录，只接收提醒。

后续可扩展：

```text
查看课表
确认课时
填写课后备注
查看待结算课酬
维护可授时间
```

#### Student / Parent，后续

V1 不登录，只接收提醒。

后续可扩展：

```text
查看课表
查看余额
查看学习路径
查看课程记录
接收续费提醒
提交调课申请
```

---

## 5. 核心业务流程

### 5.1 学生建档流程

```text
学生咨询 / 报名
  ↓
教务录入学生档案
  ↓
录入家长联系方式，选填
  ↓
录入当前学习水平
  ↓
录入学习目标
  ↓
创建学生课程报名 / 学习项目
  ↓
匹配老师
  ↓
进入排课流程
```

### 5.2 老师建档流程

```text
录入老师基础信息
  ↓
录入老师简介 / 简历 / 证书
  ↓
录入可授课程方向
  ↓
录入可教等级
  ↓
录入擅长能力标签
  ↓
录入默认课酬
  ↓
维护可授时间
```

### 5.3 学生课程报名流程

一个学生可以同时学习多个课程方向，也可以同时由多个老师授课。

例如：

```text
王同学
  ├── 日语 / JLPT / 当前 N3 / 目标 N2 / 老师A
  ├── 日语 / 会话 / 中级 / 老师B
  └── 英语 / IELTS / 目标 6.5 / 老师C
```

因此系统不采用简单的“学生绑定一个老师”模型，而采用：

```text
学生
  ↓
课程报名 / 学习项目
  ↓
老师安排
  ↓
具体课次
```

### 5.4 充值与课时流程

```text
学生 / 家长付款
  ↓
教务录入充值记录
  ↓
保存原始币种、原始金额、汇率、折算 JPY 金额
  ↓
增加学生余额 / 课程项目课时
  ↓
生成学生账户流水
```

### 5.5 排课与课前提醒流程

```text
教务创建课程
  ↓
选择学生课程报名项目
  ↓
选择实际授课老师
  ↓
选择课程方向 / 等级 / 主题
  ↓
维护上课时间与上课链接
  ↓
定时任务扫描未来课程
  ↓
课前 N 分钟发送提醒给学生 / 老师
  ↓
写入通知日志
```

### 5.6 课后确认与账务流程

```text
课程结束
  ↓
教务确认上课结果
  ↓
生成上课记录
  ↓
扣减学生余额 / 课时
  ↓
生成老师应付
  ↓
生成单课财务记录
  ↓
计算单课毛利
  ↓
判断余额不足
  ↓
必要时触发续费提醒
```

---

## 6. 课程体系与成长路径设计

### 6.1 为什么需要课程体系扩展

虽然 V1 可以靠人工判断老师与学生匹配，但系统需要预留以下能力：

```text
老师可以教多种课程类型
老师能力会随时间变化
学生可以同时上不同类型课程
学生可以跟不同老师学习不同等级内容
学生学习目标会变化
学生等级会从 N5 → N4 → N3 → N2 → N1 迁移
体育 / 艺术类课程可能按年龄、学时、考级自动升级
```

因此必须把以下概念放进数据模型：

```text
课程领域
课程方向
课程等级
能力标签
老师能力
学生课程报名
学生学习路径
等级变化事件
```

### 6.2 课程领域：course_domain

表示大类：

```text
日语
英语
足球
钢琴
编程
留学辅导
职业证书
```

字段示例：

```text
course_domain
  id
  name
  code
  type              LANGUAGE / SPORT / ART / ACADEMIC / OTHER
  enabled
```

### 6.3 课程方向：course_track

表示某领域下的学习路线。

例如日语：

```text
JLPT 备考
日常会话
商务日语
少儿日语
面试日语
作文写作
```

字段示例：

```text
course_track
  id
  domain_id
  name
  code
  enabled
```

### 6.4 课程等级：course_level

表示某条学习路线下的阶段。

例如 JLPT：

```text
入门
N5
N4
N3
N2
N1
高级
```

字段示例：

```text
course_level
  id
  track_id
  name
  code
  sort_order
  min_age
  max_age
  min_lesson_hours
  recommended_lesson_hours
  enabled
```

其中：

```text
min_age / max_age
min_lesson_hours / recommended_lesson_hours
```

主要用于体育、艺术、考级、年龄段升级类课程扩展。

### 6.5 能力标签：skill_tag

表示具体能力点。

语言类：

```text
词汇
语法
阅读
听力
口语
写作
面试
综合
```

体育类：

```text
体能
协调性
技巧
战术
比赛
考级动作
```

字段示例：

```text
skill_tag
  id
  domain_id
  name
  code
  enabled
```

---

## 7. 老师能力模型

### 7.1 老师能力不是静态字段

老师能力不能简单写成：

```text
teacher.level = N3
```

因为老师可能同时具备：

```text
能教 N5-N3 基础课
能教中级会话
不能教 N1 作文
后期经过培训后可以教 N2
某段时间暂停高级课
```

### 7.2 老师能力表：teacher_capability

```text
teacher_capability
  id
  teacher_id
  domain_id
  track_id
  level_id
  skill_tag_codes
  capability_status     ACTIVE / PAUSED / ENDED
  verified_status       UNVERIFIED / VERIFIED
  effective_from
  effective_to
  note
  created_at
  updated_at
```

示例：

```text
老师A
  ├── 日语 / JLPT / N5
  ├── 日语 / JLPT / N4
  ├── 日语 / JLPT / N3
  ├── 日语 / 会话 / 初级
  └── 日语 / 会话 / 中级
```

### 7.3 V1 实现方式

V1 不做复杂能力认证，只做：

```text
老师简介
老师证书 / 经历说明
可教课程方向
可教等级
擅长能力标签
排课时轻提示
```

排课时：

```text
学生报名项目：JLPT N2
选择老师：老师A
老师能力：最高 N3

系统提示：
该老师暂未标记支持 N2，请确认是否继续排课。
```

V1 只提示，不强制拦截。

---

## 8. 学生课程报名与多对多关系

### 8.1 核心调整

真实业务不是：

```text
学生 → 一个老师
```

而是：

```text
学生 → 多个课程报名项目 → 多个老师 → 多个课次
```

因此必须增加核心中间层：

```text
student_course_enrollment
```

### 8.2 学生课程报名表：student_course_enrollment

```text
student_course_enrollment
  id
  student_id

  domain_id
  track_id
  current_level_id
  target_level_id

  enrollment_type       ONE_TO_ONE / GROUP / TRIAL
  status                ACTIVE / PAUSED / COMPLETED / CANCELLED

  started_at
  ended_at

  default_price_jpy
  default_lesson_unit
  lesson_balance
  balance_jpy

  remark
  created_at
  updated_at
```

一个学生可以有多条 enrollment：

```text
王同学
  ├── 日语 / JLPT / 当前 N3 / 目标 N2
  ├── 日语 / 会话 / 中级
  └── 英语 / IELTS
```

### 8.3 学生老师安排表：student_teacher_assignment

老师和学生之间的绑定要基于某个课程报名项目，而不是直接绑定学生。

```text
student_teacher_assignment
  id
  enrollment_id
  student_id
  teacher_id

  role_type             MAIN / SUBSTITUTE / ASSISTANT
  status                ACTIVE / ENDED / PAUSED

  start_date
  end_date

  reason
  remark
  created_at
```

这样可以表示：

```text
王同学的 JLPT N3-N2 课程由老师A负责
王同学的会话课由老师B负责
某一周老师A请假，由老师C临时代课
```

---

## 9. 学生学习路径与等级变化

### 9.1 学习目标也不是静态字段

学生的学习方向可能变化：

```text
2026-06：当前 N4，目标 N3
2026-12：通过 N3，目标改为 N2
2027-07：转向会话强化
```

因此不能只在 student 表里放一个 target_level。

### 9.2 学生学习路径表：student_learning_path

```text
student_learning_path
  id
  enrollment_id
  student_id

  from_level_id
  current_level_id
  target_level_id

  goal_type             EXAM / CONVERSATION / BUSINESS / HOBBY / COMPETITION
  target_exam_name
  target_exam_date

  status                ACTIVE / COMPLETED / CHANGED / PAUSED

  started_at
  ended_at
  note
```

### 9.3 学生等级变化事件表：student_level_event

```text
student_level_event
  id
  student_id
  enrollment_id
  learning_path_id

  from_level_id
  to_level_id

  event_type            ASSESSMENT / EXAM_PASS / HOURS_REACHED / AGE_REACHED / MANUAL / AUTO_RULE
  event_date
  evidence_note
  operator_id
  created_at
```

示例：

```text
王同学
N3 → N2
原因：JLPT N3 合格
日期：2026-12-20
```

体育类示例：

```text
张同学
少儿基础班 → 少儿进阶班
原因：年龄满 9 岁 + 累计学时 80 小时
```

---

## 10. 小班课扩展预留

虽然 V1 主要关注一对一，但数据模型可以预留小班课。

### 10.1 班级表：class_group

```text
class_group
  id
  name
  domain_id
  track_id
  level_id

  main_teacher_id
  status                ACTIVE / PAUSED / FINISHED

  default_price_jpy
  default_teacher_pay_jpy
  max_students

  started_at
  ended_at
  remark
```

### 10.2 班级成员表：class_group_member

```text
class_group_member
  id
  class_group_id
  student_id
  enrollment_id
  join_date
  leave_date
  status
```

### 10.3 V1 处理方式

V1 可以先不做复杂班级管理，但表设计要预留。后续小班课时，lesson 可以支持：

```text
lesson.enrollment_id     一对一课程
lesson.class_group_id    小班课程
```

---

## 11. 课程安排与历史快照

### 11.1 课程安排必须保存当时状态

课程记录不能只引用学生当前等级或老师当前能力，因为这些状态会变。

每节课必须记录：

```text
当时的课程领域
当时的课程方向
当时的课程等级
当时的实际授课老师
当时的课程主题
```

### 11.2 课程表：lesson

```text
lesson
  id
  lesson_no

  enrollment_id
  class_group_id

  student_id
  teacher_id

  domain_id
  track_id
  level_id
  lesson_topic

  scheduled_start_at
  scheduled_end_at
  duration_min
  timezone

  meeting_type       WECHAT / TENCENT / ZOOM / OTHER
  meeting_link

  status             SCHEDULED / REMINDED / COMPLETED / CANCELLED
  remind_sent_at

  remark
  created_at
  updated_at
```

---

## 12. 课后确认与账务模型

### 12.1 上课记录：attendance

```text
attendance
  id
  lesson_id
  enrollment_id
  student_id
  teacher_id

  actual_start_at
  actual_end_at

  student_attended
  teacher_attended

  lesson_deducted
  teacher_note
  operator_note

  confirmed_by
  confirmed_at
```

### 12.2 学生账户流水：student_account_ledger

```text
student_account_ledger
  id
  student_id
  enrollment_id

  biz_type               RECHARGE / LESSON_DEDUCT / REFUND / ADJUST
  amount_jpy_delta
  lesson_delta

  balance_jpy_after
  lesson_balance_after

  related_payment_id
  related_lesson_id

  operator_id
  remark
  created_at
```

### 12.3 老师账户流水：teacher_account_ledger

```text
teacher_account_ledger
  id
  teacher_id

  biz_type               LESSON_PAYABLE / PAYOUT / ADJUST
  amount_jpy_delta
  unpaid_amount_after

  related_lesson_id
  related_payout_id

  operator_id
  remark
  created_at
```

### 12.4 单课财务：lesson_finance

```text
lesson_finance
  id
  lesson_id
  enrollment_id
  student_id
  teacher_id

  student_charge_jpy
  teacher_pay_jpy
  gross_profit_jpy

  created_at
```

### 12.5 账务事务原则

课后确认必须同一事务完成：

```text
写 attendance
写 student_account_ledger
写 teacher_account_ledger
写 lesson_finance
更新学生余额缓存
更新课程报名项目余额缓存
更新老师待结算金额缓存
更新 lesson 状态
```

---

## 13. 多币种费用设计

### 13.1 本位币

系统统一使用：

```text
JPY
```

### 13.2 原币事实保留

所有充值必须保留：

```text
原始金额
原始币种
当时汇率
折算后 JPY 金额
```

示例：

```text
学生微信支付 CNY 500
汇率：1 CNY = 21.80 JPY
系统入账：JPY 10,900
```

保存为：

```text
original_amount = 500.00
original_currency = CNY
fx_rate_to_jpy = 21.800000
amount_jpy = 10900.00
```

### 13.3 充值记录表：student_payment

```text
student_payment
  id
  payment_no
  student_id
  enrollment_id

  original_amount
  original_currency
  fx_rate_to_jpy
  amount_jpy

  lessons_added
  payment_method         WECHAT / PAYPAY / BANK / CASH / OTHER

  paid_at
  operator_id
  status                 CONFIRMED / VOIDED
  remark
  created_at
```

### 13.4 金额处理原则

```text
禁止使用 float64 直接处理金额
Go 内部使用 decimal 库
数据库可用 decimal 字符串或 integer minor unit
所有报表基于 JPY 汇总
原币种信息保留用于追溯
```

---

## 14. 前端架构

### 14.1 PC / iPad 前端

采用：

```text
Soybean Admin
Vue 3
Vite
TypeScript
Naive UI
Pinia
UnoCSS
ECharts / VChart
```

### 14.2 为什么选择 Soybean Admin

相比传统 RuoYi 类前端，Soybean Admin 更适合 Zedu 的产品气质：

```text
清新
柔和
现代
卡片化
适合小而美
适合工作台和报表
```

### 14.3 Mobile V1 策略

V1 不单独做移动端项目，而是在 Soybean Admin 内提供移动优先页面：

```text
/mobile/today
/mobile/confirm
/mobile/student
/mobile/payment
/mobile/report
```

手机访问时默认进入移动快捷页。

### 14.4 Mobile 页面原则

```text
少表格
多卡片
大按钮
少字段
高频操作优先
底部快捷导航
避免 PC 表格直接压缩
```

### 14.5 V1.5 Mobile 扩展

如果后续手机使用频率高，再单独拆出：

```text
/m
```

可选 UI：

```text
TDesign Mobile Vue
Varlet
NutUI
```

Vant 稳定但视觉较朴素，暂不作为第一候选。

### 14.6 Flutter 策略

Flutter 不作为 V1 主前端。仅在后续明确需要 App 时使用。

后续 App 可复用：

```text
/api/v1
```

---

## 15. 后端技术架构

### 15.1 技术栈

```text
语言：Go
Web 框架：Gin
ORM：GORM
数据库：SQLite
数据库迁移：goose
认证：JWT + Refresh Token
定时任务：robfig/cron
日志：zap 或 zerolog
配置：config.yaml + 环境变量覆盖
邮件：Resend API + SMTP
静态资源：go:embed
```

### 15.2 核心架构

```text
Browser / Mobile Browser
        ↓
Soybean Admin 前端
        ↓
Go HTTP Server
        ├── 静态资源服务
        ├── REST API
        ├── JWT Auth
        ├── 业务服务
        ├── 定时任务
        └── 数据备份
        ↓
Repository Layer
        ↓
SQLite
```

### 15.3 后期可替换方向

```text
SQLite → MySQL / PostgreSQL
Soybean 内置移动页 → 独立 Mobile Web
Web API → Flutter App / 小程序 / 第三方集成
```

---

## 16. 发布架构

### 16.1 最小发布包

```text
zedu/
  ├── zedu-server
  └── data/
      └── zedu.db
```

### 16.2 推荐生产发布包

```text
zedu/
  ├── zedu-server
  ├── config.yaml
  ├── data/
  │   └── zedu.db
  ├── logs/
  ├── backup/
  └── README.md
```

### 16.3 前端嵌入方式

```text
Soybean Admin build
  ↓
dist/
  ↓
Go embed
  ↓
zedu-server
```

最终由一个 Go Server 同时提供：

```text
前端页面
REST API
定时任务
通知服务
备份任务
```

### 16.4 访问路径

```text
/                 前端入口
/assets           前端静态资源
/api/v1           REST API
/healthz          健康检查
```

---

## 17. 代码仓库结构

建议前后端放在一个仓库中，便于一体化构建与发布。

```text
zedu/
  ├── backend/
  │   ├── cmd/
  │   │   └── zedu-server/
  │   │       └── main.go
  │   │
  │   ├── internal/
  │   │   ├── app/
  │   │   ├── auth/
  │   │   ├── user/
  │   │   ├── student/
  │   │   ├── parent/
  │   │   ├── teacher/
  │   │   ├── course/
  │   │   ├── enrollment/
  │   │   ├── lesson/
  │   │   ├── attendance/
  │   │   ├── finance/
  │   │   ├── notification/
  │   │   ├── report/
  │   │   ├── system/
  │   │   ├── job/
  │   │   ├── audit/
  │   │   └── backup/
  │   │
  │   ├── pkg/
  │   │   ├── response/
  │   │   ├── pagination/
  │   │   ├── validator/
  │   │   ├── money/
  │   │   ├── datetime/
  │   │   ├── crypto/
  │   │   └── errors/
  │   │
  │   ├── migrations/
  │   │   ├── sqlite/
  │   │   ├── mysql/
  │   │   └── postgres/
  │   │
  │   ├── web/
  │   │   └── admin-dist/
  │   │
  │   ├── config/
  │   │   └── config.example.yaml
  │   │
  │   ├── go.mod
  │   └── go.sum
  │
  ├── frontend/
  │   ├── admin/
  │   │   ├── src/
  │   │   │   ├── api/
  │   │   │   ├── views/
  │   │   │   │   ├── dashboard/
  │   │   │   │   ├── student/
  │   │   │   │   ├── teacher/
  │   │   │   │   ├── course/
  │   │   │   │   ├── enrollment/
  │   │   │   │   ├── lesson/
  │   │   │   │   ├── finance/
  │   │   │   │   ├── notification/
  │   │   │   │   ├── report/
  │   │   │   │   ├── system/
  │   │   │   │   └── mobile/
  │   │   │   ├── router/
  │   │   │   ├── store/
  │   │   │   ├── components/
  │   │   │   ├── hooks/
  │   │   │   └── utils/
  │   │   ├── package.json
  │   │   └── vite.config.ts
  │   │
  │   └── shared/
  │       ├── types/
  │       ├── constants/
  │       └── api-schema/
  │
  ├── deploy/
  │   ├── zedu.service
  │   ├── nginx.conf
  │   ├── backup.sh
  │   └── install.sh
  │
  ├── scripts/
  │   ├── build.sh
  │   ├── build.ps1
  │   ├── release.sh
  │   └── dev.sh
  │
  ├── docs/
  │   ├── architecture.md
  │   ├── api.md
  │   ├── database.md
  │   └── deployment.md
  │
  ├── Makefile
  ├── README.md
  └── .gitignore
```

---

## 18. 后端模块结构规范

每个业务模块采用统一结构：

```text
student/
  ├── model.go
  ├── dto.go
  ├── handler.go
  ├── service.go
  ├── repository.go
  ├── routes.go
  └── errors.go
```

职责划分：

```text
handler       只处理 HTTP 入参出参
service       处理业务规则与事务
repository    处理数据库访问
model         对应数据库模型
dto           对应 API 入参出参
errors        定义模块错误
routes        注册路由
```

核心原则：

```text
业务逻辑不写在 handler
数据库操作不散落在 service 外部
model 不直接暴露给前端
所有金额计算走 money 工具包
所有时间统一转换处理
```

---

## 19. API 规范

### 19.1 API 前缀

```text
/api/v1
```

后续 App、独立移动端、第三方集成都复用该 API。

### 19.2 命名规范

RESTful 风格，资源名使用复数：

```text
GET    /api/v1/students
POST   /api/v1/students
GET    /api/v1/students/{id}
PUT    /api/v1/students/{id}
DELETE /api/v1/students/{id}
```

业务动作使用子资源或 action：

```text
POST /api/v1/lessons/{id}/confirm
POST /api/v1/lessons/{id}/cancel
POST /api/v1/students/{id}/payments
POST /api/v1/teachers/{id}/payouts
POST /api/v1/notifications/{id}/resend
```

### 19.3 API 分组

```text
/api/v1/auth
/api/v1/users
/api/v1/students
/api/v1/parents
/api/v1/teachers
/api/v1/courses
/api/v1/enrollments
/api/v1/lessons
/api/v1/attendances
/api/v1/payments
/api/v1/ledgers
/api/v1/payouts
/api/v1/reports
/api/v1/notifications
/api/v1/system
```

### 19.4 响应格式

成功响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {},
  "traceId": "20260609-xxxx"
}
```

分页响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [],
    "page": 1,
    "pageSize": 20,
    "total": 128
  },
  "traceId": "20260609-xxxx"
}
```

错误响应：

```json
{
  "code": 40001,
  "message": "student balance is insufficient",
  "traceId": "20260609-xxxx"
}
```

### 19.5 API 版本演进

```text
/api/v1  当前稳定接口
/api/v2  破坏性变更时启用
```

非破坏性变更：

```text
增加字段
增加接口
增加枚举值
```

破坏性变更：

```text
删除字段
修改字段语义
修改响应结构
修改路径
```

---

## 20. 数据库设计原则

### 20.1 V1 数据库

```text
SQLite
```

### 20.2 后期可替换

```text
MySQL
PostgreSQL
```

### 20.3 跨库兼容原则

```text
避免 SQLite 独有语法
避免数据库触发器承载核心业务
避免复杂存储过程
金额统一使用 decimal 字符串或 integer minor unit
时间统一 UTC 存储，前端按 JST 显示
业务枚举使用 varchar，不依赖数据库 enum
repository 层隔离数据库访问
migration 按 dialect 分目录
```

---

## 21. 核心表清单

### 21.1 系统与权限

```text
user_account
system_config
operation_log
backup_log
```

### 21.2 学生与老师

```text
student
parent
teacher
teacher_availability
teacher_capability
```

### 21.3 课程体系

```text
course_domain
course_track
course_level
skill_tag
progression_rule
```

### 21.4 学生学习项目

```text
student_course_enrollment
student_teacher_assignment
student_learning_path
student_level_event
```

### 21.5 班级与小班课

```text
class_group
class_group_member
```

### 21.6 排课与上课

```text
lesson
attendance
```

### 21.7 账务

```text
student_payment
student_account_ledger
teacher_payout
teacher_account_ledger
lesson_finance
exchange_rate_snapshot
```

### 21.8 通知

```text
notification_template
notification_log
```

---

## 22. 定时任务设计

### 22.1 任务模块

```text
job/
  ├── lesson_reminder_job.go
  ├── balance_alert_job.go
  ├── morning_report_job.go
  ├── owner_report_job.go
  ├── notification_retry_job.go
  └── backup_job.go
```

### 22.2 课前提醒

```text
每 5 或 10 分钟扫描
  ↓
查询未来 N 分钟课程
  ↓
排除已提醒课程
  ↓
发送学生 / 老师提醒
  ↓
写 notification_log
  ↓
更新 lesson.remind_sent_at
```

### 22.3 余额预警

触发方式：

```text
课后确认后立即判断
每日定时扫描补充判断
```

触发条件：

```text
lesson_balance <= 阈值
或
balance_jpy < 下一节预计扣费
或
未来 7 天课程数 > lesson_balance
```

### 22.4 教务晨报

```text
每天 08:00
  今日课程
  待确认课程
  待续费学生
  待结款老师
  失败通知
```

### 22.5 老板摘要

```text
每日 / 每周
  收入
  成本
  毛利
  新增学生
  续费情况
  老师课时
  课程类型收入
```

---

## 23. 通知架构

### 23.1 V1 通知通道

```text
Email
  ├── Resend
  └── SMTP
```

### 23.2 后续扩展

```text
SMS
LINE
WeChat
Telegram
App Push
```

### 23.3 通知接口

```text
NotificationSender
  Send(ctx, message) error
```

实现：

```text
ResendSender
SmtpSender
```

### 23.4 发送策略

```text
优先使用 Resend
失败后 fallback 到 SMTP
失败写日志
支持手动重发
支持失败重试
```

---

## 24. 安全与权限

### 24.1 V1 角色

```text
Owner
Operator
```

### 24.2 权限边界

Owner：

```text
查看全部数据
查看经营报表
管理操作员
修改系统配置
管理备份
```

Operator：

```text
管理学生
管理老师
排课
充值
课后确认
查看基础报表
处理通知
```

### 24.3 安全措施

```text
密码哈希
JWT 过期
Refresh Token
登录失败限制
操作日志
敏感配置加密
备份文件保护
```

---

## 25. 数据备份与恢复

SQLite 备份不能简单粗暴复制正在写入的 db 文件。

建议使用：

```text
VACUUM INTO
或 SQLite backup API
```

备份策略：

```text
每日自动备份
保留最近 30 天
每周保留一份长期备份
支持手动备份
支持下载备份文件
```

备份目录：

```text
backup/
  ├── zedu_20260609_020000.db
  ├── zedu_20260610_020000.db
  └── ...
```

---

## 26. 配置文件示例

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
  driver: sqlite
  dsn: ./data/zedu.db

auth:
  jwt_secret: change-me
  access_token_minutes: 60
  refresh_token_days: 14

mail:
  primary: resend
  resend:
    api_key: re_xxx
    from_email: noreply@abitcloud.org
    from_name: Zedu
  smtp:
    host: smtp.example.com
    port: 587
    username: xxx
    password: xxx
    from_email: noreply@abitcloud.org

lesson:
  reminder_minutes: 30
  balance_alert_lessons: 3

backup:
  enabled: true
  cron: "0 0 2 * * *"
  path: ./backup
  retention_days: 30
```

---

## 27. V1 功能清单

### 27.1 必做

```text
登录
Owner / Operator 角色
工作台
学生管理
家长信息
老师管理
老师能力标签
课程领域 / 方向 / 等级基础维护
学生课程报名
学生老师安排
排课管理
课后确认
充值录入
多币种折算 JPY
学生流水
老师流水
单课毛利
课前提醒
余额不足提醒
教务晨报
老板摘要
通知日志
操作日志
基础报表
自动备份
```

### 27.2 简化实现

```text
老师匹配只提示不强制
小班课只预留基础结构
学生成长路径先基础记录
等级迁移不自动执行
手机端先用 Soybean 内置移动优先页面
```

### 27.3 V1 不做

```text
多租户
完整学生端
完整老师端
完整家长端
支付接口
复杂自动排课
自动升级规则
Flutter App
真正离线同步
复杂课程商品化
复杂班级管理
```

---

## 28. 后续演进方向

### 28.1 V1.5

```text
独立 Mobile Web
TDesign Mobile Vue / Varlet / NutUI
PWA 优化
Electron 桌面壳
更漂亮的报表
Excel 导入导出
```

### 28.2 V2

```text
老师端
学生端
家长端
调课申请
老师确认出勤
付款凭证上传
课程产品化
不同课程不同价格
不同等级不同课酬
自动推荐老师
等级升级提醒
考试计划管理
```

### 28.3 V3

```text
Flutter App
LINE / WeChat 通知
支付接口
多实例模板化部署
更完整经营分析
课程成长路径可视化
AI 老师匹配建议
AI 学习报告
```

---

## 29. 当前最终技术决策

```text
后端：
Go + Gin + GORM + SQLite

前端：
Soybean Admin

Mobile：
V1 在 Soybean 内做移动优先页面

发布：
Go embed 前端资源
一个 zedu-server 提供页面 + API + 定时任务

数据库：
SQLite 起步
保留 MySQL / PostgreSQL 替换能力

App：
暂不做
后期复用 /api/v1
```

---

## 30. 总结

Zedu V0.1 当前架构可以定义为：

> 一套小而美的前后端一体化教培教务系统。  
> 后端采用 Go + SQLite，前端采用 Soybean Admin，前端静态资源嵌入 Go 二进制，最终以一个可执行文件 + 一个数据库文件 + 一个可选配置文件发布。  
> 系统以教务运营为核心，覆盖学生、老师、课程体系、学生课程报名、排课、充值、多币种折算、课后确认、流水台账、通知提醒和基础经营报表。  
> 架构上预留 REST API 版本化、数据库替换、多端复用、未来 App 化和课程成长路径扩展能力。  

本稿确认后，下一步可继续整理：

```text
产品愿景
正式 PRD
ER 图
API 设计
页面原型
开发任务拆分
部署方案
```
