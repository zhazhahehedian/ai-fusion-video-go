# ai-fusion-video-go

融光后端 **Java → Go** 对照学习实现。与原 Java 仓共用同一 MySQL（strangler），接口契约对齐，便于随时切换对照。

- GitHub: https://github.com/zhazhahehedian/ai-fusion-video-go
- 协作规则: [AGENTS.md](AGENTS.md)
- 学习路线: [学习路线.md](学习路线.md)
- Java 对照仓: `E:/workspace/JProject/ai-fusion-video`（仅阅读对照）

## 选型

| 层 | 选择 |
|---|---|
| HTTP | chi |
| SQL | sqlc |
| 迁移 | goose |
| Redis | go-redis/v9 |
| 运行时 | WSL2 + Go 1.23+ |

## 结构

```text
cmd/server          # 进程入口
internal/           # 业务与基础设施
db/migrations       # goose
db/queries          # sqlc
dev-docs/           # 计划与交付跟踪
docker-compose.yml  # 可选 MySQL/Redis（43306/46379）
```

## 快速开始

```bash
cd /mnt/e/workspace/GoProject/ai-fusion-video-study
cp .env.example .env
make run
curl -s http://127.0.0.1:18080/healthz
```

下一步按学习路线阶段推进；当前仅保留 `/healthz` 空壳。
