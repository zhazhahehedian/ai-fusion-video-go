# 技术债

长期说明见 `dev-docs/guide/`。下表只保留还没做的项。

| 项 | 原因 |
|---|---|
| 不提供图片/视频 `submit` | 提交会创建受理并调用供应商。任务行的增删改查已经单独存在 |
| `image_item` / `video_item` 无独立增改，也还不会在生成成功后插入 | 条目应由生成 worker 写入并维护 `success_count`。现在没有真实完成写回 |
| 无合成视频、配乐回写、`statistics`、`children`、`batch` | 依赖媒体作业或另一套批量命令 |
| 无资产元数据字典和资产批量接口 | 字典是产品字段目录，不是表；批量命令会扩大权限面 |
| `model_id`、`workflow_version_id` 不校验存在性 | 本库没有 `afv_ai_model` 和 ComfyUI 工作流表 |
| 图片/视频任务没有 `version` 列 | Java 没加。乐观锁用修订表的最大 `resource_version` |
| 权限只认个人项目拥有者 | 没有团队和项目成员表。见业务讲解 01 |
| 每人远程作业额度固定为 8 | 不读 `ai_model`。模型表和供应商容量同属未做模块 |
| 带 `agentRunId` 的作业直接停止 | 没有 Agent 运行表，无法证明运行仍有效。不做成永远允许 |

已经补上、不再记为债：分镜和分镜集 `clearContent`、剧本原文替换时逐行 revision、成功 `msg=success`、`INGESTING` 不按截止时间失败、执行前检查项目拥有者和智能体运行、HTTP 超时与请求体上限、`changedFields` 稳定排序、进程输出上限、内容表会话时区与 `loc=UTC` 对齐、CI 里的 MySQL 集成测试。
