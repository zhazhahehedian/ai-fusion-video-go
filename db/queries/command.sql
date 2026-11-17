-- name: AdmitResourceCommand :exec
INSERT INTO afv_resource_command (
  id, actor_user_id, request_id, request_hash, origin,
  canvas_id, agent_run_id, tool_call_id, request_json
) VALUES (
  ?, ?, ?, ?, ?,
  ?, ?, ?, ?
) ON DUPLICATE KEY UPDATE id = id;

-- name: LockResourceCommandByRequest :one
SELECT id, request_hash, COALESCE(CAST(result_json AS CHAR), '') AS result_json
FROM afv_resource_command
WHERE actor_user_id = ? AND request_id = ?
FOR UPDATE;

-- name: CommitResourceCommand :execrows
UPDATE afv_resource_command
SET result_json = ?, undo_json = ?, committed_at = CURRENT_TIMESTAMP(6)
WHERE id = ? AND result_json IS NULL;

-- name: InsertResourceRevision :exec
INSERT INTO afv_resource_revision (
  command_id, resource_type, resource_id, project_id,
  resource_version, before_json, after_json, changed_fields
) VALUES (
  ?, ?, ?, ?,
  ?, ?, ?, ?
);
