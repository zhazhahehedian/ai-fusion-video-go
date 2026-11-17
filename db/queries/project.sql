-- name: GetProjectByID :one
SELECT
  id,
  name,
  description,
  cover_url,
  scope,
  owner_type,
  owner_id,
  status,
  COALESCE(properties, CAST('{}' AS JSON)) AS properties,
  art_style,
  art_style_description,
  art_style_image_prompt,
  art_style_image_url,
  version,
  deleted,
  create_time,
  update_time
FROM afv_project
WHERE id = ? AND deleted = 0;

-- name: GetOwnedProjectByID :one
SELECT
  id,
  name,
  description,
  cover_url,
  scope,
  owner_type,
  owner_id,
  status,
  COALESCE(properties, CAST('{}' AS JSON)) AS properties,
  art_style,
  art_style_description,
  art_style_image_prompt,
  art_style_image_url,
  version,
  deleted,
  create_time,
  update_time
FROM afv_project
WHERE id = ? AND deleted = 0
  AND owner_type = 1
  AND owner_id = ?;

-- name: LockOwnedProjectByID :one
SELECT
  id,
  name,
  description,
  cover_url,
  scope,
  owner_type,
  owner_id,
  status,
  COALESCE(properties, CAST('{}' AS JSON)) AS properties,
  art_style,
  art_style_description,
  art_style_image_prompt,
  art_style_image_url,
  version,
  deleted,
  create_time,
  update_time
FROM afv_project
WHERE id = ? AND deleted = 0
  AND owner_type = 1
  AND owner_id = ?
FOR UPDATE;

-- name: ListProjects :many
SELECT
  id,
  name,
  description,
  cover_url,
  scope,
  owner_type,
  owner_id,
  status,
  COALESCE(properties, CAST('{}' AS JSON)) AS properties,
  art_style,
  art_style_description,
  art_style_image_prompt,
  art_style_image_url,
  version,
  deleted,
  create_time,
  update_time
FROM afv_project
WHERE deleted = 0
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- 阶段2：仅个人拥有者可见（团队/成员 ACL 见技术债）
-- name: ListOwnedProjects :many
SELECT
  id,
  name,
  description,
  cover_url,
  scope,
  owner_type,
  owner_id,
  status,
  COALESCE(properties, CAST('{}' AS JSON)) AS properties,
  art_style,
  art_style_description,
  art_style_image_prompt,
  art_style_image_url,
  version,
  deleted,
  create_time,
  update_time
FROM afv_project
WHERE deleted = 0
  AND owner_type = 1
  AND owner_id = ?
ORDER BY update_time DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountOwnedProjects :one
SELECT COUNT(*)
FROM afv_project
WHERE deleted = 0
  AND owner_type = 1
  AND owner_id = ?;

-- name: ListAllOwnedProjects :many
SELECT
  id,
  name,
  description,
  cover_url,
  scope,
  owner_type,
  owner_id,
  status,
  COALESCE(properties, CAST('{}' AS JSON)) AS properties,
  art_style,
  art_style_description,
  art_style_image_prompt,
  art_style_image_url,
  version,
  deleted,
  create_time,
  update_time
FROM afv_project
WHERE deleted = 0
  AND owner_type = 1
  AND owner_id = ?
ORDER BY update_time DESC, id DESC;

-- 阶段3：项目行创建（幂等与 ensure script 由 resource 命令管道负责）
-- name: CreateOwnedProject :execlastid
INSERT INTO afv_project (
  name,
  description,
  cover_url,
  scope,
  owner_type,
  owner_id,
  status,
  properties,
  art_style,
  art_style_description,
  art_style_image_prompt,
  art_style_image_url,
  version,
  deleted
) VALUES (
  ?,
  ?,
  ?,
  2,
  1,
  ?,
  0,
  ?,
  ?,
  ?,
  ?,
  ?,
  0,
  0
);
