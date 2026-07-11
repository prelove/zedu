## ADDED Requirements

### Requirement: Excel批量导入学生与老师
系统必须支持从Excel批量导入学生和老师数据，单行失败不影响其他行，
且提供详细的导入报告。

#### Scenario: 全部成功导入
- **WHEN** 上传一份全部数据合法的学生Excel
- **THEN** 全部行导入成功，返回的导入报告显示成功数=总行数

#### Scenario: 部分行失败不影响其他行
- **WHEN** 上传的Excel中第5行邮箱重复、其余行数据合法
- **THEN** 除第5行外的其他行全部导入成功，第5行在导入报告中被标注
  跳过及原因，接口不因为第5行的问题而整体失败

#### Scenario: 邮箱大小写不敏感去重
- **WHEN** 系统已存在邮箱"a@b.com"的学生，导入的Excel中某行邮箱为
  "A@B.com"
- **THEN** 该行被判定为邮箱重复（大小写不敏感匹配），按照PRD业务
  规则R14跳过该行

#### Scenario: 学习方向不匹配时的降级处理
- **WHEN** 导入的Excel中"学习方向"列填写的名称在课程体系中找不到
  匹配的track
- **THEN** 该学生基础信息仍导入成功，但不创建课程报名，导入报告中
  明确提示该行需要人工补充课程报名

#### Scenario: 下载导入模板
- **WHEN** 调用GET /students/import-template
- **THEN** 返回一份包含正确列名和至少一行示例数据的Excel模板文件

### Requirement: 数据导出
系统必须支持将学生、老师、课次记录导出为Excel文件。

#### Scenario: 按筛选条件导出学生列表
- **WHEN** 调用GET /students/export 附带状态/课程方向筛选参数
- **THEN** 导出的Excel只包含符合筛选条件的学生记录，字段与列表页
  展示的列一致
