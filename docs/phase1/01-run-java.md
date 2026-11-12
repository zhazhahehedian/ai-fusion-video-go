# 01 · 跑通原版 Java（故事 → 分镜 → 出图）

对照路径：`JAVA_ROOT` = `E:/workspace/JProject/ai-fusion-video`（WSL: `/mnt/e/workspace/JProject/ai-fusion-video`）。

## 前置

- [ ] Docker Desktop 已启动（引擎在跑）
- [ ] 已 clone 上游（commit 基线约 `308cd8f`）

## 启动（推荐：预构建镜像）

在 **PowerShell / CMD**（或 WSL 若已接好 Docker Desktop）：

```powershell
cd E:\workspace\JProject\ai-fusion-video
copy .env.example .env
# 按需改 .env；本地学习默认即可
docker compose up -d
```

浏览器打开：http://localhost:8080

> 默认 `docker-compose.yml` **不**把 MySQL/Redis 端口暴露到宿主机。阶段1若要一边 UI 一边查表：
>
> - 方案 A：给 `mysql`/`redis` 临时取消 `ports` 注释（与原注释一致：`43306` / `46379`）
> - 方案 B：用原仓 `ai-fusion-video/docker-compose-middleware.yml` 或本学习仓 `docker-compose.yml` 单独起中间件，再按原仓文档连本地后端（阶段2 Go 也走这套端口）

## UI 主线（边走边记 notes）

按 [学习路线.md](../../学习路线.md) 业务主线：

1. 建项目 → 故事/小说生成剧本（`story_to_script`）
2. 剧本解析 / 分集（`script_full_parse` 等）
3. 分集 → 分镜（`script_to_storyboard`）
4. 角色/场景/道具出图（`asset_image_gen`）

模型密钥在管理端配置（全局 `afv_api_config` / `afv_ai_model`）。没有可用模型时，至少把项目/剧本/分镜的**表写入**走通，生成步骤可先看 IT 与 prompt。

## 查库（建议）

```text
Host: 127.0.0.1
Port: 43306
User: root（或 .env 里的账号）
Pass: 123456（默认）
DB:   ai_fusion_video
```

每走一步，对照 `notes/phase1/table-writes.md` 记下插入/更新了哪些表。

## 勾选

- [ ] `docker compose ps` 中 mysql / redis / backend / frontend / nginx 健康
- [ ] 能打开 UI 并登录/注册
- [ ] 至少完成到「分镜」或「出图」一步
- [ ] 能连上 MySQL 看到 `afv_project` 等表有数据
