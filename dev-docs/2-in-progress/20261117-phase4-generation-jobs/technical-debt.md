# 技术债

2026-11-22 之后：成功 `msg` 已改为 `success`；`INGESTING` 不再因截止时间失败；执行前会拒绝带 `agentRunId` 的作业，并复核个人项目拥有者。额度仍不读 `ai_model`。真实供应商、事件流和合成仍未做。



| 项 | 原因 | 后续 |
|---|---|---|
| stub 供应商非真实 API | 先练队列与租约 | 阶段6已把 IMAGE/VIDEO 接到任务行，完成仍是 stub |
| FFMPEG_VERSION 仅探测版本 | 先练 ProcessRunner 取消/杀进程 | ComposeVideo 多镜头 + BlobStore |
| 无 generation/media event 流 | 本阶段只验证 job 状态机 | 游标续读 API |
| 额度不读 ai_model | 无模型表 | hasCapacity 对齐 Java |
| submission.kind 的 IMAGE/VIDEO | 阶段6已写入任务行和条目。供应商仍是 stub | 接真实供应商时替换写回 |
| Docker 冒烟未跑 | 环境暂无 Docker | migrate + smoke 脚本 |
| 成功 `msg` 仍为 `ok` | 同前序阶段 | 已改为 `success` |
