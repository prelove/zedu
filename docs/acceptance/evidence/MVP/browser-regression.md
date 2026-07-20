# MVP 全量浏览器回归证据

## 结论

**PASS（技术门禁）**。2026-07-20 在隔离的 SQLite、附件和备份目录中，以真实后端和真实 Vite 开发代理完成 20 组端到端场景；zh-CN、ja-JP、en-US 各执行一次，共 **60/60 通过**（`unexpected=0`、`flaky=0`）。

这证明 M1–M6 的当前 MVP 路径可在浏览器中连通；不等同于生产发布批准，剩余 Product Owner 外部 UAT 闸门见 `release-gate.md`。

## 执行环境与约束

- 后端：`http://localhost:8080`，使用本次回归专属的 `release-candidate/zedu.db`。
- 前端：`http://localhost:5173`，经 Vite 真实代理访问后端。
- 数据、附件和备份：`output/playwright/mvp-full/release-candidate/` 下的隔离目录，未复用既有测试数据。
- 浏览器：Playwright，单 worker 串行执行以避免 SQLite 测试数据相互污染。
- 禁止并已核验：测试脚本没有 `page.route`；`/lessons`、`/notifications`、`/dashboard` 均通过真实 Vite 代理。

## 覆盖

| 范围 | 每个 locale 的场景 |
|---|---:|
| M1/M2：登录、路由守卫、Owner/Operator 权限、学生与课程基础 | 7 |
| M4a：课次创建与 Asia/Tokyo 排期 | 2 |
| M3：支付方式、充值、凭证、作废反向流水 | 3 |
| M4b/M5：通知、课后确认、余额与幂等/失败约束 | 4 |
| M6：工作台、Owner 备份、跨模块三语与敏感信息不泄露 | 4 |
| **合计** | **20 × 3 = 60** |

## 本轮发现并已修复的真实契约缺陷

1. 课程字典空集合曾序列化为 `items: null`，前端无法遍历，造成刚创建领域后不能立即创建方向。后端现稳定返回 `items: []`，并有回归测试。
2. 课次时长输入的原生 `pattern` 被双重转义，`0.5` 会被浏览器错误拒绝。现修正为可接受一至三位小数的实际 HTML 校验模式。
3. 课程字典表单在刷新关联数据期间可继续提交或保留陈旧表单。现关闭成功表单并在保存/加载期间禁用重复创建。

## 产物

- HTML 报告：`output/playwright/mvp-full/playwright-report/index.html`
- JSON 结果：`output/playwright/mvp-full/playwright-report/results.json`
- 运行日志、隔离数据库及测试数据：`output/playwright/mvp-full/release-candidate/`
- 失败时保留的 trace / 截图 / 视频：`output/playwright/mvp-full/test-results/`

结果 JSON 摘要：`expected=60`、`unexpected=0`、`flaky=0`、耗时约 7.9 分钟。

## 边界

- 本回归不发送真实邮件；Resend 真实受控收件人验证仍需 Product Owner 在安全配置的环境中执行。
- 本回归不替代人工账务核对、授权操作者备份恢复演练和 Win10 日文环境人工视觉验收。
