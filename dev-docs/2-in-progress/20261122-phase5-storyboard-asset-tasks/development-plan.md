# 阶段5：分镜、资产与图片/视频任务数据层

对齐 Java 基线表，把资源命令管道扩到这些对象的增改删，并提供 `deleted = 0` 的读接口。不接 LLM，不接真实供应商。

## 范围

| 能力 | 行为 |
|---|---|
| 迁移 `00006` | storyboard 四张表、asset / asset_item、image_task / image_item、video_task / video_item |
| 命令 | 分镜、资产的增改删走既有幂等、乐观锁、revision。删除父对象时同事务级联软删子行，并为每个子行写 revision |
| 任务删除 | 不要求 `version` / `requestId`；状态仅 2、3、4、6 可删 |
| 读 | 权限检查后返回 camelCase JSON，分页 `{list,total}` |
| 时钟 | 图片/视频任务及条目的写入使用 `UTC_TIMESTAMP()`；分镜和资产保持 `CURRENT_TIMESTAMP` |

## 推迟

见 `technical-debt.md`。合成视频、批量命令、资产元数据字典、模型表和供应商受理不在本阶段。
