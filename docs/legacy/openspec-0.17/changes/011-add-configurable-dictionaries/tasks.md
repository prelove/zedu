## 1. 支付方式管理接口

- [ ] 1.1 编写失败测试：POST /system/payment-methods 新增成功，
      code唯一时创建，重复时返回40901
      文件：backend/internal/system/payment_method_write_test.go
- [ ] 1.2 实现最小代码使测试通过
      文件：backend/internal/system/payment_method_service.go,
      payment_method_handler.go
- [ ] 1.3 编写失败测试：PUT /system/payment-methods/{code} 编辑时
      即使传入不同code也不会修改原code，只更新name等字段
- [ ] 1.4 实现代码使测试通过（编辑逻辑显式忽略请求体中的code字段）
- [ ] 1.5 编写失败测试：禁用已被历史充值引用的支付方式后，历史记录
      的展示名称仍正常，但新建充值下拉不再出现
- [ ] 1.6 验证1.5通过
- [ ] 1.7 提交：git commit -m "feat(system): payment method write API with immutable code and soft disable"

## 2. 出勤分类管理接口

- [ ] 2.1 编写失败测试：PUT /system/attendance-outcomes/{code} 修改
      suggested_charge_ratio应生效
      文件：backend/internal/system/attendance_outcome_write_test.go
- [ ] 2.2 实现最小代码使测试通过
      文件：backend/internal/system/attendance_outcome_service.go,
      attendance_outcome_handler.go
- [ ] 2.3 编写失败测试：请求体显式传suggested_deduct_lessons=null应
      清空该字段
- [ ] 2.4 实现代码使测试通过（区分"字段未传"和"字段传null"两种语义，
      使用指针类型或map解析请求体判断字段是否存在）
- [ ] 2.5 编写失败测试：不传某字段时该字段原值不变
- [ ] 2.6 验证2.5通过
- [ ] 2.7 编写失败测试：新增一个自定义code的出勤分类（不在PRD预置9种
      范围内）应能成功创建
- [ ] 2.8 验证2.7通过
- [ ] 2.9 提交：git commit -m "feat(system): attendance outcome write API with nullable suggestions and custom types"

## 3. 付款凭证上传与访问控制

- [ ] 3.1 编写失败测试：上传合法jpg(<5MB)到CONFIRMED充值应成功
      文件：backend/internal/finance/attachment_test.go
- [ ] 3.2 实现最小代码使测试通过
      文件：backend/internal/finance/attachment_model.go,
      attachment_service.go, attachment_handler.go
      要点：存储路径data/uploads/payments/{payment_id}/{uuid}_{filename}
- [ ] 3.3 编写失败测试：已有3个附件时上传第4个应返回42201
- [ ] 3.4 实现数量校验使测试通过
- [ ] 3.5 编写失败测试：伪造扩展名的可执行文件应被拒绝(40001)
- [ ] 3.6 实现文件二进制内容(magic number)校验使测试通过，不仅仅
      检查扩展名或Content-Type
- [ ] 3.7 编写失败测试：未携带JWT访问凭证下载URL应返回401
      文件：backend/internal/finance/attachment_access_test.go
- [ ] 3.8 实现鉴权中间件覆盖凭证下载路由，使测试通过
- [ ] 3.9 编写失败测试：VOIDED状态充值的附件不可删除，返回42201
- [ ] 3.10 实现代码使测试通过
- [ ] 3.11 编写失败测试：模拟数据库插入失败场景，验证已写入磁盘的
      文件被自动清理，不留孤儿文件
      文件：backend/internal/finance/attachment_atomicity_test.go
- [ ] 3.12 实现文件写入与数据库插入的一致性保证逻辑使测试通过
- [ ] 3.13 提交：git commit -m "feat(finance): payment attachment upload with access control and orphan cleanup"

## 4. 前端：字典管理页与凭证上传

- [ ] 4.1 系统设置"支付方式管理"Tab（增删改+启用禁用）
- [ ] 4.2 系统设置"出勤分类管理"Tab（含三个建议值的编辑控件，
      支持清空为"无建议"，支持新增自定义分类）
- [ ] 4.3 系统设置"本位币设置"Tab（展示当前币种和锁定状态，
      锁定后输入框禁用，复用006已实现的PUT接口做真正的写入尝试
      以便前端能准确展示是否可编辑）
- [ ] 4.4 充值表单增加"上传付款凭证"按钮（最多3张，格式/大小校验提示）
- [ ] 4.5 充值详情页展示已上传凭证缩略图，点击查看大图
- [ ] 4.6 提交：git commit -m "feat(frontend): dictionary management and attachment upload UI"

## 5. 规格场景覆盖检查表

对照本change下specs/configurable-dictionaries/spec.md的全部
Scenario，逐条标注验证task：

- [ ] 5.1 「新增支付方式」→ 1.1-1.2
- [ ] 5.2 「code重复被拒绝」→ 1.1-1.2
- [ ] 5.3 「编辑不修改code」→ 1.3-1.4
- [ ] 5.4 「禁用不影响历史引用」→ 1.5-1.6
- [ ] 5.5 「修改建议值」→ 2.1-2.2
- [ ] 5.6 「清空建议值」→ 2.3-2.4
- [ ] 5.7 「未传字段保持不变」→ 2.5-2.6
- [ ] 5.8 「新增自定义出勤分类」→ 2.7-2.8
- [ ] 5.9 「成功上传」→ 3.1-3.2
- [ ] 5.10 「超过数量限制」→ 3.3-3.4
- [ ] 5.11 「格式不支持被拒绝」→ 3.5-3.6
- [ ] 5.12 「未登录无法访问」→ 3.7-3.8
- [ ] 5.13 「已作废充值的凭证不可删除」→ 3.9-3.10
- [ ] 5.14 「上传中途失败不留孤儿文件」→ 3.11-3.12

全部勾选后才可执行`/opsx:archive add-configurable-dictionaries`。
