-- name: CreateRefreshSession :one
INSERT INTO refresh_sessions (user_id, token_hash, expires_at, user_agent, ip_address)
VALUES ($1, $2, $3, $4, sqlc.narg(ip_address)::inet)
RETURNING id, user_id, token_hash, issued_at, expires_at, revoked_at, rotated_to, user_agent, host(ip_address)::text AS ip_address, last_used_at;

-- name: GetRefreshSessionByTokenHash :one
SELECT id, user_id, token_hash, issued_at, expires_at, revoked_at, rotated_to, user_agent, host(ip_address)::text AS ip_address, last_used_at
FROM refresh_sessions
WHERE token_hash = $1;

-- name: RevokeRefreshSession :exec
UPDATE refresh_sessions
SET revoked_at = NOW()
WHERE id = $1
  AND revoked_at IS NULL;

-- name: RevokeAllRefreshSessionsForUser :exec
UPDATE refresh_sessions
SET revoked_at = NOW()
WHERE user_id = $1
  AND revoked_at IS NULL;

-- name: MarkRefreshSessionRotated :exec
UPDATE refresh_sessions
SET revoked_at = NOW(),
    rotated_to = $2
WHERE id = $1
  AND revoked_at IS NULL;
