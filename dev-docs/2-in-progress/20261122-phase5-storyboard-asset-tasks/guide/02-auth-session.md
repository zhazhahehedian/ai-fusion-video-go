# 鉴权与 Redis 会话

## 解决什么问题

浏览器和 Next.js 前端已经按 Java 的登录协议工作：access token、refresh token、一次性刷新。Go 服务要能读同一套 Redis 键，不能另造一套 JWT 字段。

## 怎么实现

会话结构在 `internal/auth/session.go`，字段与 Java `TokenService.TokenSession` 对齐：

```5:22:internal/auth/session.go
type Session struct {
	UserID        int64  `json:"userId"`
	Username      string `json:"username"`
	CurrentTeamID *int64 `json:"currentTeamId"`
}

func WithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

func SessionFrom(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionKey).(Session)
	return s, ok
}
```

中间件 `internal/httpapi/middleware_auth.go` 从 `Authorization: Bearer` 或查询参数 `access_token` 取出 access token，调用 `auth.Service.ResolveAccessToken`。这个方法先 `TokenStore.GetAccessSession`，再读 `sys_user`：用户必须仍然启用，且 id、用户名与会话一致。通过后 `WithSession` 写入 context。下游只用 `SessionFrom`，没有全局“当前用户”变量。

Redis 键和 TTL 在 `internal/auth/token.go`：

- `fusion:token:{access}`，2 小时，值是会话 JSON
- `fusion:refresh_token:{refresh}`，7 天
- access 与 refresh 互相指向：`fusion:token:{access}:refresh`、`fusion:refresh_token:{refresh}:access`

创建时四条键放在一次 pipeline 里。刷新走 `GetDel`：

```95:97:internal/auth/token.go
// Refresh atomically consumes the refresh token (GETDEL) and issues a new pair.
func (t *TokenStore) Refresh(ctx context.Context, refreshToken string) (TokenPair, bool, error) {
	raw, err := t.rdb.GetDel(ctx, refreshTokenPrefix+refreshToken).Result()
```

旧 refresh 被删掉之后才签发新的一对。并发的第二次刷新拿不到键，不能把同一张刷新票用两次。

登录、注册、刷新、登出、`/api/auth/user-info` 在 `internal/httpapi/auth_handler.go`。密码使用 bcrypt，用户行在 `sys_user`。

## 和 Java 的对应与差异

键前缀、TTL、会话 JSON 的 `userId` / `username` / `currentTeamId` 按 Java `TokenService` 保留。刷新同样是一次性消费。

Java 把登录用户放在 `SecurityContext`。Go 放在 `context.Context`。两者都不把用户 ID 写成全局变量。Go 这一阶段还没有把 `currentTeamId` 用进资源权限；读和写只认“项目 `owner_type = 1` 且 `owner_id` 是当前用户”。团队 ACL 仍是阶段3留下的缺口。

## 常见坑

- 刷新不用 `GETDEL` 而用 `GET` + `DEL`，两个请求会都看到旧票。
- 在 handler 里缓存 `*Session` 到包级变量，下一个请求会串用户。
- 把 access token 当成长效密钥。它只有 2 小时，过期后必须走 refresh。
