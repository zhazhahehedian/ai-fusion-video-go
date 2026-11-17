# 阶段2：鉴权横切 + 项目分页

## 目标

对齐学习路线「Go 练习 ①」：现有前端改 rewrite 后可登录并拉项目分页。

| 能力 | 路径 / 行为 |
|---|---|
| 登录 | `POST /api/auth/login` |
| 注册 | `POST /api/auth/register`（学习库跳过 init/开关/角色） |
| 刷新 | `POST /api/auth/refresh`（Redis `GETDEL`） |
| 登出 | `POST /api/auth/logout` |
| 当前用户 | `GET /api/auth/user-info` |
| 鉴权中间件 | Bearer / raw Authorization / `access_token` 查询参数 → context |
| 项目分页 | `GET /api/project/page?pageNo=&pageSize=` |
| 项目列表 | `GET /api/project/list`（前端仪表盘主路径，顺带对齐） |
| 创建项目 | `POST /api/project`（仅插项目行，无 workspace） |

## 契约要点

- 响应 `{code,msg,data}`，业务错误用 `AppError` + `errors.As` 映射；HTTP status 与 code 对齐（如 401）
- Token：无横线 UUID；`fusion:token:` 2h；`fusion:refresh_token:` 7d；双向关联键
- Session JSON：`{userId,username,currentTeamId}`
- 项目可见：本阶段仅 `owner_type=1 AND owner_id=当前用户`（团队/成员 ACL 见技术债）

## 包落点

- `internal/apperr` — 业务错误
- `internal/auth` — TokenService、登录服务、context Session
- `internal/project` — 列表/分页查询
- `internal/httpapi` — handler + RequireAuth 中间件
