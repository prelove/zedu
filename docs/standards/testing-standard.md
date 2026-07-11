# 测试规范

- TDD：先运行并保存失败证据，再做最小实现，重构后全量复验。
- Unit覆盖纯规则、状态机、校验、权限、i18n、幂等；新增代码≥80%，财务/认证/RBAC/上传/通知关键分支100%。
- Integration使用真实SQLite临时库、migration、fake Resend和临时文件目录，覆盖并发、故障注入与补偿。
- E2E默认Playwright；Windows原生安装/文件选择器等用computer-use补充。
- 测试不得向真实用户发信；真实Resend只使用批准测试收件箱。
- 不允许删除测试、放宽断言、跳过失败场景来让门禁通过。
