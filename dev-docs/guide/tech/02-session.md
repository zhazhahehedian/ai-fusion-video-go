# 会话

## 要解决什么

浏览器已经按 Java 的 access / refresh 协议登录。Go 要读写同一套 Redis 键，并在每个请求里把用户放进 `context.Context`。

## 在哪里

- 签发与校验：`internal/auth/service.go`、`internal/auth/token.go`
- 中间件：`internal/httpapi/middleware_auth.go`

键和 TTL：

- `fusion:token:{access}`，2 小时
- `fusion:refresh_token:{refresh}`，7 天
- 两个方向的指针。刷新使用 `GETDEL`，旧 refresh 只能用一次

`RequireAuth` 接受 `Authorization: Bearer` 或查询参数 `access_token`。通过后检查 `sys_user` 仍启用，且用户名与会话一致。下游只用 `auth.SessionFrom`。

## 还没对齐的注册门禁

Java `UserService.register` 在写用户之前还检查系统是否已初始化、是否开放注册，然后赋默认角色并加入唯一团队。本仓库没有 `afv_system_config`、角色表和团队表。注册目前只校验用户名、密码和昵称，创建启用用户后直接签发会话。这是整块会员域，没有半套开关。

`currentTeamId` 仍是空，`user-info.roles` 仍是空切片。原因相同。

## 和 Java 的对应

已对齐的是 token 键、TTL、一次性刷新和「用户被禁用后旧 token 失效」。未对齐的是注册开关、角色和团队。业务含义见 [用户、团队与权限](../business/01-user-team-permission.md)。
