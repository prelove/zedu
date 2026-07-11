# Design: 认证机制与初始化流程

## JWT结构设计
Access Token的claim中至少包含：`sub`(user_account.id)、`role`
(OWNER/OPERATOR)、`exp`(签发时间+60分钟)。Refresh Token不放业务信息，
只是一个随机字符串，存储在数据库中（可以是user_account表新增一个
refresh_token_hash字段，或者独立一张refresh_token表——考虑到V1单
用户量级不大，选择更简单的方案：直接在user_account表存储当前有效
的refresh_token的哈希值，每次刷新时生成新token并替换旧哈希，这样
天然实现了"单设备登录、旧token刷新后失效"的效果，不需要额外的表）。

## 密码与锁定策略细节
- bcrypt cost=12是在安全性和服务器CPU负担之间的平衡选择，V1用户量级
  很小（几个Operator账号），登录频率低，cost=12的计算开销可以接受
- login_fail_count的重置时机：**登录成功时清零**，而不是"锁定时间
  过后自动清零"——这意味着即使locked_until已经过期，只要用户还是
  输错密码，fail_count会继续从上次的值累加而非从0开始，这是更严格
  的防暴力策略
- locked_until过期后，下一次登录尝试（无论成功失败）都应该重新
  评估锁定状态，不应该出现"locked_until已过期但账号仍显示锁定"的
  状态残留

## Refresh Token刷新时的安全考虑
每次成功刷新，旧的refresh_token应立即失效（生成新token替换旧哈希），
这样即使旧token泄露，攻击者也只能使用一次。前端需要处理"refresh
token本身也过期或已被使用过"的情况——此时应强制跳转回登录页，不能
静默失败导致用户卡在白屏。

## 首次启动的幂等边界
"首次启动自动创建Owner账号"这个逻辑的判断条件是`user_account`表
为空，而不是"配置文件里没有admin密码"之类的判断——这样即使服务
反复重启，只要曾经创建过账号，就不会重复创建，也不会重新生成
随机密码覆盖运营者已经设置的密码。

## 初始化向导与种子数据的调用关系
001已经把日语模板/K12模板的种子SQL文件准备好，但**没有默认执行**。
本change的`POST /init/apply-template`接口，其实现本质上就是读取
对应的种子SQL文件内容并在一个事务内执行。这里有个需要注意的边界：
如果调用时course_domain已经非空（比如误操作重复调用），不应该
报错，而应该是一个空操作（no-op）并返回成功——因为001的种子数据
本身写成了幂等的INSERT OR IGNORE模式，重复执行不会产生副作用，
这个设计上的一致性需要在本change的实现中保持。
