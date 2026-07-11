# Design: 通知发送与定时任务

## NotificationSender接口抽象

```go
type NotificationSender interface {
    Send(ctx context.Context, msg *Message) error
}
```

实现：`ResendSender`（主通道，调用Resend Go SDK）、`SmtpSender`（备用，
net/smtp标准库）。上层service层的发送逻辑：先尝试ResendSender，若返回
error则尝试SmtpSender作为fallback，两者都失败才标记notification_log
为FAILED。两次尝试都要各自捕获error_msg，取最后一次失败的错误信息记录。

## 六类定时任务的触发条件与幂等设计

### 任务1：课前提醒（每10分钟轮询）
```sql
SELECT * FROM lesson
WHERE remind_sent_at IS NULL
  AND status = 'SCHEDULED'
  AND scheduled_start_at BETWEEN NOW()+20min AND NOW()+40min
```
命中的每条lesson：
- 若student.email非空，生成一条notification_log(recipient_type=STUDENT)并发送
- 若teacher.email非空，生成一条notification_log(recipient_type=TEACHER)并发送
- 无论以上两条发送成功与否，只要任务尝试过该lesson，立即更新
  `remind_sent_at = NOW()` 和 `status = 'REMINDED'`
  （这一步是幂等的关键：即使发送失败，也不应该让同一lesson在下次
  10分钟轮询时被重复选中，避免用户收到多封相同提醒。发送失败的
  单独走通知重试任务补偿，而不是让主任务重复扫描）

### 任务2：教务晨报（每天08:00，可配置）
聚合五类数据一次性生成：今日课程列表、待确认课次列表（COMPLETED前、
scheduled_end_at已过的）、余额不足学生列表、待结款老师列表、
最近24小时内的FAILED通知列表。发送给全部status=ACTIVE的Owner和
Operator账号。

### 任务3：余额预警（每天20:00，可配置）
触发条件（满足其一即可）：
```
enrollment.lesson_balance <= system_config.balance_alert_lessons
或
enrollment.balance_amount < enrollment.charge_per_lesson_amount
```
课后确认事务提交后也应实时判断一次（见006-add-attendance-confirmation
的design.md），本任务是每日兜底扫描，两者不冲突。

### 任务4：老板周报（每周一08:00，可配置）
统计上周（周一到周日）的收入/课酬/毛利/完成课次数/取消课次数/
新增学生数/各课程方向课时分布，发送给Owner。

### 任务5：通知重试（每30分钟）
```sql
SELECT * FROM notification_log
WHERE status = 'FAILED' AND retry_count < 3
```
重新调用发送逻辑，成功则status=SENT，失败则retry_count+=1。
手动点击"重发"按钮的效果等同于把retry_count重置为0后立即触发一次发送。

### 任务6：自动关闭过期课次（每天03:00）
```sql
UPDATE lesson SET status='COMPLETED'
WHERE status IN ('SCHEDULED','REMINDED')
  AND scheduled_end_at < NOW() - INTERVAL 4 HOUR
```
仅状态变更，不触发任何ledger写入（区别于Operator手动课后确认）。
4小时缓冲期是为了避免误关"确实还在上但Operator还没来得及点确认"的课次。

## 模板渲染
notification_template.subject_tpl/body_tpl使用Go的text/template语法，
`{{.VariableName}}`占位符。渲染时传入的数据结构因模板而异（如
LESSON_REMINDER_STUDENT需要StudentName/TeacherName/LessonDate/
LessonTime/MeetingType/MeetingLink等字段）。
