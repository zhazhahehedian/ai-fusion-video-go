# 技术债

| 项 | 原因 | 后续 |
|---|---|---|
| 不提供 `POST /api/generation/image/submit` 与 video submit | 提交会创建受理记录并调用供应商 | 下一步把任务行接到租约队列，仍可不接真实 HTTP |
| `image_item` / `video_item` 无独立增改命令 | 条目由生成 worker 写入；本阶段只建表、随任务删除做逻辑删除、提供按任务读取 | worker 成功后插入条目并维护 `success_count` |
| 不级联删除分镜树和资产树 | Java 的 `clearContent` / batch 是单独命令 | 删除父对象前必须先删子对象，否则返回冲突 |
| 无合成视频、配乐回写、`statistics`、`children`、`batch` | 依赖媒体作业或另一套批量命令 | 媒体写回阶段 |
| 无 `GET /api/asset/metadata/{type}` 与 `/all`、`/batch` | 元数据是产品字段字典，不是表结构；分页全库与批量创建会扩大权限面 | 字典与批量命令单独做 |
| `model_id`、`workflow_version_id` 不校验存在性 | 本库还没有 `afv_ai_model` / ComfyUI 工作流表 | 引入模型表后再校验 |
| 图片/视频任务没有 `version` 列 | Java 没给这两张表加版本；加列会让共用库漂移 | 乐观锁用 `afv_resource_revision` 的最大 `resource_version` |
| 权限只认个人项目拥有者 | 与阶段3一致，尚无团队 ACL | 成员表与团队上下文 |
| 任务表没有 `DEFAULT CURRENT_TIMESTAMP` | 避免会话时区；Go 写入显式 `UTC_TIMESTAMP()` | 共用 Flyway 已建库时，以 Java 列默认为准，Go 语句仍显式写 UTC |
