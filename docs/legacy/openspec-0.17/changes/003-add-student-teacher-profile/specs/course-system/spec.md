## ADDED Requirements

### Requirement: 课程体系四层维护
系统必须支持课程领域/方向/等级/能力标签的增删改和启用禁用，不得在
代码中硬编码具体学科名称，且禁用操作不影响历史数据的可读性。

#### Scenario: 新增自定义课程方向
- **WHEN** 在已有领域下新增一个自定义方向，code在该领域下唯一
- **THEN** 创建成功

#### Scenario: code格式与唯一性校验
- **WHEN** 提交的code包含除字母数字下划线以外的字符，或与同领域下
  已有code重复
- **THEN** 前者返回40001，后者返回40901

#### Scenario: 禁用不影响历史引用
- **WHEN** 禁用某个课程等级
- **THEN** 历史课次中引用该等级的记录仍能正常查询和显示，该等级
  只是不再出现在"新建课程报名"等前瞻性操作的可选列表中

#### Scenario: 排序调整
- **WHEN** 调整某方向下多个等级的sort_order
- **THEN** 后续查询该方向的等级列表时按新的sort_order排列

#### Scenario: 领域类型枚举校验
- **WHEN** 新增课程领域时提交type字段
- **THEN** 必须是LANGUAGE/K12/SPORT/ART/ACADEMIC/CERTIFICATE/OTHER
  之一，不属于此范围返回40001
