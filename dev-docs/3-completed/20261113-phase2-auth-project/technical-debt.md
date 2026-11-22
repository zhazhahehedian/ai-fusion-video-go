# 技术债

2026-11-22 之后：成功 `msg` 已改为 `success`。注册门禁、角色、团队和成员 ACL 仍未做，因为缺少对应的表，没有单独加开关。



| 项 | 原因 | 后续 |
|---|---|---|
| 项目 ACL 仅个人 owner | 学习库无 `afv_project_member` / `afv_team*` | 移植 `CanvasAccessMapper.readableProjects` |
| `currentTeamId` 恒为 null | 无团队表/查询 | 接 `TeamService.getCurrentTeamIdByUser` |
| `user-info.roles` 空切片 | 无角色表 | 接 `loadUserRoles` |
| register 跳过系统 init / 开放注册开关 / 赋角色 / 入单团队 | 学习库无对应表 | 对齐 `UserService.register` + init |
| create project 只插 `afv_project` | 未接 resource-command；无 member/script/storyboard 表 | 阶段3 幂等命令 + workspace ensure |
| create 校验 `requestId` 但不做幂等 | 无 `afv_resource_command` | 阶段3 按 `request_id` 重放 |
| 响应无 `version` | 学习迁移 `00001` 无该列 | 同步 Java 增量迁移后再写入/返回 |
| 未移植 legacy `fusion:user_token:` 刷新旁路 | 新签发已有双向关联键 | 若要对齐老 token 再补 |
| 成功 `msg` 仍为 `ok` | 前端只看 `code===0` | 已改为 Java 的 `success` |
