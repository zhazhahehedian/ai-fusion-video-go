-- name: GetUserByID :one
SELECT
  id,
  username,
  password,
  nickname,
  avatar,
  email,
  phone,
  status,
  deleted,
  create_time,
  update_time
FROM sys_user
WHERE id = ? AND deleted = 0;

-- name: GetUserByUsername :one
SELECT
  id,
  username,
  password,
  nickname,
  avatar,
  email,
  phone,
  status,
  deleted,
  create_time,
  update_time
FROM sys_user
WHERE username = ? AND deleted = 0;

-- name: CreateUser :execlastid
INSERT INTO sys_user (
  username,
  password,
  nickname,
  status,
  deleted
) VALUES (
  ?,
  ?,
  ?,
  1,
  0
);
