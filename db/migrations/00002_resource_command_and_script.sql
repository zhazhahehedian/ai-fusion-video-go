-- +goose Up
-- 阶段3：项目版本号 + 剧本/分集 + 资源命令与修订（对齐 V1.1.3.0.1 子集；不含 event/feed/canvas）

ALTER TABLE `afv_project`
  ADD COLUMN `version` INT NOT NULL DEFAULT 0 COMMENT '业务内容版本';

CREATE TABLE `afv_script` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_id` bigint NOT NULL COMMENT '所属项目ID',
  `title` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '剧本标题',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '剧本正文内容（格式化后）',
  `raw_content` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '剧本原始内容（用户粘贴的原文）',
  `total_episodes` int DEFAULT '0' COMMENT '总集数',
  `story_synopsis` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '故事梗概',
  `characters_json` json DEFAULT NULL COMMENT '角色列表JSON',
  `source_type` int DEFAULT '0' COMMENT '来源类型：0-手动创建 1-文件导入 2-AI生成',
  `parsing_status` int DEFAULT '0' COMMENT '解析状态：0-未解析 1-解析中 2-解析完成 3-解析失败',
  `parsing_progress` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '解析进度描述',
  `summary` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'AI生成的剧本摘要',
  `genre` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '剧本类型/题材',
  `target_audience` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '目标受众',
  `duration_estimate` int DEFAULT NULL COMMENT '预估总时长（分钟）',
  `scope` int DEFAULT '3' COMMENT '可见范围：1-公开 2-私有 3-仅团队可见',
  `owner_type` int DEFAULT NULL COMMENT '拥有者类型：1-个人 2-团队',
  `owner_id` bigint DEFAULT NULL COMMENT '拥有者ID',
  `ai_generated` tinyint DEFAULT '0' COMMENT '是否由AI生成',
  `version` int DEFAULT '0' COMMENT '乐观锁版本号',
  `status` int DEFAULT '0' COMMENT '状态：0-草稿 1-正常',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_script_project` (`project_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='剧本表';

CREATE TABLE `afv_script_episode` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `script_id` bigint NOT NULL COMMENT '所属剧本ID',
  `episode_number` int DEFAULT NULL COMMENT '集号（从1开始）',
  `title` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '本集标题',
  `synopsis` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '本集剧情梗概',
  `raw_content` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '本集原始剧本内容',
  `duration_estimate` int DEFAULT NULL COMMENT '预估时长（分钟）',
  `total_scenes` int DEFAULT '0' COMMENT '本集总场次数',
  `source_type` int DEFAULT '0' COMMENT '来源类型：0-AI解析 1-手动添加',
  `sort_order` int DEFAULT '0' COMMENT '排列顺序',
  `parsing_status` int DEFAULT '0' COMMENT '解析状态：0-未解析 1-解析中 2-解析完成 3-解析失败',
  `status` int DEFAULT '0' COMMENT '状态：0-草稿 1-正常',
  `version` int DEFAULT '0' COMMENT '乐观锁版本号',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_episode_script` (`script_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='分集剧本表';

CREATE TABLE `afv_resource_command` (
  `id` char(36) CHARACTER SET ascii COLLATE ascii_bin NOT NULL COMMENT '业务命令标识',
  `actor_user_id` bigint NOT NULL COMMENT '经过认证的操作用户编号',
  `request_id` varchar(128) CHARACTER SET ascii COLLATE ascii_bin NOT NULL COMMENT '用户作用域内的幂等标识',
  `request_hash` char(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL COMMENT '命令内容与执行作用域摘要',
  `origin` varchar(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL COMMENT '命令来源',
  `canvas_id` char(36) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL COMMENT '发起命令的画布标识',
  `agent_run_id` varchar(64) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL COMMENT '可信智能体运行标识',
  `tool_call_id` varchar(128) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL COMMENT '可信工具调用标识',
  `request_json` json NOT NULL COMMENT '已验证的业务命令内容',
  `result_json` json DEFAULT NULL COMMENT '已提交的幂等回执',
  `undo_json` json DEFAULT NULL COMMENT '授权恢复所需的变更前值',
  `create_time` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '命令创建时间',
  `committed_at` datetime(6) DEFAULT NULL COMMENT '命令完成时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_resource_command_request` (`actor_user_id`,`request_id`),
  KEY `idx_resource_command_agent` (`agent_run_id`,`tool_call_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='跨页面与智能体共用的业务命令审计';

CREATE TABLE `afv_resource_revision` (
  `command_id` char(36) CHARACTER SET ascii COLLATE ascii_bin NOT NULL COMMENT '所属已提交业务命令',
  `resource_type` varchar(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL COMMENT '业务资源类型',
  `resource_id` bigint NOT NULL COMMENT '业务对象编号',
  `project_id` bigint DEFAULT NULL COMMENT '所属项目编号',
  `resource_version` int NOT NULL COMMENT '本次命令提交后的对象版本',
  `before_json` json DEFAULT NULL COMMENT '修改前的完整业务内容，新增对象为空',
  `after_json` json NOT NULL COMMENT '修改后的完整业务内容，包含删除状态',
  `changed_fields` json NOT NULL COMMENT '本次实际修改的业务字段',
  `create_time` datetime(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '版本记录时间',
  PRIMARY KEY (`command_id`,`resource_type`,`resource_id`),
  KEY `idx_resource_revision_object` (`resource_type`,`resource_id`,`resource_version`),
  KEY `idx_resource_revision_project` (`project_id`,`create_time`,`command_id`),
  CONSTRAINT `fk_resource_revision_command` FOREIGN KEY (`command_id`) REFERENCES `afv_resource_command` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='业务内容前后版本与条件恢复依据';

-- +goose Down
DROP TABLE IF EXISTS `afv_resource_revision`;
DROP TABLE IF EXISTS `afv_resource_command`;
DROP TABLE IF EXISTS `afv_script_episode`;
DROP TABLE IF EXISTS `afv_script`;
ALTER TABLE `afv_project` DROP COLUMN `version`;
