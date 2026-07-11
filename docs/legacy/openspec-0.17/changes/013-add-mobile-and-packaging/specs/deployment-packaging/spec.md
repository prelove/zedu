## ADDED Requirements

### Requirement: 跨平台构建
系统必须能在一台开发机上产出Windows/Linux/macOS三个平台的独立
可执行文件，且不依赖任何平台特定的交叉编译工具链。

#### Scenario: 一体化构建脚本
- **WHEN** 执行构建脚本
- **THEN** 依次完成前端构建、go:embed嵌入、三平台Go二进制编译，
  产出三个独立文件且均可在对应平台独立运行

#### Scenario: 无CGO依赖验证
- **WHEN** 检查构建产物的依赖
- **THEN** 三个二进制文件均不依赖任何外部C库（验证modernc.org/sqlite
  的纯Go特性生效）

### Requirement: 服务化部署脚本
系统必须提供Windows和Linux平台的服务安装脚本，且服务异常退出时
能自动重启。

#### Scenario: Windows服务安装
- **WHEN** 以管理员身份执行install-service.bat
- **THEN** 服务被注册为Windows服务并自动启动，服务异常退出后
  由WinSW自动重启

#### Scenario: Linux服务安装
- **WHEN** 执行install.sh
- **THEN** 服务被注册为systemd服务，开机自启，异常退出后自动重启

### Requirement: 云备份的条件化启用
系统必须支持可选的Litestream云备份，且未启用时不产生任何外部
连接尝试。

#### Scenario: 备份范围包含上传文件
- **WHEN** 启用Litestream备份
- **THEN** 备份范围应包含data/uploads/目录（付款凭证），不能只
  备份SQLite数据库文件

#### Scenario: 未启用时不产生副作用
- **WHEN** config.yaml中litestream_enabled=false
- **THEN** 系统启动时不尝试连接任何S3/R2服务，不产生相关错误日志
