-- +goose Up
-- 阶段5：分镜、资产、生图/生视频任务（对齐 Java 基线 B1.1.0.0.8 及之后落在这些表上的列）。
-- 业务内容表的 create_time/update_time 保持 CURRENT_TIMESTAMP，与剧本表一致。
-- 图片/视频任务及条目是作业记录：不使用会话时区的 CURRENT_TIMESTAMP，由写入语句调用 UTC_TIMESTAMP()。
-- version / compose_job_id 对应 V1.1.3.0.1；bgm_* 对应 V1.1.2.0.0；workflow_version_id 对应 V1.1.1.0.0。

CREATE TABLE `afv_storyboard` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `project_id` bigint DEFAULT NULL COMMENT '所属项目ID',
  `script_id` bigint DEFAULT NULL COMMENT '关联剧本ID',
  `title` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分镜标题',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '分镜描述',
  `custom_columns` json DEFAULT NULL COMMENT '自定义列配置JSON',
  `scope` int DEFAULT '3' COMMENT '可见范围：1-公开 2-私有 3-仅团队可见',
  `owner_type` int DEFAULT NULL COMMENT '拥有者类型：1-个人 2-团队',
  `owner_id` bigint DEFAULT NULL COMMENT '拥有者ID',
  `total_duration` int DEFAULT NULL COMMENT '预估总时长（秒）',
  `status` int DEFAULT '0' COMMENT '状态：0-草稿 1-正常',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int NOT NULL DEFAULT '0' COMMENT '业务内容版本',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_storyboard_project` (`project_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='分镜脚本表';

CREATE TABLE `afv_storyboard_episode` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `storyboard_id` bigint NOT NULL COMMENT '所属分镜ID',
  `script_episode_id` bigint DEFAULT NULL COMMENT '关联的剧本分集ID',
  `episode_number` int DEFAULT NULL COMMENT '集号',
  `title` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '集标题',
  `synopsis` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '本集梗概',
  `sort_order` int DEFAULT '0' COMMENT '排列顺序',
  `status` int DEFAULT '0' COMMENT '状态：0-草稿 1-正常',
  `deleted` tinyint DEFAULT '0' COMMENT '逻辑删除标志',
  `deleted_id` bigint NOT NULL DEFAULT '0' COMMENT '逻辑删除隔离标识，0-未删除，删除后为记录ID',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `composed_video_url` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '本集合成视频URL',
  `compose_status` tinyint NOT NULL DEFAULT '0' COMMENT '合成状态: 0未开始 1合成中 2已完成 3失败',
  `compose_error_msg` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '合成失败原因',
  `composed_at` datetime DEFAULT NULL COMMENT '合成完成时间',
  `bgm_task_id` varchar(64) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Sonilo 配乐任务ID（如需查询该音轨的授权记录，凭此ID向 Sonilo 查询）',
  `bgm_audio_url` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '配乐音轨URL（持久化存档）',
  `bgm_sfx_audio_url` varchar(512) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '音效音轨URL（持久化存档）',
  `bgm_error_msg` varchar(1024) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '配乐失败原因（配乐失败不影响本集合成结果）',
  `version` int NOT NULL DEFAULT '0' COMMENT '业务内容版本',
  `compose_job_id` char(36) CHARACTER SET ascii COLLATE ascii_bin DEFAULT NULL COMMENT '最近一次媒体合成任务标识',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_sb_episode_script_episode` (`storyboard_id`,`script_episode_id`,`deleted_id`) USING BTREE,
  KEY `idx_sb_episode_storyboard` (`storyboard_id`) USING BTREE,
  KEY `idx_sb_episode_script_episode` (`script_episode_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='分镜集表';

CREATE TABLE `afv_storyboard_scene` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `episode_id` bigint NOT NULL COMMENT '所属分镜集ID',
  `storyboard_id` bigint NOT NULL COMMENT '所属分镜ID（冗余）',
  `scene_number` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '场次编号',
  `scene_heading` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '场景标头',
  `location` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '场景地点',
  `time_of_day` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '时间段',
  `int_ext` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '内外景标识',
  `sort_order` int DEFAULT '0' COMMENT '排列顺序',
  `status` int DEFAULT '0' COMMENT '状态：0-草稿 1-正常',
  `deleted` tinyint DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_time` datetime DEFAULT NULL COMMENT '更新时间',
  `version` int NOT NULL DEFAULT '0' COMMENT '业务内容版本',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_sb_scene_episode` (`episode_id`) USING BTREE,
  KEY `idx_sb_scene_storyboard` (`storyboard_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='分镜场次表';

CREATE TABLE `afv_storyboard_item` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `storyboard_id` bigint NOT NULL COMMENT '所属分镜ID',
  `storyboard_episode_id` bigint DEFAULT NULL COMMENT '所属分镜集ID',
  `storyboard_scene_id` bigint DEFAULT NULL COMMENT '所属分镜场次ID',
  `sort_order` int DEFAULT '0' COMMENT '排列顺序',
  `shot_number` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '镜号',
  `auto_shot_number` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '自动编号（系统生成）',
  `image_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '用户上传参考图片URL',
  `reference_image_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '外部参考图片URL',
  `video_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '视频URL（最终成品）',
  `generated_image_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'AI生成的图片URL',
  `first_frame_image_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '首帧参考图片URL',
  `last_frame_image_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '尾帧参考图片URL',
  `first_frame_prompt` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'AI生成首帧时使用的提示词',
  `last_frame_prompt` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'AI生成尾帧时使用的提示词',
  `generated_video_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'AI生成的视频URL',
  `shot_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '景别：远景/全景/中景/近景/特写',
  `duration` decimal(10,2) DEFAULT NULL COMMENT '预估时长（秒）',
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '画面内容描述',
  `scene_expectation` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '画面期望描述（AI生图提示）',
  `sound` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '声音描述',
  `dialogue` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '台词/旁白',
  `sound_effect` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '音效',
  `music` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '配乐建议',
  `camera_movement` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '镜头运动：推/拉/摇/移/跟/升/降',
  `camera_angle` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '镜头角度：平视/俯视/仰视',
  `camera_equipment` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '摄像机装备',
  `focal_length` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '镜头焦段',
  `transition` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '转场效果：切/淡入/淡出/溶/划',
  `character_ids` json DEFAULT NULL COMMENT '出场角色子资产ID列表 JSON (List<Long> of AssetItem.id)',
  `scene_asset_item_id` bigint DEFAULT NULL COMMENT '场景子资产ID (AssetItem.id)',
  `prop_ids` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '道具子资产ID列表 JSON (List<Long> of AssetItem.id)',
  `remark` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `custom_data` json DEFAULT NULL COMMENT '自定义扩展数据JSON',
  `ai_generated` tinyint DEFAULT '0' COMMENT '是否由AI生成',
  `status` int DEFAULT '0' COMMENT '状态：0-草稿 1-正常',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `video_prompt` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'AI生成视频时使用的提示词（保存以便复用和手动调整）',
  `version` int NOT NULL DEFAULT '0' COMMENT '业务内容版本',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_sb_item_storyboard` (`storyboard_id`) USING BTREE,
  KEY `idx_sb_item_scene` (`storyboard_scene_id`) USING BTREE,
  KEY `idx_sb_item_episode` (`storyboard_episode_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='分镜条目表';

CREATE TABLE `afv_asset` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` bigint DEFAULT NULL COMMENT '创建者用户ID',
  `project_id` bigint DEFAULT NULL COMMENT '所属项目ID',
  `type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '资产类型：character/scene/prop/vehicle/building/costume/effect',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '资产名称',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '资产描述',
  `cover_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '封面图URL',
  `properties` json DEFAULT NULL COMMENT '动态属性JSON（如角色的appearance、age等）',
  `tags` json DEFAULT NULL COMMENT '标签列表JSON',
  `source_type` int DEFAULT '1' COMMENT '来源类型：1-用户上传 2-AI生成',
  `ai_prompt` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'AI生成时使用的提示词',
  `owner_type` int DEFAULT NULL COMMENT '拥有者类型：1-个人 2-团队',
  `owner_id` bigint DEFAULT NULL COMMENT '拥有者ID',
  `status` int DEFAULT '1' COMMENT '状态：0-草稿 1-正常',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int NOT NULL DEFAULT '0' COMMENT '业务内容版本',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_asset_project` (`project_id`) USING BTREE,
  KEY `idx_asset_owner` (`owner_type`,`owner_id`) USING BTREE,
  KEY `idx_asset_type` (`project_id`,`type`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='资产表';

CREATE TABLE `afv_asset_item` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `asset_id` bigint NOT NULL COMMENT '所属主资产ID',
  `item_type` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '子资产类型：front/side/back/detail/expression/pose/variant/original',
  `name` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '子资产名称',
  `image_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '图片URL',
  `thumbnail_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '缩略图URL',
  `properties` json DEFAULT NULL COMMENT '动态属性JSON',
  `sort_order` int DEFAULT '0' COMMENT '排列顺序',
  `source_type` int DEFAULT '1' COMMENT '来源类型：1-用户上传 2-AI生成',
  `ai_prompt` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT 'AI生成时使用的提示词',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `version` int NOT NULL DEFAULT '0' COMMENT '业务内容版本',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_asset_item_asset` (`asset_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='子资产表';

CREATE TABLE `afv_image_task` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `task_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务唯一标识',
  `user_id` bigint NOT NULL COMMENT '发起用户ID',
  `project_id` bigint DEFAULT NULL COMMENT '关联项目ID',
  `prompt` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '生图提示词',
  `prompt_template_id` bigint DEFAULT NULL COMMENT '提示词模板ID',
  `ref_image_urls` json DEFAULT NULL COMMENT '参考图片URL列表JSON',
  `ratio` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '画面比例（如16:9）',
  `resolution` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分辨率（如1920x1080）',
  `aspect_ratio` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '宽高比描述',
  `width` int DEFAULT NULL COMMENT '图片宽度（像素）',
  `height` int DEFAULT NULL COMMENT '图片高度（像素）',
  `count` int DEFAULT '1' COMMENT '生成数量',
  `success_count` int DEFAULT '0' COMMENT '已成功生成数量',
  `status` int DEFAULT '0' COMMENT '任务状态：0-排队中 1-处理中 2-已完成 3-失败',
  `error_msg` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '失败错误信息',
  `model_id` bigint DEFAULT NULL COMMENT '使用的AI模型ID',
  `workflow_version_id` bigint DEFAULT NULL COMMENT '固定的 ComfyUI 工作流版本标识',
  `category` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '任务分类标签',
  `owner_type` int DEFAULT NULL COMMENT '拥有者类型：1-个人 2-团队',
  `owner_id` bigint DEFAULT NULL COMMENT '拥有者ID',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL COMMENT '创建时间（UTC）',
  `update_time` datetime NOT NULL COMMENT '更新时间（UTC）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `task_id` (`task_id`) USING BTREE,
  KEY `idx_image_task_workflow_version` (`workflow_version_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='生图任务表';

CREATE TABLE `afv_image_item` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `task_id` bigint NOT NULL COMMENT '所属生图任务ID',
  `platform_task_id` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '平台侧任务ID',
  `image_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '生成的图片URL',
  `thumbnail_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '缩略图URL',
  `width` int DEFAULT NULL COMMENT '图片宽度（像素）',
  `height` int DEFAULT NULL COMMENT '图片高度（像素）',
  `file_size` bigint DEFAULT NULL COMMENT '文件大小（字节）',
  `status` int DEFAULT '0' COMMENT '状态：0-生成中 1-成功 2-失败',
  `error_msg` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '失败错误信息',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL COMMENT '创建时间（UTC）',
  `update_time` datetime NOT NULL COMMENT '更新时间（UTC）',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_image_item_task` (`task_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='生图条目表';

CREATE TABLE `afv_video_task` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `task_id` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '任务唯一标识',
  `user_id` bigint NOT NULL COMMENT '发起用户ID',
  `project_id` bigint DEFAULT NULL COMMENT '关联项目ID',
  `prompt` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '生视频提示词',
  `prompt_template_id` bigint DEFAULT NULL COMMENT '提示词模板ID',
  `generate_mode` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '生成模式：text2video/image2video',
  `first_frame_image_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '首帧参考图片URL',
  `last_frame_image_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '尾帧参考图片URL',
  `reference_image_urls` json DEFAULT NULL COMMENT '参考图片URL列表JSON',
  `reference_video_urls` json DEFAULT NULL COMMENT '参考视频URL列表 JSON',
  `reference_audio_urls` json DEFAULT NULL COMMENT '参考音频URL列表 JSON',
  `ratio` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '画面比例（如16:9）',
  `resolution` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分辨率（如1920x1080）',
  `duration` int DEFAULT NULL COMMENT '视频时长（秒）',
  `watermark` tinyint DEFAULT '0' COMMENT '是否添加水印',
  `generate_audio` tinyint DEFAULT '0' COMMENT '是否生成配音',
  `seed` bigint DEFAULT NULL COMMENT '随机种子（用于复现）',
  `camera_fixed` tinyint DEFAULT '0' COMMENT '是否固定镜头',
  `count` int DEFAULT '1' COMMENT '生成数量',
  `success_count` int DEFAULT '0' COMMENT '已成功生成数量',
  `status` int DEFAULT '0' COMMENT '任务状态：0-排队中 1-处理中 2-已完成 3-失败',
  `error_msg` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '失败错误信息',
  `model_id` bigint DEFAULT NULL COMMENT '使用的AI模型ID',
  `workflow_version_id` bigint DEFAULT NULL COMMENT '固定的 ComfyUI 工作流版本标识',
  `category` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '任务分类标签',
  `owner_type` int DEFAULT NULL COMMENT '拥有者类型：1-个人 2-团队',
  `owner_id` bigint DEFAULT NULL COMMENT '拥有者ID',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL COMMENT '创建时间（UTC）',
  `update_time` datetime NOT NULL COMMENT '更新时间（UTC）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `task_id` (`task_id`) USING BTREE,
  KEY `idx_video_task_workflow_version` (`workflow_version_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='生视频任务表';

CREATE TABLE `afv_video_item` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `task_id` bigint NOT NULL COMMENT '所属生视频任务ID',
  `platform_task_id` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '平台侧任务ID',
  `video_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '生成的视频URL',
  `cover_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '视频封面图URL',
  `duration` int DEFAULT NULL COMMENT '视频时长（秒）',
  `file_size` bigint DEFAULT NULL COMMENT '文件大小（字节）',
  `status` int DEFAULT '0' COMMENT '状态：0-生成中 1-成功 2-失败',
  `error_msg` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '失败错误信息',
  `first_frame_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '视频首帧图片URL',
  `last_frame_url` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '视频尾帧图片URL',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL COMMENT '创建时间（UTC）',
  `update_time` datetime NOT NULL COMMENT '更新时间（UTC）',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_video_item_task` (`task_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='生视频条目表';

-- +goose Down
DROP TABLE IF EXISTS `afv_video_item`;
DROP TABLE IF EXISTS `afv_video_task`;
DROP TABLE IF EXISTS `afv_image_item`;
DROP TABLE IF EXISTS `afv_image_task`;
DROP TABLE IF EXISTS `afv_asset_item`;
DROP TABLE IF EXISTS `afv_asset`;
DROP TABLE IF EXISTS `afv_storyboard_item`;
DROP TABLE IF EXISTS `afv_storyboard_scene`;
DROP TABLE IF EXISTS `afv_storyboard_episode`;
DROP TABLE IF EXISTS `afv_storyboard`;
