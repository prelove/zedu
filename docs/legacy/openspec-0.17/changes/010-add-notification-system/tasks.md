## 1. NotificationSender接口与双通道实现

- [ ] 1.1 编写失败测试：ResendSender.Send()调用真实/mock的Resend API成功时
      返回nil error
      文件：backend/internal/notification/sender_test.go
- [ ] 1.2 实现ResendSender使测试通过
      文件：backend/internal/notification/resend_sender.go
- [ ] 1.3 编写失败测试：SmtpSender.Send()通过mock SMTP server成功发送
- [ ] 1.4 实现SmtpSender使测试通过
      文件：backend/internal/notification/smtp_sender.go
- [ ] 1.5 编写失败测试：service层在ResendSender失败时自动降级到SmtpSender，
      降级成功则notification_log.status=SENT
      文件：backend/internal/notification/service_test.go
- [ ] 1.6 实现降级逻辑使测试通过
      文件：backend/internal/notification/service.go
- [ ] 1.7 编写失败测试：两个通道都失败时status=FAILED且error_msg被记录
- [ ] 1.8 实现代码使测试通过
- [ ] 1.9 提交：git commit -m "feat(notification): dual-channel sender with fallback"

## 2. 通知模板种子数据与渲染引擎

- [ ] 2.1 编写失败测试：5个notification_template记录应存在且
      subject_tpl/body_tpl含正确占位符
      文件：backend/internal/notification/template_test.go
- [ ] 2.2 编写种子数据迁移文件（参照PRD第二十一章21.3节模板原文）
      文件：backend/migrations/seed/notification_templates.sql
- [ ] 2.3 编写失败测试：给定模板code和数据结构，渲染函数应正确替换
      {{.VariableName}}占位符
- [ ] 2.4 实现渲染函数使测试通过（Go text/template）
      文件：backend/internal/notification/render.go
- [ ] 2.5 提交：git commit -m "feat(notification): template seed data and rendering"

## 3. 任务1 课前提醒（幂等，核心）

- [ ] 3.1 编写失败测试：scheduled_start_at在当前时间+20~40分钟窗口内、
      remind_sent_at为NULL的课次，应被选中并触发发送
      文件：backend/internal/job/lesson_reminder_test.go
- [ ] 3.2 实现最小代码使测试通过
      文件：backend/internal/job/lesson_reminder.go
- [ ] 3.3 编写失败测试：发送后remind_sent_at被更新，同一课次
      不会在下次调用时被重复选中（模拟连续两次调用任务函数）
- [ ] 3.4 实现代码使测试通过（这是幂等性的核心验证，不可跳过）
- [ ] 3.5 编写失败测试：学生email为空时，仅向老师发送，不报错
- [ ] 3.6 实现代码使测试通过
- [ ] 3.7 提交：git commit -m "feat(job): idempotent lesson reminder scan"

## 4. 任务2/3/4 晨报/余额预警/周报

- [ ] 4.1 编写失败测试：晨报聚合结果应包含五个部分（今日课程/待确认/
      余额不足/待结款/失败通知）
      文件：backend/internal/job/morning_report_test.go
- [ ] 4.2 实现最小代码使测试通过
      文件：backend/internal/job/morning_report.go
- [ ] 4.3 编写失败测试：lesson_balance<=阈值 或 balance_amount<单次收费
      两个条件任一满足即应出现在余额预警列表
      文件：backend/internal/job/balance_alert_test.go
- [ ] 4.4 实现代码使测试通过
      文件：backend/internal/job/balance_alert.go
- [ ] 4.5 编写失败测试：周报仅发送给ACTIVE的Owner账号，不发给Operator
      文件：backend/internal/job/weekly_report_test.go
- [ ] 4.6 实现代码使测试通过
      文件：backend/internal/job/weekly_report.go
- [ ] 4.7 提交：git commit -m "feat(job): morning report, balance alert, weekly report"

## 5. 任务5 通知重试

- [ ] 5.1 编写失败测试：FAILED且retry_count<3的记录应被重试任务选中
      文件：backend/internal/job/retry_test.go
- [ ] 5.2 实现最小代码使测试通过
      文件：backend/internal/job/retry.go
- [ ] 5.3 编写失败测试：retry_count达到3后不再被任务选中
- [ ] 5.4 验证5.3通过
- [ ] 5.5 编写失败测试：POST /notifications/logs/{id}/resend 应将
      retry_count重置为0并立即触发发送
      文件：backend/internal/notification/resend_handler_test.go
- [ ] 5.6 实现代码使测试通过
- [ ] 5.7 提交：git commit -m "feat(notification): retry job and manual resend"

## 6. 任务6 自动关闭过期课次

- [ ] 6.1 编写失败测试：scheduled_end_at超过4小时且status为SCHEDULED/
      REMINDED的课次应被更新为COMPLETED
      文件：backend/internal/job/auto_close_test.go
- [ ] 6.2 实现最小代码使测试通过
      文件：backend/internal/job/auto_close.go
- [ ] 6.3 编写失败测试：确认该操作不产生attendance或任何ledger记录
- [ ] 6.4 验证6.3通过
- [ ] 6.5 编写失败测试：4小时缓冲期内的课次不会被误关
- [ ] 6.6 验证6.5通过
- [ ] 6.7 提交：git commit -m "feat(job): auto-close overdue lessons without ledger writes"

## 7. 定时任务注册与集成测试

- [ ] 7.1 用robfig/cron注册全部六个任务，Cron表达式来自system_config
- [ ] 7.2 集成测试：临时把课前提醒间隔改为10秒，验证同一课次
      不会收到两次提醒邮件（真实定时器行为验证，非纯单元测试）
- [ ] 7.3 集成测试：模拟一次发送失败，验证30分钟(测试时改短)后自动重试
- [ ] 7.4 提交：git commit -m "feat(job): register all cron jobs"

## 8. 前端：通知管理页面

- [ ] 8.1 提醒规则配置页（课前提醒分钟数/余额预警阈值/晨报时间/
      周报发送日时间/通知语言）
- [ ] 8.2 通知日志页（筛选+表格+失败行"重发"按钮）
- [ ] 8.3 提交：git commit -m "feat(frontend): notification rules and logs pages"

## 9. 模板渲染健壮性与语言支持

- [ ] 9.1 编写失败测试：渲染模板时数据结构缺少某个引用变量，
      不应panic或返回error，该变量应渲染为空字符串
      文件：backend/internal/notification/render_robustness_test.go
- [ ] 9.2 实现代码使测试通过（使用text/template的MissingKey选项
      配置为zero值而非报错）
- [ ] 9.3 编写失败测试：language=BILINGUAL的模板渲染结果应同时
      包含中文和日文内容
      文件：backend/internal/notification/bilingual_test.go
- [ ] 9.4 实现双语模板渲染逻辑使测试通过
- [ ] 9.5 编写失败测试：系统语言配置为ZH但只有JA版本模板时，
      应有明确降级策略而非直接报错
- [ ] 9.6 实现降级查找逻辑使测试通过
- [ ] 9.7 提交：git commit -m "feat(notification): robust template rendering with language fallback"

## 10. 规格场景覆盖检查表

对照本change下specs/notification/spec.md和specs/lesson-scheduling
（MODIFIED）spec.md的全部Scenario，逐条标注验证task：

- [ ] 10.1 「Resend发送成功」→ 1.1-1.2
- [ ] 10.2 「Resend失败降级到SMTP」→ 1.5-1.6
- [ ] 10.3 「两个通道都失败」→ 1.7-1.8
- [ ] 10.4 「首次命中窗口发送提醒」→ 3.1-3.2
- [ ] 10.5 「已发送过不再重复」→ 3.3-3.4
- [ ] 10.6 「收件人缺失邮箱时跳过」→ 3.5-3.6
- [ ] 10.7 「课时余额触发」→ 4.3-4.4
- [ ] 10.8 「金额余额触发」→ 4.3-4.4
- [ ] 10.9 「自动重试」→ 5.1-5.2
- [ ] 10.10 「达到重试上限」→ 5.3-5.4
- [ ] 10.11 「手动重发重置计数」→ 5.5-5.6
- [ ] 10.12 「晨报内容完整」→ 4.1-4.2
- [ ] 10.13 「周报仅发给Owner」→ 4.5-4.6
- [ ] 10.14 「双语模板渲染」→ 9.3-9.4
- [ ] 10.15 「模板变量缺失时优雅降级」→ 9.1-9.2
- [ ] 10.16 「语言配置影响模板选择」→ 9.5-9.6
- [ ] 10.17 「超时自动关闭」→ 6.1-6.2
- [ ] 10.18 「缓冲期内不误关」→ 6.5-6.6

全部勾选后才可执行`/opsx:archive add-notification-system`。
