-- name: CreateScriptForProject :execlastid
INSERT INTO afv_script (
  project_id, title, scope, owner_type, owner_id, version, status, deleted, total_episodes
) VALUES (
  ?, ?, ?, ?, ?, 0, 0, 0, 0
);

-- name: GetScriptByProjectID :one
SELECT
  id, project_id, title, content, raw_content, total_episodes, story_synopsis,
  COALESCE(characters_json, CAST('[]' AS JSON)) AS characters_json, source_type, parsing_status, parsing_progress, summary, genre,
  target_audience, duration_estimate, scope, owner_type, owner_id, ai_generated,
  version, status, deleted, create_time, update_time
FROM afv_script
WHERE project_id = ? AND deleted = 0
ORDER BY id DESC
LIMIT 1;

-- name: GetScriptByID :one
SELECT
  id, project_id, title, content, raw_content, total_episodes, story_synopsis,
  COALESCE(characters_json, CAST('[]' AS JSON)) AS characters_json, source_type, parsing_status, parsing_progress, summary, genre,
  target_audience, duration_estimate, scope, owner_type, owner_id, ai_generated,
  version, status, deleted, create_time, update_time
FROM afv_script
WHERE id = ? AND deleted = 0;

-- name: LockScriptByID :one
SELECT
  id, project_id, title, content, raw_content, total_episodes, story_synopsis,
  COALESCE(characters_json, CAST('[]' AS JSON)) AS characters_json, source_type, parsing_status, parsing_progress, summary, genre,
  target_audience, duration_estimate, scope, owner_type, owner_id, ai_generated,
  version, status, deleted, create_time, update_time
FROM afv_script
WHERE id = ? AND deleted = 0
FOR UPDATE;

-- name: BumpScriptVersionAndEpisodeCount :execrows
UPDATE afv_script
SET version = version + 1,
    total_episodes = ?,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND deleted = 0;

-- name: CountEpisodesByScript :one
SELECT COUNT(*)
FROM afv_script_episode
WHERE script_id = ? AND deleted = 0;

-- name: CreateScriptEpisode :execlastid
INSERT INTO afv_script_episode (
  script_id, episode_number, title, synopsis, raw_content, duration_estimate,
  total_scenes, source_type, sort_order, parsing_status, status, version, deleted
) VALUES (
  ?, ?, ?, ?, ?, ?,
  0, ?, ?, 0, ?, 0, 0
);

-- name: GetScriptEpisodeByID :one
SELECT
  id, script_id, episode_number, title, synopsis, raw_content, duration_estimate,
  total_scenes, source_type, sort_order, parsing_status, status, version, deleted,
  create_time, update_time
FROM afv_script_episode
WHERE id = ? AND deleted = 0;

-- name: LockScriptEpisodeByID :one
SELECT
  id, script_id, episode_number, title, synopsis, raw_content, duration_estimate,
  total_scenes, source_type, sort_order, parsing_status, status, version, deleted,
  create_time, update_time
FROM afv_script_episode
WHERE id = ? AND deleted = 0
FOR UPDATE;

-- name: ListEpisodesByScript :many
SELECT
  id, script_id, episode_number, title, synopsis, raw_content, duration_estimate,
  total_scenes, source_type, sort_order, parsing_status, status, version, deleted,
  create_time, update_time
FROM afv_script_episode
WHERE script_id = ? AND deleted = 0
ORDER BY sort_order ASC, episode_number ASC, id ASC;

-- name: UpdateScriptEpisode :execrows
UPDATE afv_script_episode
SET episode_number = ?,
    title = ?,
    synopsis = ?,
    raw_content = ?,
    duration_estimate = ?,
    source_type = ?,
    sort_order = ?,
    status = ?,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND deleted = 0;

-- name: SoftDeleteScriptEpisode :execrows
UPDATE afv_script_episode
SET deleted = 1,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND deleted = 0;

-- name: FindEpisodeNumberConflict :one
SELECT id
FROM afv_script_episode
WHERE script_id = ? AND episode_number = ? AND deleted = 0
  AND id <> ?
LIMIT 1;

-- name: UpdateScript :execrows
UPDATE afv_script
SET content = ?,
    raw_content = ?,
    story_synopsis = ?,
    characters_json = ?,
    source_type = ?,
    parsing_status = ?,
    parsing_progress = ?,
    summary = ?,
    genre = ?,
    target_audience = ?,
    duration_estimate = ?,
    status = ?,
    total_episodes = ?,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND deleted = 0;

-- name: SoftDeleteEpisodesByScript :exec
UPDATE afv_script_episode
SET deleted = 1,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE script_id = ? AND deleted = 0;

-- name: BumpEpisodeSceneCount :execrows
UPDATE afv_script_episode
SET total_scenes = ?,
    version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND deleted = 0;

-- name: BumpScriptVersionOnly :execrows
UPDATE afv_script
SET version = version + 1,
    update_time = CURRENT_TIMESTAMP
WHERE id = ? AND version = ? AND deleted = 0;
