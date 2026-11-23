-- name: LockGenerationDispatch :one
SELECT id FROM afv_generation_dispatch_lock WHERE id = 1 FOR UPDATE;

-- name: DatabaseNow :one
SELECT UTC_TIMESTAMP(6) AS now;

-- name: InsertGenerationSubmission :exec
INSERT INTO afv_generation_submission (
  id, request_id, scope_key, request_hash, kind, task_id, public_task_id,
  actor_user_id, actor_team_id, project_id, canvas_node_run_id,
  context_json, model_snapshot_json, api_config_id, api_config_hash,
  state, cancel_requested, revision, created_at, updated_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?,
  ?, ?, ?, ?,
  ?, 0, 1, UTC_TIMESTAMP(6), UTC_TIMESTAMP(6)
) ON DUPLICATE KEY UPDATE id = id;

-- name: GetGenerationSubmissionByScopeRequest :one
SELECT
  id, request_id, scope_key, request_hash, kind, task_id, public_task_id,
  actor_user_id, actor_team_id, project_id, canvas_node_run_id,
  context_json, model_snapshot_json, api_config_id, api_config_hash,
  state, cancel_requested, revision, created_at, updated_at
FROM afv_generation_submission
WHERE scope_key = ? AND request_id = ?;

-- name: LockGenerationSubmissionByScopeRequest :one
SELECT
  id, request_id, scope_key, request_hash, kind, task_id, public_task_id,
  actor_user_id, actor_team_id, project_id, canvas_node_run_id,
  context_json, model_snapshot_json, api_config_id, api_config_hash,
  state, cancel_requested, revision, created_at, updated_at
FROM afv_generation_submission
WHERE scope_key = ? AND request_id = ?
FOR UPDATE;

-- name: GetGenerationSubmissionByID :one
SELECT
  id, request_id, scope_key, request_hash, kind, task_id, public_task_id,
  actor_user_id, actor_team_id, project_id, canvas_node_run_id,
  context_json, model_snapshot_json, api_config_id, api_config_hash,
  state, cancel_requested, revision, created_at, updated_at
FROM afv_generation_submission
WHERE id = ?;

-- name: LockGenerationSubmissionByID :one
SELECT id FROM afv_generation_submission WHERE id = ? FOR UPDATE;

-- name: UpdateGenerationSubmissionState :execrows
UPDATE afv_generation_submission
SET state = ?, revision = revision + 1, updated_at = UTC_TIMESTAMP(6)
WHERE id = ?;

-- name: InsertRemoteGenerationJob :exec
INSERT INTO afv_remote_generation_job (
  id, submission_id, item_id, item_index, model_id, actor_user_id, project_id,
  state, phase, input_json, next_poll_at, deadline_at, created_at, updated_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?,
  'QUEUED', 'NEW', ?, UTC_TIMESTAMP(6),
  TIMESTAMPADD(HOUR, 2, UTC_TIMESTAMP(6)), UTC_TIMESTAMP(6), UTC_TIMESTAMP(6)
);

-- name: ListJobsBySubmissionID :many
SELECT
  id, submission_id, item_id, item_index, model_id, actor_user_id, project_id,
  state, phase, provider_task_id, active_attempt_id, input_json, COALESCE(result_json, CAST('null' AS JSON)) AS result_json,
  cancel_requested, cancel_sent, permit_held, owner_instance_id, owner_epoch,
  lease_until, next_poll_at, deadline_at, poll_attempts, consecutive_errors,
  error_code, error_message, revision, created_at, updated_at
FROM afv_remote_generation_job
WHERE submission_id = ?
ORDER BY item_index ASC;

-- name: GetRemoteGenerationJob :one
SELECT
  id, submission_id, item_id, item_index, model_id, actor_user_id, project_id,
  state, phase, provider_task_id, active_attempt_id, input_json, COALESCE(result_json, CAST('null' AS JSON)) AS result_json,
  cancel_requested, cancel_sent, permit_held, owner_instance_id, owner_epoch,
  lease_until, next_poll_at, deadline_at, poll_attempts, consecutive_errors,
  error_code, error_message, revision, created_at, updated_at
FROM afv_remote_generation_job
WHERE id = ?;

-- name: SelectClaimCandidates :many
SELECT id, model_id, actor_user_id, project_id, phase, permit_held, cancel_requested
FROM afv_remote_generation_job
WHERE (
    (state IN ('READY', 'WAITING') AND next_poll_at <= UTC_TIMESTAMP(6))
    OR (state = 'RUNNING' AND lease_until <= UTC_TIMESTAMP(6))
    OR state = 'NEEDS_REVIEW'
    OR (state = 'UNKNOWN_SUBMISSION' AND (permit_held = TRUE OR phase <> 'FINISHED'))
  )
ORDER BY next_poll_at, id
LIMIT 128
FOR UPDATE SKIP LOCKED;

-- name: DeferJobForCapacity :exec
UPDATE afv_remote_generation_job
SET next_poll_at = TIMESTAMPADD(SECOND, 2, UTC_TIMESTAMP(6))
WHERE id = ?;

-- name: ClaimRemoteGenerationJob :execrows
UPDATE afv_remote_generation_job
SET state = 'RUNNING',
    owner_instance_id = ?,
    owner_epoch = owner_epoch + 1,
    lease_until = TIMESTAMPADD(SECOND, 90, UTC_TIMESTAMP(6)),
    permit_held = ?,
    revision = revision + 1,
    updated_at = UTC_TIMESTAMP(6)
WHERE id = ?;

-- name: CountPermitHeldByActor :one
SELECT COUNT(*) FROM afv_remote_generation_job
WHERE actor_user_id = ? AND permit_held = TRUE;

-- name: HeartbeatRemoteGenerationJob :execrows
UPDATE afv_remote_generation_job
SET lease_until = TIMESTAMPADD(SECOND, 90, UTC_TIMESTAMP(6))
WHERE id = ?
  AND owner_instance_id = ?
  AND owner_epoch = ?
  AND state = 'RUNNING'
  AND lease_until > UTC_TIMESTAMP(6);

-- name: SelectLeasedJobForUpdate :one
SELECT
  id, submission_id, item_id, item_index, model_id, actor_user_id, project_id,
  state, phase, provider_task_id, active_attempt_id, input_json, COALESCE(result_json, CAST('null' AS JSON)) AS result_json,
  cancel_requested, cancel_sent, permit_held, owner_instance_id, owner_epoch,
  lease_until, next_poll_at, deadline_at, poll_attempts, consecutive_errors,
  error_code, error_message, revision, created_at, updated_at
FROM afv_remote_generation_job
WHERE id = ?
  AND owner_instance_id = ?
  AND owner_epoch = ?
  AND state = 'RUNNING'
  AND lease_until > UTC_TIMESTAMP(6)
FOR UPDATE;

-- name: ScheduleJobPoll :exec
UPDATE afv_remote_generation_job
SET state = 'WAITING',
    phase = sqlc.arg(phase),
    lease_until = NULL,
    next_poll_at = TIMESTAMPADD(SECOND, sqlc.arg(delay_seconds), UTC_TIMESTAMP(6)),
    poll_attempts = poll_attempts + 1,
    consecutive_errors = sqlc.arg(consecutive_errors),
    error_code = sqlc.arg(error_code),
    error_message = sqlc.arg(error_message),
    revision = revision + 1,
    updated_at = UTC_TIMESTAMP(6)
WHERE id = sqlc.arg(id);

-- name: PrepareJobAttempt :exec
UPDATE afv_remote_generation_job
SET active_attempt_id = ?,
    phase = 'PREPARING',
    revision = revision + 1,
    updated_at = UTC_TIMESTAMP(6)
WHERE id = ?;

-- name: BeginJobSubmission :exec
UPDATE afv_remote_generation_job
SET provider_task_id = ?,
    phase = 'SUBMITTING',
    revision = revision + 1,
    updated_at = UTC_TIMESTAMP(6)
WHERE id = ?;

-- name: FinishRemoteGenerationJob :exec
UPDATE afv_remote_generation_job
SET state = ?,
    phase = 'FINISHED',
    permit_held = FALSE,
    lease_until = NULL,
    result_json = ?,
    error_code = ?,
    error_message = ?,
    revision = revision + 1,
    updated_at = UTC_TIMESTAMP(6)
WHERE id = ?;

-- name: MarkJobsReadyBySubmission :exec
UPDATE afv_remote_generation_job
SET state = 'READY'
WHERE submission_id = ? AND state = 'QUEUED';

-- name: InsertGenerationAttempt :exec
INSERT INTO afv_generation_attempt (
  id, job_id, attempt_no, request_json, created_at
) VALUES (
  ?, ?, ?, ?, UTC_TIMESTAMP(6)
);

-- name: AcknowledgeGenerationAttempt :exec
UPDATE afv_generation_attempt
SET provider_task_id = ?,
    submitted_at = UTC_TIMESTAMP(6),
    acknowledged_at = UTC_TIMESTAMP(6)
WHERE id = ?;

-- name: InsertAsyncOutbox :exec
INSERT INTO afv_async_outbox (
  event_key, event_type, aggregate_id, payload_json, next_attempt_at, created_at
) VALUES (
  ?, ?, ?, ?, UTC_TIMESTAMP(6), UTC_TIMESTAMP(6)
);

-- name: NextPendingGenerationOutbox :one
SELECT o.id, o.event_type, o.aggregate_id, o.payload_json
FROM afv_async_outbox o
WHERE o.state = 'PENDING'
  AND o.event_type IN ('GENERATION_ACCEPTED', 'GENERATION_CHANGED', 'MEDIA_JOB_ACCEPTED')
  AND o.next_attempt_at <= UTC_TIMESTAMP(6)
  AND NOT EXISTS (
    SELECT 1 FROM afv_async_outbox prior
    WHERE prior.aggregate_id = o.aggregate_id
      AND prior.state = 'PENDING'
      AND prior.id < o.id
  )
ORDER BY o.id
LIMIT 1
FOR UPDATE;

-- name: MarkOutboxDelivered :exec
UPDATE afv_async_outbox
SET state = 'DELIVERED',
    attempts = attempts + 1,
    last_error = NULL,
    delivered_at = UTC_TIMESTAMP(6)
WHERE id = ?;

-- name: BackoffOutbox :execrows
UPDATE afv_async_outbox
SET next_attempt_at = TIMESTAMPADD(SECOND, LEAST(60, POW(2, LEAST(attempts + 1, 6))), UTC_TIMESTAMP(6)),
    attempts = attempts + 1,
    last_error = ?
WHERE id = ? AND state = 'PENDING';
