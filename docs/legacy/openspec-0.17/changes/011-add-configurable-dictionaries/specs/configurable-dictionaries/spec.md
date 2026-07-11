## ADDED Requirements

### Requirement: 支付方式管理
系统必须支持后台增删改支付方式字典，已被引用的code不可修改，
禁用而非删除以保护历史数据。

#### Scenario: 新增支付方式
- **WHEN** 提交一个新的支付方式，code在字典中唯一
- **THEN** 创建成功，充值表单的下拉选项立即包含新增项

#### Scenario: code重复被拒绝
- **WHEN** 提交的code与已有支付方式重复
- **THEN** 返回40901

#### Scenario: 编辑不修改code
- **WHEN** 对已有支付方式提交编辑请求，请求体中包含code字段且与
  原值不同
- **THEN** code保持不变，仅name/sort_order/enabled等其他字段被更新

#### Scenario: 禁用不影响历史引用
- **WHEN** 禁用一个已被历史充值记录引用的支付方式
- **THEN** 该支付方式不再出现在新建充值的选择列表中，但历史充值
  记录中该字段对应的名称仍能正常显示

### Requirement: 出勤分类管理
系统必须支持后台增删改出勤分类字典及其建议值，且能正确区分
"清空建议值"与"不修改建议值"两种操作语义。

#### Scenario: 修改建议值
- **WHEN** 编辑某个出勤分类的suggested_charge_ratio为新数值
- **THEN** 更新成功，后续课后确认页选择该分类时应带出新的建议值

#### Scenario: 清空建议值
- **WHEN** 编辑请求体显式包含suggested_deduct_lessons字段且值为null
- **THEN** 该字段被更新为NULL，前端后续选择该分类时不再自动填充
  建议扣课时

#### Scenario: 未传字段保持不变
- **WHEN** 编辑请求体不包含suggested_teacher_pay_ratio字段
- **THEN** 该字段的原值保持不变，不会被误置空

#### Scenario: 新增自定义出勤分类
- **WHEN** 运营者需要一个PRD预置9种之外的出勤分类（如"设备故障"）
- **THEN** 系统允许新增自定义code的出勤分类，可自行设定建议值或
  留空为纯手工判断

### Requirement: 付款凭证上传与访问控制
系统必须支持为充值记录上传付款凭证，严格限制数量、格式和访问权限，
且不因上传中途失败而产生孤儿数据。

#### Scenario: 成功上传
- **WHEN** 对一笔CONFIRMED状态的充值上传一张jpg图片(小于5MB)
- **THEN** 上传成功，payment_attachment新增一条记录

#### Scenario: 超过数量限制
- **WHEN** 同一笔充值已有3个附件，尝试上传第4个
- **THEN** 返回42201，提示已达上传数量上限

#### Scenario: 格式不支持被拒绝
- **WHEN** 上传一个.exe文件（即使伪造扩展名为.jpg）
- **THEN** 系统通过文件二进制内容(magic number)校验识别真实类型
  并拒绝，返回40001，而非仅凭扩展名或声明的Content-Type判断

#### Scenario: 未登录无法访问
- **WHEN** 未携带有效JWT直接请求凭证文件的下载URL
- **THEN** 返回401，不返回文件内容

#### Scenario: 已作废充值的凭证不可删除
- **WHEN** 尝试删除一笔status=VOIDED的充值下的附件
- **THEN** 返回42201，保留该凭证作为作废记录的取证材料

#### Scenario: 上传中途失败不留孤儿文件
- **WHEN** 文件已写入磁盘但数据库记录插入失败（模拟）
- **THEN** 系统应清理已写入的物理文件，磁盘上不遗留无对应数据库
  记录的孤儿文件
