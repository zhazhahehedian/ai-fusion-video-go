# 阶段4：异步生成任务与媒体（进行中）

对齐学习路线「Go 练习 ③」：MySQL 租约队列 + `FOR UPDATE SKIP LOCKED` + `owner_epoch` fencing；worker ticker / 信号量 / heartbeat / 优雅停机；`exec.CommandContext` 跑外部进程（ffmpeg）。

## 范围

### A. 远端生成（已完成）

| 能力 | 行为 |
|---|---|
| 迁移 `00004` | submission / remote_job / attempt / async_outbox / dispatch_lock |
| 受理 | 幂等 submission → job(QUEUED) → outbox `GENERATION_ACCEPTED` |
| Outbox / Claim / Worker | READY 领取；stub 供应商 prepare→WAITING→SUCCEEDED |
| HTTP | `POST /api/generation/echo`；`GET /api/generation/submissions/{id}` |

### B. 媒体处理（本切片）

| 能力 | 行为 |
|---|---|
| 迁移 `00005` | `afv_media_job` |
| 受理 | 幂等 media job(QUEUED) → outbox `MEDIA_JOB_ACCEPTED` → WAITING |
| Claim | SKIP LOCKED + epoch；heartbeat / leased fencing |
| ProcessRunner | `CommandContext` + 每秒 onTick（鉴租约）+ 取消 Kill |
| Worker | 单槽 busy；`FFMPEG_VERSION` 跑 `ffmpeg -version` 写入 checkpoint |
| HTTP | `POST /api/media/jobs`；`GET /api/media/jobs/{id}` |
| 停机 | 与 generation 同 `errgroup` |

## 不做（技术债）

- 真实 image/video 策略、ComposeVideo 多镜头拼接、MediaBlobStore
- generation_event / media_job_event 游标
- 按 ai_model 完整 hasCapacity
- Docker 冒烟（暂缓，本地无 Docker 时跳过）

## 包落点

- `internal/generation` — 远端生成队列
- `internal/media` — media job + ProcessRunner + worker
- `db/migrations/00004*`,`00005*`；`db/queries/{generation,media}.sql`
