# M1-KIMI-01：前端与三语基础（给 Kimi）

## 执行身份与目标

你是本工单唯一实现者。目标是建立最小Vue 3/Vite/TypeScript strict前端、健康状态页、三语资源及可执行测试。不得拉取或复制整个Soybean Admin，也不得实现业务页面。

从`origin/main`创建分支`m1/kimi-frontend-foundation`。如与GLM并行，使用独立clone/worktree；不得在对方工作目录切换分支。

## 开始前必须读取

1. `AGENTS.md`
2. `docs/status/PROJECT_STATUS.md`
3. `docs/governance/GOVERNANCE.md`
4. `docs/standards/coding-standard.md`
5. `docs/standards/testing-standard.md`
6. `docs/standards/i18n-and-encoding.md`
7. 当前OpenSpec proposal/design/spec/tasks全部文件

执行前运行并记录：

```powershell
openspec instructions apply --change establish-engineering-foundation --json
openspec validate --all --strict --no-interactive
node --version
npm --version
```

## 写入范围

只允许新增/修改：

```text
frontend/package.json
frontend/package-lock.json
frontend/index.html
frontend/tsconfig*.json
frontend/vite.config.*
frontend/src/**
frontend/tests/**
```

禁止修改根目录、`backend/`、`openspec/`、`docs/`、`.github/`、`scripts/`及任务勾选。

## 功能契约

- 页面只能展示应用名、当前locale切换、后端健康状态（loading/healthy/unavailable）和版本占位信息。
- locale固定`zh-CN`、`ja-JP`、`en-US`；三份key集合完全一致。
- API错误使用稳定内部状态映射为本地化文案，不能向用户展示原始异常堆栈。
- 日期/金额基础格式器需显式locale与Asia/Tokyo，不依赖Windows系统语言。
- 必须支持中文/日文/emoji文本及中日文工程路径。

## TDD执行步骤

1. 先写失败测试：三语key parity、fallback、locale切换、健康状态三态、日期/JPY格式化。
2. 保存失败输出后实现最小代码。
3. 开启TypeScript strict，禁止`any`逃逸和未解释的lint禁用。
4. 测试后端不可达时页面仍正常渲染并显示本地化“服务不可用”。
5. 不引入业务导航、登录页、Dashboard图表或结款入口。

## 必须执行的验收命令

```powershell
Set-Location frontend
npm ci
npm run lint
npm run typecheck
npm run test:unit
npm run build
```

测试至少覆盖三种locale各一次；新增代码行覆盖率目标≥80%。

## 依赖约束

- 使用稳定、固定版本，不写`latest`、`*`或无上限范围。
- 只引入构建Vue/Vite/TS、i18n、测试和lint所需最小依赖。
- Naive UI若健康页没有真实需求可暂不引入；禁止为未来页面预装大批依赖。

## 禁止事项

- 不得拉取Soybean Admin main/master或复制完整后台模板。
- 不得实现认证、人员、课程、财务、通知、上传、正式结款页面。
- 不得写假API、Mock业务数据或隐藏业务路由。
- 不得修改共享进度文档或自称`ACCEPTED`。

## 交付格式

提交一个Lore commit，并回复：commit SHA、改动文件、红灯/绿灯证据、依赖及版本理由、覆盖率、三语验证结果、已知风险和未测试项。需要修改共享契约时停止并标记`BLOCKED`。
