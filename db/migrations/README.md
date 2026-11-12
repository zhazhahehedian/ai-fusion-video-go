# goose migrations（MySQL）

与 Java Flyway **共用同一 MySQL 库** `ai_fusion_video`（strangler）。

1. **软链 / 同步脚本**：从原仓 `ai-fusion-video/src/main/resources/db/migration/` 同步 `.sql`，再用 goose 执行（必要时把 Flyway 文件名改成 goose 可识别的 `NNNNNN_name.sql`）。
2. **最小子集拷贝**：只拷当前阶段需要的表，在本目录维护增量。

```bash
goose -dir db/migrations mysql "$DATABASE_URL" status
```
