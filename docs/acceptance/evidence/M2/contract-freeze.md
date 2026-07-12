# M2 共享契约与依赖冻结记录

- 日期：2026-07-12
- 依据：PRD v3.1-r1、ADR-007、OpenSpec change `add-m2-core-management`
- 状态：ACCEPTED（规划门禁）
- 批准人：Product Owner

## 精确依赖

| 依赖 | 版本 | 用途 | Owner |
|---|---:|---|---|
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` | access token 签发与验证 | GLM / M2-GLM-01 |
| `golang.org/x/crypto` | `v0.54.0` | bcrypt 密码哈希 | GLM / M2-GLM-01 |
| `vue-router` | `v5.1.0` | 前端受保护路由 | Kimi / M2-KIMI-01 |

不得新增 Pinia；M2 使用 Vue 内置响应式状态。不得使用 `latest`、`*` 或自行实现 JWT/bcrypt。

## HTTP 外层与错误码

- 成功：`{ "code": 0, "data": <payload> }`
- 失败：`{ "code": <业务码>, "message": <稳定错误键>, "requestId": <string> }`
- 列表 data：`{ "items": [], "page": 1, "pageSize": 20, "total": 0 }`；`page` 从 1 开始，`pageSize` 为 1–100。
- 固定错误：`40101` 未认证或 token 失效、`40102` 登录失败、`40103` 锁定、`40301` 权限不足、`40401` 不存在、`40901` 数据冲突、`42201` 状态不允许。

## 路由与角色矩阵

| 路由 | 方法 | Owner | Operator | 说明 |
|---|---|:---:|:---:|---|
| `/auth/login` | POST | 公开 | 公开 | 返回 access token，设置 refresh cookie |
| `/auth/refresh` | POST | refresh cookie | refresh cookie | 轮换 refresh session |
| `/auth/logout` | POST | 是 | 是 | 撤销当前 session |
| `/auth/me` | GET | 是 | 是 | 当前账号 |
| `/users` | GET/POST | 是 | 否 | Operator 账号管理 |
| `/users/{id}/disable` | POST | 是 | 否 | 禁用并撤销 session |
| `/onboarding/initialize` | POST | 是 | 否 | 显式首次初始化 |
| `/onboarding/reset` | POST | 是 | 否 | 仅无业务数据时 |
| `/students`、`/students/{id}` | GET/POST/PATCH | 是 | 是 | 学生资料 |
| `/students/{id}/parents`、`/students/{id}/parents/{parentId}` | GET/POST/PATCH | 是 | 是 | 家长联系人 |
| `/teachers`、`/teachers/{id}` | GET/POST/PATCH | 是 | 是 | 老师资料 |
| `/teachers/{id}/capabilities`、`/teachers/{id}/availability` | GET/POST/PATCH | 是 | 是 | 能力和可授时间 |
| `/course-domains`、`/tracks`、`/levels`、`/capability-tags` | GET/POST/PATCH | 是 | 是 | 课程字典 |
| `/students/{id}/enrollments`、`/enrollments/{id}` | GET/POST/PATCH | 是 | 是 | 报名 |
| `/enrollments/{id}/assignments`、`/assignments/{id}/end` | GET/POST | 是 | 是 | 师生安排 |

所有 M2 业务路由均要求 Bearer access token；Owner 包含 Operator 权限。refresh token 仅使用 `HttpOnly; Secure; SameSite=Strict` cookie，JSON、日志与 operation_log 绝不包含 refresh token、access token、密码或密码哈希。

## 已冻结业务规则

1. `student.email` 可为空；非空时全局唯一，含软删除记录。创建或更新冲突均返回 HTTP 409 / `40901`；没有“仍然新建”旁路。
2. `teacher_capability` 以 `(teacher_id, track_id, level_id)` 唯一；结束记录写 `effective_to`，不得删除历史。
3. `student → enrollment → assignment` 是 M2 的关系链；一个 enrollment 最多一个 ACTIVE assignment；替换必须单事务结束旧记录并创建新记录。
4. 每个成功业务写操作与 `operation_log` 同事务；失败、冲突、未授权不得留下成功审计记录。

## 禁止范围（负面验收）

M2 禁止新增或暴露下列任一 API、迁移、导航、页面或按钮：lesson、attendance、payment、payment evidence、notification、backup、report、payout、正式结款、学生/老师/家长登录。发现即退回任务，不以隐藏菜单或 TODO 规避。
