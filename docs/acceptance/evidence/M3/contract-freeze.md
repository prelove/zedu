# M3 契约冻结证据

- OpenSpec change：`add-m3-recharge-ledger-evidence`
- 冻结日期：2026-07-19
- 验证：`openspec validate add-m3-recharge-ledger-evidence --strict` 通过。

## 冻结内容

1. 角色：Owner/Operator 可录入、查询、作废充值和管理凭证；仅 Owner 可修改本位币与支付方式。
2. 成功/错误信封：沿用既有 `{code,data}` / `{code,message,requestId}`；权限 `40101/40301`、状态/验证 `42201`、冲突 `40901`、数据库/事务失败 `50002`，不得新增错误码。
3. 财务事实：原始金额/币种/汇率快照不可覆盖；`paymentNo` 是幂等键；确认与作废必须写 payment、ledger、余额、审计（及首次本位币锁定）于同一事务。
4. 文件：每 payment 最多 3 个；仅 jpg/jpeg/png/webp/pdf，最大 5 MiB；认证下载、受控路径、无匿名直链；发布失败必须补偿。
5. 负面范围：没有 payout、teacher ledger、refund、adjust、lesson、attendance、report、notification、backup API/UI。

## 事实来源

- PRD v3.1 §7.1–7.4、§9.6、§10.8、§13.9/13.12、§14.1–14.3（R8/R16/R18/R20）、§15.3、§16.5、§20.1、§23.2、§24.3–24.6。
- OpenSpec proposal/design/specs/tasks（本 change）。
