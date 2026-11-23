# 进度

- [x] `POST /api/generation/image/submit` 与 `POST /api/generation/video/submit`
- [x] 任务、条目、受理、远端作业同一事务，幂等冲突回滚任务行
- [x] stub 写回条目与 `success_count`；失败关闭仍在排队的任务
- [x] 领取作业时 `result_json` 为空也能扫描
- [x] 字段与幂等单测；MySQL 8 上的回放、并发、失败关闭、非拥有者项目
- [x] 讲解 `guide/`，并更新 `dev-docs/guide/`
