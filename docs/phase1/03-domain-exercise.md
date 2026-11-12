# 03 · 领域练习（边学边敲）

目标：不看代码也能讲清「一个项目里剧本/分镜/资产/媒体怎么挂」。答案写在 `notes/phase1/`，提交进仓方便回顾。

## 练习 A · 画出核心树

在 `notes/phase1/domain-tree.md` 用缩进或 mermaid 写出：

```text
sys_user / afv_team
afv_project
  afv_script → episode → scene_item
  afv_storyboard → episode → scene → item(镜头)
  afv_asset → asset_item
  afv_canvas → node / edge
  afv_media_resource ← afv_media_binding
```

补全字段级直觉（不必背全列）：

- `afv_project.owner_type` / `owner_id` / `scope`
- 逻辑删除：`deleted` + `deleted_id`
- 乐观锁：`version`
- 无租户拦截器：权限在 `ResourceAccessPolicy` 一类策略里

## 练习 B · UI 步骤 × 表写入

跑 [01-run-java.md](01-run-java.md) 时，每完成一步在 `notes/phase1/table-writes.md` 记：

| 步骤 | 你做了什么 | 观察到的表（INSERT/UPDATE） | 备注 |
|---|---|---|---|
| 1 | | | |
| … | | | |

## 练习 C · 一句话架构

用自己的话改写（写到 `notes/phase1/architecture.md`）：

> MySQL 是真相源；Redis 做 token/缓存/锁/唤醒；Agent 事件先落库再 SSE；生成进度多靠 REST 轮询。

对照 [学习路线.md](../../学习路线.md) 「0. 一句话架构」自我批改。

## 勾选

- [ ] domain-tree.md
- [ ] table-writes.md 至少 3 行真实观察
- [ ] architecture.md
- [ ] （加分）对照 `CanvasFilmJourneyIT` 标出与 UI 步骤的对应关系
