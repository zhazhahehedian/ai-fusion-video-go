# goose migrations

阶段1 **不**把 Java Flyway 基线整份拷进本仓（体积大、双源易漂）。

阶段2 起推荐两种其一（选一种贯彻到底）：

1. **软链 / 同步脚本**：从 `JAVA_ROOT/ai-fusion-video/src/main/resources/db/migration/` 同步 `.sql`，goose 使用 `goose mysql ... up`（必要时把 Flyway 文件名改成 goose 可识别的 `NNNNNN_name.sql`）。
2. **最小子集拷贝**：只拷阶段2需要的表（`sys_user`、token 无关表结构、`afv_project` 等），在本目录维护增量。

约定：Go 与 Java **共用同一个 MySQL 库** `ai_fusion_video`，做到随时对照切换。

本阶段本目录可为空；`sqlc.yaml` 的 `schema: migrations` 在有 SQL 文件后再跑 `make sqlc`。
