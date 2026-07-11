# Zedu 开发规范与桌面端封装方案

> 用途：作为系统提示词/规范文档提供给 Windsurf / Devin / GLM / Codex 等 AI 编程工具遵循
> 整理日期：2026-07-03
> 配套文档：Zedu-PRD-v2.0-完整实装版.md（本文档是其技术规范的补充与细化，不重复已有内容）

---

## 第一部分：后端架构规范（AI编码工具必须遵循）

### 1.1 参照标准声明

本项目**不采用**任何现成的 Go 后台管理框架（如 gin-vue-admin、go-admin、soybean-admin-go 等），原因是这些框架普遍默认绑定 MySQL/PostgreSQL + Redis，与本项目"SQLite + 单文件 All-in-One"的部署目标冲突。

本项目的代码组织方式参照以下两个社区公认标准，AI 编码工具在生成代码时须遵循这两份规范的精神，而非自由发挥：

1. **`golang-standards/project-layout`**（Go 项目目录布局的事实标准）
   - 项目地址：github.com/golang-standards/project-layout
   - 采纳的约定：`cmd/`（入口）、`internal/`（私有业务代码）、`pkg/`（可复用工具包）三段式目录结构

2. **Ardan Labs Service 架构**（生产级 Go 服务分层参考）
   - 采纳的约定：每个业务模块内部严格分离 handler（HTTP层）/ service（业务逻辑层）/ repository（数据访问层），层与层之间通过接口解耦，不允许跨层直接调用

### 1.2 SQLite 相关的强制约束

本项目**必须**使用 `modernc.org/sqlite` 作为 SQLite 驱动，**禁止**使用 `mattn/go-sqlite3`。

原因：`mattn/go-sqlite3` 依赖 CGO，交叉编译需要为每个目标平台单独配置 C 交叉编译器（可参考 Gitea 项目文档中关于跨平台构建 SQLite 支持的复杂度说明），这会破坏"一台开发机一键编出三平台二进制"的核心目标。`modernc.org/sqlite` 是纯 Go 实现，无 CGO 依赖，交叉编译只需标准的 `GOOS`/`GOARCH` 环境变量。

```go
// 正确
import _ "modernc.org/sqlite"
db, err := sql.Open("sqlite", "./data/zedu.db")

// 禁止
import _ "github.com/mattn/go-sqlite3"
```

### 1.3 模块内部结构强制规范

每个 `internal/` 下的业务模块必须遵循以下文件划分，AI 生成代码时严格对应：

```
internal/<module>/
├── model.go       仅包含 GORM struct 定义，不包含任何业务逻辑方法
├── dto.go         API 请求/响应结构体，字段须与 model 解耦（不允许直接把 model 作为 handler 返回值）
├── handler.go      仅做：解析请求参数 → 调用 service → 包装响应。禁止在此层写 SQL 或业务判断
├── service.go      所有业务规则、状态机校验、事务编排必须在此层完成
├── repository.go   所有 GORM 查询语句封装于此，handler/service 不得直接操作 *gorm.DB
├── routes.go       仅做路由注册，不包含逻辑
└── errors.go       模块级业务错误定义（错误码见PRD第八章8.3）
```

**AI工具生成代码时的检查清单**（每次生成新模块代码后自查）：

```
□ handler.go 中是否出现了 SQL 语句或 db.Where(...) 调用？→ 若有，违规，需移到 repository.go
□ service.go 中是否直接调用了 gin.Context？→ 若有，违规，service层不应感知HTTP细节
□ model.go 中是否包含了业务方法（如计算金额）？→ 若有，违规，应移到 service.go
□ 金额运算是否使用了 float64？→ 若有，违规，必须用 shopspring/decimal
□ 时间处理是否直接用了 time.Now() 而未考虑时区？→ 若有，须通过 pkg/datetime 统一处理
□ 涉及多表写入的操作是否包裹在 db.Transaction() 中？→ 账务相关操作必须有事务
```

### 1.4 事务边界规范（重中之重）

以下操作**必须**在单一数据库事务内完成，AI 生成代码时不允许拆分成多次独立提交：

```
课后确认（PRD 3.6节）：attendance写入 + student_account_ledger + teacher_account_ledger +
                        lesson_finance + enrollment余额更新 + teacher待结算更新 + lesson状态更新

充值确认：student_payment写入 + student_account_ledger + enrollment余额更新

充值作废：student_payment状态更新 + student_account_ledger冲正记录 + enrollment余额更新

结款提交：teacher_payout写入 + teacher_account_ledger + teacher.unpaid_amount更新
```

参考代码模式（PRD 10.3节已给出示例，此处强调：**任何AI生成的账务相关函数，必须能够回答"如果第3步失败，前2步的写入是否会自动回滚"这个问题，答案必须是"是"**）。

### 1.5 幂等性规范

以下场景必须实现幂等保护，AI生成代码时需主动加入判断，而非事后补充：

```
课前提醒任务：以 lesson.remind_sent_at IS NULL 作为幂等门控
充值/结款接口：接口层面支持业务单号（payment_no/payout_no）作为幂等键，
              重复提交相同单号应返回已有记录而非重复创建
```

---

## 第二部分：桌面端封装技术方案

### 2.1 技术选型：Wails v2（非v3）

| 候选方案 | 是否采纳 | 理由 |
|---------|:---:|------|
| **Wails v2** | ✅ 采纳 | 纯Go技术栈，与现有后端语言一致；v2为稳定版，生产可用；架构与现有go:embed方案高度契合 |
| Wails v3 | ❌ 暂不采纳 | 截至2026年仍为Alpha阶段，API可能变动，不适合作为生产依赖 |
| Tauri v2 | ❌ 不采纳 | 需要引入Rust工具链作为第二语言；需要sidecar模式管理外部Go进程生命周期，增加复杂度；仅在极致追求安装包体积（3MB级别）时才值得考虑 |
| Electron | ❌ 不采纳 | 安装包150MB+，与项目"轻量"定位完全矛盾 |

### 2.2 架构设计

**核心原则：桌面封装不改动任何业务代码，只增加一个可选的构建目标（build target）。**

```
现有架构（Web部署）：
  zedu-server 二进制（go:embed前端 + Gin HTTP Server）
  ↓
  监听 :8080，浏览器访问

桌面封装架构（增量方案）：
  zedu-server 二进制（同一套代码）
  ↓
  Wails 在同一进程内启动原生窗口
  ↓
  窗口内部WebView直接指向内嵌的HTTP Server（同进程，无需跨进程通信）
  ↓
  用户看到的是一个原生桌面窗口，而非浏览器标签页
```

关键点：**不采用 Tauri 式的"sidecar外部进程"模式**，因为 Wails 本身就是 Go 程序，可以在同一个 `main.go` 里：

```go
func main() {
    if isDesktopMode() {
        // Wails桌面模式：在goroutine中启动现有的Gin server（监听本地端口）
        go startGinServer(":18080")
        // Wails窗口指向本地server
        runWailsApp("http://localhost:18080")
    } else {
        // Web部署模式：现有逻辑不变
        startGinServer(":8080")
    }
}
```

这样做的好处：**业务代码（handler/service/repository）完全不感知运行模式**，桌面封装是纯粹的"外壳"工作，不产生技术债。

### 2.3 需要新增的工作项（供开发任务拆分参考）

```
□ 引入 wails.io CLI 工具，`wails init` 生成桌面壳工程骨架
□ 实现端口动态分配（避免与用户本机其他服务冲突），启动时探测可用端口
□ 实现单实例锁（防止用户重复点击图标启动多个进程）
□ 系统托盘图标 + 右键菜单（打开主窗口/退出）
□ 窗口关闭行为定义（是完全退出，还是最小化到托盘——建议提供用户可切换的偏好设置）
□ 三平台图标资源准备（.ico for Windows / .icns for macOS / .png for Linux）
□ Windows：确认WebView2运行时检测逻辑（Win11通常已预装，需为Win10准备bootstrapper）
□ 打包产物体积和启动时间验证（目标：安装包<30MB，冷启动<2秒）
```

### 2.4 已知限制与应对

```
限制：WebView2（Windows）不支持Cookie
应对：本项目认证方案本来就是JWT Bearer Header（存于Pinia store+localStorage替代方案），
      不依赖Cookie，此限制对本项目无影响，无需额外处理

限制：Wails v2的自动更新机制不如Tauri成熟
应对：V1阶段桌面版本更新频率低（本地工具性质），可接受"手动下载新版覆盖安装"，
      暂不实现自动更新，V2再评估是否需要
```

### 2.5 与移动端的关系澄清（避免混淆）

```
桌面封装（Wails）解决的问题：让Windows/Mac用户获得"原生应用"体验，而非打开浏览器
移动端支持（已在PRD中，无需额外方案）：手机浏览器直接访问部署的Web地址，
                                        已有响应式 /mobile/* 页面覆盖
若想要移动端"类App"体验（V1.5计划）：PWA（manifest.json + service worker），
                                     与Wails/Tauri无关，是纯Web技术方案

原生移动封装（Tauri Mobile / Wails v3 Android桥接）：均为Alpha阶段，V1不采用，
                                                      如V3阶段有强需求再评估
```

---

## 第三部分：跨端UI一致性检查清单

由于桌面端（Wails窗口内WebView）和移动端（手机浏览器）在不同操作系统上使用不同的渲染引擎，需要在测试阶段覆盖以下矩阵，而非增加新的设计规范：

| 平台 | 渲染引擎 | 测试要点 |
|------|---------|---------|
| Windows桌面（Wails）| WebView2（Chromium内核）| 与开发时用Chrome调试的表现基本一致，风险最低 |
| macOS桌面（Wails）| WebKit（Safari内核）| 需额外测试：CSS Grid/Flexbox边界情况、UnoCSS生成的原子类是否有WebKit前缀需求 |
| iOS移动浏览器 | WebKit（Safari内核）| 同上，且需测试触摸区域大小（≥44×44px，PRD 15.4已定义）|
| Android移动浏览器 | Chromium内核 | 风险较低 |
| Linux桌面（Wails，如有需要）| WebKitGTK | 优先级最低，若无Linux用户可暂不测试 |

**结论：不新增视觉设计工作，只需在Sprint 7测试阶段（PRD第十七章）的测试矩阵中加入"macOS Safari内核渲染验证"这一项即可。**

---

*本文档配合 Zedu-PRD-v2.0-完整实装版.md 使用，建议在Sprint 0启动时，将第一部分"后端架构规范"整体作为system prompt提供给AI编程工具（Windsurf/Devin/GLM/Codex），确保生成代码从第一行起就符合规范，而非事后重构。*
