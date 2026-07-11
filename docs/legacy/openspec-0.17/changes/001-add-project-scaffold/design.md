# Design: 工程脚手架技术方案

## 为什么选modernc.org/sqlite而非mattn/go-sqlite3

这不是一个随意的技术选型，而是直接服务于PRD第十九章"All-in-One单文件
部署"这个核心产品定位的关键决策。mattn/go-sqlite3依赖CGO，意味着交叉
编译Windows/Linux/macOS三个平台的二进制时，需要在对应平台准备C交叉
编译工具链——这类似Gitea项目（一个成熟的Go+SQLite单二进制服务）在其
构建文档中专门花篇幅说明的CGO交叉编译复杂度。modernc.org/sqlite是
纯Go实现，`GOOS=windows GOARCH=amd64 go build`这样简单的环境变量切换
就能在一台开发机上编出全部三平台产物，这是我们能够承诺"7天MVP"节奏
的技术前提之一。

## 目录结构设计原则

严格按照PRD16.3节的cmd/internal/pkg三段式：
- `cmd/zedu-server/main.go`：只负责组装（读配置→连数据库→注册路由→
  启动定时任务→监听端口），不写业务逻辑
- `internal/<module>/`：每个业务能力一个包，包内再按
  model/dto/handler/service/repository/routes/errors七个文件严格分层
  （详见PRD16.4节），这是从002开始所有change都要遵守的强制约束，
  本change负责把这个目录骨架先摆出来，即使还没有业务代码
- `pkg/`：跨模块复用的纯工具函数（money、datetime、response、
  pagination、validator、crypto、errors），这些包不应该反向依赖任何
  internal/下的业务包

## 数据库迁移机制

使用goose而非GORM的AutoMigrate，原因：AutoMigrate只能新增字段/表，
不能安全处理字段类型变更、字段重命名、索引调整这些场景，且没有
"回滚"能力。goose的每个迁移文件包含`-- +goose Up`和`-- +goose Down`
两部分，本change要求：
- 00001_init_schema.sql：建表+索引+外键（Up），DROP TABLE（Down，
  按外键依赖的反序删除，避免外键约束报错）
- 种子数据用独立的迁移文件（而非糅合进建表迁移），放在
  migrations/seed/目录下，理由：种子数据未来可能需要针对不同部署
  实例（日语机构 vs K12机构）选择性应用，与"建表"这个所有实例都
  必须执行的动作在生命周期上不同

## 种子数据的幂等性设计

种子数据的SQL必须写成幂等的（重复执行不产生重复记录或报错），
使用`INSERT OR IGNORE INTO ... WHERE NOT EXISTS`的模式，或者依赖
表上已有的UNIQUE约束+`INSERT OR IGNORE`语法。这是因为：
1. 开发环境可能会重复跑迁移做调试
2. 001-add-project-scaffold的种子数据（日语模板+K12模板+字典+配置）
   和002-add-auth-and-init的初始化向导逻辑要能配合工作——向导检测
   course_domain为空时才提示初始化，但本change的种子数据其实已经
   把日语模板写进去了，所以"是否要在Day 0就直接写入日语模板种子数据，
   还是等002的向导来触发"需要明确：**本change只负责让种子数据SQL
   文件"存在且正确"，实际是否在迁移阶段就执行、还是留给002的
   /init/apply-template接口按需调用，由002决定**，本change的验收
   标准只关注种子数据SQL本身的正确性（可以在测试环境里手动执行
   验证），不涉及"什么时机自动执行"这个业务判断

## PRAGMA配置为什么是启动强制项而非可选配置

`journal_mode=WAL`直接影响多个Operator同时操作时的并发读写能力，
`foreign_keys=ON`是SQLite的一个反直觉的默认行为陷阱（SQLite默认不
强制外键约束，必须显式开启），如果遗漏这一条，后续所有change里
写的外键约束测试都会静默失效（能插入违反外键的脏数据而不报错），
这是一个特别容易被AI工具遗漏、且遗漏后果具有隐蔽性的配置项，因此
本change必须把它做成服务启动流程里不可跳过的一步，而不是配置文件
里的一个可选项。
