## 1. 移动优先页面前端

- [ ] 1.1 编写/mobile/today页面：调用现有课次列表接口，卡片式展示
      文件：frontend/admin/src/views/mobile/today/index.vue
- [ ] 1.2 实现"复制链接"按钮（浏览器剪贴板API）
- [ ] 1.3 编写/mobile/confirm页面，内嵌简化版课后确认表单
      文件：frontend/admin/src/views/mobile/confirm/index.vue
- [ ] 1.4 验证确认提交后卡片正确从列表消失（复用006的接口）
- [ ] 1.5 编写/mobile/recharge页面
      文件：frontend/admin/src/views/mobile/recharge/index.vue
- [ ] 1.6 编写/mobile/alerts页面
      文件：frontend/admin/src/views/mobile/alerts/index.vue
- [ ] 1.7 人工核对四个页面全部可点击元素的CSS最小尺寸≥44×44px
      （无自动化测试，人工用浏览器开发者工具逐一测量确认）
- [ ] 1.8 提交：git commit -m "feat(mobile): mobile-first pages reusing existing APIs"

## 2. 跨平台构建脚本

- [ ] 2.1 编写build.sh：先执行前端npm run build，再依次编译三平台
      Go二进制，产出到dist/目录
      文件：scripts/build.sh
- [ ] 2.2 编写build.ps1（Windows PowerShell版本，供Windows开发环境使用）
      文件：scripts/build.ps1
- [ ] 2.3 编写失败测试（人工执行验证，非自动化单元测试）：在一台
      机器上运行build.sh，验证产出三个文件且各自能在对应平台
      （或对应平台的虚拟机/容器）独立启动
- [ ] 2.4 若2.3失败（如CGO依赖导致交叉编译失败），检查是否有代码
      路径意外引入了mattn/go-sqlite3或其他CGO依赖
- [ ] 2.5 提交：git commit -m "chore: cross-platform build scripts"

## 3. 服务化部署脚本

- [ ] 3.1 编写zedu-service.xml（WinSW配置，含onfailure restart）
      文件：deploy/zedu-service.xml
- [ ] 3.2 编写install-service.bat
      文件：deploy/install-service.bat
- [ ] 3.3 人工验证：在Windows环境执行安装脚本，验证服务正确注册
      并可通过服务管理器查看状态
- [ ] 3.4 编写zedu.service（systemd unit文件，含Restart=always）
      文件：deploy/zedu.service
- [ ] 3.5 编写install.sh
      文件：deploy/install.sh
- [ ] 3.6 人工验证：在Linux环境执行安装脚本，验证systemctl status
      显示服务正常运行，且kill进程后能自动重启
- [ ] 3.7 提交：git commit -m "chore: windows and linux service installation scripts"

## 4. Litestream云备份集成（可选）

- [ ] 4.1 编写litestream.yml模板，备份范围包含data/zedu.db和
      data/uploads/目录
      文件：deploy/litestream.yml
- [ ] 4.2 编写失败测试：config.yaml中litestream_enabled=false时，
      服务启动不应产生任何S3/R2连接尝试
      文件：backend/internal/backup/litestream_test.go
- [ ] 4.3 实现条件化启动逻辑使测试通过
- [ ] 4.4 若已配置真实S3/R2账号，人工验证litestream restore命令
      能正确恢复数据库和uploads目录
- [ ] 4.5 提交：git commit -m "feat(backup): conditional litestream integration covering uploads dir"

## 5. 规格场景覆盖检查表

- [ ] 5.1 「今日课程页面」→ 1.1-1.2
- [ ] 5.2 「待确认课次页面内嵌确认表单」→ 1.3-1.4
- [ ] 5.3 「快速充值页面」→ 1.5
- [ ] 5.4 「触摸区域达标」→ 1.7
- [ ] 5.5 「一体化构建脚本」→ 2.1-2.3
- [ ] 5.6 「无CGO依赖验证」→ 2.3-2.4
- [ ] 5.7 「Windows服务安装」→ 3.1-3.3
- [ ] 5.8 「Linux服务安装」→ 3.4-3.6
- [ ] 5.9 「备份范围包含上传文件」→ 4.1、4.4
- [ ] 5.10 「未启用时不产生副作用」→ 4.2-4.3

全部勾选后才可执行`/opsx:archive add-mobile-and-packaging`。
