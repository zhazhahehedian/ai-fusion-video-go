# 发件箱、退避与数据库 UTC

## 解决什么问题

业务状态和“可以开始执行”的通知必须一起提交。通知投递失败不能把业务事务滚回去，也不能让失败的事件立刻被再次选中，压住后面的事件。

另外，作业时间如果用 `CURRENT_TIMESTAMP`，它跟随会话时区。一个连接是 `+08:00`，另一个是 UTC，租约和退避会相差 8 小时。

## 怎么实现

受理和作业插入在同一事务里写 `afv_async_outbox`。`internal/generation/outbox.go` 每 500 毫秒拉取一条到期的 `PENDING` 事件。`NextPendingGenerationOutbox` 用 `FOR UPDATE`，并且要求同一 `aggregate_id` 没有更早的未投递行，这样同一聚合按 id 顺序投递。

投递失败时，业务事务回滚。退避写在另一个事务里，成功后再把投递错误返回给循环。循环看到 `deliveryFailure` 会继续后面的事件；如果退避事务自己失败，返回 `backoffFailure`，这一轮停下来，避免在库故障时紧循环。

退避间隔是数据库算的，上限 60 秒：

```237:239:db/queries/generation.sql
UPDATE afv_async_outbox
SET next_attempt_at = TIMESTAMPADD(SECOND, LEAST(60, POW(2, LEAST(attempts + 1, 6))), UTC_TIMESTAMP(6)),
```

`attempts + 1` 先被 `LEAST(..., 6)` 截断，所以指数是 2、4、8、16、32、60，不会因为次数很大而溢出。

作业列上的时间全部用 `UTC_TIMESTAMP(6)` 写入和比较：`created_at`、`lease_until`、`next_poll_at`、`next_attempt_at`。迁移里这些列是 `datetime(6)`，没有 `DEFAULT CURRENT_TIMESTAMP`。集成测试把会话设成 `+08:00`，再比较 `UTC_TIMESTAMP()` 与 `CURRENT_TIMESTAMP()` 相差 8 小时，用来证明写入没有跟会话走。

阶段5的图片/视频任务沿用同一规则，但列类型保持 Java 的 `datetime`（没有小数秒）。插入和更新写 `UTC_TIMESTAMP()`，不设会话默认值。分镜和资产是内容表，和剧本一样保留 `CURRENT_TIMESTAMP`，因为 Java 这些列就是会话时间，共用库时默认值要一致。Go 进程在 `platform.WithUTCTimeZone` 里把没有显式 `time_zone` 的 DSN 补成 `+00:00`，这样 `loc=UTC` 标出的墙钟就是 UTC，而不是把会话本地时间误标成 UTC。集成测试仍把会话设成 `+08:00`，用来证明作业写入不跟会话走。

## 和 Java 的对应与差异

Java 发件箱也是业务事务内插入、后台按到期时间投递。Go 把退避从“投递失败却把整次事务回滚、下次立刻再抢到同一行”改成了独立事务里的指数退避。这是 PR #1 相对最初实现的修正。

Java 业务表普遍 `DEFAULT CURRENT_TIMESTAMP`。Go 只对作业时间改成显式 UTC。如果一张表已经由 Flyway 建成并带有 `ON UPDATE CURRENT_TIMESTAMP`，Go 的 `SET update_time = UTC_TIMESTAMP()` 仍会覆盖这次更新的值；没被 Go 更新语句碰到的列，继续按 Java 的默认值走。

## 常见坑

- 在投递失败的同一个事务里更新 `next_attempt_at`，然后返回错误触发回滚。退避被一起回滚，事件马上又能被选中。
- 用应用服务器的 `time.Now()` 写租约，再拿数据库的 `UTC_TIMESTAMP` 比较。两台机器的时钟差会让租约提前过期或永不超时。
- 把分镜 `create_time` 也改成 UTC。那样和 Java 内容表、以及本仓库剧本表的会话时钟不一致。
