-- name: CreateConversation :one
INSERT INTO conversations (user_id, wallet_id, title)
VALUES ($1, $2, $3)
RETURNING id, user_id, wallet_id, title, message_count, last_message_preview, created_at, updated_at;

-- name: GetConversationByID :one
SELECT id, user_id, wallet_id, title, message_count, last_message_preview, created_at, updated_at
FROM conversations
WHERE id = $1;

-- name: ListConversationsByUser :many
SELECT id, user_id, wallet_id, title, message_count, last_message_preview, created_at, updated_at
FROM conversations
WHERE user_id = $1
  AND (sqlc.narg('wallet_id')::uuid IS NULL OR wallet_id = sqlc.narg('wallet_id'))
  AND (
    sqlc.narg('cursor_at')::timestamptz IS NULL
    OR (updated_at, id) < (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY updated_at DESC, id DESC
LIMIT $2;

-- name: BumpConversationOnMessage :exec
UPDATE conversations
SET message_count = message_count + 1,
    last_message_preview = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: InsertConversationMessage :one
INSERT INTO conversation_messages (
    conversation_id, role, status, content, response, tool_calls, error_code, error_message
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, conversation_id, role, status, content, response, tool_calls, error_code, error_message, created_at;

-- name: UpdateConversationMessage :exec
UPDATE conversation_messages
SET status = $2,
    content = $3,
    response = $4,
    tool_calls = $5,
    error_code = $6,
    error_message = $7
WHERE id = $1;

-- name: ListConversationMessages :many
SELECT id, conversation_id, role, status, content, response, tool_calls, error_code, error_message, created_at
FROM conversation_messages
WHERE conversation_id = $1
  AND (
    sqlc.narg('cursor_at')::timestamptz IS NULL
    OR (created_at, id) < (sqlc.narg('cursor_at')::timestamptz, sqlc.narg('cursor_id')::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT $2;

-- name: ListRecentConversationMessages :many
SELECT id, conversation_id, role, status, content, response, tool_calls, error_code, error_message, created_at
FROM conversation_messages
WHERE conversation_id = $1
ORDER BY created_at DESC
LIMIT $2;
