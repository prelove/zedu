# Zedu 轻量级教培教务管理系统
## 产品需求文档（PRD）— 正式定案版

> **文档版本**：v3.1-r1（Final / MVP范围修订版；含 2026-07-12 学生邮箱唯一性决策）
> **状态**：正式定案，作为后续所有开发、设计、测试工作的唯一依据
> **整理日期**：2026-07-04
> **项目代号**：Zedu（Zero-friction Education）
> **文档谱系**：本文档融合并取代此前所有讨论稿（v0.1业务讨论稿 → v0.2部署方案 → v1.0正式PRD → v2.0完整实装版 → v2.1配置化修订说明 → v2.2业务配置化收敛版），是唯一权威版本。此前所有文档转为历史参考，不再单独使用。

---

## 文档说明

### 本文档的地位

这是 Zedu 项目的**事实文档（Source of Truth）**。此后任何设计变更、开发决策、范围调整，都应该：

1. 优先检查本文档是否已有相关规定
2. 若需要变更，在本文档对应章节直接修订，并在文末"变更记录"中留痕
3. 不再另外维护多份并行的 PRD 草稿——所有讨论结论最终都收敛回这一份文档

### 本版相对此前版本的定位

v2.0 提供了完整的技术实现深度（DDL、API、页面规格、状态机），但在课程/定价/币种/出勤这几个业务弹性维度上，把"业务澄清清单的建议默认值"误当成了"架构约束"。v2.1 和 v2.2 独立发现并纠正了这个问题，明确了"结构通用、默认可给、业务不写死"的原则。本文档在 v2.2 确立的业务原则基础上，重新构建了完整的技术实现层，确保业务弹性的原则真正落到每一张表、每一个字段、每一个接口上，而不是停留在文字描述层面。

---

## 目录

**第一部分：产品定位与原则**
1. 产品定位与一句话定义
2. 产品设计五大原则
3. 用户角色与核心场景
4. 产品范围（V1做什么、不做什么）

**第二部分：核心业务概念**
5. 课程体系（四层配置化结构）
6. 定价与课酬模型
7. 币种与支付设计
8. 课时与出勤处理

**第三部分：业务流程**
9. 核心业务流程详解（含异常分支）

**第四部分：功能与界面**
10. 功能模块与页面级UI规格
11. 移动端策略

**第五部分：数据与接口**
12. 数据模型设计（完整DDL）
13. API接口规范
14. 业务规则与校验总表
15. 状态机设计

**第六部分：技术架构**
16. 后端技术架构与开发规范
17. 前端技术架构
18. 桌面端封装方案
19. 部署与发布方案
20. 数据安全与备份
21. 通知与提醒系统
22. 非功能性需求

**第七部分：项目管理**
23. 开发计划与任务拆分
24. 测试计划与验收标准
25. 风险登记表
26. 版本演进规划

**附录**

---

# 第一部分：产品定位与原则

## 第一章 产品定位与一句话定义

### 1.1 一句话定义

> Zedu 是一套面向小型教培机构和兼职教务运营者的轻量级教务与经营闭环系统，帮助运营者用最低操作负担完成学生建档、课程报名、老师安排、排课提醒、课后确认、充值扣费、老师结款、经营报表和数据备份。

### 1.2 核心差异化价值

Zedu 不追求"功能堆满"，差异化价值在于：

```
课程可以自由配置，不绑定任何单一学科
学生可以同时学多个课程项目，由不同老师负责
价格和课酬可以灵活约定，系统给建议、人做最终决定
每节课的实际收费和课酬都落成事实快照，历史不因未来调整而改变
多币种收款可以追溯（原始金额、汇率、折算金额三者并存）
付款截图可以留证，减少"到底有没有付"的纠纷
请假/缺课扣费可以人工灵活处理，不被僵硬规则捆绑
系统自动做提醒这类"不涉及金钱决策"的事，但绝不自动执行扣费/结款这类财务动作
数据单文件部署、备份、复制、恢复，运维成本极低
```

一句话总结：**Zedu 是小型教培机构的"经营闭环账本 + 教务提醒工作台"，不是重型 SaaS 平台。它必须足够灵活以适应真实小机构的业务随意性，也必须足够克制以便一周内跑起来并长期好维护。**

### 1.3 首个部署实例背景

本文档的首个真实落地场景：一位日语教育从业者，以兼职形式撮合日语学习者与老师，组织一对一/小班课程，赚取学费与课酬之间的差价，目前用 Excel + 人工管理，核心痛点是提醒遗漏、续费被动、账务繁琐、信息分散。

首个实例的初始化选择：**应用"日语培训模板"作为种子数据**（模板机制详见第五章），本位币选定为 **JPY**。这两项是本次部署的具体配置值，不是系统架构的限制——同一套系统换一个部署实例，完全可以选择"K12学科模板"+ CNY 本位币启动，不需要改一行代码。

---

## 第二章 产品设计五大原则

这五条原则贯穿本文档所有后续章节，是所有设计决策的最高判断依据。

### 原则一：结构上通用，配置上灵活

系统不写死日语，也不写死任何学科。课程体系必须通过四层配置化结构表达（领域/方向/等级/能力标签），日语只是初始化模板，不是系统边界。同样的克制适用于支付方式、出勤分类——凡是小机构业务中经常变化的枚举值，一律做成可后台维护的字典，而不是写进数据库 CHECK 约束或代码逻辑判断。

### 原则二：默认值可以给，业务不能写死

系统可以且应该预置默认课程模板、默认课时单位、默认支付方式、默认出勤分类、默认扣费建议——这些默认值的存在是为了让 Operator 少打字、快速上手，**不是为了限制业务**。机构必须能在后台自由调整：课程领域/方向/等级/标签、支付方式、本位币（仅初始化阶段）、出勤结果分类及其建议值、单个学生项目的收费、单个老师或师生安排的课酬。

### 原则三：财务上落事实，不依赖回算

所有关键财务动作必须形成不可变的事实快照：充值保存原始币种、原始金额、汇率、折算后金额；课后确认保存实际扣课时、实际学生收费、实际老师课酬；单课财务保存当时的收入、成本、毛利。**历史课次不因未来调价、换老师、改等级、修改出勤分类建议值而改变**——这是保证账目可信的底线原则。

### 原则四：系统给建议，人做最终决定

小型教培机构的真实业务充满例外：熟人价、促销价、临时代课、当天请假扣半节、提前请假不扣、老师到场仍给课酬……这些很难用规则完全覆盖。V1 的处理方式统一为：系统根据配置给出建议值 → Operator 可以修改为任意实际值 → 最终落库的是实际值，不是建议值 → 系统保留调整理由、备注和操作日志。**任何"建议值"字段，在业务语义上都必须是可覆盖的默认值，不能设计成强制生效的计算结果。**

### 原则五：自动化用于提醒，财务动作必须人工确认

V1 可以完全自动化：课前提醒、余额预警、教务晨报、老板周报、失败通知重试。但以下动作必须由人明确确认后才执行：课后确认（触发扣费）、充值确认、退款、作废、结款、数据恢复。**系统不能替运营者自动决定任何涉及金钱的最终结果。**

---

## 第三章 用户角色与核心场景

### 3.1 目标用户画像

Zedu 面向以下类型的小型教培机构或个人运营者：

```
兼职日语/英语课程撮合者
小型 K12 学科辅导工作室
少儿兴趣课/体育课/艺术课小团队
留学辅导、考试辅导、职业培训小机构
目前仍用 Excel、微信、人工提醒、手工核账的运营者
```

共同特点：业务规则灵活，价格和课酬经常靠约定；人员少，系统不能太复杂；老师/学生/家长沟通多靠微信/邮件；账务不一定复杂，但必须可追溯；不想一开始就承担 SaaS、多租户、复杂权限和运维成本。

### 3.2 用户角色与权限

| 角色代码 | 角色名称 | V1是否登录 | 核心诉求 |
|---|:---:|:---:|---|
| `OWNER` | 老板/经营者 | ✅ | 看收入、成本、毛利、待续费、待结款、备份状态；管理系统配置和账号 |
| `OPERATOR` | 教务/运营者 | ✅ | 建档、排课、充值、确认、扣费、结款、发提醒 |
| `TEACHER` | 老师 | ❌ | 接收课程提醒，V2可扩展登录 |
| `STUDENT` | 学生 | ❌ | 接收课程提醒，V2可扩展登录 |
| `PARENT` | 家长 | ❌ | 作为联系人和付款人记录，V2可扩展通知 |

**权限模型**：V1采用简单二级RBAC，Owner权限完全包含Operator（Owner能做Operator能做的一切，外加账号管理、系统配置、备份恢复、完整报表）。所有写操作无条件记录`operation_log`，用审计弥补权限粒度较粗的风险。

**V1明确边界**：Teacher/Student/Parent没有账号、不能登录，只作为通知接收方存在。所有交互通过邮件单向触达，任何变更需求（请假、改时间）仍通过线下渠道联系Operator代为操作。

### 3.3 三个核心用户旅程

#### 场景A：新学生成交

```
Operator新建学生档案
  → 新建课程报名项目（如"初中数学/中考冲刺/初二"）
  → 选择或新建老师，创建师生安排
  → 录入首次充值，上传付款截图
  → 系统形成余额、课时、默认价格、默认老师课酬上下文
  → 后续可直接排课
```

#### 场景B：今天上完课

```
Operator在今日课程或手机确认页看到待确认课次
  → 选择出勤结果分类（正常上课/当天请假/无故缺席/已改期等）
  → 系统带出建议扣课时、建议收费、建议老师课酬
  → Operator根据实际情况修改为实际值
  → 提交后系统在同一事务内写出勤、学生流水、老师流水、单课财务、更新课次状态
```

#### 场景C：月底结款和复盘

```
Owner或Operator查看老师待结款金额
  → 选择老师和结算周期
  → 系统列出未结算课次明细
  → Operator勾选结算范围，填写实付金额和备注
  → 系统写老师结款记录和老师账户流水
  → Owner在报表中查看收入、课酬、毛利、课程方向分布、待续费学生
```

---

## 第四章 产品范围

### 4.1 V1 必须包含（18项）

```
1.  登录与基础权限：Owner / Operator
2.  初始化向导：选择日语模板 / K12学科模板 / 空白模板
3.  学生管理：档案、家长信息、状态、来源备注
4.  老师管理：档案、简介、能力标签、可授时间、默认课酬
5.  课程体系配置：领域、方向、等级、能力标签（后台可维护字典）
6.  学生课程报名：一个学生支持多个课程项目并行
7.  师生安排：支持主老师、代课老师、换老师历史
8.  排课管理：创建/编辑/取消课次、今日课程、时间冲突提示
9.  课后确认：出勤分类、建议值、实际扣课时/收费/课酬
10. 充值管理：多币种、汇率快照、支付方式字典、付款凭证上传
11. 学生流水：充值、扣课、退款、调整、作废冲正
12. 老师流水与结款：应付、结款、调整
13. 单课财务快照：每节课收入、成本、毛利
14. 通知系统：课前提醒、余额预警、教务晨报、老板周报、失败重发
15. 工作台：今日课程、待确认、待续费、待结款、失败通知、基础图表
16. 系统设置：本位币、支付方式、出勤分类、提醒规则、邮件配置、备份配置
17. 数据备份：数据库+配置+付款凭证一并备份
18. All-in-One发布：一个可执行文件 + 一个数据库文件 + 配置/数据目录
```

### 4.2 V1 明确不做（12项）

```
1.  多租户SaaS架构
2.  学生端/老师端/家长端独立登录
3.  自动排课算法（AI匹配、冲突自动求解）
4.  复杂价格规则引擎（price_plan价格模板库）
5.  复杂老师课酬规则引擎（teacher_pay_agreement约定表）
6.  复杂请假规则引擎（多条件自动判责）
7.  在线支付API对接（PayPay/Stripe/微信支付）
8.  Flutter/React Native原生App
9.  微信小程序
10. 完整班级管理和拼班管理（数据结构预留，UI不做）
11. 正式发票/税务申报系统
12. 完整课程内容、题库、作业系统
```

**关于"暂不做"的说明**：第4、5、6项并非"做不到"，而是判断V1阶段用"默认值+手工覆盖+事实快照"的机制已经能达到同等的业务灵活性，只是操作效率上多几次手动录入。触发条件详见第六章6.4、6.5节。

---

# 第二部分：核心业务概念

## 第五章 课程体系（四层配置化结构）

### 5.1 四层结构定义

```
课程领域 course_domain   最上层分类：日语、英语、小学数学、初中物理、钢琴、足球……
    └ 课程方向 course_track   领域下的学习路线：JLPT备考、中考冲刺、少儿兴趣班……
         └ 课程等级 course_level   方向下的阶段：N5-N1、小一到初三、初级/中级/高级……
              └ 能力标签 skill_tag   具体能力点：语法、函数、力学、体能……
```

**设计约束（架构层面的硬性要求）**：

```
不允许在代码中判断"日语""N1""数学"等具体业务词
业务判断必须基于 domain_id/track_id/level_id/skill_tag 的配置数据，而非硬编码枚举
所有模板数据都可修改，不作为内置常量依赖
禁用某个课程配置项不影响历史课次的显示（历史课次引用的是当时的ID快照，不因配置被禁用而消失）
```

### 5.2 初始化模板机制

首次启动时，若 `course_domain` 表为空，系统展示初始化向导：

```
请选择初始化模板：
  ○ 日语培训模板（预置日语/JLPT/会话/商务/少儿）
  ○ K12学科辅导模板（预置数学/物理/化学，按年级分级）
  ○ 空白模板（不预置任何数据，完全自行配置）
  [确认并初始化]
```

初始化模板只是一次性种子数据（通过 `migrations/seed/` 目录下的SQL脚本实现），选择后可以后续在课程体系维护页面自由增删改、启用禁用、改名排序。**首个部署实例（日语教务场景）直接应用日语模板，无需在向导中人工选择——但向导机制作为通用能力保留在系统中，供未来复制部署给其他机构时使用。**

### 5.3 日语模板种子数据

```
课程领域：日语（type=LANGUAGE）
课程方向：JLPT备考 / 日常会话 / 商务日语 / 少儿日语
JLPT等级：入门 → N5 → N4 → N3 → N2 → N1
会话等级：初级 → 中级 → 高级
能力标签：词汇 / 语法 / 阅读 / 听力 / 口语 / 写作 / 综合 / 面试技巧 / 商务敬语
```

### 5.4 K12模板种子数据（保留供未来使用，本次部署不启用）

```
课程领域：小学数学 / 初中数学 / 初中物理 / 初中化学
课程方向：同步辅导 / 期末冲刺 / 中考冲刺 / 专题强化
等级：按年级配置（小一至小六 / 初一至初三）
能力标签：基础概念 / 计算 / 应用题 / 函数 / 几何 / 力学 / 电学 / 实验 / 错题整理
```

### 5.5 老师能力与学生等级变化

老师能力是多条 `teacher_capability` 记录（领域+方向+等级+能力标签+有效期+认证状态），不是老师档案上的单一字段，因为老师可能同时具备多种能力且能力会随时间变化。排课时若检测到老师未标记支持当前课程等级，仅做软提示，不强制拦截——小机构常有临时安排和熟人老师，系统不能把业务卡死。

学生等级变化（如 N3→N2）通过 `student_level_event` 记录事件，保留完整历史，不直接覆盖 `enrollment.current_level_id`。

---

## 第六章 定价与课酬模型

### 6.1 设计原则

小型教培机构价格极其灵活：标准价、熟人价、促销价、试听免费、老客户长期优惠、老师课酬因人而异、代课临时另算、请假扣半节……真实业务不可能用一套统一定价规则覆盖。**V1不做复杂价格规则引擎，但必须支持"默认值 + 手工覆盖 + 事实快照"这一最小闭环。**

### 6.2 三层价格模型

```
第一层：课程报名默认收费
  student_course_enrollment.charge_per_lesson_amount
  表示该学生该课程项目未来课次的默认收费。修改此字段只影响未来新建课次，不影响已产生的历史快照。

第二层：师生安排默认课酬
  student_teacher_assignment.rate_amount
  表示某课程项目下该老师的默认课酬。若为空，取老师档案 teacher.default_rate_amount 作为兜底。

第三层：单课实际快照
  lesson_finance.charge_amount / teacher_pay_amount / gross_profit_amount
  表示本节课最终事实。Operator在课后确认时可覆盖为任意实际金额，提交后此快照不因未来调价而改变。
```

**示例**：老师A默认课酬1200 → 王同学项目单独约定老师A每节1300（写入assignment.rate_amount）→ 某次因病请假找了代课老师B，当次课后确认时手动填入1600（写入lesson_finance.teacher_pay_amount，不影响assignment上的1300）。三层各自独立，互不覆盖，历史可追溯每一层当时的取值。

### 6.3 暂不做的两项（及触发升级条件）

**暂不做价格模板库（`price_plan`）**

原因：小机构初期手工输入更灵活；价格表过早系统化会增加操作负担；`enrollment.charge_per_lesson_amount`已能覆盖大部分需求；单课快照已保证历史正确。

*触发V1.5升级条件*：当每天有大量新签约、手工输入价格成为明显效率瓶颈时，新增可复用的价格模板供快速选择。

**暂不做课酬约定规则表（`teacher_pay_agreement`）**

原因：老师默认课酬 + 师生安排覆盖值 + 单课快照三层已能表达任意课酬场景，只是当一个老师同时带多个学生的同一门课时，需要在每个assignment上分别设置，比"按课程类型设一次自动套用"多几次操作。

*触发V1.5升级条件*：当老师数量增多（如单个老师带课学生超过15~20人）、课酬经常需要按课程方向/等级批量调整时，新增按（老师,领域,方向,等级）维度的课酬约定表。

---

## 第七章 币种与支付设计

### 7.1 系统本位币（Base Currency）

本位币用于余额、报表、毛利统计和结款汇总的统一口径。**本位币不写死为JPY，而是在系统初始化时选择**（JPY/CNY/USD/其他），存储于 `system_config` 表的 `base_currency` 键。

**本位币锁定规则（重要架构决策）**：本位币只能在系统尚未产生任何财务记录（充值、学生流水、老师流水、单课财务、老师结款）时自由设置。**一旦产生第一条财务记录，本位币立即锁定，系统设置页面不再允许直接修改**（对应 `system_config` 中的 `base_currency_locked` 标志位在首次财务写入时自动置为true）。若未来确有必要变更本位币，需要走单独设计的数据迁移工具，V1不提供在线迁移能力。

**本次部署配置**：本位币 = JPY。

### 7.2 每笔充值必须保留原币事实

无论本位币是什么，每笔充值记录必须完整保存：

```
原始金额 original_amount（发生时的实际支付金额，可能是任意币种）
原始币种 original_currency
到本位币的汇率 fx_rate_to_base（发生时的汇率快照，非实时API取值）
折算后本位币金额 amount_base（= original_amount × fx_rate_to_base）
支付方式、付款时间、操作员、凭证附件
```

**示例**：家长微信转账 CNY 500，系统本位币JPY，Operator录入当时汇率21.8，系统计算并保存 `amount_base = 10900`，同时 `original_amount=500.00`、`original_currency=CNY`、`fx_rate_to_base=21.8` 三者永久保留，供任何时候追溯查账。

### 7.3 支付方式字典化

支付方式不用数据库CHECK约束写死，改为独立字典表 `payment_method`，支持后台增删改、启用禁用。默认初始化：微信支付(WECHAT)、支付宝(ALIPAY)、PayPay(PAYPAY)、银行转账(BANK)、现金(CASH)、其他(OTHER)。业务表保存的是`payment_method_code`这个稳定技术值；code一旦创建不建议修改（避免历史记录失去关联），但显示名称`name`可以随时改。

### 7.4 付款凭证附件

每笔充值支持上传付款凭证（截图/PDF），解决小机构现金/转账收款场景下"到底有没有付、付了多少"的纠纷隐患。规则：

```
每笔充值最多3个附件
支持格式：jpg / jpeg / png / webp / pdf
单文件最大5MB
存储路径：data/uploads/payments/{payment_id}/
文件访问必须经过登录鉴权（不允许匿名直链访问）
备份包必须包含 uploads 目录，不能只备份数据库
```

---

## 第八章 课时与出勤处理

### 8.1 课时时长不写死

`lesson.duration_min` 是自由整数字段，允许10~480分钟范围内任意取值，覆盖30分钟辅导课到120分钟强化课等各种真实场景，不假设固定60分钟。

### 8.2 课时扣减支持小数

`attendance.lesson_deducted` 是 `REAL` 类型，V1建议最小颗粒度为0.5课时（如90分钟课按1.5课时扣，请假只扣半节），但数据库层面不限制只能是0.5的倍数，允许任意合理小数。

### 8.3 出勤分类配置化（不用CHECK约束）

`attendance.outcome_type` **不使用数据库CHECK约束**，而是引用一张独立的配置字典表 `attendance_outcome_type`，该表同时承载"建议扣课时/建议收费比例/建议课酬比例"三个建议值列，后台可自由增删改：

| code | 名称 | 建议扣课时 | 建议收费比例 | 建议课酬比例 |
|---|---|---:|---:|---:|
| NORMAL | 正常上课 | 1 | 1.0 | 1.0 |
| STUDENT_LEAVE_EARLY | 学生提前请假 | 0 | 0 | 0 |
| STUDENT_LEAVE_SAMEDAY | 学生当天请假 | 0.5 | 0.5 | 1.0 |
| STUDENT_NOSHOW | 学生无故缺席 | 1 | 1.0 | 1.0 |
| STUDENT_LATE_OR_LEAVE_EARLY | 学生迟到/早退 | （手工）| （手工）| 1.0 |
| TEACHER_LEAVE | 老师请假 | 0 | 0 | 0 |
| RESCHEDULED | 已改期 | 0 | 0 | 0 |
| TECHNICAL_ISSUE | 技术问题无法上课 | 0 | 0 | 0 |
| OTHER | 其他 | （手工）| （手工）| （手工）|

"建议比例"字段为空表示"无建议，前端不自动带值，完全由Operator手动填写"。**建议值仅用于前端表单自动带出默认填充，不构成后端强制校验，Operator在课后确认时可以随意改写为任意实际值。**

### 8.4 课后确认表单字段（更新版，替代此前版本）

```
出勤结果分类：[下拉选择，来自attendance_outcome_type字典]
计划时长：[只读，来自lesson.duration_min]
实际时长：[数字输入，默认带入计划时长]

建议扣课时：[选择分类后自动计算并显示，仅供参考]
实际扣课时：[数字输入，默认带入建议值，可任意修改]

建议学生收费：[= enrollment.charge_per_lesson_amount × 建议收费比例，仅供参考]
实际学生收费：[数字输入，默认带入建议值，可任意修改]

建议老师课酬：[= assignment.rate_amount或teacher.default_rate_amount × 建议课酬比例，仅供参考]
实际老师课酬：[数字输入，默认带入建议值，可任意修改]

老师备注：[多行文本，选填]
运营备注：[多行文本，选填]

[确认提交]（提交后不可直接删除或撤销，二次确认弹窗提示；若需修正，走"人工调整流水"流程并记录原因）
```

**审计要求**：`attendance` 表须同时保存"当时展示给Operator的建议值"和"Operator最终提交的实际值"两组数据（详见第十二章DDL），确保未来能追溯"系统建议了什么、人做了什么决定"，这对处理纠纷和优化建议规则都有价值。


---

# 第三部分：业务流程

## 第九章 核心业务流程详解（含异常分支）

### 9.1 学生建档与课程报名

**主流程**：

```
学生咨询/报名 → Operator录入学生档案（姓名*/邮箱/电话/国籍/时区/备注）
  → （可选）录入家长联系方式，可多条，标记主联系人
  → 创建课程报名项目：选课程领域→方向→当前等级→目标等级→报名类型（一对一/小班/试听）
  → 设置每课次默认收费（可先留空，充值时再确定）
  → 匹配老师，创建师生安排
  → 录入首次充值，上传付款截图
  → 进入排课流程
```

**异常分支**：

| 情况 | 处理方式 |
|---|---|
| 邮箱与已有学生重复 | 阻止创建或更新，返回`40901`并提示已有学生；不得提供“仍然新建”绕过路径 |
| 学生暂无邮箱 | 允许保存，但明确提示"未填写邮箱，将无法接收自动提醒" |
| 课程方向下暂无合适老师 | 允许先创建报名项目、暂不指定老师，排课时若无老师则不允许创建课次 |
| 试听转正式 | `enrollment_type`从TRIAL改为ONE_TO_ONE，历史课次不受影响 |

### 9.2 老师建档与能力维护

```
录入老师基础信息（姓名*/邮箱/电话）→ 录入简介与教学经历
  → 新增能力记录（可多条：领域+方向+等级+能力标签）→ 设置默认课酬
  → 维护可授时间（可多条：周几+起止时间）
  → 后续能力变化时，新增或结束能力记录（effective_to置为今日），不直接覆盖历史
```

**异常分支**：老师能力记录唯一性由`(teacher_id, track_id, level_id)`组合约束，重复添加需走编辑而非新增；老师暂停接单时`status=PAUSED`，排课候选列表自动排除，历史数据保留。

### 9.3 课程报名与师生匹配（多对多模型）

**核心设计**：学生可同时报名多个课程方向，每个方向可配不同老师，不采用"学生绑定单一老师"的简化模型。核心关系链：`学生 → 课程报名(enrollment) → 师生安排(assignment) → 具体课次(lesson)`。

**换老师主流程**：

```
Operator在课程报名详情页点击"更换老师"
  → 弹窗选择新老师，系统显示新老师能力标签供参考（软提示，不强制拦截）
  → 填写更换原因（选填）
  → 确认更换：旧assignment.status=ENDED记录end_date；新assignment.status=ACTIVE
  → enrollment上的余额和课时不变（随学生走，不随老师走）
  → 若存在未来SCHEDULED状态课次，提示是否批量更新为新老师（Operator逐条确认或批量处理）
  → 历史（已完成）课次不受影响，保留原老师快照
```

### 9.4 排课与课前提醒

```
Operator创建课次：选学生→自动列出ACTIVE课程报名项目→选项目
  → 自动代入该项目当前ACTIVE的assignment对应老师（可改选）
  → 选上课时间（前端按系统时区展示，后端存UTC）、时长、上课方式、链接、（可选）本节课主题
  → 保存，lesson.status=SCHEDULED
  → 定时任务每10分钟扫描，课前N分钟（默认30，可配置）发送提醒邮件给学生+老师
  → 更新remind_sent_at（幂等标记，防止重复发送），写通知日志
```

**异常分支**：

| 情况 | 处理方式 |
|---|---|
| 排课时余额不足 | 前端警示（黄色提示条），不强制拦截创建 |
| 老师能力不匹配当前等级 | 软提示，不拦截 |
| 学生/老师在同一时段已有其他课次 | 提示时间冲突详情，Operator可选择"仍然创建" |
| 学生/老师无邮箱 | 提醒任务扫描时跳过无邮箱一方，仅通知有邮箱一方 |
| 提醒发送失败 | 写入notification_log(status=FAILED)，30分钟后自动重试，最多3次，仍失败进入晨报异常列表 |

### 9.5 课后确认与账务（核心事务）

```
课程结束 → Operator选择出勤结果分类 → 系统带出建议扣课时/收费/课酬 → Operator修改为实际值 → 提交
  ↓（同一数据库事务内完成，任一步失败全部回滚）
  ├─ 写attendance（含建议值快照与实际值）
  ├─ 若实际扣课时>0：写student_account_ledger（LESSON_DEDUCT），更新enrollment余额缓存
  ├─ 若实际老师课酬>0：写teacher_account_ledger（LESSON_PAYABLE），更新teacher待结款缓存
  ├─ 写lesson_finance（实际收费/实际课酬/毛利快照）
  └─ 更新lesson.status=COMPLETED
  ↓（事务提交后）
  判断enrollment是否触发余额预警条件，若已为0或负数可立即触发即时提醒，否则等每日20:00扫描
```

**异常分支**：

| 情况 | 处理方式 |
|---|---|
| 学生请假是否扣费 | 由出勤分类的建议比例带出默认值，Operator最终决定，系统不做强制判责 |
| 老师缺席 | 不扣学生课时，不产生老师应付款，可能需另行安排补课（新建课次） |
| 提交时系统崩溃/网络中断 | 依赖数据库事务原子性自动回滚，不会出现"扣了课时但没记课酬"的半成功状态 |
| 误操作确认了错误课次 | V1不支持撤销课后确认，走"人工调整流水"（biz_type=ADJUST）修正并详细记录原因 |
| 忘记确认，课次一直挂起 | 定时任务在课次结束4小时后自动将lesson.status置为COMPLETED，但不自动触发账务，工作台"待确认"列表持续提醒 |

### 9.6 充值与余额

```
学生/家长线下付款 → Operator打开学生详情页"新建充值"
  → 选课程报名项目 → 选支付方式（来自payment_method字典）
  → 输入原始金额+原始币种 → 若非本位币，输入汇率，系统自动计算折算金额
  → 选课时套餐或手动输入课时数 → 录入付款日期，上传付款截图（最多3张）
  → 提交：写student_payment（status=CONFIRMED）+ student_account_ledger（RECHARGE）+ 更新enrollment余额缓存
```

**异常分支**：充值录错不允许直接编辑/删除，须"作废"（status=VOIDED，同时生成反向ledger冲正）后重新录入；退款走`biz_type=REFUND`，允许负向金额调整，需备注原因。

### 9.7 结款

```
Operator打开老师详情页"新建结款" → 选结算周期
  → 系统自动查询该周期内未被结算的teacher_account_ledger(LESSON_PAYABLE)记录
  → 列表展示明细，Operator可勾选排除个别记录 → 确认或调整实付金额
  → 提交：写teacher_payout + teacher_account_ledger（PAYOUT，负向冲抵）+ 更新teacher.unpaid_amount
```

---


# 第四部分：功能与界面

## 第十章 功能模块与页面级UI规格

页面遵循Soybean Admin设计语言（卡片化、圆角、留白充分），字段级规格供前端开发直接依据。

### 10.1 全局导航结构

```
顶部栏：Logo+系统名 | 通知铃铛（未读角标）| 用户头像下拉
左侧栏：工作台 / 学员管理(学生列表+充值记录) / 教师管理(老师列表+结款记录) /
       课程体系 / 课程管理(排课列表+日历视图) /
       财务管理(收支概览+学生账单+老师账单+报表导出) / 数据图表 /
       通知管理(提醒规则+通知日志) / 系统设置
```

移动端（<768px）：左侧栏收起为底部Tab Bar：`今日` `确认` `学生` `更多`。

### 10.2 工作台（Dashboard）— `/dashboard`

**Owner视图**：本月收入/本月课酬支出/本月毛利/活跃学生数（指标卡）+ 待续费学生 + 待结款老师 + 失败通知 + 最近备份时间 + 图表区（月度收入趋势折线图/课程方向占比饼图/学生增长趋势面积图）

**Operator视图**：今日课程数/待续费学生数/待结款老师数/待确认课次数（指标卡）+ 今日课程列表（时间/学生/老师/上课方式/状态/[查看链接][手动发提醒][确认出勤]）+ 待续费学生卡片 + 待结款老师卡片 + 快速充值入口 + 移动端确认入口

失败通知存在时，橙色警示条置顶显示。

### 10.3 学生列表 — `/students`

顶部：[+新建学生][导入Excel][导出Excel] + 搜索框（姓名/邮箱/电话）+ 筛选（状态/课程方向）

表格列：姓名/邮箱/电话/学习方向(标签)/当前活跃课程项目数/总剩余课时/总余额/最近上课时间/状态/操作(查看/编辑/快速充值)

### 10.4 学生详情页 — `/students/:id`

**左侧基础信息卡**：姓名/日文名/邮箱/电话/国籍/时区/状态(可切换)/备注/来源渠道(选填) + 家长信息子卡(可多条)

**右侧Tab页签**：
- Tab1「学习项目」：课程报名卡片列表（方向+等级/当前老师/剩余课时/余额/默认收费/状态/[查看课次][换老师][暂停恢复][编辑]）
- Tab2「账务」：充值记录表格(含凭证查看入口) + 账户流水表格 + [新建充值][导出账单PDF]
- Tab3「上课记录」：历史课次表格，按项目/日期筛选，显示出勤分类
- Tab4「通知记录」：发送记录，失败可[重发]

### 10.5 老师列表/详情页 — `/teachers`、`/teachers/:id`

列表额外列：可教方向/在带学生数/累计应付未结款(红色高亮>0)

详情页Tab：能力与时间（能力标签列表+可授时间周视图）/ 账务（应付概览+结款记录+课时流水）/ 带课记录 / 通知记录

### 10.6 课程体系维护 — `/courses`

三栏联动：左-课程领域列表 / 中-选中领域的方向列表 / 右-选中方向的等级列表+该领域能力标签管理。每项支持增删改、启用禁用、拖拽排序。首次进入若无数据，触发初始化模板向导（详见第五章5.2节）。

### 10.7 排课管理 — `/lessons`、`/lessons/calendar`

**列表视图**：[+新建课次] + 筛选（日期范围/学生/老师/状态多选）
表格列：日期时间/学生/老师/课程方向-等级/上课方式/状态(SCHEDULED蓝/REMINDED橙/COMPLETED绿/CANCELLED灰)/操作
操作：[查看][编辑](仅SCHEDULED)[取消][确认出勤](时间已过未确认时高亮)

**日历视图**：月/周切换，色块展示，点击弹详情浮层；移动端自动降级为按日分组列表

**新建课次表单**：学生*→课程报名项目(联动)* / 老师*(自动带入，可改选) / 上课日期*+开始时间*+时长(默认60分钟，10-480范围)* / 上课方式*(微信群/腾讯会议/Zoom/其他) / 上课链接 / 本节课主题(选填) / 备注

**课后确认表单**：详见第八章8.4节完整字段规格

### 10.8 充值记录页 — `/finance/payments`

[+新建充值] + 筛选（学生/日期范围/支付方式/状态）
表格列：单号/日期/学生/项目/原始金额币种/折算本位币/课时/支付方式/凭证图标/状态/操作
VOIDED状态整行置灰仅显示[查看]；CONFIRMED状态显示[查看][作废](二次确认+填写原因)

**新建充值表单**：选学生→选课程报名项目 / 支付方式(下拉，来自payment_method字典) / 原始金额+原始币种 / 汇率(非本位币时必填，系统自动计算折算金额) / 课时套餐或手动输入课时数 / 付款日期 / 上传凭证(最多3张，jpg/jpeg/png/webp/pdf，单文件≤5MB) / 备注

### 10.9 结款记录页 — `/finance/payouts`

结构对称充值记录页，新建结款先"预览"待结算明细列表（可勾选排除），确认后提交。

### 10.10 财务报表 — `/finance/report`

日期范围选择器（本月/上月/本季度/本年/自定义）+ 汇总卡片（总收入/总课酬支出/总毛利/毛利率）+ 明细表格（按月/周/日切换）+ [导出Excel][导出PDF]

### 10.11 数据图表页 — `/reports`

学生增长曲线(折线) / 月度课时完成率(堆叠柱状图) / 收入差价走势(三线折线图) / 各老师带课分布(横向柱状图) / 各课程方向收入占比(环形图)，每图表支持[导出PNG]

### 10.12 通知管理 — `/notifications`

Tab1「提醒规则」：课前提醒分钟数(默认30) / 余额预警阈值(默认3课时) / 晨报时间(默认08:00) / 周报发送日+时间 / 通知语言(中文/日文/双语)

Tab2「通知日志」：筛选（类型/状态/日期/收件人）+ 表格 + 失败行[重发]

### 10.13 系统设置 — `/settings`

```
Tab「基本信息」：机构名称/Logo/联系方式/系统时区
Tab「本位币设置」：显示当前本位币；若已产生财务记录则字段禁用编辑并提示已锁定
Tab「支付方式管理」：payment_method字典的增删改/启用禁用
Tab「出勤分类管理」：attendance_outcome_type字典的增删改，含建议扣课时/建议收费比例/建议课酬比例三列
Tab「课程模板」：查看/重新应用初始化模板（仅限尚无课程数据时可用）
Tab「课时套餐」：套餐管理列表（名称/金额/课时数/启用状态）
Tab「邮件配置」：Resend API Key(脱敏)/发件人名称/发件邮箱/SMTP备用配置
Tab「数据备份」(仅Owner)：自动备份开关+Cron配置/Litestream云备份配置/备份历史+[立即备份][下载][恢复]
Tab「账号管理」(仅Owner)：Operator账号列表+[新建账号][禁用][重置密码]
Tab「操作日志」：全量操作记录，筛选+搜索
```

### 10.14 首次启动初始化向导 — `/onboarding`

系统检测到无课程数据时的一次性引导页，字段规格详见第五章5.2节。

---

## 第十一章 移动端策略

V1不单独开发移动端项目，在Soybean Admin内提供移动优先响应式页面：

```
/mobile/today      今日课程（卡片列表，大触摸区域，[复制链接][确认出勤]大按钮）
/mobile/confirm    待确认课次（内嵌简化版课后确认表单：出勤分类下拉+关键金额+提交）
/mobile/recharge   快速充值（学生搜索大输入框+极简表单）
/mobile/alerts     待续费学生（[一键拨号][复制微信号][去充值]快捷按钮）
```

设计原则：少表格、多卡片、大按钮、少字段、底部快捷导航，触摸区域≥44×44px。


---

# 第五部分：数据与接口

## 第十二章 数据模型设计（完整DDL）

### 12.1 设计原则

```
金额字段统一使用整数（本位币最小单位）存储，字段名统一为 *_amount 或 amount_base，
  不使用 _jpy 等币种后缀（本位币由system_config.base_currency决定，可能是任意币种）
Go层使用shopspring/decimal处理金额运算
所有时间字段存储UTC，应用层转时区
业务枚举优先使用"字典表+外键引用"而非数据库CHECK约束，尤其是支付方式、出勤分类这类
  小机构高频需要自定义的字段；只有真正稳定不变的状态类枚举（如订单状态）才用CHECK
每张业务表预留 extra_json TEXT 字段
软删除使用 deleted_at 字段
外键显式声明，启动时 PRAGMA foreign_keys=ON
```

### 12.2 完整DDL（SQLite方言）

```sql
-- ========== 系统与权限 ==========

CREATE TABLE user_account (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL CHECK(role IN ('OWNER','OPERATOR')),
  display_name TEXT NOT NULL,
  email TEXT UNIQUE,
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','DISABLED')),
  last_login_at DATETIME,
  login_fail_count INTEGER NOT NULL DEFAULT 0,
  locked_until DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);

-- 系统配置（key-value），承载 base_currency / base_currency_locked / 各类提醒规则等
CREATE TABLE system_config (
  config_key TEXT PRIMARY KEY,
  config_value TEXT NOT NULL,
  description TEXT,
  updated_by INTEGER REFERENCES user_account(id),
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- 初始化关键配置项（示例，实际由初始化向导写入）：
-- ('base_currency', 'JPY', '系统经营本位币')
-- ('base_currency_locked', '0', '本位币是否已锁定，产生首条财务记录后自动置1')
-- ('lesson_reminder_minutes', '30', '课前提醒分钟数')
-- ('balance_alert_lessons', '3', '余额预警课时阈值')

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

-- ========== 课程体系（配置化字典） ==========

CREATE TABLE course_domain (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  code TEXT NOT NULL UNIQUE,
  type TEXT NOT NULL CHECK(type IN ('LANGUAGE','K12','SPORT','ART','ACADEMIC','CERTIFICATE','OTHER')),
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

-- ========== 学生与老师 ==========

CREATE TABLE student (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  name_local TEXT,  -- 本地语言姓名（如日文名），字段名泛化，不叫name_jp
  email TEXT,
  phone TEXT,
  nationality TEXT,
  timezone TEXT NOT NULL DEFAULT 'Asia/Tokyo',
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','ENDED')),
  source_channel TEXT,  -- 来源渠道/介绍人，选填
  note TEXT,
  extra_json TEXT,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
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
  name_local TEXT,
  email TEXT,
  phone TEXT,
  bio TEXT,
  default_rate_amount INTEGER NOT NULL DEFAULT 0,  -- 默认课酬，本位币最小单位
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
  skill_tag_codes TEXT,  -- JSON数组
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
  charge_per_lesson_amount INTEGER NOT NULL DEFAULT 0,  -- 第一层价格：默认每课次收费
  lesson_balance REAL NOT NULL DEFAULT 0,
  balance_amount INTEGER NOT NULL DEFAULT 0,
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
  rate_amount INTEGER,  -- 第二层价格：该安排下老师课酬，NULL则取teacher.default_rate_amount
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

-- ========== 班级（预留，V1数据结构保留，UI不做） ==========

CREATE TABLE class_group (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  domain_id INTEGER NOT NULL REFERENCES course_domain(id),
  track_id INTEGER REFERENCES course_track(id),
  level_id INTEGER REFERENCES course_level(id),
  main_teacher_id INTEGER REFERENCES teacher(id),
  status TEXT NOT NULL DEFAULT 'ACTIVE' CHECK(status IN ('ACTIVE','PAUSED','FINISHED')),
  charge_per_lesson_amount INTEGER NOT NULL DEFAULT 0,
  rate_per_lesson_amount INTEGER NOT NULL DEFAULT 0,
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
  duration_min INTEGER NOT NULL DEFAULT 60 CHECK(duration_min BETWEEN 10 AND 480),
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

-- 出勤结果分类字典（配置化，非CHECK约束）
CREATE TABLE attendance_outcome_type (
  code TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  suggested_deduct_lessons REAL,        -- NULL=无建议，前端不自动带值
  suggested_charge_ratio REAL,          -- 相对enrollment.charge_per_lesson_amount的比例，NULL=无建议
  suggested_teacher_pay_ratio REAL,     -- 相对老师课酬的比例，NULL=无建议
  sort_order INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- 初始化默认数据（见第八章8.3节表格）

CREATE TABLE attendance (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  lesson_id INTEGER NOT NULL UNIQUE REFERENCES lesson(id),
  enrollment_id INTEGER REFERENCES student_course_enrollment(id),
  student_id INTEGER NOT NULL REFERENCES student(id),
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  outcome_type TEXT NOT NULL REFERENCES attendance_outcome_type(code),  -- 字典引用，非CHECK
  actual_start_at DATETIME,
  actual_end_at DATETIME,
  actual_duration_min INTEGER,
  student_attended INTEGER NOT NULL DEFAULT 1,
  teacher_attended INTEGER NOT NULL DEFAULT 1,
  -- 审计用：提交时系统展示的建议值快照
  suggested_deduct_lessons REAL,
  suggested_charge_amount INTEGER,
  suggested_teacher_pay_amount INTEGER,
  -- 最终落库的实际值（真正影响流水的字段）
  lesson_deducted REAL NOT NULL DEFAULT 0,
  charge_amount INTEGER NOT NULL DEFAULT 0,
  teacher_pay_amount INTEGER NOT NULL DEFAULT 0,
  teacher_note TEXT,
  operator_note TEXT,
  confirmed_by INTEGER REFERENCES user_account(id),
  confirmed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ========== 账务体系 ==========

-- 支付方式字典（配置化，非CHECK约束）
CREATE TABLE payment_method (
  code TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- 初始化：WECHAT微信支付 / ALIPAY支付宝 / PAYPAY / BANK银行转账 / CASH现金 / OTHER其他

CREATE TABLE student_payment (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  payment_no TEXT NOT NULL UNIQUE,
  student_id INTEGER NOT NULL REFERENCES student(id),
  enrollment_id INTEGER NOT NULL REFERENCES student_course_enrollment(id),
  original_amount TEXT NOT NULL,        -- decimal字符串，原始金额
  original_currency TEXT NOT NULL,      -- 原始币种，不假设JPY
  fx_rate_to_base TEXT NOT NULL DEFAULT '1',  -- 到本位币汇率快照
  amount_base INTEGER NOT NULL,         -- 折算后本位币金额（整数最小单位）
  lessons_added REAL NOT NULL DEFAULT 0,
  package_name TEXT,
  payment_method_code TEXT NOT NULL REFERENCES payment_method(code),
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

-- 付款凭证附件
CREATE TABLE payment_attachment (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  payment_id INTEGER NOT NULL REFERENCES student_payment(id),
  file_name TEXT NOT NULL,
  file_path TEXT NOT NULL,
  file_type TEXT,
  file_size INTEGER,
  uploaded_by INTEGER REFERENCES user_account(id),
  uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_attachment_payment ON payment_attachment(payment_id);

CREATE TABLE teacher_payout (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  payout_no TEXT NOT NULL UNIQUE,
  teacher_id INTEGER NOT NULL REFERENCES teacher(id),
  period_start DATE NOT NULL,
  period_end DATE NOT NULL,
  lesson_count REAL NOT NULL DEFAULT 0,
  amount_base INTEGER NOT NULL DEFAULT 0,        -- 应付金额
  actual_amount_base INTEGER NOT NULL DEFAULT 0, -- 实付金额
  payment_method_code TEXT REFERENCES payment_method(code),
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
  amount_delta INTEGER NOT NULL,        -- 本位币变动金额（可负）
  lesson_delta REAL NOT NULL DEFAULT 0,
  balance_after INTEGER NOT NULL,
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
  amount_delta INTEGER NOT NULL,
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
  charge_amount INTEGER NOT NULL DEFAULT 0,        -- 第三层价格：单课实际收费
  teacher_pay_amount INTEGER NOT NULL DEFAULT 0,   -- 第三层价格：单课实际课酬
  gross_profit_amount INTEGER NOT NULL DEFAULT 0,  -- = charge_amount - teacher_pay_amount
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_finance_created ON lesson_finance(created_at);

CREATE TABLE fx_rate_snapshot (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  from_currency TEXT NOT NULL,
  to_currency TEXT NOT NULL,   -- 动态等于system_config.base_currency，不硬编码JPY
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

### 12.3 启动时必须执行的PRAGMA

```sql
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;
PRAGMA synchronous=NORMAL;
```

### 12.4 表清单汇总（31张）

```
系统与权限(4)：user_account / system_config / operation_log / backup_log
课程体系(4)：course_domain / course_track / course_level / skill_tag
学生与老师(5)：student / parent / teacher / teacher_availability / teacher_capability
学习项目(4)：student_course_enrollment / student_teacher_assignment /
             student_learning_path / student_level_event
班级预留(2)：class_group / class_group_member
排课上课(3)：lesson / attendance_outcome_type / attendance
账务体系(7)：payment_method / student_payment / payment_attachment /
             teacher_payout / student_account_ledger / teacher_account_ledger /
             lesson_finance / fx_rate_snapshot
通知体系(2)：notification_template / notification_log
```

（说明：相比v2.0的29张，本版净增支付方式字典、付款凭证、出勤分类字典3张表，减少独立的price_plan/enrollment_pricing_agreement/teacher_pay_agreement/attendance_policy_template等4张此前设想但已判定暂不需要的表）

### 12.5 ER关系图（文字版）

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
lesson 1─1 attendance ─N─1 attendance_outcome_type
lesson 1─1 lesson_finance
student_course_enrollment 1─N student_payment ─N─1 payment_method
student_payment 1─N payment_attachment
student 1─N student_account_ledger
teacher 1─N teacher_account_ledger
teacher 1─N teacher_payout ─N─1 payment_method
lesson 1─N notification_log
```


## 第十三章 API接口规范

### 13.1 通用约定

```
Base URL: /api/v1
认证: Authorization: Bearer <JWT access_token>
时间格式: ISO8601 UTC
分页: page(默认1)、pageSize(默认20，最大100)
统一响应: {code, message, data, traceId}
```

未来App、独立mobile、老师端、学生端都复用同一套`/api/v1`，不分裂多套API。

### 13.2 认证与用户

```
POST /auth/login              登录
POST /auth/refresh            刷新Token
POST /auth/logout             登出
GET  /auth/me                 当前用户信息
POST /auth/change-password    修改密码
GET  /users                   Operator账号列表（仅Owner）
POST /users                   新建Operator账号（仅Owner）
POST /users/{id}/disable      禁用账号（仅Owner）
POST /users/{id}/reset-password  重置密码（仅Owner）
```

### 13.3 初始化向导（新增）

```
GET  /init/status             检测是否需要初始化（course_domain是否为空）
GET  /init/templates          获取可选模板列表（日语/K12/空白）
POST /init/apply-template     应用选定模板，写入种子数据
```

### 13.4 学生与家长

```
GET    /students                       列表（分页/搜索/筛选）
POST   /students                       新建
GET    /students/{id}                  详情
PUT    /students/{id}                  编辑
POST   /students/{id}/status           变更状态
GET    /students/{id}/enrollments      课程报名列表
GET    /students/{id}/lessons          课次记录
GET    /students/{id}/ledger           账户流水
GET    /students/{id}/notifications    通知记录
POST   /students/import                Excel批量导入
GET    /students/export                Excel导出
GET    /students/import-template       下载导入模板
GET    /students/{studentId}/parents   家长列表
POST   /students/{studentId}/parents   新建家长
PUT    /parents/{id}                   编辑
DELETE /parents/{id}                   删除
POST   /parents/{id}/set-primary       设为主联系人
```

### 13.5 老师

```
GET    /teachers                       列表
POST   /teachers                       新建
GET    /teachers/{id}                  详情
PUT    /teachers/{id}                  编辑
POST   /teachers/{id}/status           变更状态
GET    /teachers/{id}/capabilities     能力列表
POST   /teachers/{id}/capabilities     新建能力
PUT    /teacher-capabilities/{id}      编辑能力
DELETE /teacher-capabilities/{id}      删除能力
GET    /teachers/{id}/availability     可授时间
POST   /teachers/{id}/availability     新建可授时间
GET    /teachers/{id}/students         带课学生列表
GET    /teachers/{id}/ledger           账务流水
POST   /teachers/import                Excel批量导入
GET    /teachers/export                Excel导出
```

### 13.6 课程体系

```
GET  /courses/domains                    领域列表
POST /courses/domains                    新建领域
PUT  /courses/domains/{id}               编辑
GET  /courses/domains/{id}/tracks        方向列表
POST /courses/tracks                     新建方向
PUT  /courses/tracks/{id}                编辑
GET  /courses/tracks/{id}/levels         等级列表
POST /courses/levels                     新建等级
PUT  /courses/levels/{id}                编辑
GET  /courses/domains/{id}/skill-tags    能力标签列表
POST /courses/skill-tags                 新建标签
PUT  /courses/skill-tags/{id}            编辑
```

### 13.7 课程报名与师生安排

```
GET  /enrollments                                          列表
POST /enrollments                                          新建
GET  /enrollments/{id}                                     详情
PUT  /enrollments/{id}                                     编辑（含charge_per_lesson_amount）
POST /enrollments/{id}/status                              变更状态
GET  /enrollments/{id}/assignments                         师生安排历史
POST /enrollments/{id}/assignments/change-teacher          换老师
GET  /enrollments/{id}/learning-paths                      学习路径历史
POST /enrollments/{id}/learning-paths                      新建学习路径
GET  /enrollments/{id}/level-events                        等级变化事件
POST /enrollments/{id}/level-events                        新建等级变化事件
```

### 13.8 排课与上课

```
GET  /lessons                     列表
GET  /lessons/calendar            日历视图数据
POST /lessons                     新建课次
GET  /lessons/{id}                详情
PUT  /lessons/{id}                编辑（仅SCHEDULED）
POST /lessons/{id}/cancel         取消
POST /lessons/{id}/confirm        课后确认（见下方请求体）
POST /lessons/{id}/remind         手动触发提醒
GET  /lessons/{id}/conflicts      时间冲突检测
```

**课后确认请求体（更新版）**：

```json
POST /api/v1/lessons/9001/confirm
{
  "outcomeType": "STUDENT_LEAVE_SAMEDAY",
  "actualDurationMin": 50,
  "studentAttended": false,
  "teacherAttended": true,
  "lessonDeducted": 0.5,
  "chargeAmount": 1000,
  "teacherPayAmount": 1500,
  "teacherNote": "学生当天请假，老师已等待",
  "operatorNote": "按约定扣半节"
}
```

**响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "lessonId": 9001,
    "status": "COMPLETED",
    "studentBalanceAfter": 19200,
    "lessonBalanceAfter": 6,
    "teacherUnpaidAfter": 16800
  },
  "traceId": "20260704-abc123"
}
```

### 13.9 账务

```
GET  /finance/payments                          充值记录列表
POST /finance/payments                          新建充值
GET  /finance/payments/{id}                     详情
POST /finance/payments/{id}/void                作废
POST /finance/payments/{id}/attachments         上传付款凭证
GET  /finance/payments/{id}/attachments         凭证列表
GET  /finance/payouts                           结款记录列表
POST /finance/payouts/preview                   预览待结算明细
POST /finance/payouts                           提交结款
GET  /finance/ledger/student/{studentId}        学生流水
GET  /finance/ledger/teacher/{teacherId}        老师流水
POST /finance/ledger/adjust                     人工调整流水
GET  /finance/report                            财务汇总报表
GET  /finance/report/export                     导出报表
GET  /finance/packages                          课时套餐列表
POST /finance/packages                          新建套餐
PUT  /finance/packages/{id}                     编辑套餐
```

**新建充值请求体（更新版）**：

```json
POST /api/v1/finance/payments
{
  "studentId": 1001,
  "enrollmentId": 5001,
  "originalAmount": "500.00",
  "originalCurrency": "CNY",
  "fxRateToBase": "21.80",
  "lessonsAdded": 10,
  "packageName": "10课时标准包",
  "paymentMethodCode": "WECHAT",
  "paidAt": "2026-07-04T10:00:00Z",
  "note": "暑期续费"
}
```

### 13.10 报表

```
GET /reports/dashboard              工作台聚合数据
GET /reports/revenue-trend          月度收入趋势
GET /reports/student-growth         学生增长曲线
GET /reports/track-distribution     课程方向分布
GET /reports/teacher-workload       老师带课分布
GET /reports/completion-rate        课时完成率
```

### 13.11 通知

```
GET /notifications/logs                通知日志列表
POST /notifications/logs/{id}/resend   重发
GET /notifications/templates           模板列表
PUT /notifications/templates/{code}    编辑模板
GET /notifications/rules               提醒规则配置
PUT /notifications/rules               更新提醒规则
```

### 13.12 系统设置（新增/调整）

```
GET  /system/config                     获取系统配置
PUT  /system/config                     更新系统配置（不含base_currency）
GET  /system/base-currency              获取本位币及锁定状态
PUT  /system/base-currency              修改本位币（若已锁定则返回42201错误）
GET  /system/payment-methods            支付方式字典列表
POST /system/payment-methods            新增支付方式
PUT  /system/payment-methods/{code}     编辑支付方式
GET  /system/attendance-outcomes        出勤分类字典列表
POST /system/attendance-outcomes        新增出勤分类
PUT  /system/attendance-outcomes/{code} 编辑出勤分类（含建议值）
GET  /system/operation-logs             操作日志
GET  /system/fx-rates                   参考汇率列表
PUT  /system/fx-rates                   更新参考汇率
```

### 13.13 备份

```
GET  /backup/logs               备份历史列表
POST /backup/trigger            立即手动备份
GET  /backup/{id}/download      下载备份文件
POST /backup/{id}/restore       恢复数据（需二次确认参数confirmText="CONFIRM"）
```

### 13.14 健康检查

```
GET /healthz   { "status": "ok", "version": "3.0.0", "uptime": 12345 }
```

---

## 第十四章 业务规则与校验总表

### 14.1 字段校验规则速查

| 字段 | 规则 |
|---|---|
| student.name | 必填，1~50字符 |
| student.email | 选填，标准邮箱格式；填写时全局唯一，重复返回40901 |
| teacher.default_rate_amount | 必填，≥0整数 |
| enrollment.charge_per_lesson_amount | 必填，≥0整数 |
| lesson.duration_min | 必填，10~480之间整数 |
| lesson.scheduled_start_at | 必填，合法日期时间 |
| student_payment.original_amount | 必填，>0，decimal字符串，最多2位小数 |
| student_payment.fx_rate_to_base | 必填，>0，decimal字符串；若original_currency=base_currency则固定为"1" |
| student_payment.lessons_added | 必填，>0 |
| attendance.actual_duration_min | 选填，若填写须≤lesson.duration_min×2（防异常录入）|
| attendance.lesson_deducted | 必填，≥0，允许小数 |
| user_account.password | ≥8位，须含字母+数字 |
| course_domain/track/level.code | 必填，同层级内唯一，仅字母数字下划线 |
| payment_method.code | 必填，唯一，创建后不建议修改 |
| attendance_outcome_type.code | 必填，唯一，创建后不建议修改 |

### 14.2 关键业务规则清单

```
R1  本位币在产生首条财务记录后自动锁定（base_currency_locked=1），系统设置页面禁止直接修改
R2  学生status=ENDED后不可再新建课次或课程报名，历史数据只读可查
R3  老师status=ENDED后不可再被指派新课次，历史数据只读可查
R4  enrollment.status=CANCELLED/COMPLETED后不可再排课
R5  同一enrollment下同一role_type=MAIN只能有一个ACTIVE的assignment
R6  lesson.status=COMPLETED后不可再编辑基础信息（时间/老师），仅可编辑note
R7  attendance一旦创建不可删除、不可修改（修正走ADJUST流水）
R8  student_payment.status=VOIDED的记录不计入任何余额和报表统计
R9  charge_per_lesson_amount可随时修改，只影响未来新建课次，不影响历史lesson_finance快照
R10 teacher_capability的(teacher_id,track_id,level_id)组合唯一
R11 通知重试最多3次，超过后需人工手动重发
R12 一个lesson只能对应一条attendance记录（唯一约束）
R13 备份恢复操作会完全覆盖当前数据库，操作前自动先做一次快照备份
R14 Excel导入时若邮箱已存在，整行跳过并在导入报告中标注，不做自动合并
R15 结款预览时，已被其他payout关联的LESSON_PAYABLE记录不重复出现
R16 payment_method.code和attendance_outcome_type.code创建后不建议修改（避免历史记录关联失效），仅显示名称name可修改
R17 attendance_outcome_type的建议值（suggested_*字段）仅用于前端表单默认填充，不构成后端强制校验，不会覆盖Operator手动填写的实际值
R18 付款凭证仅支持单笔充值最多3个附件，不做多级图片管理/裁剪/压缩等增强功能
R19 attendance表的suggested_*快照字段与实际落库字段(lesson_deducted/charge_amount/teacher_pay_amount)必须同时存在，前者用于审计追溯"系统建议了什么"，后者是真正生效的业务事实
R20 所有账务相关的多表写入（课后确认、充值确认、充值作废、结款提交）必须在单一数据库事务内完成
```

### 14.3 业务错误码总表

| 错误码 | HTTP状态 | 说明 |
|---|---|---|
| 0 | 200 | 成功 |
| 40001 | 400 | 参数校验失败 |
| 40002 | 400 | 学生余额不足 |
| 40003 | 400 | 时间冲突 |
| 40101 | 401 | 未登录或Token过期 |
| 40102 | 401 | 用户名或密码错误 |
| 40103 | 401 | 账号已被锁定 |
| 40301 | 403 | 权限不足 |
| 40401 | 404 | 资源不存在 |
| 40901 | 409 | 数据冲突（邮箱重复、能力记录重复）|
| 42201 | 422 | 状态不允许该操作（含"本位币已锁定，禁止修改"）|
| 50001 | 500 | 服务器内部错误 |
| 50002 | 500 | 数据库事务失败 |
| 50301 | 503 | 邮件服务暂时不可用 |

---

## 第十五章 状态机设计

### 15.1 课次状态机

```
SCHEDULED ──课前提醒任务触发──▶ REMINDED
SCHEDULED/REMINDED ──Operator取消──▶ CANCELLED
SCHEDULED/REMINDED ──自动关闭任务(超时未确认)──▶ COMPLETED（仅状态变更，无账务）
SCHEDULED/REMINDED ──Operator课后确认──▶ COMPLETED（含账务事务）
COMPLETED, CANCELLED 为终态
```

### 15.2 课程报名状态机

```
ACTIVE ──暂停──▶ PAUSED ──恢复──▶ ACTIVE
ACTIVE ──达成目标──▶ COMPLETED
ACTIVE/PAUSED ──终止──▶ CANCELLED
COMPLETED, CANCELLED 为终态
```

### 15.3 充值记录状态机

```
CONFIRMED ──作废(需填写原因)──▶ VOIDED
VOIDED为终态，不可恢复为CONFIRMED
```

### 15.4 师生安排状态机

```
ACTIVE ──换老师──▶ ENDED
ACTIVE ──临时暂停──▶ PAUSED ──恢复──▶ ACTIVE
ENDED为终态
```

### 15.5 通知状态机

```
PENDING ──发送成功──▶ SENT
PENDING ──发送失败──▶ FAILED ──重试成功──▶ SENT
FAILED ──重试达最大次数仍失败──▶ FAILED（终态，需人工处理）
PENDING/FAILED ──课次被取消──▶ CANCELLED
```

### 15.6 本位币锁定状态机（新增）

```
UNLOCKED（初始状态，无任何财务记录）──首条财务记录写入──▶ LOCKED
LOCKED为终态，V1不提供在线迁移解锁能力
```


---

# 第六部分：技术架构

## 第十六章 后端技术架构与开发规范

### 16.1 技术栈

| 层次 | 选型 | 说明 |
|---|---|---|
| 语言 | Go 1.22+ | 单二进制，~30MB内存，跨平台编译 |
| Web框架 | Gin | 主流，中间件生态完善 |
| ORM | GORM | 支持SQLite/MySQL/PostgreSQL，迁移无缝 |
| 数据库 | SQLite（`modernc.org/sqlite`）| 纯Go驱动，**禁止**使用`mattn/go-sqlite3`（依赖CGO，破坏跨平台一键编译）|
| 数据库迁移 | goose | SQL文件版本管理，按dialect分目录 |
| 认证 | JWT + Refresh Token | Access 60分钟，Refresh 14天 |
| 金额运算 | shopspring/decimal | 精确decimal，避免浮点 |
| 定时任务 | robfig/cron v3 | 标准Cron表达式 |
| 邮件主通道 | Resend Go SDK | 零配置，API Key驱动 |
| 邮件备用 | net/smtp | 标准库Fallback |
| Excel | excelize | Go生态最成熟的Excel库 |
| PDF | gofpdf | 报表/账单PDF生成 |
| 静态嵌入 | go:embed | 前端dist打包进二进制 |
| 日志 | zerolog | 结构化日志 |
| 配置 | viper | YAML+环境变量覆盖 |
| 备份（可选）| Litestream | SQLite流式备份到S3/R2 |
| 桌面封装（可选）| Wails v2 | 详见第十八章 |

### 16.2 架构参照标准（供AI编程工具遵循）

本项目不直接采用任何现成Go后台管理框架（gin-vue-admin、go-admin、soybean-admin-go等），因为它们普遍默认绑定MySQL/PostgreSQL+Redis，与本项目"SQLite+单文件All-in-One"目标冲突。代码组织参照两个社区公认标准：

```
golang-standards/project-layout —— cmd/internal/pkg三段式目录结构
Ardan Labs Service架构 —— handler/service/repository严格分层，接口解耦
```

### 16.3 工程目录结构

```
zedu/
├── backend/
│   ├── cmd/zedu-server/main.go
│   ├── internal/
│   │   ├── app/ auth/ user/ student/ parent/ teacher/ course/
│   │   ├── enrollment/ lesson/ finance/ notification/ report/
│   │   ├── system/ job/ audit/ backup/ init/(初始化向导)
│   │   └── payment/(支付方式与凭证)
│   ├── pkg/
│   │   ├── response/ pagination/ validator/ money/ datetime/ crypto/ errors/
│   ├── migrations/
│   │   ├── sqlite/ mysql/ postgres/
│   │   └── seed/（初始化模板种子数据：japanese.sql / k12.sql）
│   ├── web/admin-dist/（前端构建产物，go:embed目标）
│   ├── config/config.example.yaml
│   ├── go.mod / go.sum
├── frontend/admin/（Soybean Admin工程）
├── deploy/（systemd/WinSW/nginx/litestream配置）
├── scripts/（多平台构建脚本）
└── docs/（本PRD及配套文档）
```

### 16.4 模块内部结构规范（强制）

```
internal/<module>/
├── model.go       仅GORM struct定义，不含业务逻辑方法
├── dto.go         API入参/出参，与model解耦
├── handler.go     仅解析请求→调用service→包装响应，禁止写SQL或业务判断
├── service.go     所有业务规则、状态机校验、事务编排
├── repository.go  所有GORM查询封装，handler/service不直接操作*gorm.DB
├── routes.go      仅路由注册
└── errors.go      模块级业务错误码
```

**AI生成代码自查清单**：

```
□ handler.go中是否出现SQL语句或db.Where()调用？→ 违规，移到repository.go
□ service.go是否直接调用gin.Context？→ 违规，service层不应感知HTTP细节
□ model.go是否包含业务方法（如计算金额）？→ 违规，移到service.go
□ 金额运算是否使用float64？→ 违规，必须用shopspring/decimal
□ 涉及多表写入是否包裹在db.Transaction()中？→ 账务相关操作必须有事务
□ 是否使用了modernc.org/sqlite而非mattn/go-sqlite3？→ 后者禁止使用
□ 支付方式/出勤分类是否用了CHECK约束？→ 违规，必须用字典表+外键
```

### 16.5 事务边界规范（最高优先级）

以下操作必须在单一数据库事务内完成：

```
课后确认：attendance写入 + student_account_ledger + teacher_account_ledger +
          lesson_finance + enrollment余额更新 + teacher待结算更新 + lesson状态更新
充值确认：student_payment写入 + student_account_ledger + enrollment余额更新
充值作废：student_payment状态更新 + student_account_ledger冲正记录 + enrollment余额更新
结款提交：teacher_payout写入 + teacher_account_ledger + teacher.unpaid_amount更新
本位币锁定：财务记录写入 + system_config.base_currency_locked置1（同一事务）
```

### 16.6 幂等性规范

```
课前提醒任务：以lesson.remind_sent_at IS NULL作为幂等门控
充值/结款接口：接口层支持业务单号(payment_no/payout_no)作为幂等键
```

---

## 第十七章 前端技术架构

### 17.1 技术选型

```
框架：Vue 3 + Vite + TypeScript
UI组件库：Naive UI（Soybean Admin内置）
Admin模板：Soybean Admin（Vue3+Naive UI版本）
图表：ECharts 5
状态管理：Pinia
CSS：UnoCSS
HTTP客户端：axios
```

**选择Soybean Admin的理由**：相比RuoYi-Vue3，Soybean Admin视觉更现代（卡片化、动态主题）、移动端适配更好、ECharts内置支持完整、TypeScript全量支持、不绑定特定后端框架。

### 17.2 关键组件复用清单

```
<StudentSelector />       学生选择器
<TeacherSelector />       老师选择器
<EnrollmentSelector />    课程报名项目选择器（联动学生）
<CourseTrackCascader />   课程领域/方向/等级级联选择器
<MoneyInput />            金额输入框（自动格式化，币种符号来自base_currency配置）
<PaymentMethodSelect />   支付方式下拉（来自字典API）
<AttendanceOutcomeSelect /> 出勤分类下拉（选择后自动查询建议值）
<AttachmentUploader />    付款凭证上传（限3张，格式校验）
<DateTimePicker />        日期时间选择器（统一时区处理）
<StatusTag />             状态标签
<ConfirmDialog />         二次确认弹窗
```

### 17.3 请求层与状态管理

```typescript
// api/request.ts - 统一拦截401自动刷新token、统一错误提示
// store/auth.ts - 登录态、用户信息
// store/app.ts - 全局配置（base_currency、系统时区等，启动时拉取一次）
// store/dictionary.ts - 课程体系/支付方式/出勤分类字典缓存，减少重复请求
```

---

## 第十八章 桌面端封装方案

### 18.1 技术选型：Wails v2

| 候选 | 是否采纳 | 理由 |
|---|:---:|---|
| **Wails v2** | ✅ | 纯Go技术栈；v2稳定生产可用；与现有go:embed架构高度契合 |
| Wails v3 | ❌ | 仍为Alpha，不适合生产依赖 |
| Tauri v2 | ❌ | 需引入Rust作为第二语言；需sidecar管理外部进程生命周期 |
| Electron | ❌ | 安装包150MB+，与轻量定位矛盾 |

### 18.2 架构设计原则

**桌面封装不改动任何业务代码，只增加一个可选构建目标。** 因为现有架构本来就用`go:embed`把前端打进二进制自己起HTTP Server，这与Wails的设计理念几乎是同一件事：

```go
func main() {
    if isDesktopMode() {
        go startGinServer(":18080")  // 同进程内启动现有Gin server
        runWailsApp("http://localhost:18080")  // Wails窗口指向本地server
    } else {
        startGinServer(":8080")  // Web部署模式，现有逻辑不变
    }
}
```

业务代码（handler/service/repository）完全不感知运行模式，桌面封装是纯粹的外壳工作。

### 18.3 已知限制与应对

```
WebView2(Windows)不支持Cookie → 本项目认证方案本来就是JWT Bearer Header，不受影响
Wails v2自动更新机制不如Tauri成熟 → V1阶段桌面更新频率低，可接受手动覆盖安装
```

### 18.4 跨引擎兼容性QA要求

Mac/Linux上Wails使用WebKit内核（非Windows的WebView2/Chromium内核），测试阶段（第二十四章）须在Safari内核环境额外验证Naive UI和UnoCSS的渲染一致性。这是QA检查项，不是新增的设计任务——三端UI本来就是同一套Soybean Admin代码。

### 18.5 与移动端的关系澄清

```
桌面封装（Wails）解决：Windows/Mac用户获得原生应用体验
移动端支持（已在第十一章）：手机浏览器访问部署URL，已有响应式页面覆盖
移动端"类App"体验（V1.5）：PWA（manifest.json+service worker），与Wails/Tauri无关
原生移动封装：均为Alpha技术，V1/V1.5不采用
```

---

## 第十九章 部署与发布方案

### 19.1 All-in-One设计

核心理念：一个压缩包解决所有问题。`go:embed`打包前端 + `modernc.org/sqlite`纯Go驱动无CGO依赖 + 内嵌HTTP服务器与定时任务 + 首次运行自动建库执行migration。

**发布包结构**：

```
zedu/
├── zedu-server(.exe)   主程序
├── config.yaml         配置文件（首次运行自动生成模板）
└── data/
    ├── zedu.db          SQLite数据库（首次运行自动创建）
    └── uploads/         付款凭证等上传文件
```

### 19.2 四种部署模式

**模式A：Windows双击运行** —— 双击exe，浏览器访问localhost:8080，适合个人本地使用

**模式B：Windows Service** —— 用WinSW（单exe+xml配置，无需.NET）注册为系统服务，开机自启

**模式C：Linux/AWS云端**（推荐生产）—— systemd管理，配合Nginx反向代理+Let's Encrypt SSL

**模式D：Wails桌面应用**（可选）—— 详见第十八章

### 19.3 跨平台编译

```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o dist/zedu_windows_amd64.exe .
GOOS=linux   GOARCH=amd64 go build -ldflags="-w -s" -o dist/zedu_linux_amd64 .
GOOS=darwin  GOARCH=arm64 go build -ldflags="-w -s" -o dist/zedu_darwin_arm64 .
```

因使用`modernc.org/sqlite`（纯Go无CGO），一台机器可一次性编出三平台二进制。

### 19.4 配置文件示例

```yaml
server:
  host: 0.0.0.0
  port: 8080
  public_url: https://zedu.abitcloud.org

app:
  name: Zedu
  timezone: Asia/Tokyo

database:
  driver: sqlite
  dsn: ./data/zedu.db

auth:
  jwt_secret: change-me-in-production
  access_token_minutes: 60
  refresh_token_days: 14

mail:
  primary: resend
  resend:
    api_key: re_xxxxxxxxxx
    from_email: noreply@abitcloud.org
    from_name: Zedu教务
  smtp:
    host: smtp.gmail.com
    port: 587

upload:
  max_file_size_mb: 5
  max_attachments_per_payment: 3
  allowed_types: [jpg, jpeg, png, webp, pdf]
  storage_path: ./data/uploads

backup:
  auto_enabled: true
  cron: "0 0 2 * * *"
  path: ./backup
  retention_days: 30
  litestream_enabled: false
```

**说明**：`base_currency`不在config.yaml中配置，而是通过初始化向导写入数据库`system_config`表，因为它是产生首条财务记录后需要锁定的业务数据，不适合放在可随意编辑的配置文件里。

---

## 第二十章 数据安全与备份

### 20.1 安全措施

| 项目 | 实现 |
|---|---|
| 密码存储 | bcrypt哈希，cost=12 |
| 登录防暴力 | 连续失败5次锁定15分钟 |
| Token管理 | JWT(60分钟)+Refresh Token(14天)|
| 操作审计 | 所有写操作写入operation_log |
| 隐私保护 | 手机号脱敏（138****1234）|
| 文件访问 | 付款凭证访问必须经过登录鉴权，禁止匿名直链 |
| HTTPS | 生产环境必须启用 |

### 20.2 数据合规（日本APPI）

学生/老师联系方式不在邮件正文完整暴露；系统V1不对公众开放；付款凭证等敏感文件不可匿名访问。

### 20.3 备份策略

| 方式 | 实现 | 适用场景 |
|---|---|---|
| 手动备份 | 系统设置页一键下载.zip（db+config+uploads）| 随时按需 |
| 自动本地备份 | 每天02:00，`VACUUM INTO`方式，保留30天 | 本地运行 |
| Litestream流式备份 | 实时同步S3/R2，每5秒一次 | 云端强烈推荐 |

**关键要求**：备份包**必须包含uploads目录**（付款凭证），不能只备份数据库文件——这是本版相对v2.0的重要修正，因为v2.0设计时还没有付款凭证这个概念。恢复操作前系统自动先对当前状态做一次快照备份。

---

## 第二十一章 通知与提醒系统

### 21.1 通知渠道

V1仅支持邮件：主通道Resend API，备用通道SMTP（Fallback）。

### 21.2 六类定时任务

```
任务1 课前提醒（每10分钟轮询）：remind_sent_at IS NULL且课前20~40分钟窗口内 → 发邮件 → 更新remind_sent_at
任务2 教务晨报（每天08:00）：今日课程/待确认课次/余额不足学生/待结款老师/失败通知
任务3 余额预警（每天20:00）：lesson_balance≤阈值 或 balance_amount<下节课预计费用
任务4 老板周报（每周一08:00）：收入/课酬/毛利/完成课次/新增学生/课程方向分布
任务5 通知重试（每30分钟）：FAILED且retry_count<3 → 重新发送
任务6 自动关闭过期课次（每天03:00）：SCHEDULED/REMINDED且结束4小时后 → COMPLETED（仅状态变更）
```

### 21.3 邮件模板（必须实现的5个）

| 模板代码 | 发送对象 | 场景 |
|---|---|---|
| LESSON_REMINDER_STUDENT | 学生 | 课前提醒（含上课链接）|
| LESSON_REMINDER_TEACHER | 老师 | 课前提醒（含学生信息）|
| BALANCE_ALERT_OPERATOR | Operator | 余额不足预警汇总 |
| MORNING_REPORT_OPERATOR | Operator | 每日晨报 |
| OWNER_WEEKLY_REPORT | Owner | 老板周报 |

模板存储于数据库，支持后台编辑，变量用`{{.VariableName}}`占位符，邮件中手机号脱敏。

---

## 第二十二章 非功能性需求

### 22.1 性能要求

| 指标 | 目标值 |
|---|---|
| 工作台首屏加载 | <2秒 |
| API响应时间(P95) | <500ms |
| 列表分页查询(1000条内) | <300ms |
| 并发用户支持 | ≥10 |

### 22.2 可用性与国际化

```
单实例部署，无高可用集群要求
V1界面语言：中文为主；通知邮件语言可配置中文/日文/双语
时区：系统级默认Asia/Tokyo，可配置
金额显示：千分位分隔符，币种符号来自base_currency动态渲染
```

### 22.3 浏览器兼容性

支持Chrome/Edge/Safari最新两个大版本，移动端Safari(iOS)/Chrome(Android)；不支持IE。


---

# 第七部分：项目管理

## 第二十三章 开发计划与任务拆分

### 23.1 三段式演进节奏

本项目采用"7天MVP → 4周V1 → 按需V1.5"的节奏，而非一次性追求完整V1范围。这样能尽快让运营者用上核心闭环，用真实数据验证设计，再决定后续投入优先级。

### 23.2 MVP（目标7天，最小经营闭环）

> 2026-07-11决策：7天是进度目标，不是牺牲财务、安全、权限或恢复质量的硬承诺。MVP完成以第二十四章门禁为准。

```
□ 登录（Owner/Operator）
□ 初始化向导（应用日语模板）
□ 学生管理（基础CRUD）
□ 老师管理（基础CRUD）
□ 课程体系配置（四层字典，已预置日语模板数据）
□ 课程报名（enrollment基础CRUD）
□ 师生安排（assignment基础CRUD）
□ 排课（lesson基础CRUD，无冲突检测）
□ 课后确认（出勤分类+建议值+实际值，含账务事务）
□ 充值（含多币种折算、付款凭证上传与鉴权访问）
□ Resend邮件通知（课次相关通知、课前提醒、通知日志、幂等、失败重试与人工重放）
□ 学生流水查看
□ 老师应付查看（暂不做正式结款流程）
□ 本地手动备份（数据库+配置+付款凭证，含manifest和恢复演练）
□ 极简工作台（今日课程+待续费+老师应付+通知失败，无图表）
```

MVP验收标准：能完整走通"新建学生→报名→充值/凭证→排课→通知→确认→查流水/应付→备份恢复"主链路；账务核对零差异；通知失败不回滚排课；凭证不可越权访问；正式老师结款入口为零。

### 23.3 V1（MVP之后3~4周，正式可运营版本）

在MVP基础上补充：

```
Sprint 1：老师结款完整流程 + 换老师流程 + 时间冲突检测 + 余额不足软提示
Sprint 2：SMTP降级 + 晨报/周报等完整通知自动化
Sprint 3：支付方式字典管理 + 出勤分类字典管理页面
Sprint 4：工作台ECharts图表 + 财务报表 + Excel导入导出 + 操作日志页面
Sprint 5：移动优先页面(/mobile/*) + Litestream云备份 + All-in-One多平台打包
Sprint 6：测试(第二十四章) + Bug修复 + 真实数据迁移 + 正式上线
```

### 23.4 V1.5（体验增强，按需触发）

```
独立Mobile Web(/m) / PWA
Wails桌面壳打包
价格模板库（触发条件：见第六章6.3节）
老师课酬约定表（触发条件：见第六章6.3节）
课时包有效期
PDF正式收据
更丰富报表
```

### 23.5 V2（多端扩展，暂不规划具体时间）

```
老师端/学生端登录
家长通知
小班课完整管理
支付API对接
更复杂请假规则引擎
AI辅助总结和匹配建议
```

### 23.6 AI辅助开发策略

```
1. 给定第十二章DDL → AI生成GORM Model + Gin Handler + Service骨架
2. 在骨架上填充业务逻辑（差价计算/课时扣减/幂等控制），严格遵循第十六章16.4节自查清单
3. 参考Soybean Admin示例 → AI生成Vue页面组件，对照第十章UI规格逐项核对
4. 邮件模板 → AI生成HTML，Resend控制台预览效果
```

---

## 第二十四章 测试计划与验收标准

### 24.1 分阶段业务闭环验收（对应用户旅程）

**MVP闭环**必须能完整走通：

```
1. 新建学生 → 2. 新建老师 → 3. 配置课程 → 4. 新建学生课程报名 → 5. 安排老师
→ 6. 录入充值(上传付款截图) → 7. 创建课次 → 8. 发送提醒 → 9. 课后确认(手工覆盖实际扣费和课酬)
→ 10. 查看学生流水/老师应付/单课财务 → 11. 查看工作台汇总 → 12. 备份并恢复演练
```

MVP负向验收：正式老师结款的菜单、路由、API、直链与隐藏开关均不存在或不可执行。

**V1闭环**在MVP基础上增加：正式老师结款、完整通知自动化、报表与数据导入导出、移动/打包及真实数据迁移。结款财务语义和权限须在V1启动前单独批准。

### 24.2 配置化验收（本版新增重点）

```
1. 可以初始化日语模板
2. 可以初始化K12模板
3. 可以空白初始化后手工建课程
4. 可以新增支付方式（如支付宝）
5. 可以新增出勤分类
6. 可以修改出勤分类的建议值
7. 可以使用非JPY本位币初始化（如CNY）
8. 产生财务记录后，本位币不可通过系统设置页修改（验证42201错误码）
9. 单课实际收费可以不同于enrollment默认收费
10. 单课实际课酬可以不同于assignment/teacher默认课酬
```

### 24.3 账务事务测试（最高优先级）

```
TC-01 课后确认后，学生余额正确扣减，老师应付正确增加
TC-02 课后确认事务中途模拟异常，验证全部回滚（无部分写入）
TC-03 学生请假选择"提前请假"分类，验证不扣课时、不产生老师课酬
TC-04 老师缺席选择"老师请假"分类，验证不扣学生课时、不产生应付款
TC-05 充值作废后，验证余额正确冲正，ledger生成VOID记录
TC-06 多币种充值，验证汇率折算金额精度正确
TC-07 attendance表的suggested_*字段与实际字段分别正确落库，两者可以不同
```

### 24.4 配置化边界测试

```
TC-08 尝试在有财务记录后修改base_currency，验证返回42201且未实际修改
TC-09 新增支付方式后，充值表单下拉立即出现新选项
TC-10 修改出勤分类建议值后，课后确认页自动带出的建议值相应变化
TC-11 禁用某个课程等级后，历史引用该等级的课次仍能正常显示
```

### 24.5 幂等性与状态机测试

```
TC-12 课前提醒任务重复执行，验证同一课次不会收到两次提醒
TC-13 尝试编辑已COMPLETED的课次，验证返回42201
TC-14 换老师后，验证旧assignment正确ENDED，新assignment正确ACTIVE
```

### 24.6 数据安全测试

```
TC-15 未登录直接访问付款凭证文件URL，验证被拒绝
TC-16 手动触发备份，验证备份包含uploads目录
TC-17 恢复操作前，验证自动生成当前状态快照
```

### 24.7 跨平台部署测试

```
TC-18 Windows双击运行，浏览器可正常访问
TC-19 Linux systemd服务安装，开机自启验证
TC-20 三平台编译产物均可独立运行（modernc.org/sqlite无CGO依赖验证）
TC-21 Wails桌面壳在Windows(WebView2)和Mac(WebKit)均正常渲染
```

---

## 第二十五章 风险登记表

| 编号 | 风险描述 | 影响 | 可能性 | 应对措施 |
|---|---|:---:|:---:|---|
| RISK-01 | SQLite并发写入锁冲突 | 中 | 低 | WAL模式+busy_timeout+应用层写串行化 |
| RISK-02 | Resend邮件送达率不达预期 | 中 | 中 | 验证发件域名SPF/DKIM，监控失败率 |
| RISK-03 | 账务事务设计遗漏边界情况 | 高 | 中 | 按24.3节测试用例充分测试+人工调整流水兜底 |
| RISK-04 | 时区处理错误导致提醒不准 | 中 | 中 | 统一UTC存储+集成测试专项验证 |
| RISK-05 | 出勤分类建议值配置不合理导致日常操作体验差 | 低 | 中 | 上线后收集实际使用反馈，允许运营者随时调整建议值 |
| RISK-06 | 运营者遗忘课后确认，账务滞后 | 中 | 高 | 工作台"待确认"醒目提示+晨报持续提醒 |
| RISK-07 | 备份未包含uploads目录导致凭证丢失 | 高 | 低 | 明确写入第二十章2.3节要求，测试用例TC-16专项验证 |
| RISK-08 | 本位币锁定机制实现有误，导致本应锁定时仍可修改 | 高 | 低 | 测试用例TC-08专项验证 |
| RISK-09 | 单机部署无高可用 | 中 | 低 | 云端部署+Litestream实时备份 |
| RISK-10 | 过早引入price_plan等复杂表导致MVP延期 | 中 | 中 | 严格按第六章6.3节的触发条件判断，不提前实现 |

---

## 第二十六章 版本演进规划

```
MVP  （7天）      最小经营闭环，验证核心链路可用
V1   （3~4周）    正式可运营版本，完整通知/结款/报表/备份
V1.5 （按需触发）  独立移动端/PWA/Wails桌面版/价格模板/课酬约定表
V2   （远期）      多端登录/家长通知/小班课/支付API/AI辅助
```

### 26.1 扩展时机参考

| 触发条件 | 建议扩展 | 版本 |
|---|---|:---:|
| 学生>200，续费跟进成本高 | 学生自助端 | V2 |
| 老师>30，排课沟通成本高 | 老师自助端 | V2 |
| 每天大量新签约，手工定价成瓶颈 | 价格模板库 | V1.5 |
| 老师带生多、课酬常按类型批量调 | 课酬约定表 | V1.5 |
| 手机操作频率高 | 独立移动端 | V1.5 |
| 月流水达规模 | 支付API接口 | V2 |
| 计划复制部署给K12/体育机构 | 直接使用现有K12模板+空白模板机制 | 已支持，无需额外开发 |

---

# 附录

## 附录A：术语表

| 术语 | 英文/代码 | 说明 |
|---|---|---|
| 课程领域 | course_domain | 最顶层课程分类 |
| 课程方向 | course_track | 领域下的学习路线 |
| 课程等级 | course_level | 方向下的阶段 |
| 课程报名 | enrollment | 学生对某课程方向的报名记录 |
| 师生安排 | assignment | enrollment下具体由哪位老师授课 |
| 课次 | lesson | 一次具体的上课安排 |
| 出勤确认 | attendance | 课后确认产生的记录，触发账务 |
| 出勤结果分类 | outcome_type | 配置化的出勤情形分类（正常/请假/缺席等）|
| 本位币 | base_currency | 系统统一记账和报表的币种，初始化后锁定 |
| 建议值 | suggested_* | 仅供前端表单默认填充，不强制生效 |
| 实际值 | 落库字段 | 最终生效、影响账务的真实数据 |
| 幂等 | Idempotent | 同一操作重复执行结果一致 |
| RBAC | Role-Based Access Control | 基于角色的权限控制 |
| WAL | Write-Ahead Logging | SQLite日志模式，提升并发性能 |

## 附录B：技术依赖速查

**后端**：gin-gonic/gin · gorm.io/gorm · modernc.org/sqlite · pressly/goose · robfig/cron/v3 · resend/resend-go/v2 · shopspring/decimal · golang-jwt/jwt/v5 · spf13/viper · rs/zerolog · xuri/excelize/v2 · jung-kurt/gofpdf · wailsapp/wails/v2

**前端**：vue@3 · vite · typescript · naive-ui · pinia · axios · echarts · unocss

## 附录C：当前仍需运营者确认的问题

以下问题不阻塞架构实现，但影响本次实际上线的具体配置，建议上线前至少确认前5项：

```
1. 首个部署实例使用日语模板（已按此假设写入本文档，如有变化需告知）
2. 系统本位币确认为JPY（已按此假设写入本文档且将被锁定，请确认无误）
3. 当前常用支付方式：微信/PayPay/银行转账/现金是否已覆盖？是否需要支付宝/楽天Pay？
4. 付款凭证保留几张合适？默认3张是否够用？
5. 默认课时单位：30/50/60分钟中最常用的是哪个？
6. 现有Excel数据规模（学生/老师各多少）及字段结构，用于设计导入模板
7. 出勤分类的默认建议值（第八章8.3节表格）是否符合实际运营习惯？
8. 是否需要导入历史上课记录/充值记录，还是只导入当前余额快照？
9. 系统主要通过电脑还是手机访问，决定移动端页面开发优先级
10. 首个Operator账号是否只有运营者本人，还是需要预留多个账号位？
```

## 附录D：变更记录

| 版本 | 日期 | 变更摘要 |
|---|---|---|
| v0.1-v0.2 | 2026-06-08 | 初始业务讨论稿，明确All-in-One部署理念 |
| v1.0 | 2026-06-09 | 首份正式PRD，完整数据模型与API规范纲要 |
| v2.0 | 2026-06-09 | 完整实装版，补充全量DDL/API/页面级UI规格/测试用例 |
| v2.1 | 2026-07-04 | 配置化修订说明，独立评估Codex意见，采纳5项低成本高价值修改 |
| v2.2 | 2026-07-04 | Codex业务配置化收敛版，确立五大设计原则和三层定价模型 |
| v3.0 | 2026-07-04 | 正式定案版，融合v2.0技术深度与v2.2业务原则，重写完整DDL/API/UI规格以贯彻配置化原则，作为唯一权威文档 |
| v3.1 | 2026-07-11 | 调整MVP边界：Resend通知与付款凭证前移；正式老师结款留在V1；增加分阶段闭环、附件备份恢复及无结款入口门禁 |
| v3.1-r1 | 2026-07-12 | 产品决策：学生邮箱填写时必须唯一；删除重复邮箱的“仍然新建”分支，统一以40901拒绝冲突写入 |

---

*Zedu PRD v3.1-r1（MVP范围修订版）· 更新日期 2026-07-12*

*本文档是Zedu项目的唯一事实文档。后续所有开发、设计、测试工作均以此为准。若有变更需求，请直接在对应章节修订并更新附录D变更记录，不再另外维护并行的讨论稿。*
