## 1. 仓库与目录骨架

- [ ] 1.1 初始化Git仓库，建立backend/frontend/docs三层目录
- [ ] 1.2 backend: `go mod init`，建立cmd/internal/pkg/migrations四段目录，
      migrations下再建sqlite/mysql/postgres/seed四个子目录
- [ ] 1.3 frontend: 拉取Soybean Admin(Naive UI版)官方脚手架
- [ ] 1.4 建立.gitignore（排除data/、logs/、backup/）
- [ ] 1.5 建立Makefile：dev/build/test/migrate-up/migrate-down五个命令
- [ ] 1.6 编写失败测试：GET /healthz 应返回200和{"status":"ok"}
      文件：backend/internal/app/health_test.go
- [ ] 1.7 实现最小Gin服务使测试通过
      文件：backend/cmd/zedu-server/main.go, backend/internal/app/health.go
- [ ] 1.8 编写失败测试：pkg/money等工具包不应import任何internal/包
      文件：backend/pkg/depcheck_test.go（用go/ast或简单的import扫描实现）
- [ ] 1.9 实现代码结构使1.8测试通过（若失败调整import关系）
- [ ] 1.10 提交：git commit -m "chore: project scaffold with layered architecture"

## 2. 数据库迁移

- [ ] 2.1 编写失败测试：迁移后应能查到31张表
      文件：backend/migrations/migration_test.go
- [ ] 2.2 将PRD第十二章12.2节DDL转为goose迁移文件的Up部分
      文件：backend/migrations/sqlite/00001_init_schema.sql
- [ ] 2.3 编写对应的Down部分（按外键依赖反序DROP TABLE）
- [ ] 2.4 迁移文件Up部分开头加入三条PRAGMA(journal_mode=WAL/
      foreign_keys=ON/busy_timeout=5000)
- [ ] 2.5 实现迁移执行逻辑使2.1测试通过
      要点：使用modernc.org/sqlite驱动，禁止mattn/go-sqlite3
- [ ] 2.6 编写失败测试：插入违反外键约束的记录应报错
- [ ] 2.7 验证2.6通过（若失败检查PRAGMA foreign_keys是否生效）
- [ ] 2.8 编写失败测试：执行goose down后全部31张表被删除且无报错
- [ ] 2.9 验证2.8通过（若失败检查Down部分的DROP顺序）
- [ ] 2.10 编写失败测试：服务启动后查询PRAGMA状态应为WAL/ON/≥5000
      文件：backend/internal/app/pragma_test.go
- [ ] 2.11 实现启动时校验/设置PRAGMA的逻辑使测试通过
- [ ] 2.12 提交：git commit -m "feat(db): initial schema migration with rollback and pragma enforcement"

## 3. 种子数据

- [ ] 3.1 编写失败测试：日语模板应用后course_domain应包含"日语"及其下
      4个方向、JLPT6个等级+会话3个等级、9个能力标签
      文件：backend/migrations/seed/japanese_template_test.go
- [ ] 3.2 编写日语模板种子SQL（PRD第五章5.3节）
      文件：backend/migrations/seed/japanese_template.sql
- [ ] 3.3 编写K12模板种子SQL（PRD第五章5.4节，本change只需存在文件，
      不需要默认执行）
      文件：backend/migrations/seed/k12_template.sql
- [ ] 3.4 编写失败测试：payment_method应有6条记录
- [ ] 3.5 编写支付方式种子SQL（PRD第七章7.3节）
      文件：backend/migrations/seed/payment_methods.sql
- [ ] 3.6 编写失败测试：attendance_outcome_type应有9条记录且建议值正确
- [ ] 3.7 编写出勤分类种子SQL（PRD第八章8.3节表格）
      文件：backend/migrations/seed/attendance_outcome_types.sql
- [ ] 3.8 编写失败测试：system_config应包含base_currency=JPY、
      base_currency_locked=0、lesson_reminder_minutes=30、
      balance_alert_lessons=3
- [ ] 3.9 编写系统配置种子SQL
      文件：backend/migrations/seed/system_config.sql
- [ ] 3.10 编写失败测试：对已应用日语模板的数据库重复执行种子SQL，
       course_domain等表记录数不应翻倍
- [ ] 3.11 实现幂等逻辑（INSERT OR IGNORE模式）使3.10测试通过
- [ ] 3.12 验证3.1/3.4/3.6/3.8全部测试通过
- [ ] 3.13 提交：git commit -m "feat(db): idempotent seed data for templates and dictionaries"

## 4. AI编码规范文档落地

- [ ] 4.1 将openspec/project.md内容整理为可直接喂给AI工具的system prompt文档
- [ ] 4.2 用规范生成一个测试性模块骨架（如parent模块），人工核对是否符合
      handler/service/repository分层要求
- [ ] 4.3 提交：git commit -m "docs: AI coding standards reference"

## 5. 规格场景覆盖检查表

对照本change下specs/data-foundation/spec.md的全部Scenario，逐条
标注验证task：

- [ ] 5.1 「本地启动」→ 1.6-1.7
- [ ] 5.2 「目录结构符合规范」→ 1.1-1.2
- [ ] 5.3 「迁移执行」→ 2.1、2.5
- [ ] 5.4 「迁移可回滚」→ 2.8-2.9
- [ ] 5.5 「PRAGMA配置生效」→ 2.4、2.10-2.11
- [ ] 5.6 「日语模板种子数据校验」→ 3.1-3.2
- [ ] 5.7 「K12模板种子数据存在但不默认应用」→ 3.3
- [ ] 5.8 「字典与配置数据校验」→ 3.4-3.9
- [ ] 5.9 「种子数据幂等」→ 3.10-3.11
- [ ] 5.10 「依赖方向校验」→ 1.8-1.9

全部勾选后才可执行`/opsx:archive add-project-scaffold`。
