# 作业队列

## 要解决什么

生成和媒体任务可以被多个进程领取，崩溃后能再领，但不能两个进程同时把同一次提交送给供应商。

## 在哪里

- 领取：`internal/generation/repo.go` 的 `Claim`
- 执行前判断：`internal/generation/worker.go` 的 `gateExecution`
- 身份：`Repository.ExecutionPermitted`

领取使用 `FOR UPDATE SKIP LOCKED`，写入 `owner_instance_id` 和 `owner_epoch`。心跳必须仍持有未过期租约。丢失租约的写入返回 `LEASE_LOST`。

## 截止时间

`gateExecution` 在阶段不是 `INGESTING` 且数据库 UTC 时间晚于 `deadline_at` 时停止任务。准备阶段记为 `FAILED`，文案是「准备超过等待期限，未提交供应商」。已经提交过的阶段记为 `UNKNOWN_SUBMISSION`，文案是「远端结果未知，已结束本地等待，原请求不会自动重发」。

`INGESTING` 表示供应商结果已经回到本地，正在入库。这时再按截止时间判失败，会把已经拿到的结果丢掉。Java `RemoteGenerationWorker` 同样排除这个阶段。

## 执行身份

执行前读取受理记录的 `context_json`：

- 有 `agentRunId`：本仓库没有 Agent 运行表，无法证明那次运行仍有效，任务停止。
- 有 `projectId`：发起人必须仍是该个人项目的拥有者，否则停止。
- 两者都没有：继续。当前 echo 受理就是这种。

准备阶段的失效沿用「已取消排队，尚未提交供应商」。更后的阶段使用 `EXECUTION_NOT_PERMITTED`。Java 对后一种情况是 `requestStop`，再靠供应商轮询收尾。这里没有真实轮询，所以直接停止，避免 stub 把失效任务标成成功。

## 额度

同一用户同时持有的远程作业许可仍是常量 `MaxUserRemoteJobs = 8`。Java 按 `ai_model` 的容量判断。模型表和真实供应商都不在这次范围内，所以没有做「读不到就按 8」的开关。echo 受理的模型快照也没有容量字段。

`INGESTING` 不占新的许可，这和 Java `needsPermit` 一致。
