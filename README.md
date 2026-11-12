# ai-fusion-video-study

融光（ai-fusion-video）**Java → Go** 对照学习仓。按 [学习路线.md](学习路线.md) 分阶段实现；与原 Java 仓共用同一 MySQL（strangler）。

- **Origin（私有）**: https://origin.cursor.com/yuigakki-w/ai-fusion-video-study
- **Java 对照根目录**: `JAVA_ROOT=E:/workspace/JProject/ai-fusion-video`

路径前缀：`B/` / `R/` / `W/` 见学习路线文首。

## 选型（已定死）

| 层 | 选择 |
|---|---|
| HTTP | chi |
| SQL | sqlc |
| 迁移 | goose |
| Redis | go-redis/v9（阶段2起） |
| 运行时 | WSL2 + Go 1.23+ |

为什么是 sqlc+goose、为什么选 chi 不选 gin：见对话结论与阶段2文档；一句话——**SQL 与中间件都贴近原项目/标准库，方便对照学习**。

## 仓库结构

```text
cmd/server          # 入口（阶段1: /healthz）
internal/httpapi    # 统一 {code,msg,data}
internal/auth       # 阶段2
internal/project    # 阶段2
db/migrations       # goose（阶段1 仅说明）
db/queries          # sqlc
docs/phase1         # 阶段1清单
notes/phase1        # 你边学边写
docker-compose.yml  # MySQL 43306 + Redis 46379
```

## 第一天怎么开干

1. 读 [docs/phase1/00-wsl-toolchain.md](docs/phase1/00-wsl-toolchain.md)，确认 WSL 里 `go`/`sqlc`/`goose` 可用。
2. `make run` → `curl http://127.0.0.1:18080/healthz` 应返回 `code:0`。
3. 按 [docs/phase1/01-run-java.md](docs/phase1/01-run-java.md) 起原版 Docker，UI 走「故事 → 分镜 → 出图」。
4. 按 [docs/phase1/02-read-list.md](docs/phase1/02-read-list.md) 阅读，摘要写进 `notes/phase1/readings.md`。
5. 做 [docs/phase1/03-domain-exercise.md](docs/phase1/03-domain-exercise.md)。

阶段1 **不写** login / project API；那是阶段2。

## 常用命令（WSL）

```bash
cd /mnt/e/workspace/GoProject/ai-fusion-video-study
cp .env.example .env   # 按需
make docker-up         # 仅中间件；若 Java 栈已占用同端口则跳过
make run
make tidy
```

## 阶段索引

| 阶段 | 内容 | 状态 |
|---|---|---|
| 1 | 跑原版 + 读懂领域 + Go 空壳 | 进行中（本仓已脚手架） |
| 2 | auth / 统一错误 / project 分页（对接现有前端） | 未开始 |
| 3 | ResourceCommand 幂等与修订 | 未开始 |
| 4 | MySQL 队列生成任务 | 未开始 |
| 5 | Agent ReAct + SSE | 未开始 |
| 6 | 画布 DAG（可选） | 未开始 |
