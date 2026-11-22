-- name: CreateImageTask :execlastid
INSERT INTO afv_image_task (
  task_id, user_id, project_id, prompt, prompt_template_id, ref_image_urls,
  ratio, resolution, aspect_ratio, width, height, count, success_count, status,
  error_msg, model_id, workflow_version_id, category, owner_type, owner_id, deleted,
  create_time, update_time
) VALUES (
  ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?, ?, 0, ?,
  ?, ?, ?, ?, 1, ?, 0,
  UTC_TIMESTAMP(), UTC_TIMESTAMP()
);

-- name: GetImageTaskByID :one
SELECT
  id, task_id, user_id, project_id, prompt, prompt_template_id,
  COALESCE(ref_image_urls, CAST('null' AS JSON)) AS ref_image_urls,
  ratio, resolution, aspect_ratio, width, height, count, success_count, status,
  error_msg, model_id, workflow_version_id, category, owner_type, owner_id, deleted,
  create_time, update_time
FROM afv_image_task
WHERE id = ? AND deleted = 0;

-- name: GetImageTaskByPublicID :one
SELECT
  id, task_id, user_id, project_id, prompt, prompt_template_id,
  COALESCE(ref_image_urls, CAST('null' AS JSON)) AS ref_image_urls,
  ratio, resolution, aspect_ratio, width, height, count, success_count, status,
  error_msg, model_id, workflow_version_id, category, owner_type, owner_id, deleted,
  create_time, update_time
FROM afv_image_task
WHERE task_id = ? AND deleted = 0;

-- name: LockImageTaskByID :one
SELECT
  id, task_id, user_id, project_id, prompt, prompt_template_id,
  COALESCE(ref_image_urls, CAST('null' AS JSON)) AS ref_image_urls,
  ratio, resolution, aspect_ratio, width, height, count, success_count, status,
  error_msg, model_id, workflow_version_id, category, owner_type, owner_id, deleted,
  create_time, update_time
FROM afv_image_task
WHERE id = ? AND deleted = 0
FOR UPDATE;

-- name: PageImageTasksByUser :many
SELECT
  id, task_id, user_id, project_id, prompt, prompt_template_id,
  COALESCE(ref_image_urls, CAST('null' AS JSON)) AS ref_image_urls,
  ratio, resolution, aspect_ratio, width, height, count, success_count, status,
  error_msg, model_id, workflow_version_id, category, owner_type, owner_id, deleted,
  create_time, update_time
FROM afv_image_task
WHERE user_id = ? AND deleted = 0
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CountImageTasksByUser :one
SELECT COUNT(*) FROM afv_image_task
WHERE user_id = ? AND deleted = 0;

-- name: UpdateImageTask :execrows
UPDATE afv_image_task
SET prompt = ?,
    prompt_template_id = ?,
    ref_image_urls = ?,
    ratio = ?,
    resolution = ?,
    aspect_ratio = ?,
    width = ?,
    height = ?,
    count = ?,
    status = ?,
    error_msg = ?,
    model_id = ?,
    workflow_version_id = ?,
    category = ?,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND deleted = 0;

-- name: SoftDeleteImageTask :execrows
UPDATE afv_image_task
SET deleted = 1,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND deleted = 0;

-- name: SoftDeleteImageItemsByTask :exec
UPDATE afv_image_item
SET deleted = 1,
    update_time = UTC_TIMESTAMP()
WHERE task_id = ? AND deleted = 0;

-- name: ListImageItemsByTask :many
SELECT * FROM afv_image_item
WHERE task_id = ? AND deleted = 0
ORDER BY id ASC;

-- name: CreateVideoTask :execlastid
INSERT INTO afv_video_task (
  task_id, user_id, project_id, prompt, prompt_template_id, generate_mode,
  first_frame_image_url, last_frame_image_url, reference_image_urls, reference_video_urls, reference_audio_urls,
  ratio, resolution, duration, watermark, generate_audio, seed, camera_fixed,
  count, success_count, status, error_msg, model_id, workflow_version_id, category,
  owner_type, owner_id, deleted, create_time, update_time
) VALUES (
  ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?, ?, ?,
  ?, 0, ?, ?, ?, ?, ?,
  1, ?, 0, UTC_TIMESTAMP(), UTC_TIMESTAMP()
);

-- name: GetVideoTaskByID :one
SELECT
  id, task_id, user_id, project_id, prompt, prompt_template_id, generate_mode,
  first_frame_image_url, last_frame_image_url,
  COALESCE(reference_image_urls, CAST('null' AS JSON)) AS reference_image_urls,
  COALESCE(reference_video_urls, CAST('null' AS JSON)) AS reference_video_urls,
  COALESCE(reference_audio_urls, CAST('null' AS JSON)) AS reference_audio_urls,
  ratio, resolution, duration, watermark, generate_audio, seed, camera_fixed,
  count, success_count, status, error_msg, model_id, workflow_version_id, category,
  owner_type, owner_id, deleted, create_time, update_time
FROM afv_video_task
WHERE id = ? AND deleted = 0;

-- name: GetVideoTaskByPublicID :one
SELECT
  id, task_id, user_id, project_id, prompt, prompt_template_id, generate_mode,
  first_frame_image_url, last_frame_image_url,
  COALESCE(reference_image_urls, CAST('null' AS JSON)) AS reference_image_urls,
  COALESCE(reference_video_urls, CAST('null' AS JSON)) AS reference_video_urls,
  COALESCE(reference_audio_urls, CAST('null' AS JSON)) AS reference_audio_urls,
  ratio, resolution, duration, watermark, generate_audio, seed, camera_fixed,
  count, success_count, status, error_msg, model_id, workflow_version_id, category,
  owner_type, owner_id, deleted, create_time, update_time
FROM afv_video_task
WHERE task_id = ? AND deleted = 0;

-- name: LockVideoTaskByID :one
SELECT
  id, task_id, user_id, project_id, prompt, prompt_template_id, generate_mode,
  first_frame_image_url, last_frame_image_url,
  COALESCE(reference_image_urls, CAST('null' AS JSON)) AS reference_image_urls,
  COALESCE(reference_video_urls, CAST('null' AS JSON)) AS reference_video_urls,
  COALESCE(reference_audio_urls, CAST('null' AS JSON)) AS reference_audio_urls,
  ratio, resolution, duration, watermark, generate_audio, seed, camera_fixed,
  count, success_count, status, error_msg, model_id, workflow_version_id, category,
  owner_type, owner_id, deleted, create_time, update_time
FROM afv_video_task
WHERE id = ? AND deleted = 0
FOR UPDATE;

-- name: PageVideoTasksByUser :many
SELECT
  id, task_id, user_id, project_id, prompt, prompt_template_id, generate_mode,
  first_frame_image_url, last_frame_image_url,
  COALESCE(reference_image_urls, CAST('null' AS JSON)) AS reference_image_urls,
  COALESCE(reference_video_urls, CAST('null' AS JSON)) AS reference_video_urls,
  COALESCE(reference_audio_urls, CAST('null' AS JSON)) AS reference_audio_urls,
  ratio, resolution, duration, watermark, generate_audio, seed, camera_fixed,
  count, success_count, status, error_msg, model_id, workflow_version_id, category,
  owner_type, owner_id, deleted, create_time, update_time
FROM afv_video_task
WHERE user_id = ? AND deleted = 0
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CountVideoTasksByUser :one
SELECT COUNT(*) FROM afv_video_task
WHERE user_id = ? AND deleted = 0;

-- name: UpdateVideoTask :execrows
UPDATE afv_video_task
SET prompt = ?,
    prompt_template_id = ?,
    generate_mode = ?,
    first_frame_image_url = ?,
    last_frame_image_url = ?,
    reference_image_urls = ?,
    reference_video_urls = ?,
    reference_audio_urls = ?,
    ratio = ?,
    resolution = ?,
    duration = ?,
    watermark = ?,
    generate_audio = ?,
    seed = ?,
    camera_fixed = ?,
    count = ?,
    status = ?,
    error_msg = ?,
    model_id = ?,
    workflow_version_id = ?,
    category = ?,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND deleted = 0;

-- name: SoftDeleteVideoTask :execrows
UPDATE afv_video_task
SET deleted = 1,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND deleted = 0;

-- name: SoftDeleteVideoItemsByTask :exec
UPDATE afv_video_item
SET deleted = 1,
    update_time = UTC_TIMESTAMP()
WHERE task_id = ? AND deleted = 0;

-- name: ListVideoItemsByTask :many
SELECT * FROM afv_video_item
WHERE task_id = ? AND deleted = 0
ORDER BY id ASC;
