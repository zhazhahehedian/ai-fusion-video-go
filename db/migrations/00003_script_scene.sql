-- +goose Up
-- 阶段3：剧本分场次表（对齐基线 afv_script_scene_item）

CREATE TABLE `afv_script_scene_item` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `episode_id` bigint NOT NULL COMMENT '所属分集ID',
  `script_id` bigint NOT NULL COMMENT '所属剧本ID',
  `scene_number` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '场次编号（如1-1表示第1集第1场）',
  `scene_heading` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '场景标头（如"内景 客厅 夜"）',
  `location` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '场景地点',
  `time_of_day` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '时间段：日/夜/黄昏/清晨等',
  `int_ext` varchar(16) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '内外景标识：内景/外景/内外景',
  `characters` json DEFAULT NULL COMMENT '出场角色名列表JSON',
  `character_asset_ids` json DEFAULT NULL COMMENT '出场角色资产ID列表JSON',
  `scene_asset_id` bigint DEFAULT NULL COMMENT '场景资产ID',
  `prop_asset_ids` json DEFAULT NULL COMMENT '道具资产ID列表JSON',
  `scene_description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '场景氛围/环境描述',
  `dialogues` json DEFAULT NULL COMMENT '对白/动作元素列表JSON',
  `sort_order` int DEFAULT '0' COMMENT '排列顺序',
  `status` int DEFAULT '0' COMMENT '状态：0-草稿 1-正常',
  `version` int DEFAULT '0' COMMENT '乐观锁版本号',
  `deleted` tinyint NOT NULL DEFAULT '0' COMMENT '逻辑删除标志',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_script_scene_episode` (`episode_id`) USING BTREE,
  KEY `idx_script_scene_script` (`script_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='剧本分场次表';

-- +goose Down
DROP TABLE IF EXISTS `afv_script_scene_item`;
