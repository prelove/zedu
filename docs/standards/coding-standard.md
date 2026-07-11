# 编码规范

- Go：`gofmt`、`go vet`；handler只处理HTTP，service负责业务/事务，repository封装GORM，DTO与model分离。
- TypeScript开启strict；用户文案使用i18n key，稳定英文标识不本地化。
- 金额使用decimal或明确最小货币单位，禁止float；时间存UTC，展示按Asia/Tokyo。
- 所有业务写操作有审计；日志不得包含密码、token、API Key、完整邮箱或凭证内容。
- 新依赖必须有用途、版本、许可证、替代方案和批准记录。
- API使用统一错误码；数据库迁移必须可重放并验证up/down/up。
- 代码和文档UTF-8/LF；不得依赖CP932/GBK。PowerShell读写必须显式编码。
