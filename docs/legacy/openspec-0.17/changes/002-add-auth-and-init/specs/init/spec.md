## ADDED Requirements

### Requirement: 初始化状态检测
系统必须能检测是否需要展示初始化向导。

#### Scenario: 已有课程数据
- **WHEN** GET /init/status 且 course_domain表非空
- **THEN** 返回 needsInit=false

#### Scenario: 无课程数据
- **WHEN** GET /init/status 且 course_domain表为空
- **THEN** 返回 needsInit=true 及可选模板列表

### Requirement: 应用种子模板
系统必须支持将预置模板数据写入课程体系表，且提供多个模板选项以
体现架构的通用性。

#### Scenario: 应用日语模板
- **WHEN** POST /init/apply-template 且 templateCode=japanese
- **THEN** course_domain/track/level/skill_tag被写入日语模板的完整种子数据
  （见PRD第五章5.3节），且该操作幂等（重复调用不产生重复数据）

#### Scenario: 应用K12模板
- **WHEN** POST /init/apply-template 且 templateCode=k12
- **THEN** course_domain/track/level/skill_tag被写入K12模板的完整种子数据
  （见PRD第五章5.4节），证明系统架构不绑定单一学科

#### Scenario: 选择空白模板
- **WHEN** POST /init/apply-template 且 templateCode=blank
- **THEN** 不写入任何课程体系种子数据，仅将初始化状态标记为已完成，
  由运营者后续在课程体系维护页面自行创建

#### Scenario: 重复应用不产生副作用
- **WHEN** 已应用过日语模板后，再次调用相同的apply-template请求
- **THEN** 不产生重复记录，返回成功（幂等）
