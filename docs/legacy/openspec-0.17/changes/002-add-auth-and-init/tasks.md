## 1. 用户登录接口

- [ ] 1.1 编写失败测试：POST /auth/login 正确凭证应返回200和两个token
      文件：backend/internal/auth/service_test.go
- [ ] 1.2 实现最小代码使测试通过
      文件：backend/internal/auth/model.go, dto.go, service.go, handler.go, routes.go
      要点：密码用bcrypt(cost=12)校验；JWT用golang-jwt/jwt/v5签发，
      claim包含sub/role/exp
- [ ] 1.3 编写失败测试：错误密码应返回401且login_fail_count递增
- [ ] 1.4 实现代码使1.3通过
- [ ] 1.5 编写失败测试：连续5次失败应锁定账号，第6次返回40103
- [ ] 1.6 实现代码使1.5通过
- [ ] 1.7 编写失败测试：登录成功后login_fail_count应重置为0
- [ ] 1.8 实现代码使1.7通过
- [ ] 1.9 编写失败测试：locked_until过期后应允许新的登录尝试
- [ ] 1.10 实现代码使1.9通过（检查是否有状态残留判断错误）
- [ ] 1.11 提交：git commit -m "feat(auth): login with lockout and fail-count reset"

## 2. Token刷新与登出

- [ ] 2.1 编写失败测试：POST /auth/refresh 用有效refreshToken应返回新的accessToken和新refreshToken
- [ ] 2.2 实现代码使测试通过（refresh_token哈希存储于user_account）
- [ ] 2.3 编写失败测试：用已经刷新过一次的旧refreshToken再次刷新应返回401
- [ ] 2.4 实现token轮换逻辑使测试通过
- [ ] 2.5 编写失败测试：POST /auth/logout 应使refreshToken失效
- [ ] 2.6 实现代码使测试通过
- [ ] 2.7 提交：git commit -m "feat(auth): token refresh rotation and logout"

## 3. 首次启动默认账号

- [ ] 3.1 编写失败测试：全新user_account表启动后应自动创建一条role=OWNER记录
      文件：backend/internal/auth/bootstrap_test.go
- [ ] 3.2 实现最小代码使测试通过
      要点：随机密码仅打印到控制台/日志，不通过任何API返回
- [ ] 3.3 编写失败测试：user_account表已有记录时重启不应重复创建
- [ ] 3.4 验证3.3通过
- [ ] 3.5 编写失败测试：用初始密码登录后，访问非change-password接口应被拦截
- [ ] 3.6 实现代码使测试通过（可用一个must_change_password标志位实现）
- [ ] 3.7 编写失败测试：改密码时未提交正确旧密码应被拒绝
- [ ] 3.8 实现代码使测试通过
- [ ] 3.9 提交：git commit -m "feat(auth): bootstrap default owner account with forced password change"

## 4. 初始化向导接口

- [ ] 4.1 编写失败测试：course_domain非空时GET /init/status返回needsInit=false
- [ ] 4.2 实现最小代码使测试通过
- [ ] 4.3 编写失败测试：POST /init/apply-template(japanese)应写入正确数量的
      种子数据(4个方向、6个JLPT+3个会话等级、9个能力标签)
- [ ] 4.4 实现最小代码使测试通过，确保幂等(重复调用不重复插入)
- [ ] 4.5 编写失败测试：POST /init/apply-template(k12)应写入K12模板数据，
      证明模板机制不绑定单一学科
- [ ] 4.6 实现代码使测试通过（复用001已准备的K12种子SQL）
- [ ] 4.7 编写失败测试：POST /init/apply-template(blank)不应写入任何
      课程数据，但应标记初始化完成
- [ ] 4.8 实现代码使测试通过
- [ ] 4.9 提交：git commit -m "feat(init): onboarding template application with three options"

## 5. 前端：登录页与初始化向导

- [ ] 5.1 登录页组件 + Pinia auth store，对接1-2的接口
- [ ] 5.2 路由守卫：未登录跳转/login，未改密码跳转改密码页，
      accessToken过期时自动尝试refresh，refresh也失败则跳转登录页
- [ ] 5.3 初始化向导页面，对接4的接口（本次部署固定选日语模板，无需
      用户手动选择，见PRD第五章5.2节说明；但向导本身作为通用组件保留
      三个选项的完整UI，供未来其他部署实例使用）
- [ ] 5.4 提交：git commit -m "feat(frontend): login and onboarding pages"

## 6. 规格场景覆盖检查表

对照本change下specs/auth/spec.md和specs/init/spec.md的全部Scenario，
逐条标注验证task：

- [ ] 6.1 「正确凭证登录成功」→ 1.1-1.2
- [ ] 6.2 「错误密码」→ 1.3-1.4
- [ ] 6.3 「连续失败锁定」→ 1.5-1.6
- [ ] 6.4 「登录成功清零失败计数」→ 1.7-1.8
- [ ] 6.5 「锁定期满后允许重试」→ 1.9-1.10
- [ ] 6.6 「正常刷新」→ 2.1-2.2
- [ ] 6.7 「旧token刷新后失效」→ 2.3-2.4
- [ ] 6.8 「登出使token失效」→ 2.5-2.6
- [ ] 6.9 「全新环境启动」→ 3.1-3.2
- [ ] 6.10 「非首次启动不重复创建」→ 3.3-3.4
- [ ] 6.11 「强制改密码」→ 3.5-3.6
- [ ] 6.12 「改密码需验证旧密码」→ 3.7-3.8
- [ ] 6.13 「已有课程数据」→ 4.1-4.2
- [ ] 6.14 「无课程数据」→ 4.1-4.2
- [ ] 6.15 「应用日语模板」→ 4.3-4.4
- [ ] 6.16 「应用K12模板」→ 4.5-4.6
- [ ] 6.17 「选择空白模板」→ 4.7-4.8
- [ ] 6.18 「重复应用不产生副作用」→ 4.4（幂等性验证包含在内）

全部勾选后才可执行`/opsx:archive add-auth-and-init`。
