## ADDED Requirements

### Requirement: 用户登录
系统必须支持用户名密码登录，返回JWT access token和refresh token。

#### Scenario: 正确凭证登录成功
- **WHEN** 用户提交正确的username和password
- **THEN** 返回200，data中包含accessToken(60分钟有效)和refreshToken(14天有效)

#### Scenario: 错误密码
- **WHEN** 用户提交错误密码
- **THEN** 返回401，login_fail_count加1

#### Scenario: 连续失败锁定
- **WHEN** 同一账号连续5次登录失败
- **THEN** 账号locked_until设置为当前时间+15分钟，第6次尝试返回40103

#### Scenario: 登录成功清零失败计数
- **WHEN** 用户曾有过若干次失败尝试(未达到锁定阈值)，随后一次登录成功
- **THEN** login_fail_count被重置为0

#### Scenario: 锁定期满后允许重试
- **WHEN** locked_until已经过期，用户再次尝试登录
- **THEN** 系统不再拒绝该次尝试（即锁定状态不会因为过期而残留阻挡）

### Requirement: Token刷新与轮换
系统必须支持用refresh token换取新的access token，且旧refresh token
在刷新后立即失效。

#### Scenario: 正常刷新
- **WHEN** 提交有效且未过期的refreshToken
- **THEN** 返回新的accessToken和新的refreshToken

#### Scenario: 旧token刷新后失效
- **WHEN** 使用已经被用于刷新过一次的旧refreshToken再次尝试刷新
- **THEN** 返回401，提示需要重新登录

#### Scenario: 登出使token失效
- **WHEN** 调用登出接口
- **THEN** 当前refreshToken立即失效，无法再用于刷新

### Requirement: 首次启动默认账号
系统首次启动检测到user_account表为空时，必须自动创建一个Owner账号。

#### Scenario: 全新环境启动
- **WHEN** 服务首次启动且user_account表无记录
- **THEN** 自动创建一个role=OWNER的账号，随机生成16位密码并仅打印到
  控制台和logs/init.log，不写入任何可被前端读取的接口

#### Scenario: 非首次启动不重复创建
- **WHEN** 服务重启但user_account表已有记录
- **THEN** 不会再次创建默认账号，也不会重新生成密码

#### Scenario: 强制改密码
- **WHEN** 使用初始密码首次登录成功
- **THEN** 除POST /auth/change-password外的其他接口应返回提示"需先修改密码"

#### Scenario: 改密码需验证旧密码
- **WHEN** 提交改密码请求
- **THEN** 必须同时提交正确的旧密码，否则拒绝，即使是首次强制改密码流程
