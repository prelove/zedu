## ADDED Requirements

### Requirement: 受限付款凭证上传
授权 Owner/Operator MUST 只能向 CONFIRMED payment 上传最多三份付款凭证。每份文件 MUST 是 `jpg`、`jpeg`、`png`、`webp` 或 `pdf`，且不超过 5 MiB；系统 MUST 忽略客户端文件名作为存储路径，MUST 不在 operation_log 或应用日志写入文件内容、临时路径或敏感凭证数据。

#### Scenario: 上传第三份有效凭证
- **WHEN** payment 已有两份有效附件且 Operator 上传符合格式与大小的第三份文件
- **THEN** 系统 MUST 创建受控 attachment 元数据和文件，并返回可列举的附件信息

#### Scenario: 超限或恶意文件被拒绝
- **WHEN** payment 已有三份附件，或上传文件超出 5 MiB、扩展名/内容类型不被允许、文件名含路径片段
- **THEN** 系统 MUST 拒绝请求且不增加 attachment 记录、不发布文件、不产生成功审计

### Requirement: 凭证鉴权下载与失败补偿
系统 MUST 通过认证端点列举和下载 attachment，禁止静态或匿名直链。下载端点 MUST 校验 attachment 与 payment 的归属并只读取受控根目录下的登记文件。文件发布、attachment 元数据和审计发生失败时 MUST 进行补偿，使系统不把未完成上传报告为成功。

#### Scenario: 未认证或错误归属下载
- **WHEN** 未认证用户访问凭证，或 attachmentId 不属于 path 中 paymentId
- **THEN** 系统 MUST 返回 `40101` 或 `40401`，且不泄露文件内容、绝对路径或存储结构

#### Scenario: 文件发布失败
- **WHEN** 附件 metadata 事务提交后无法将临时文件发布到受控路径
- **THEN** 系统 MUST 返回 `50002` 并补偿 metadata，最终列表中不存在该附件且受控目录没有可访问的半成品
