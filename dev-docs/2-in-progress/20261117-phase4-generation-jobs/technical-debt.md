# 技术债

| 项 | 原因 | 后续 |
|---|---|---|
| stub 供应商非真实 API | 先练队列与租约 | 接 OpenAI-compatible image strategy |
| FFMPEG_VERSION 仅探测版本 | 先练 ProcessRunner 取消/杀进程 | ComposeVideo 多镜头 + BlobStore |
| 无 generation/media event 流 | 本阶段只验证 job 状态机 | 游标续读 API |
| 额度不读 ai_model | 无模型表 | hasCapacity 对齐 Java |
| submission.kind 用 ECHO | 阶段5已建 image_task/video_task，但受理仍不写这两张表 | 把 IMAGE/VIDEO 受理接到任务行，仍不接真实供应商 |
| Docker 冒烟未跑 | 环境暂无 Docker | migrate + smoke 脚本 |
| 成功 `msg` 仍为 `ok` | 同前序阶段 | 可选对齐 `success` |
