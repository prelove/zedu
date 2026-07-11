## ADDED Requirements

### Requirement: 移动优先页面
系统必须提供四个针对手机浏览器优化的简化页面，且复用现有API接口
而不新增专属后端能力。

#### Scenario: 今日课程页面
- **WHEN** 手机浏览器访问/mobile/today
- **THEN** 展示今日课次的卡片列表，每张卡片可一键复制上课链接，
  时间已过的课次显示"确认出勤"按钮

#### Scenario: 待确认课次页面内嵌确认表单
- **WHEN** 在/mobile/confirm页面对某张卡片提交出勤确认
- **THEN** 调用与PC端相同的POST /lessons/{id}/confirm接口，
  成功后该卡片从列表中消失

#### Scenario: 快速充值页面
- **WHEN** 在/mobile/recharge搜索学生并提交极简充值表单
- **THEN** 调用与PC端相同的充值接口，创建成功后返回搜索页

#### Scenario: 触摸区域达标
- **WHEN** 检查四个移动页面的可点击元素样式
- **THEN** 全部可点击元素的最小尺寸不低于44×44px
