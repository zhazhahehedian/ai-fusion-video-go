-- +goose Up
-- 阶段4续：媒体处理任务表（对齐 V1.1.3.0.3 afv_media_job；不含 preview/comfy/blob 附属表）

CREATE TABLE `afv_media_job` (
  `id` char(36) CHARACTER SET ascii COLLATE ascii_bin NOT NULL COMMENT '媒体任务标识',
  `kind` varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务类型',
  `project_id` bigint DEFAULT NULL COMMENT '项目标识，独立任务可为空',
  `target_id` bigint DEFAULT NULL COMMENT '需要正式写回的业务对象标识',
  `target_version` int DEFAULT NULL COMMENT '等待写回的业务版本',
  `actor_user_id` bigint NOT NULL COMMENT '发起用户',
  `request_id` varchar(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL COMMENT '客户端幂等请求标识',
  `request_hash` char(64) CHARACTER SET ascii COLLATE ascii_general_ci NOT NULL COMMENT '请求与身份摘要',
  `context_json` json NOT NULL COMMENT '经认证的执行上下文',
  `input_json` json NOT NULL COMMENT '固定的输入来源版本与参数',
  `checkpoint_json` json NOT NULL COMMENT '已持久化的阶段结果',
  `status` varchar(32) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'QUEUED' COMMENT '任务状态',
  `phase` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '等待处理' COMMENT '当前处理阶段',
  `error_msg` text COLLATE utf8mb4_unicode_ci COMMENT '终止或待处理原因',
  `owner_instance_id` varchar(128) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '当前执行实例',
  `owner_epoch` bigint NOT NULL DEFAULT '0' COMMENT '执行代数',
  `lease_until` datetime(6) DEFAULT NULL COMMENT '执行租约截止时间',
  `next_run_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '下次执行时间',
  `deadline_at` datetime(6) NOT NULL COMMENT '任务总时限',
  `cancel_requested` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否请求取消',
  `revision` bigint NOT NULL DEFAULT '0' COMMENT '状态版本',
  `created_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '创建时间',
  `updated_at` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_media_job_request` (`actor_user_id`,`request_id`),
  KEY `idx_media_job_schedule` (`status`,`next_run_at`,`lease_until`),
  KEY `idx_media_job_target` (`project_id`,`target_id`,`created_at`),
  KEY `idx_media_job_kind_schedule` (`kind`,`status`,`next_run_at`,`lease_until`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='可恢复媒体处理任务';

-- +goose Down
DROP TABLE IF EXISTS `afv_media_job`;
