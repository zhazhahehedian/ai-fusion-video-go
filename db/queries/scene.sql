-- name: CreateScriptScene :execlastid
INSERT INTO afv_script_scene_item (
  episode_id, script_id, scene_number, scene_heading, location, time_of_day, int_ext,
  characters, character_asset_ids, scene_asset_id, prop_asset_ids, scene_description,
  dialogues, sort_order, status, version, deleted
) VALUES (
  ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?,
  ?, ?, ?, 0, 0
);

-- name: GetScriptSceneByID :one
SELECT
  id, episode_id, script_id, scene_number, scene_heading, location, time_of_day, int_ext,
  COALESCE(characters, CAST('[]' AS JSON)) AS characters,
  COALESCE(character_asset_ids, CAST('[]' AS JSON)) AS character_asset_ids,
  scene_asset_id,
  COALESCE(prop_asset_ids, CAST('[]' AS JSON)) AS prop_asset_ids,
  scene_description,
  COALESCE(dialogues, CAST('[]' AS JSON)) AS dialogues,
  sort_order, status, version, deleted, create_time, update_time
FROM afv_script_scene_item
WHERE id = ? AND deleted = 0;

-- name: LockScriptSceneByID :one
SELECT
  id, episode_id, script_id, scene_number, scene_heading, location, time_of_day, int_ext,
  COALESCE(characters, CAST('[]' AS JSON)) AS characters,
  COALESCE(character_asset_ids, CAST('[]' AS JSON)) AS character_asset_ids,
  scene_asset_id,
  COALESCE(prop_asset_ids, CAST('[]' AS JSON)) AS prop_asset_ids,
  scene_description,
  COALESCE(dialogues, CAST('[]' AS JSON)) AS dialogues,
  sort_order, status, version, deleted, create_time, update_time
FROM afv_script_scene_item
WHERE id = ? AND deleted = 0
FOR UPDATE;

-- name: ListScenesByEpisode :many
SELECT
  id, episode_id, script_id, scene_number, scene_heading, location, time_of_day, int_ext,
  COALESCE(characters, CAST('[]' AS JSON)) AS characters,
  COALESCE(character_asset_ids, CAST('[]' AS JSON)) AS character_asset_ids,
  scene_asset_id,
  COALESCE(prop_asset_ids, CAST('[]' AS JSON)) AS prop_asset_ids,
  scene_description,
  COALESCE(dialogues, CAST('[]' AS JSON)) AS dialogues,
  sort_order, status, version, deleted, create_time, update_time
FROM afv_script_scene_item
WHERE episode_id = ? AND deleted = 0
ORDER BY sort_order ASC, id ASC;

-- name: ListScenesByScript :many
SELECT
  id, episode_id, script_id, scene_number, scene_heading, location, time_of_day, int_ext,
  COALESCE(characters, CAST('[]' AS JSON)) AS characters,
  COALESCE(character_asset_ids, CAST('[]' AS JSON)) AS character_asset_ids,
  scene_asset_id,
  COALESCE(prop_asset_ids, CAST('[]' AS JSON)) AS prop_asset_ids,
  scene_description,
  COALESCE(dialogues, CAST('[]' AS JSON)) AS dialogues,
  sort_order, status, version, deleted, create_time, update_time
FROM afv_script_scene_item
WHERE script_id = ? AND deleted = 0
ORDER BY id ASC;

-- name: CountScenesByEpisode :one
SELECT COUNT(*)
FROM afv_script_scene_item
WHERE episode_id = ? AND deleted = 0;

-- name: UpdateScriptScene :execrows
UPDATE afv_script_scene_item
SET scene_number = ?,
    scene_heading = ?,
    location = ?,
    time_of_day = ?,
    int_ext = ?,
    characters = ?,
    character_asset_ids = ?,
    scene_asset_id = ?,
    prop_asset_ids = ?,
    scene_description = ?,
    dialogues = ?,
    sort_order = ?,
    status = ?,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND deleted = 0;

-- name: SoftDeleteScriptScene :execrows
UPDATE afv_script_scene_item
SET deleted = 1,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND deleted = 0;

-- name: SoftDeleteScenesByEpisode :exec
UPDATE afv_script_scene_item
SET deleted = 1,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE episode_id = ? AND deleted = 0;

-- name: SoftDeleteScenesByScript :exec
UPDATE afv_script_scene_item
SET deleted = 1,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE script_id = ? AND deleted = 0;
