# 国际化与编码规范

- locale：zh-CN、ja-JP、en-US；API错误用稳定code，UI/邮件本地化。
- 缺key在CI失败；生产fallback不得展示空字符串或原始占位符。
- 源码/Markdown/JSON/YAML/SQL为UTF-8无BOM、LF；bat/cmd为CRLF。
- 禁止用系统ANSI保存文件；Windows PowerShell 5.1写UTF-8无BOM需使用`.NET UTF8Encoding(false)`，不能依赖重定向默认编码。
- 对内CSV为UTF-8；面向日本Excel下载使用UTF-8 BOM、RFC4180、CRLF。CP932导入必须显式选择，禁止静默猜测。
- 测试中文、日文、emoji、全角、逗号、引号、换行及中日文路径；必要时Unicode NFC规范化。
