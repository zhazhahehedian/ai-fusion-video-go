# 02 · 阅读清单（阶段1）

路径前缀与 [学习路线.md](../../学习路线.md) 一致：

- `JAVA_ROOT` = `E:/workspace/JProject/ai-fusion-video`
- `B/` = `JAVA_ROOT/ai-fusion-video/src/main/java/com/stonewu/fusion/`
- `R/` = `JAVA_ROOT/ai-fusion-video/src/main/resources/`
- `W/` = `JAVA_ROOT/ai-fusion-video-web/`

## 必读

| # | 材料 | 你要带走的结论 |
|---|---|---|
| 1 | `JAVA_ROOT/AGENTS.md` | 「禁止 fallback」工程哲学；前后端约束边界 |
| 2 | `JAVA_ROOT/docker-compose.yml` + `.env.example` | 进程拓扑；MySQL 真相源、Redis 角色 |
| 3 | `R/db/migration/B1.1.0.0.8__baseline_1.0.0.sql` 中 project / script* / storyboard* / asset* | 领域表骨架、逻辑删除、`version` |
| 4 | `R/prompts/agents/` 任选 3 个 `.md`（建议下方） | 产品在「聊什么、产出什么」 |
| 5 | `.../integration/CanvasFilmJourneyIT.java` | 故事→成片端到端主线，**反复回看** |

建议的 3 个 prompt：

1. `script-story-to-script.system.md`
2. `script-to-storyboard.system.md`
3. `asset-image-generation.system.md`

## 读法

1. 先读 prompt 建立「剧情」，再翻表结构，最后对着 IT 看断言顺序。
2. 在 `notes/phase1/readings.md` 用自己的话写 5～10 行摘要，禁止大段粘贴。

## 勾选

- [ ] AGENTS.md
- [ ] compose + env
- [ ] 基线 SQL 核心表
- [ ] 3 个 agent prompt
- [ ] CanvasFilmJourneyIT 通读一遍（不懂实现细节没关系，先跟步骤）
