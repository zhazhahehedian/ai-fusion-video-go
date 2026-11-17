# 技术债

| 项 | 原因 | 后续 |
|---|---|---|
| 无 resource_event / feed | 本阶段只验证命令幂等与剧本写入 | 接 `/api/resource-events` 游标轮询 |
| 无 storyboard / member | 学习路线以剧本命令为主 | INITIALIZE_WORKSPACE + storyboard |
| 无 `MANAGE_SCRIPT_SCENES` | 单条 scene CRUD 已够练习管道 | 对齐批量场次命令 |
| 无 agent / canvas 上下文校验 | 无对应表与 fence | 对齐 `ResourceAgentFence` |
| ACL 仍仅个人 owner | 无 member 表 | 移植 `CanvasAccessPolicy` |
| 成功 `msg` 仍为 `ok` | 同阶段2 | 可选对齐 `success` |
