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
  properties,
  art_style,
  art_style_description,
  art_style_image_prompt,
  art_style_image_url,
  deleted,
  create_time,
  update_time
FROM afv_project
WHERE id = ? AND deleted = 0;

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
  properties,
  art_style,
  art_style_description,
  art_style_image_prompt,
  art_style_image_url,
  deleted,
  create_time,
  update_time
FROM afv_project
WHERE deleted = 0
ORDER BY id DESC
LIMIT ? OFFSET ?;
