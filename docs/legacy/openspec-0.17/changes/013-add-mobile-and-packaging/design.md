# Design: 移动页面简化原则与跨平台构建流水线

## 移动页面为什么是独立路由而非PC页面的响应式降级

Soybean Admin的PC页面（如排课列表、充值表单）信息密度高，字段多，
如果单纯靠CSS媒体查询让同一个组件在窄屏下"挤"成手机可用的样子，
往往会牺牲可用性（字段挤在一起、按钮太小难点击）。因此本change的
四个移动页面是**独立的、专门设计的简化视图**，只调用现有API的一
个子集（不需要为移动端新增任何后端接口，四个页面复用的都是004~007
已经实现的接口），前端展示逻辑完全独立于PC对应页面。

- `/mobile/today`：调用GET /lessons?date=today，卡片列表展示，
  每张卡片信息量控制在"时间+学生+老师+状态"四项，[复制链接]按钮
  直接调用浏览器剪贴板API
- `/mobile/confirm`：调用GET /lessons?status=SCHEDULED,REMINDED&
  scheduledEndBefore=now，每张卡片内嵌简化版的出勤分类下拉+关键
  金额字段+提交按钮，复用006的POST /lessons/{id}/confirm接口
- `/mobile/recharge`：学生搜索框调用GET /students?keyword=，
  选中后的极简充值表单复用007的POST /finance/payments接口
- `/mobile/alerts`：调用GET /reports/dashboard中的待续费列表部分

## 触摸区域与视觉密度约束

所有可点击元素（按钮、卡片本身若可点击、下拉选择器）的最小尺寸
不低于44×44px（CSS中通常体现为`min-height: 44px`或等效的padding
设置），这是PRD第十五章15.4节明确的移动端可用性基本要求，也是
苹果Human Interface Guidelines和谷歌Material Design两方都建议的
最小触摸目标尺寺，本change在实现每个移动页面组件时都需要显式核对
这一点，而非依赖框架默认样式碰巧达标。

## 跨平台构建流水线设计

构建脚本的核心逻辑非常直接，因为modernc.org/sqlite不需要CGO：
```bash
GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o dist/zedu_windows_amd64.exe .
GOOS=linux   GOARCH=amd64 go build -ldflags="-w -s" -o dist/zedu_linux_amd64 .
GOOS=darwin  GOARCH=arm64 go build -ldflags="-w -s" -o dist/zedu_darwin_arm64 .
```
`-ldflags="-w -s"`去除调试符号表以减小二进制体积。构建前必须先执行
前端的`npm run build`并确保产物被go:embed正确嵌入（构建脚本应该是
"先构建前端→再构建后端"这个顺序的一体化脚本，而不是要求开发者手动
分两步执行并容易遗忘顺序）。

## WinSW与systemd安装脚本的设计要点

WinSW配置文件(zedu-service.xml)中的`onfailure action="restart"`
设置是必要的容错机制——如果服务因为未预期的panic而退出，Windows
服务管理器应该自动重启它，而不是让服务永久停止直到有人发现并手动
重启。systemd的install.sh同理配置`Restart=always`和`RestartSec=5`。

## Litestream集成的条件化设计

Litestream是可选功能，取决于运营者是否配置了S3/R2账号（config.yaml
中的`backup.litestream_enabled`开关）。本change的实现应该保证：当
该开关为false时，系统完全不应该尝试连接任何外部存储服务，不能因为
"代码里有这段逻辑"就在未配置的情况下产生连接失败的错误日志噪音。
Litestream的备份范围必须包含data/uploads/目录（付款凭证文件），
不能只备份SQLite数据库本身——这是对PRD第二十章20.3节要求的具体
技术落实。
