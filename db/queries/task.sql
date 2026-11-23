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

-- name: InsertImageItemPending :execlastid
INSERT INTO afv_image_item (
  task_id, status, deleted, create_time, update_time
) VALUES (
  ?, 0, 0, UTC_TIMESTAMP(), UTC_TIMESTAMP()
);

-- name: InsertVideoItemPending :execlastid
INSERT INTO afv_video_item (
  task_id, status, deleted, create_time, update_time
) VALUES (
  ?, 0, 0, UTC_TIMESTAMP(), UTC_TIMESTAMP()
);

-- name: MarkImageTaskProcessing :execrows
UPDATE afv_image_task
SET status = 1,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND deleted = 0 AND status IN (0, 1);

-- name: MarkVideoTaskProcessing :execrows
UPDATE afv_video_task
SET status = 1,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND deleted = 0 AND status IN (0, 1);

-- name: FinishImageItemStub :execrows
UPDATE afv_image_item
SET platform_task_id = ?,
    image_url = ?,
    thumbnail_url = ?,
    width = ?,
    height = ?,
    status = 1,
    error_msg = NULL,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND task_id = ? AND deleted = 0 AND status = 0;

-- name: FinishVideoItemStub :execrows
UPDATE afv_video_item
SET platform_task_id = ?,
    video_url = ?,
    cover_url = ?,
    duration = ?,
    status = 1,
    error_msg = NULL,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND task_id = ? AND deleted = 0 AND status = 0;

-- 一次提交只对应一个远端作业，成功数等于该作业写回的条目数。
-- name: CompleteImageTaskStub :execrows
UPDATE afv_image_task
SET success_count = ?,
    status = 2,
    error_msg = NULL,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND deleted = 0 AND status IN (0, 1);

-- name: CompleteVideoTaskStub :execrows
UPDATE afv_video_task
SET success_count = ?,
    status = 2,
    error_msg = NULL,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND deleted = 0 AND status IN (0, 1);

-- name: FailImageItemsPending :execrows
UPDATE afv_image_item
SET status = 2,
    error_msg = ?,
    update_time = UTC_TIMESTAMP()
WHERE task_id = ? AND deleted = 0 AND status = 0;

-- name: FailVideoItemsPending :execrows
UPDATE afv_video_item
SET status = 2,
    error_msg = ?,
    update_time = UTC_TIMESTAMP()
WHERE task_id = ? AND deleted = 0 AND status = 0;

-- name: FailImageTaskOpen :execrows
UPDATE afv_image_task
SET status = 3,
    error_msg = ?,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND deleted = 0 AND status IN (0, 1);

-- name: FailVideoTaskOpen :execrows
UPDATE afv_video_task
SET status = 3,
    error_msg = ?,
    update_time = UTC_TIMESTAMP()
WHERE id = ? AND deleted = 0 AND status IN (0, 1);
