# 提交与写回

## 要解决什么

Java 的生图、生视频按钮调用的是 `POST /api/generation/image/submit` 和 `POST /api/generation/video/submit`，不是资源命令上的创建接口。提交要同时留下业务任务、排队中的结果条目和可以重试的远端作业。只建任务行，工作者不会知道该写哪几条结果。

## 请求

`scope.userId` 必须等于当前登录用户。Java 还要比对当前团队。本仓库会话没有团队，所以请求里只要带了 `teamId`，就返回 409「当前账号或团队已变化，请回到原提交范围确认任务」。

提示词、数量、参考地址这些字段走 `ApplyImageTaskFields` / `ApplyVideoTaskFields`，和 `POST /api/generation/image` 创建任务是同一套规则。参考地址可以是 JSON 数组，也可以是 Java 提交体里的 JSON 数组字符串。单独一个地址字符串会被拒绝。

`workflowVersionId` 在创建接口里可以先存着。提交会执行，所以这里直接拒绝。没有工作流表时把它标成成功，会把 ComfyUI 请求伪装成已经生成。

有 `projectId` 时，项目必须是当前用户的个人项目。别人的项目返回 403，并且不留下任务行。

## 一次事务里写什么

`internal/generation/accept.go` 的顺序是：

1. 插入任务，状态 0，公开标识是带连字符的 UUID。
2. 按 `count` 插入条目，状态 0。
3. 插入受理。`kind` 是 `IMAGE` 或 `VIDEO`，`task_id` 是业务表的数字主键。
4. 用 `scope_key + request_id` 锁住受理。如果这个标识已经属于另一条内容，返回 409，事务回滚，刚插入的任务和条目一起消失。
5. 如果是同一次内容的重复提交，同样回滚新行，响应里返回原来的公开标识。
6. 插入一条远端作业和 `GENERATION_ACCEPTED` 发件箱事件。

作业输入里的 `_resultItemIds` 是这批条目的主键。宽高和时长也冻结在输入里，稍后的人工修改不会改掉这次写回。

## 工作者写回

`internal/generation/linked.go`：

| 时机 | 任务 | 条目 |
|---|---|---|
| 提交刚完成 | 0 排队 | 0 生成中 |
| stub 已向“供应商”登记 | 1 处理中 | 仍是 0 |
| stub 成功 | 2 完成，`success_count` 等于条目数 | 1 成功，地址为 `stub://image/{公开标识}/{条目}` 或 `stub://video/...` |
| 截止时间或身份失效，且还在排队或处理中 | 3 失败 | 仍为 0 的条目改为 2，并写入原因 |

`stub://` 只说明写回发生了。它不是上传结果，也不能当图片地址去拉取。

echo 作业不走这张表。`result_json` 在领取前是空的，查询时用 `COALESCE` 变成 JSON `null`，否则 `json.RawMessage` 扫不到 NULL，工作者会在领到作业之前失败。

## 和 Java 的差别

Java `GenerationSubmissionService` 会冻结模型、校验供应商协议、按协议批量拆作业，并在成功后写入真实媒体地址。这里没有模型表和供应商，所以一次提交只有一个作业，完成地址是 stub。数量上限、提示词长度、团队和水印默认值的差别记在 `technical-debt.md`。
