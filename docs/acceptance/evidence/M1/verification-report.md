# M1 集成验证报告

- 日期：2026-07-11
- 环境：Windows 10 日文区域环境；Go 1.23.3 windows/amd64；Node v24.8.0；npm 11.7.0
- 集成分支：`m1/integration-glm-kimi`（从 `origin/main` 创建）
- 合并顺序：GLM 后端 → Kimi 前端（按 M1_EXECUTION_BOARD 集成顺序）

## 合并来源

| 工单 | 分支 | Commit SHA | 改动文件数 |
|---|---|---|---|
| M1-GLM-01 | `m1/glm-backend-foundation` | `909c518` | 14 文件 (+943 行) |
| M1-KIMI-01 | `m1/kimi-frontend-foundation` | `c919671` | 29 文件 (+7462 行) |

合并无冲突（GLM 只写 `backend/`，Kimi 只写 `frontend/`，写入范围无交集）。

## 后端测试证据（GLM）

| 检查 | 命令 | 结果 |
|---|---|---|
| 格式化 | `go fmt ./...` | 退出码 0，无输出 |
| 静态检查 | `go vet ./...` | 退出码 0，无输出 |
| 单元测试 | `go test ./... -count=1 -v` | 8/8 PASS（migration up/down/up、PRAGMA、外键强制、UTF-8 往返、健康检查、依赖边界、日志脱敏、请求ID） |
| 稳定性 | `go test ./... -run "TestHealth\|TestMigration\|TestPragma\|TestUTF8\|TestRedaction" -count=20` | 退出码 0，20×8 全部 PASS |
| 构建 | `go build ./cmd/zedu-server` | 退出码 0，二进制生成成功 |

### BLOCKED: `-race` 标志

- 命令：`go test ./... -race -count=1`
- 结果：`go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`
- 原因：Windows 工具链默认 CGO_ENABLED=0，`-race` 需要 CGO
- 绕行：已运行非 race 版本全部通过；建议 CI 在 Ubuntu 上启用 `-race` 验证

## 前端测试证据（Kimi）

| 检查 | 命令 | 结果 |
|---|---|---|
| 依赖安装 | `npm ci` | 退出码 0，346 packages installed |
| Lint | `npm run lint` (`eslint . --max-warnings 0`) | 退出码 0，无警告 |
| 类型检查 | `npm run typecheck` (`vue-tsc --noEmit`) | 退出码 0，TS strict 通过 |
| 单元测试 | `npm run test:unit` (`vitest run`) | 57/57 PASS（8 个测试文件） |
| 覆盖率 | `npm run test:coverage` | 行/语句/分支/函数覆盖率均 100%（目标 ≥80%） |
| 构建 | `npm run build` (`vue-tsc --noEmit && vite build`) | 退出码 0，41 模块转换成功 |

### 三语验证

- locale 固定 `zh-CN`、`ja-JP`、`en-US`
- `ja-JP.ts` 和 `en-US.ts` 均使用 `LocaleSchema` 类型约束，编译期保证 key parity
- `i18n.test.ts` 递归比对三语 key 集合完全一致
- 测试覆盖三种 locale 各至少一次（health.test.ts 中 zh-CN/ja-JP/en-US 分别验证）
- CJK 和 emoji 字符验证通过（`i18n.test.ts` 正则匹配）
- 日期/JPY 格式化使用显式 locale 和 `Asia/Tokyo`，不依赖 Windows 系统语言

## 依赖审查

### 后端依赖

| 依赖 | 版本 | 理由 |
|---|---|---|
| `modernc.org/sqlite` | v1.29.10 | 纯 Go SQLite 驱动，禁止 CGO 和 mattn/go-sqlite3（任务要求） |
| 间接依赖（humanize/uuid/lru 等） | 固定版本 | modernc.org/sqlite 的传递依赖，非直接引入 |

### 前端依赖

| 依赖 | 版本 | 理由 |
|---|---|---|
| `vue` | 3.5.13 | Vue 3 框架（任务要求） |
| `vue-i18n` | 11.1.1 | 三语 i18n 支持（任务要求） |
| `vite` | 6.0.7 | 构建工具（任务要求） |
| `typescript` | 5.7.3 | TS strict 模式（任务要求） |
| `vitest` | 2.1.8 | 单元测试框架 |
| `@vue/test-utils` | 2.4.6 | Vue 组件测试工具 |
| `eslint` / `@vue/eslint-config-typescript` | 9.17.0 / 14.2.0 | Lint（`no-explicit-any: error`） |
| `jsdom` | 25.0.1 | 测试 DOM 环境 |
| `@vitest/coverage-v8` | 2.1.8 | 覆盖率收集 |

所有版本均为固定版本，无 `latest`、`*` 或无上限范围。未引入 Naive UI（健康页无真实需求）。未拉取 Soybean Admin。

## 已知风险

1. **API 路径不匹配**：前端 `HealthStatus.vue` 请求 `/api/healthz`，后端服务 `/healthz`。Vite proxy 将 `/api` 转发到 `http://localhost:8080` 但不重写路径，导致 `/api/healthz` → `http://localhost:8080/api/healthz`（后端无此路由）。这是 M1-CODEX-01 集成验收需要处理的契约对齐问题。前端测试使用 mock fetch，后端测试独立验证，各自通过但端到端集成需修复路径映射。
2. **`-race` 不可用**：Windows 默认 CGO_ENABLED=0，`-race` 需要 CGO。非 race 测试全部通过，建议 CI 在 Linux 上启用 `-race`。
3. **npm audit 漏洞**：`npm ci` 报告 10 个漏洞（2 low, 4 moderate, 2 high, 2 critical），均为 dev 依赖中的传递依赖。不影响生产构建，但应在 M1-CODEX-01 中评估是否需要 `npm audit fix`。

## 未测试项

- 端到端集成（前端→后端健康检查实际 HTTP 请求）——因 API 路径不匹配未验证
- `-race` 并发检测——Windows CGO 限制
- CI 流水线（GitHub Actions）——M1-CODEX-01 范围
- OpenSpec strict 校验重跑——需在集成分支上重新运行

## 状态

- 集成合并：完成，无冲突
- 全部测试：通过（后端 8/8 + 前端 57/57）
- 构建：通过（后端二进制 + 前端 dist）
- 独立 Reviewer 签署：待定
- OpenSpec tasks 勾选：待独立 Reviewer 签署后更新
