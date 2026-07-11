## Context

旧001把31张表、前后端脚手架和种子一次完成，变更过大且难以回滚。本设计只建立最小工程与共享基础，业务表随能力增量迁移。

## Goals / Non-Goals

目标：一条命令验证构建、迁移、测试和编码；形成稳定目录与依赖方向。非目标：任何业务能力和MVP外结款入口。

## Decisions

- 后端按 `cmd/internal/pkg/migrations`；业务模块采用model/dto/handler/service/repository/routes/errors分层。
- SQLite使用modernc驱动，启动强制foreign_keys、WAL和busy_timeout；migration支持up/down/up。
- 前端采用Vue3/Vite/TypeScript strict；只建立shell、i18n和健康页，不复制未锁定模板主分支。Naive UI延后到真实页面能力需要时再评估。
- 文本UTF-8/LF；日期UTC；locale为zh-CN/ja-JP/en-US；CI校验key parity。
- 日志结构化并含request/correlation ID，默认脱敏。

被拒方案：一次创建全部业务表（耦合大）；直接拉取脚手架main（不可复现）；依赖系统CP932（必然乱码）。

## Risks / Trade-offs

增量迁移增加文件数但显著降低跨能力冲突。Windows与Ubuntu双CI增加时间，但能提前发现路径、换行和编码问题。

## Migration / Rollback

先创建无业务数据的基础迁移；每次验证up/down/up。失败可删除开发临时库并回退该迁移；不得修改已发布迁移内容。
