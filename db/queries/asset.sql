-- name: CreateAsset :execlastid
INSERT INTO afv_asset (
  user_id, project_id, type, name, description, cover_url, properties, tags,
  source_type, ai_prompt, owner_type, owner_id, status, deleted, version
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?,
  ?, ?, 1, ?, ?, 0, 0
);

-- name: GetAssetByID :one
SELECT
  id, user_id, project_id, type, name, description, cover_url,
  COALESCE(properties, CAST('null' AS JSON)) AS properties,
  COALESCE(tags, CAST('null' AS JSON)) AS tags,
  source_type, ai_prompt, owner_type, owner_id, status, deleted,
  create_time, update_time, version
FROM afv_asset
WHERE id = ? AND deleted = 0;

-- name: LockAssetByID :one
SELECT
  id, user_id, project_id, type, name, description, cover_url,
  COALESCE(properties, CAST('null' AS JSON)) AS properties,
  COALESCE(tags, CAST('null' AS JSON)) AS tags,
  source_type, ai_prompt, owner_type, owner_id, status, deleted,
  create_time, update_time, version
FROM afv_asset
WHERE id = ? AND deleted = 0
FOR UPDATE;

-- name: ListAssetsByOwner :many
SELECT a.*
FROM afv_asset a
INNER JOIN afv_project p
  ON p.id = a.project_id AND p.deleted = 0 AND p.owner_type = 1 AND p.owner_id = sqlc.arg(owner_id)
WHERE a.deleted = 0
  AND (sqlc.narg(project_id) IS NULL OR a.project_id = sqlc.narg(project_id))
  AND (sqlc.arg(asset_type) = '' OR a.type = sqlc.arg(asset_type))
  AND (sqlc.arg(keyword) = '' OR a.name LIKE sqlc.arg(keyword_pattern) ESCAPE '\\')
ORDER BY a.id DESC;

-- name: UpdateAsset :execrows
UPDATE afv_asset
SET type = ?,
    name = ?,
    description = ?,
    cover_url = ?,
    properties = ?,
    tags = ?,
    source_type = ?,
    ai_prompt = ?,
    status = ?,
    version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;

-- name: SoftDeleteAsset :execrows
UPDATE afv_asset
SET deleted = 1,
    version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;

-- name: BumpAssetVersion :execrows
UPDATE afv_asset
SET version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;

-- name: CreateAssetItem :execlastid
INSERT INTO afv_asset_item (
  asset_id, item_type, name, image_url, thumbnail_url, properties,
  sort_order, source_type, ai_prompt, deleted, version
) VALUES (
  ?, ?, ?, ?, ?, ?,
  ?, ?, ?, 0, 0
);

-- name: GetAssetItemByID :one
SELECT
  id, asset_id, item_type, name, image_url, thumbnail_url,
  COALESCE(properties, CAST('null' AS JSON)) AS properties,
  sort_order, source_type, ai_prompt, deleted, create_time, update_time, version
FROM afv_asset_item
WHERE id = ? AND deleted = 0;

-- name: LockAssetItemByID :one
SELECT
  id, asset_id, item_type, name, image_url, thumbnail_url,
  COALESCE(properties, CAST('null' AS JSON)) AS properties,
  sort_order, source_type, ai_prompt, deleted, create_time, update_time, version
FROM afv_asset_item
WHERE id = ? AND deleted = 0
FOR UPDATE;

-- name: ListAssetItems :many
SELECT
  id, asset_id, item_type, name, image_url, thumbnail_url,
  COALESCE(properties, CAST('null' AS JSON)) AS properties,
  sort_order, source_type, ai_prompt, deleted, create_time, update_time, version
FROM afv_asset_item
WHERE asset_id = ? AND deleted = 0
ORDER BY sort_order ASC, id ASC;

-- name: LockAssetItemsByAsset :many
SELECT
  id, asset_id, item_type, name, image_url, thumbnail_url,
  COALESCE(properties, CAST('null' AS JSON)) AS properties,
  sort_order, source_type, ai_prompt, deleted, create_time, update_time, version
FROM afv_asset_item
WHERE asset_id = ? AND deleted = 0
ORDER BY id ASC
FOR UPDATE;

-- name: ClearAssetCover :execrows
UPDATE afv_asset
SET cover_url = NULL,
    version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;

-- name: CountAssetItems :one
SELECT COUNT(*) FROM afv_asset_item
WHERE asset_id = ? AND deleted = 0;

-- name: UpdateAssetItem :execrows
UPDATE afv_asset_item
SET item_type = ?,
    name = ?,
    image_url = ?,
    thumbnail_url = ?,
    properties = ?,
    sort_order = ?,
    source_type = ?,
    ai_prompt = ?,
    version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;

-- name: SoftDeleteAssetItem :execrows
UPDATE afv_asset_item
SET deleted = 1,
    version = version + 1
WHERE id = ? AND version = ? AND deleted = 0;
