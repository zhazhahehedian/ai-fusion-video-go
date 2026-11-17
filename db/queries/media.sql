-- name: InsertMediaJob :exec
INSERT INTO afv_media_job (
  id, kind, project_id, target_id, target_version, actor_user_id,
  request_id, request_hash, context_json, input_json, checkpoint_json,
  status, phase, deadline_at, next_run_at, created_at, updated_at
) VALUES (
  ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?, ?,
  'QUEUED', ?, TIMESTAMPADD(HOUR, 2, UTC_TIMESTAMP(6)), UTC_TIMESTAMP(6),
  UTC_TIMESTAMP(6), UTC_TIMESTAMP(6)
) ON DUPLICATE KEY UPDATE id = id;

-- name: GetMediaJobByActorRequest :one
SELECT
  id, kind, project_id, target_id, target_version, actor_user_id,
  request_id, request_hash, context_json, input_json, checkpoint_json,
  status, phase, error_msg, owner_instance_id, owner_epoch, lease_until,
  next_run_at, deadline_at, cancel_requested, revision, created_at, updated_at
FROM afv_media_job
WHERE actor_user_id = ? AND request_id = ?;

-- name: LockMediaJobByActorRequest :one
SELECT
  id, kind, project_id, target_id, target_version, actor_user_id,
  request_id, request_hash, context_json, input_json, checkpoint_json,
  status, phase, error_msg, owner_instance_id, owner_epoch, lease_until,
  next_run_at, deadline_at, cancel_requested, revision, created_at, updated_at
FROM afv_media_job
WHERE actor_user_id = ? AND request_id = ?
FOR UPDATE;

-- name: GetMediaJobByID :one
SELECT
  id, kind, project_id, target_id, target_version, actor_user_id,
  request_id, request_hash, context_json, input_json, checkpoint_json,
  status, phase, error_msg, owner_instance_id, owner_epoch, lease_until,
  next_run_at, deadline_at, cancel_requested, revision, created_at, updated_at
FROM afv_media_job
WHERE id = ?;

-- name: MarkMediaJobWaiting :exec
UPDATE afv_media_job
SET status = 'WAITING',
    phase = sqlc.arg(phase),
    revision = revision + 1,
    updated_at = UTC_TIMESTAMP(6)
WHERE id = sqlc.arg(id) AND status = 'QUEUED';

-- name: SelectMediaClaimCandidate :one
SELECT id, kind, cancel_requested
FROM afv_media_job
WHERE (
    (status IN ('QUEUED', 'WAITING') AND next_run_at <= UTC_TIMESTAMP(6))
    OR (status = 'RUNNING' AND lease_until <= UTC_TIMESTAMP(6))
  )
ORDER BY next_run_at, id
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: ClaimMediaJob :execrows
UPDATE afv_media_job
SET status = 'RUNNING',
    phase = sqlc.arg(phase),
    owner_instance_id = sqlc.arg(owner_instance_id),
    owner_epoch = owner_epoch + 1,
    lease_until = TIMESTAMPADD(SECOND, 90, UTC_TIMESTAMP(6)),
    revision = revision + 1,
    updated_at = UTC_TIMESTAMP(6)
WHERE id = sqlc.arg(id);

-- name: HeartbeatMediaJob :execrows
UPDATE afv_media_job
SET lease_until = TIMESTAMPADD(SECOND, 90, UTC_TIMESTAMP(6))
WHERE id = ?
  AND owner_instance_id = ?
  AND owner_epoch = ?
  AND status = 'RUNNING'
  AND lease_until > UTC_TIMESTAMP(6);

-- name: SelectLeasedMediaJobForUpdate :one
SELECT
  id, kind, project_id, target_id, target_version, actor_user_id,
  request_id, request_hash, context_json, input_json, checkpoint_json,
  status, phase, error_msg, owner_instance_id, owner_epoch, lease_until,
  next_run_at, deadline_at, cancel_requested, revision, created_at, updated_at
FROM afv_media_job
WHERE id = ?
  AND owner_instance_id = ?
  AND owner_epoch = ?
  AND status = 'RUNNING'
  AND lease_until > UTC_TIMESTAMP(6)
FOR UPDATE;

-- name: CheckpointMediaJob :exec
UPDATE afv_media_job
SET checkpoint_json = ?,
    phase = ?,
    revision = revision + 1,
    updated_at = UTC_TIMESTAMP(6)
WHERE id = ?;

-- name: FinishMediaJob :exec
UPDATE afv_media_job
SET status = ?,
    phase = ?,
    checkpoint_json = ?,
    error_msg = ?,
    owner_instance_id = NULL,
    lease_until = NULL,
    revision = revision + 1,
    updated_at = UTC_TIMESTAMP(6)
WHERE id = ?;
