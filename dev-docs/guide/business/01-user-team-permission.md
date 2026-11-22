# 用户、团队与权限

## 已实现

用户名 3 到 32 个字符，密码 6 到 32 个字符，两次密码必须一致。昵称为空时用用户名。用户名唯一。密码只存 bcrypt。

登录时用户必须 `status = 1`。密码错误和用户不存在都返回「用户名或密码错误」，不暴露账号是否存在。登录和注册成功后签发 access 与 refresh。

读项目、剧本、分镜、资产时，要求 `afv_project.owner_type = 1` 且 `owner_id` 是当前用户。不是拥有者返回 403「无权访问该项目」。实现集中在 `requireOwnedProject` 和各查询的 owner 条件。

生成作业如果带了项目，执行前再查一次这个拥有者关系。项目被删或已经不属于发起人时，任务停止。

## 为什么先只做个人拥有者

Java 的可见性在 `CanvasAccessPolicy`：个人拥有者、团队成员、项目成员都能按角色读写。Go 库里没有团队表和项目成员表。只判断个人拥有者是一条完整规则，不会在缺表时把别人的项目放进来。

## 尚未实现

| 规则 | Java | 为什么这次不做 |
|---|---|---|
| 系统已初始化、开放注册才允许注册 | `UserService.register` | 需要系统配置表，不是单独一个 if |
| 注册后赋默认角色、加入唯一团队 | 同上 | 需要角色表和团队成员表 |
| 团队上下文 `currentTeamId` | `TeamService.getCurrentTeamIdByUser` | 同上 |
| 成员可读可写 | `CanvasAccessMapper.readableProjects` | 成员表和角色一起才有意义 |

这些仍记在各阶段 `technical-debt.md`。
