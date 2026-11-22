# 融光 Go 后端技术讲解

阶段目录里的这份讲解保留作对照。长期维护的技术讲解和业务逻辑讲解在 `dev-docs/guide/`。

这份讲解只描述仓库里已经落地的代码。阶段2到阶段4的进程、鉴权和队列仍在运行；阶段5把分镜、资产、图片/视频任务补进同一套数据层。

| 文档 | 内容 |
|---|---|
| [01-架构与进程](01-architecture.md) | chi、sqlc、errgroup 里的 HTTP / outbox / worker |
| [02-鉴权与 Redis 会话](02-auth-session.md) | 与 Java TokenService 对齐的 access / refresh |
| [03-资源命令管道](03-resource-command.md) | 幂等、请求哈希、乐观锁、revision |
| [04-租约作业队列](04-lease-queue.md) | `SKIP LOCKED`、`owner_epoch`、心跳、额度、状态机 |
| [05-发件箱、退避与 UTC](05-outbox-and-utc.md) | outbox 与为什么作业时间用数据库 UTC |
| [06-媒体进程、日志与错误](06-media-logging-errors.md) | ProcessRunner、slog、`request_id`、`{code,msg,data}` |
| [07-分镜资产与任务](07-storyboard-asset-task.md) | 本次新增的表、接口和取舍 |

阅读顺序可以按编号。每一节都按同一组问题写：要解决什么、代码在哪、和 Java 哪里一样、哪里不一样、容易踩的坑。
