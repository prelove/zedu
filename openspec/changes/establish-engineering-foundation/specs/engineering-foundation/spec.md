## ADDED Requirements

### Requirement: 可运行的分层工程

系统 MUST 提供可独立构建的 Go 后端和 Vue/TypeScript 前端，并保持共享包不反向依赖业务模块。

#### Scenario: 最小健康检查
- **WHEN** 在全新开发环境安装锁定依赖并启动后端
- **THEN** `GET /healthz` MUST 返回200和稳定健康状态，前端构建 MUST 成功

#### Scenario: 禁止业务空壳
- **WHEN** 完成本change并扫描路由和页面
- **THEN** 系统 MUST NOT 注册认证、财务、通知、凭证或正式结款的假实现入口

### Requirement: 可逆的SQLite迁移

系统 MUST 使用modernc SQLite并通过增量迁移管理结构，连接 MUST 启用外键、WAL和busy timeout。

#### Scenario: 迁移往返
- **WHEN** 对全新临时数据库执行up、down、up
- **THEN** 每步 MUST 成功，最终schema MUST 与预期一致且外键约束生效

### Requirement: 三语和Win10日文环境基础

仓库和运行时 MUST 支持UTF-8、zh-CN/ja-JP/en-US及包含中日文的Windows路径。

#### Scenario: 编码往返
- **WHEN** 在Windows日文区域设置下读写中文、日文、emoji和全角字符
- **THEN** 源码、JSON、SQLite和UI MUST 无替换字符或乱码，三语key集合 MUST 一致

### Requirement: 可复现质量门禁

项目 MUST 在Windows和Ubuntu CI使用锁定工具版本运行OpenSpec strict、格式、静态检查、测试、迁移和构建。

#### Scenario: CI拒绝不合规变更
- **WHEN** 规格无MUST/SHALL、i18n缺key、迁移不可回滚或构建失败
- **THEN** 对应CI任务 MUST 失败并提供可定位输出
