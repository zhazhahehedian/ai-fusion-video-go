# 阶段3：资源命令日志（剧本域完整）

对齐学习路线「Go 练习 ②」：`CommandHandler{Authorize, Apply}` + `(resourceType, op)` 注册表；剧本树增删改走命令管道；同一 `request_id` 重放返回同一回执。

## 范围

| 能力 | 行为 |
|---|---|
| 迁移 | `00002` version + script/episode + command/revision；`00003` scene |
| 命令管道 | admit → Authorize → Apply → commit（result + revision） |
| Handler | `PROJECT/CREATE`；`SCRIPT` PATCH / REPLACE_SCRIPT_SOURCE；`SCRIPT_EPISODE` CRUD；`SCRIPT_SCENE` CRUD |
| 读接口 | script by project/id；episodes；episode；scenes；scene |
| 写接口 | 对齐 Java `ScriptController` + 项目 create 经命令 |
| 幂等 | `(actor_user_id, request_id)`；hash 冲突 409；已提交回放 |

## 不做（技术债）

- `afv_resource_event` / feed / canvas 投递
- 团队 ACL、agent fence
- storyboard / INITIALIZE_WORKSPACE 完整容器
- `MANAGE_SCRIPT_SCENES` 批量命令

## 包落点

- `internal/resource` — 命令模型、Registry、Service、ScriptQuery、VO
- `internal/resource/handlers` — Project / Script / Episode / Scene
- `db/queries` — command / script / scene
- `internal/httpapi` — script_handler + project create
