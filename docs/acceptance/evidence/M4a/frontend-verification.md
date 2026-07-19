# M4a 前端验证证据

日期：2026-07-19（Asia/Tokyo）
范围：受认证保护的 `/lessons` 页面与 M4a API adapter。

已验证：

- adapter 使用冻结路径 `/lessons`、`/lessons/{id}` 和 `/lessons/{id}/cancel`，不使用历史 `/api` 前缀。
- 创建与更新请求体分离：更新不得携带 `enrollmentId` 或 `assignmentId`。
- 课程页面支持创建、状态筛选、分页及仅 `SCHEDULED` 状态的取消操作。
- `zh-CN`、`ja-JP`、`en-US` 均补齐 M4a `nav.lessons` 与 `lessons.*` 词条。

执行门禁：

```text
npm run lint
npm run typecheck
npm run test:unit -- --run
npm run build
```

结果：通过。全系统人工 UAT 与 Kimi 浏览器回归按项目决策汇入后续 MVP 总验收，不作为 M4a 自动化门禁的替代。
