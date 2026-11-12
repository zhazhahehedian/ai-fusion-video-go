# ai-fusion-video-go

融光后端 **Java → Go** 对照实现。与原 Java 仓共用同一 MySQL（strangler），接口契约对齐。

- GitHub: https://github.com/zhazhahehedian/ai-fusion-video-go
- 协作规则: [AGENTS.md](AGENTS.md)

## 技术栈

| 层 | 选择 |
|---|---|
| HTTP | chi |
| SQL | sqlc |
| 迁移 | goose |
| 数据库 | MySQL 8 |
| 任务队列 | MySQL 租约队列 |
| Redis | go-redis/v9 |

## 结构

```text
cmd/server          # 进程入口
internal/           # 业务与基础设施
db/migrations       # goose
db/queries          # sqlc
dev-docs/           # 需求跟踪
docker-compose.yml  # MySQL + Redis（43306/46379）
```

## 快速开始

```bash
cp .env.example .env
make docker-up
make run
curl -s http://127.0.0.1:18080/healthz
```
