-- name: CreateStoryboard :execlastid
INSERT INTO afv_storyboard (
  project_id, script_id, title, description, custom_columns, scope,
  owner_type, owner_id, status, deleted, version
) VALUES (
  ?, ?, ?, ?, ?, ?,
  1, ?, ?, 0, 0
);

-- name: GetStoryboardByID :one
SELECT * FROM afv_storyboard
WHERE id = ? AND deleted = 0;

-- name: LockStoryboardByID :one
SELECT * FROM afv_storyboard
WHERE id = ? AND deleted = 0
FOR UPDATE;

-- name: GetLatestStoryboardByProject :one
SELECT * FROM afv_storyboard
WHERE project_id = ? AND deleted = 0
ORDER BY id DESC
LIMIT 1;

-- name: UpdateStoryboard :execrows
UPDATE afv_storyboard
SET description = ?,
    custom_columns = ?,
    status = ?,
    version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;

-- name: SoftDeleteStoryboard :execrows
UPDATE afv_storyboard
SET deleted = 1,
    version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;

-- name: BumpStoryboardVersion :execrows
UPDATE afv_storyboard
SET version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;

-- name: CreateStoryboardEpisode :execlastid
INSERT INTO afv_storyboard_episode (
  storyboard_id, script_episode_id, episode_number, title, synopsis, sort_order, status,
  deleted, deleted_id, create_time, update_time, version
) VALUES (
  ?, ?, ?, ?, ?, ?, ?,
  0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0
);

-- name: GetStoryboardEpisodeByID :one
SELECT * FROM afv_storyboard_episode
WHERE id = ? AND IFNULL(deleted, 0) = 0;

-- name: LockStoryboardEpisodeByID :one
SELECT * FROM afv_storyboard_episode
WHERE id = ? AND IFNULL(deleted, 0) = 0
FOR UPDATE;

-- name: ListStoryboardEpisodes :many
SELECT * FROM afv_storyboard_episode
WHERE storyboard_id = ? AND IFNULL(deleted, 0) = 0
ORDER BY sort_order ASC, episode_number ASC, id ASC;

-- name: CountStoryboardEpisodes :one
SELECT COUNT(*) FROM afv_storyboard_episode
WHERE storyboard_id = ? AND IFNULL(deleted, 0) = 0;

-- name: UpdateStoryboardEpisode :execrows
UPDATE afv_storyboard_episode
SET script_episode_id = ?,
    episode_number = ?,
    title = ?,
    synopsis = ?,
    status = ?,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND IFNULL(deleted, 0) = 0;

-- name: SoftDeleteStoryboardEpisode :execrows
UPDATE afv_storyboard_episode
SET deleted = 1,
    deleted_id = id,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND IFNULL(deleted, 0) = 0;

-- name: BumpStoryboardEpisodeVersion :execrows
UPDATE afv_storyboard_episode
SET version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND IFNULL(deleted, 0) = 0;

-- name: CreateStoryboardScene :execlastid
INSERT INTO afv_storyboard_scene (
  episode_id, storyboard_id, scene_number, scene_heading, location, time_of_day, int_ext,
  sort_order, status, deleted, create_time, update_time, version
) VALUES (
  ?, ?, ?, ?, ?, ?, ?,
  ?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0
);

-- name: GetStoryboardSceneByID :one
SELECT * FROM afv_storyboard_scene
WHERE id = ? AND IFNULL(deleted, 0) = 0;

-- name: LockStoryboardSceneByID :one
SELECT * FROM afv_storyboard_scene
WHERE id = ? AND IFNULL(deleted, 0) = 0
FOR UPDATE;

-- name: ListStoryboardScenesByEpisode :many
SELECT * FROM afv_storyboard_scene
WHERE episode_id = ? AND IFNULL(deleted, 0) = 0
ORDER BY sort_order ASC, id ASC;

-- name: ListStoryboardScenesByBoard :many
SELECT * FROM afv_storyboard_scene
WHERE storyboard_id = ? AND IFNULL(deleted, 0) = 0
ORDER BY sort_order ASC, id ASC;

-- name: CountStoryboardScenesByEpisode :one
SELECT COUNT(*) FROM afv_storyboard_scene
WHERE episode_id = ? AND IFNULL(deleted, 0) = 0;

-- name: UpdateStoryboardScene :execrows
UPDATE afv_storyboard_scene
SET scene_number = ?,
    scene_heading = ?,
    location = ?,
    time_of_day = ?,
    int_ext = ?,
    status = ?,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND IFNULL(deleted, 0) = 0;

-- name: SoftDeleteStoryboardScene :execrows
UPDATE afv_storyboard_scene
SET deleted = 1,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND IFNULL(deleted, 0) = 0;

-- name: BumpStoryboardSceneVersion :execrows
UPDATE afv_storyboard_scene
SET version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND IFNULL(deleted, 0) = 0;

-- name: CreateStoryboardItem :execlastid
INSERT INTO afv_storyboard_item (
  storyboard_id, storyboard_episode_id, storyboard_scene_id, sort_order,
  shot_number, auto_shot_number, image_url, reference_image_url, video_url,
  generated_image_url, first_frame_image_url, last_frame_image_url,
  first_frame_prompt, last_frame_prompt, generated_video_url,
  shot_type, duration, content, scene_expectation, sound, dialogue,
  sound_effect, music, camera_movement, camera_angle, camera_equipment, focal_length, transition,
  character_ids, scene_asset_item_id, prop_ids, remark, custom_data,
  ai_generated, status, video_prompt, version
) VALUES (
  ?, ?, ?, ?,
  ?, ?, ?, ?, ?,
  ?, ?, ?,
  ?, ?, ?,
  ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?,
  ?, ?, ?, 0
);

-- name: GetStoryboardItemByID :one
SELECT * FROM afv_storyboard_item
WHERE id = ? AND deleted = 0;

-- name: LockStoryboardItemByID :one
SELECT * FROM afv_storyboard_item
WHERE id = ? AND deleted = 0
FOR UPDATE;

-- name: ListStoryboardItemsByBoard :many
SELECT * FROM afv_storyboard_item
WHERE storyboard_id = ? AND deleted = 0
ORDER BY sort_order ASC, id ASC;

-- name: ListStoryboardItemsByScene :many
SELECT * FROM afv_storyboard_item
WHERE storyboard_scene_id = ? AND deleted = 0
ORDER BY sort_order ASC, id ASC;

-- name: CountStoryboardItemsByScene :one
SELECT COUNT(*) FROM afv_storyboard_item
WHERE storyboard_scene_id = ? AND deleted = 0;

-- name: UpdateStoryboardItem :execrows
UPDATE afv_storyboard_item
SET shot_number = ?,
    image_url = ?,
    reference_image_url = ?,
    video_url = ?,
    generated_image_url = ?,
    first_frame_image_url = ?,
    last_frame_image_url = ?,
    first_frame_prompt = ?,
    last_frame_prompt = ?,
    generated_video_url = ?,
    shot_type = ?,
    duration = ?,
    content = ?,
    scene_expectation = ?,
    sound = ?,
    dialogue = ?,
    sound_effect = ?,
    music = ?,
    camera_movement = ?,
    camera_angle = ?,
    camera_equipment = ?,
    focal_length = ?,
    transition = ?,
    character_ids = ?,
    scene_asset_item_id = ?,
    prop_ids = ?,
    remark = ?,
    custom_data = ?,
    ai_generated = ?,
    status = ?,
    video_prompt = ?,
    version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;

-- name: SoftDeleteStoryboardItem :execrows
UPDATE afv_storyboard_item
SET deleted = 1,
    version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;
