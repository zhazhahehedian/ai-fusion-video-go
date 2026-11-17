# 进度

## A. 远端生成

- [x] 迁移 `00004` + sqlc + enqueue/outbox/claim/worker/HTTP
- [x] 单测（hash）+ `go test` / `go build`
- [x] 讲解笔记（生成队列）
- [ ] Docker 冒烟（暂缓）

## B. 媒体 / ffmpeg

- [x] 迁移 `00005` `afv_media_job`
- [x] sqlc + enqueue + outbox `MEDIA_JOB_ACCEPTED`
- [x] claim / heartbeat / fencing
- [x] `ProcessRunner`（CommandContext + onTick + Kill）
- [x] media worker（`FFMPEG_VERSION`）
- [x] HTTP 入队/查询 + main 接线
- [x] 单测：取消/onTick 杀进程（不依赖 Docker）
- [x] 讲解笔记补充
