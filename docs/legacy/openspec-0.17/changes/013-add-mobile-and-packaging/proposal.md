# Proposal: 移动端页面与跨平台部署打包

## Why
运营者日常有相当比例的操作发生在手机上（外出时确认课次、快速查看
待续费学生），PC端的Soybean Admin页面在手机浏览器上虽然响应式布局
能用，但信息密度和交互方式并不是为触屏优化的。同时，PRD反复强调的
"All-in-One单文件部署"承诺，需要在本change中真正兑现为可执行的跨
平台构建产物和安装脚本，否则前面11个change写的所有代码都只能停留在
"能在开发机上跑"的状态。

## 业务背景
根据PRD第十一章，移动端策略不是重新做一套独立的前端项目，而是在
现有Soybean Admin工程内新增几个"移动优先"的响应式页面路由，这几个
页面遵循"少表格、多卡片、大按钮、少字段"的原则，且触摸区域不小于
44×44px（这是移动端可用性的行业基本标准，来自苹果和谷歌各自的
人机界面指南）。

跨平台打包能落地依赖的关键前提，是001-add-project-scaffold中就已经
确定使用modernc.org/sqlite（纯Go，无CGO依赖）——本change正是要验证
这个早期技术决策的正确性：如果不是纯Go驱动，本change的"一台开发机
编出三平台二进制"这个目标根本无法达成。

## What Changes
- 新增移动优先页面：/mobile/today、/mobile/confirm、/mobile/recharge、
  /mobile/alerts
- 新增三平台交叉编译脚本
- 新增Windows(WinSW)和Linux(systemd)的服务安装脚本
- 新增Litestream云备份集成（可选功能，取决于是否配置S3/R2）

## Non-Goals
- 不实现原生移动App（Flutter/React Native均为PRD明确的V3范围，
  本change的"移动端"完全是响应式Web页面，不涉及任何原生开发）
- 不实现PWA（manifest.json+service worker属于PRD的V1.5范围，
  本change只做响应式页面本身，PWA化是后续独立的改进）
- 不实现Wails桌面壳打包（属于PRD第十八章的可选能力，若后续需要
  再作为独立change提出，本change聚焦"部署"而非"桌面封装"）
- 不实现自动化的Litestream健康检查告警（本change只做集成配置，
  监控告警属于运维范畴，不在本项目V1范围）

## Impact
- Affected specs: mobile-web（新增）、deployment-packaging（新增）
- Affected code: frontend/admin/src/views/mobile/、
  scripts/build.sh、scripts/build.ps1、deploy/zedu-service.xml、
  deploy/install.sh、deploy/install-service.bat、
  deploy/litestream.yml
- 依赖：004（enrollment）、005（lesson）、006（attendance-confirmation）、
  007（payment）都需要先完成，因为四个移动页面分别是这些capability
  的简化视图，没有对应的后端接口就无法实现前端页面
- 被依赖：无（本change是MVP之后的体验与运维完善，不阻塞014的测试
  与上线工作，但014的跨平台部署测试用例依赖本change产出的构建脚本）
