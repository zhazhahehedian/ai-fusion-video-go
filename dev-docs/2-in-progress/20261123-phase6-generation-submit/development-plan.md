# 阶段6：图片与视频提交接到受理

把 Java `POST /api/generation/image/submit` 与 `POST /api/generation/video/submit` 接到已有的任务表和租约队列。不调用真实供应商，不建模型表。

## 范围

| 能力 | 行为 |
|---|---|
| 提交 | 同一事务写入任务、`count` 条状态为 0 的条目、受理和一条远端作业。返回公开任务标识 |
| 幂等 | 同一用户、同一 `requestId`、同一内容返回原标识。内容不同返回 409，并回滚这次多写的任务 |
| 范围 | `scope.userId` 必须是当前用户。没有团队上下文时，带 `teamId` 返回 409 |
| 项目 | 有 `projectId` 时必须仍是个人项目拥有者 |
| 写回 | stub 成功把条目标成 1，地址为 `stub://`，任务标成 2，`success_count` 等于条目数。截止或身份失效把任务标成 3 |
| 工作流 | 带 `workflowVersionId` 直接拒绝 |

## 不做

真实供应商、模型容量、取消接口、execution 查询、ComfyUI。见 `technical-debt.md`。
