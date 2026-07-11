## 1. 排课后端（基础版）

- [ ] 1.1 编写失败测试：POST /lessons 创建课次成功，status默认SCHEDULED，
      lesson_no格式正确
      文件：backend/internal/lesson/service_test.go
- [ ] 1.2 实现最小代码使测试通过
      文件：backend/internal/lesson/model.go, dto.go, service.go, handler.go, repository.go
- [ ] 1.3 编写失败测试：同一天创建多个课次，lesson_no应各不相同且递增
      文件：backend/internal/lesson/lesson_no_test.go
- [ ] 1.4 实现lesson_no生成逻辑使测试通过（事务内查询当日最大序号+1）
- [ ] 1.5 编写失败测试：duration_min超出10~480范围应返回40001
- [ ] 1.6 实现校验逻辑使测试通过
- [ ] 1.7 编写失败测试：meeting_type=WECHAT且meeting_link为空或非URL文本
      应能正常创建，不触发格式校验错误
- [ ] 1.8 验证1.7通过
- [ ] 1.9 编写失败测试：基于COMPLETED/CANCELLED状态的enrollment创建课次
      应返回42201
- [ ] 1.10 实现代码使测试通过
- [ ] 1.11 编写失败测试：COMPLETED状态课次尝试编辑上课时间应返回42201，
      仅note可修改
- [ ] 1.12 实现状态校验使测试通过
- [ ] 1.13 编写失败测试：以Asia/Tokyo时区创建19:00的课次，数据库中
      scheduled_start_at应为对应的正确UTC时间
- [ ] 1.14 实现时区转换逻辑使测试通过
- [ ] 1.15 提交：git commit -m "feat(lesson): basic CRUD with lesson-no generation and status guard"

## 2. 前端：排课列表与新建表单

- [ ] 2.1 排课列表页（学生/老师/时间/上课方式/状态）
- [ ] 2.2 新建课次表单（学生→enrollment联动→老师自动带入可改选）
- [ ] 2.3 提交：git commit -m "feat(frontend): lesson list and creation form"

## 3. 规格场景覆盖检查表

对照本change下specs/lesson-scheduling/spec.md的全部Scenario，逐条
标注验证task：

- [ ] 3.1 「新建课次」→ 1.1-1.2
- [ ] 3.2 「时长超出范围被拒绝」→ 1.5-1.6
- [ ] 3.3 「lesson_no唯一且可读」→ 1.3-1.4
- [ ] 3.4 「上课链接选填」→ 1.7-1.8
- [ ] 3.5 「基于终态enrollment创建课次被拒绝」→ 1.9-1.10
- [ ] 3.6 「SCHEDULED状态可编辑」→ 1.11（编辑测试的前置状态）
- [ ] 3.7 「COMPLETED状态不可编辑基础信息」→ 1.11-1.12
- [ ] 3.8 「存储与显示分离（时区）」→ 1.13-1.14

全部勾选后才可执行`/opsx:archive add-lesson-scheduling`。
