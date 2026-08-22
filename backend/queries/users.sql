-- name: CreateUser :one
INSERT INTO users (kind, email, display_name)
VALUES ($1, $2, $3)
RETURNING id, kind, email, display_name, created_at, updated_at, deleted_at;

-- name: GetUserByID :one
SELECT id, kind, email, display_name, created_at, updated_at, deleted_at
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: UpgradeUser :one
UPDATE users
SET kind = 'REGISTERED',
    email = COALESCE(sqlc.narg(email), email),
    display_name = COALESCE(sqlc.narg(display_name), display_name),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id, kind, email, display_name, created_at, updated_at, deleted_at;
