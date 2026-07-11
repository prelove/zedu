## 1. 仓库与工具

- [x] 1.1 [依赖:M0] 固定Go/Node/OpenSpec及前后端依赖版本；输出工具版本证据
- [x] 1.2 建立backend/frontend/scripts目录和最小任务入口；不得创建业务空壳
- [x] 1.3 编写失败测试验证`GET /healthz`和包依赖方向，再实现最小服务

## 2. 数据基础

- [x] 2.1 编写迁移up/down/up失败测试及PRAGMA/外键测试
- [x] 2.2 实现modernc SQLite连接和最小基础迁移，使2.1通过
- [x] 2.3 建立幂等模板seed框架，但本change只放框架级样例

## 3. 前端、i18n与日志

- [x] 3.1 建立Vue/Vite/TS strict shell和健康页，编写构建/单测
- [x] 3.2 建立三语资源及缺key失败检查，验证中日文路径和内容往返
- [x] 3.3 建立脱敏结构化日志、request/correlation ID及测试

## 4. CI与验收

- [x] 4.1 建立Windows/Ubuntu CI：OpenSpec strict、格式、静态检查、测试、迁移、构建
- [x] 4.2 运行全部门禁并保存到`docs/acceptance/evidence/M1/`
- [x] 4.3 更新追踪矩阵、项目状态、风险和路线图；独立Reviewer签署

## 5. 场景覆盖

- [x] 5.1 最小健康检查→1.3、3.1
- [x] 5.2 禁止业务空壳→1.2、4.2
- [x] 5.3 迁移往返→2.1-2.2
- [x] 5.4 编码往返→3.2、4.1
- [x] 5.5 CI拒绝不合规变更→4.1-4.2
