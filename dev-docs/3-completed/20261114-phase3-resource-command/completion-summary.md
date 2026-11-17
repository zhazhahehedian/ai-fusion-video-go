# 阶段3完成摘要：资源命令日志（剧本域）

## 交付

| 能力 | 落点 |
|---|---|
| 迁移 `00002`/`00003` | version + script/episode/command/revision + scene |
| 命令管道 admit → Authorize → Apply → commit | `internal/resource/service.go` |
| `(type, op)` Registry | `handler.go`；`main` 装配 |
| PROJECT/CREATE；SCRIPT PATCH/SOURCE；EPISODE/SCENE CRUD | `internal/resource/handlers` |
| HTTP 对齐 ScriptController 主路径 | `script_handler.go` + project create |
| 幂等 `(actor, request_id)` + hash 冲突 | `command.sql` + 冒烟 |

讲解：`notes/phase3/资源命令管道讲解.md`。

## 验证

- `go test ./...` / `go build` 通过
- `.local/verification/2026-11-14-phase3/smoke-full.sh`：项目幂等、分集/场次 CRUD、source replace 清空树

## 刻意未做（移交后续）

- resource_event / feed / canvas 投递
- storyboard / member / INITIALIZE_WORKSPACE
- `MANAGE_SCRIPT_SCENES` 批量
- 团队 ACL、agent fence
