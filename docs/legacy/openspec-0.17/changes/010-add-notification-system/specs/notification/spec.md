## ADDED Requirements

### Requirement: 通知发送双通道
系统必须支持Resend为主通道、SMTP为备用降级通道发送邮件。

#### Scenario: Resend发送成功
- **WHEN** ResendSender.Send()调用成功
- **THEN** notification_log.status = SENT，sent_at记录当前时间

#### Scenario: Resend失败降级到SMTP
- **WHEN** ResendSender.Send()返回错误
- **THEN** 系统自动尝试SmtpSender.Send()，若该次成功则status=SENT

#### Scenario: 两个通道都失败
- **WHEN** ResendSender和SmtpSender均发送失败
- **THEN** notification_log.status = FAILED，error_msg记录最后一次失败原因

### Requirement: 课前提醒幂等发送
系统必须保证同一课次的课前提醒只发送一次，即使定时任务被重复触发。

#### Scenario: 首次命中窗口发送提醒
- **WHEN** 课次scheduled_start_at落在当前时间+20~40分钟窗口内，
  且remind_sent_at为NULL
- **THEN** 生成提醒通知并将remind_sent_at更新为当前时间

#### Scenario: 已发送过不再重复
- **WHEN** 同一课次的remind_sent_at已非NULL
- **THEN** 下次任务扫描不会再次选中该课次

#### Scenario: 收件人缺失邮箱时跳过
- **WHEN** 学生或老师的email字段为空
- **THEN** 系统仅向有邮箱的一方发送提醒，不因缺失邮箱而报错或跳过整条记录

### Requirement: 余额预警双条件触发
系统必须能识别学生课时余额或金额余额不足的情况。

#### Scenario: 课时余额触发
- **WHEN** enrollment.lesson_balance <= 配置的阈值
- **THEN** 该enrollment出现在余额预警列表中

#### Scenario: 金额余额触发
- **WHEN** enrollment.balance_amount < enrollment.charge_per_lesson_amount
- **THEN** 该enrollment出现在余额预警列表中，即使lesson_balance尚未达到阈值

### Requirement: 通知失败重试
系统必须对失败的通知自动重试，且有次数上限。

#### Scenario: 自动重试
- **WHEN** notification_log.status=FAILED 且 retry_count<3
- **THEN** 定时任务重新尝试发送，成功则status变SENT，
  失败则retry_count加1

#### Scenario: 达到重试上限
- **WHEN** retry_count已达3且仍失败
- **THEN** 记录保持FAILED状态，不再自动重试，需人工手动重发

#### Scenario: 手动重发重置计数
- **WHEN** 用户对FAILED记录点击"重发"
- **THEN** retry_count重置为0并立即触发一次新的发送尝试

### Requirement: 教务晨报与老板周报聚合
系统必须能定时生成聚合报告并发送给对应角色。

#### Scenario: 晨报内容完整
- **WHEN** 晨报任务触发
- **THEN** 邮件内容包含今日课程、待确认课次、余额不足学生、
  待结款老师、近24小时失败通知五个部分，发送给全部ACTIVE的
  Owner和Operator

#### Scenario: 周报仅发给Owner
- **WHEN** 周报任务触发
- **THEN** 邮件发送给全部ACTIVE的Owner账号，不发送给Operator

### Requirement: 通知语言与模板渲染健壮性
系统必须支持配置通知语言，且模板渲染在数据缺失时不应导致任务崩溃。

#### Scenario: 双语模板渲染
- **WHEN** system_config中通知语言配置为BILINGUAL，且模板language
  字段对应为BILINGUAL
- **THEN** 邮件正文包含中文和目标语言（如日文）两个版本的内容，
  中文在前

#### Scenario: 模板变量缺失时优雅降级
- **WHEN** 渲染某模板时，传入的数据结构缺少模板中引用的某个变量
  （如{{.LessonNote}}但该课次未填写备注）
- **THEN** 该变量渲染为空字符串而非导致渲染失败或任务崩溃，
  其余变量正常渲染

#### Scenario: 语言配置影响模板选择
- **WHEN** 系统配置语言为ZH，但某个场景只存在JA语言的模板
- **THEN** 系统应有明确的降级策略（优先寻找BILINGUAL版本，
  否则退回该模板code下任意可用语言版本），而非直接报错不发送
