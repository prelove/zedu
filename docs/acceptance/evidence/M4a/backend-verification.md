# M4a 后端验证证据

日期：2026-07-19（Asia/Tokyo）
范围：`add-m4a-lesson-scheduling`，仅基础排课。

已验证：

- 迁移 `006_m4a_lesson` 的 SQLite up/down/up，课程表与索引可重复创建。
- 课程创建要求 Owner/Operator、ACTIVE 报名与 ACTIVE 师生安排；生成唯一 `lesson_no`，默认 `SCHEDULED`。
- `Asia/Tokyo` 的本地 `19:00` 规范化为 UTC `10:00`，同时保留 `Asia/Tokyo`。
- 时长、时区与 WeChat HTTPS 链接无效时返回 42201。
- 更新、取消仅适用于 `SCHEDULED`；取消后再次更新返回 42201。
- 创建、更新和取消与对应 `operation_log` 在同一事务内；未写入 M3 payment 数据。
- 未认证请求返回 40101，非 Owner/Operator 写入返回 40301。

执行门禁：

```text
GOTOOLCHAIN=local go test ./internal/app/lesson ./internal/platform/database -count=1 -v
GOTOOLCHAIN=local go vet ./...
GOTOOLCHAIN=local go build ./cmd/zedu-server
```

结果：通过。Linux `-race` 继续由 CI/MVP 总验收覆盖；本地 Windows 纯 Go SQLite 环境不执行该项。
