# 安全与隐私规范

- 密钥只来自环境/secret store，不进入DB、UI、日志、导出或备份；dev/test/prod隔离。
- 所有资源同时验证认证和对象授权，防IDOR；拒绝响应不泄露存在性与物理路径。
- 凭证存Web根外，随机名；扩展名/Content-Type/magic一致；限制大小数量；安全Content-Disposition、nosniff及适用sandbox。
- 文件与DB不能假装具有同一事务：使用临时文件、fsync、原子rename、状态和补偿/孤儿扫描。
- 备份manifest记录相对路径、大小、SHA-256；临时恢复验证成功后原子切换。
