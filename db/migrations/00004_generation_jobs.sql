-- +goose Up
-- 阶段4：生成受理 / 远端 job / attempt / outbox / dispatch 锁（对齐 V1.1.3.0.3 子集；不含 media/comfy/canvas/event 流）

CREATE TABLE `afv_generation_submission` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '受理记录标识',
  `request_id` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '调用方幂等标识',
  `scope_key` char(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '用户团队与入口作用域摘要',
  `request_hash` char(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '规范化请求摘要',
  `kind` varchar(24) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '生成类别',
  `task_id` bigint NOT NULL COMMENT '业务任务数据库标识',
  `public_task_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '业务任务公开标识',
  `actor_user_id` bigint NOT NULL COMMENT '发起用户',
  `actor_team_id` bigint DEFAULT NULL COMMENT '发起团队上下文',
  `project_id` bigint DEFAULT NULL COMMENT '所属项目',
  `canvas_node_run_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '画布节点运行标识',
  `context_json` json NOT NULL COMMENT '经校验的内部执行身份',
  `model_snapshot_json` json NOT NULL COMMENT '模型与能力的不可变快照',
  `api_config_id` bigint NOT NULL COMMENT '供应商配置标识',
  `api_config_hash` char(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '供应商配置一致性摘要',
  `state` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '受理及聚合执行状态',
  `cancel_requested` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否请求停止后续执行',
  `revision` bigint NOT NULL DEFAULT '1' COMMENT '状态版本',
  `created_at` datetime(6) NOT NULL COMMENT '受理时间（UTC）',
  `updated_at` datetime(6) NOT NULL COMMENT '最近状态变化时间（UTC）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_generation_request` (`scope_key`,`request_id`),
  UNIQUE KEY `uk_generation_task` (`kind`,`task_id`),
  KEY `idx_generation_actor` (`actor_user_id`,`created_at`),
  KEY `idx_generation_project_state` (`project_id`,`state`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生成任务原子受理与幂等回执';

CREATE TABLE `afv_remote_generation_job` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '远端执行记录标识',
  `submission_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '所属受理记录',
  `item_id` bigint NOT NULL COMMENT '对应的结果条目标识',
  `item_index` int NOT NULL COMMENT '请求中的稳定结果顺序',
  `model_id` bigint NOT NULL COMMENT '占用额度的模型',
  `actor_user_id` bigint NOT NULL COMMENT '占用额度的用户',
  `project_id` bigint DEFAULT NULL COMMENT '占用额度的项目',
  `state` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '远端执行状态',
  `phase` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '当前持久化执行阶段',
  `provider_task_id` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '供应商受理标识',
  `active_attempt_id` char(36) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '当前显式提交尝试标识',
  `input_json` json NOT NULL COMMENT '固定生成参数与稳定媒体引用',
  `result_json` json DEFAULT NULL COMMENT '供应商结果接收检查点',
  `cancel_requested` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否请求远端取消',
  `cancel_sent` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否已发送远端取消请求',
  `permit_held` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否仍占用远端并发额度',
  `owner_instance_id` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '当前工作者实例',
  `owner_epoch` bigint NOT NULL DEFAULT '0' COMMENT '工作者领取代数',
  `lease_until` datetime(6) DEFAULT NULL COMMENT '租约到期时间（UTC）',
  `next_poll_at` datetime(6) NOT NULL COMMENT '下次执行时间（UTC）',
  `deadline_at` datetime(6) NOT NULL COMMENT '自动等待截止时间（UTC）',
  `poll_attempts` int NOT NULL DEFAULT '0' COMMENT '单次查询累计次数',
  `consecutive_errors` int NOT NULL DEFAULT '0' COMMENT '连续传输错误次数',
  `error_code` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '稳定错误码',
  `error_message` text COLLATE utf8mb4_unicode_ci COMMENT '用户可读执行原因',
  `revision` bigint NOT NULL DEFAULT '1' COMMENT '状态版本',
  `created_at` datetime(6) NOT NULL COMMENT '创建时间（UTC）',
  `updated_at` datetime(6) NOT NULL COMMENT '最近变更时间（UTC）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_remote_generation_item` (`submission_id`,`item_index`),
  KEY `idx_remote_generation_due` (`state`,`next_poll_at`,`lease_until`),
  KEY `idx_remote_generation_model` (`model_id`,`permit_held`),
  KEY `idx_remote_generation_actor` (`actor_user_id`,`permit_held`),
  KEY `idx_remote_generation_project` (`project_id`,`permit_held`),
  CONSTRAINT `fk_remote_generation_submission` FOREIGN KEY (`submission_id`) REFERENCES `afv_generation_submission` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='单次供应商请求与可恢复执行状态';

CREATE TABLE `afv_generation_attempt` (
  `id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '提交尝试标识',
  `job_id` char(36) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '远端执行记录',
  `attempt_no` int NOT NULL COMMENT '经明确决策创建的尝试序号',
  `request_json` json DEFAULT NULL COMMENT '可查询协议的预备请求检查点',
  `provider_task_id` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '不可变供应商受理回执',
  `result_json` json DEFAULT NULL COMMENT '同步供应商的不可变响应结果',
  `resolution` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '未知受理的核实结论',
  `resolution_note` text COLLATE utf8mb4_unicode_ci COMMENT '核实依据与明确重试说明',
  `resolved_by` bigint DEFAULT NULL COMMENT '作出核实决定的用户',
  `created_at` datetime(6) NOT NULL COMMENT '调用前提交意图时间（UTC）',
  `submitted_at` datetime(6) DEFAULT NULL COMMENT '准备完成后、付费调用之前提交的意图时间（UTC）',
  `acknowledged_at` datetime(6) DEFAULT NULL COMMENT '回执入库时间（UTC）',
  `resolved_at` datetime(6) DEFAULT NULL COMMENT '核实决定时间（UTC）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_generation_attempt_no` (`job_id`,`attempt_no`),
  CONSTRAINT `fk_generation_attempt_job` FOREIGN KEY (`job_id`) REFERENCES `afv_remote_generation_job` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外部副作用意图、回执与人工核实审计';

CREATE TABLE `afv_async_outbox` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '投递记录标识',
  `event_key` varchar(191) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '事件幂等标识',
  `event_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '已登记投递类型',
  `aggregate_id` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '关联聚合标识',
  `payload_json` json NOT NULL COMMENT '不含凭据的投递内容',
  `state` varchar(16) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'PENDING' COMMENT '待投递或已投递',
  `attempts` int NOT NULL DEFAULT '0' COMMENT '投递尝试次数',
  `next_attempt_at` datetime(6) NOT NULL COMMENT '下次投递时间（UTC）',
  `last_error` text COLLATE utf8mb4_unicode_ci COMMENT '最近投递错误',
  `created_at` datetime(6) NOT NULL COMMENT '与业务状态一起提交的创建时间（UTC）',
  `delivered_at` datetime(6) DEFAULT NULL COMMENT '投递确认时间（UTC）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_async_outbox_event` (`event_key`),
  KEY `idx_async_outbox_due` (`state`,`next_attempt_at`,`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='业务任务可靠投递发件箱';

CREATE TABLE `afv_generation_dispatch_lock` (
  `id` int NOT NULL COMMENT '全局额度检查锁标识',
  `version` bigint NOT NULL DEFAULT '0' COMMENT '调度版本',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='跨实例生成额度与事件顺序的短事务锁';

INSERT INTO `afv_generation_dispatch_lock` (`id`, `version`) VALUES (1, 0);

-- +goose Down
DELETE FROM `afv_generation_dispatch_lock` WHERE `id` = 1;
DROP TABLE IF EXISTS `afv_generation_attempt`;
DROP TABLE IF EXISTS `afv_remote_generation_job`;
DROP TABLE IF EXISTS `afv_async_outbox`;
DROP TABLE IF EXISTS `afv_generation_submission`;
DROP TABLE IF EXISTS `afv_generation_dispatch_lock`;
