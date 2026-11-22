# 时间

## 要解决什么

作业的租约和退避必须跟数据库 UTC 比较。内容表要和 Java 共用时，列默认值仍是 `CURRENT_TIMESTAMP`。驱动参数 `loc=UTC` 会把读到的 DATETIME 标成 UTC。如果 MySQL 会话其实是东八区，内容时间的墙钟是东八区，标签却是 UTC。

## 在哪里

- 作业写入：`db/queries/generation.sql`、`media.sql`、`task.sql` 使用 `UTC_TIMESTAMP`
- 内容表：剧本、分镜、资产的 SQL 仍写 `CURRENT_TIMESTAMP`
- 连接：`internal/platform/mysql.go` 的 `WithUTCTimeZone`

`OpenMySQL` 在 DSN 没有 `time_zone` 时追加 `time_zone='+00:00'`。这样内容表的 `CURRENT_TIMESTAMP` 写入 UTC 墙钟，`loc=UTC` 的标签和值一致。调用方已经指定 `time_zone` 时不改写。集成测试故意使用 `+08:00`，确认作业列不跟随会话。

Java 进程的会话时区由部署决定，常常是东八区。共用库上，Java 写入的内容时间和 Go 写入的内容时间可能差 8 小时。这是标明了的差异，不用在读路径上猜时区。

作业列没有 `DEFAULT CURRENT_TIMESTAMP`。比较和心跳都用 `UTC_TIMESTAMP`，不依赖会话。
