# Zedu 项目交付上下文快照

- 任务：审查并修订现有 PRD、OpenSpec proposal/design/spec/tasks，建立工程、质量与验收规范，形成路线图，随后实施、验收并推送至 GitHub。
- 期望结果：得到可由人类及 GLM、Devin、Kimi、Codex 等 AI 编码工具稳定执行的单一事实源、可追溯任务链、工程骨架及持续验收证据。
- 用户指定方法：OpenSpec + Superpowers；中文文档优先；系统支持中/日/英国际化；兼顾 Windows 10 日文版开发与编码兼容性。
- 当前证据：仓库尚未初始化 Git；根目录有 README、5 份 Markdown 调研/PRD文档、1 份 DOCX、2 份 SVG、OpenSpec 001-014 共 14 组 change；尚无业务代码。
- 当前工具：OpenSpec CLI 0.17.2；未发现名为 `superpowers` 的独立 CLI；OMX 0.11.12；Go 1.23.3；Node 24.8.0；Git 2.53.0.windows.1。
- 关键约束：不得在需求未澄清时锁定业务架构；不得凭空补业务规则；账务一致性、幂等、国际化、Windows 编码和可复现测试均需成为强制门禁。
- 新增证据：`docs/2_prd/Zedu-PRD-Final-v3.1.md` 存在并明确声明为唯一事实源；其通过配置化原则收敛了此前多数业务假设，并列出 V1 的 18 项范围、12 项非目标、MVP/V1 路线和 21 项验收用例。
- 已发现风险：README 与 project.md 对最终 PRD 的路径引用错误；两处还引用不存在的 `Zedu-OpenSpec-Superpowers执行方案.md`；最终 PRD 附录 C 仍有 10 项上线配置待运营者确认；MVP（7天）与完整 V1（后续3~4周）均有定义，但当前开发委托未明确本轮交付终点。
- 待确认：本轮首先交付可验收 MVP，还是直接以完整 V1 为当前完成条件；附录 C 配置项的确认时点；哪些架构/产品决策可由项目代理自主决定；GitHub 远程认证状态与推送策略。
- 范围决议：先完成并验收 MVP，再进入 V1；总路线图必须覆盖全生命周期并持续保存任务、依赖、进度与验收定位。
- 决策边界：代理可自主修订文档结构、OpenSpec格式、任务/依赖、代码组织、测试、CI、编码兼容及不改变业务语义的技术方案；产品范围、财务语义、角色权限、个人信息处理、持续费用服务、生产部署和真实数据迁移必须人工确认；附录C配置可暂用默认值开发但须在MVP上线验收前确认。
- 工具核验：本机 OpenSpec 0.17.2，npm 官方包最新为 1.6.0；现有 14 个 change 在当前 CLI 下全部校验失败（0/14），需升级后按新格式逐项修复，不能直接据此开工。官方 Superpowers 仓库最新 tag 为 v6.1.1，需按实际 AI harness 分别安装，不存在通用 `superpowers` CLI 命令。
- 校验根因样本：001、006、014 的 Requirement 正文未使用 OpenSpec 强制要求的 SHALL/MUST 规范词；属于系统性格式缺陷而非个别文件损坏，14组均需迁移。
- 可能触点：`docs/`、`openspec/`、根 README、未来的 `backend/`、`frontend/`、CI、测试与发布目录。
