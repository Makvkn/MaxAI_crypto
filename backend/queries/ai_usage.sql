-- name: GetAIUsage :one
SELECT user_id, usage_date, used, plan, updated_at
FROM ai_usage
WHERE user_id = $1 AND usage_date = $2;

-- name: UpsertAIUsage :one
INSERT INTO ai_usage (user_id, usage_date, used, plan)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, usage_date) DO UPDATE
SET used = EXCLUDED.used,
    plan = EXCLUDED.plan,
    updated_at = NOW()
RETURNING user_id, usage_date, used, plan, updated_at;

-- name: GetAIUsageOperationByKey :one
SELECT id, user_id, usage_date, operation, idempotency_key, created_at
FROM ai_usage_operations
WHERE idempotency_key = $1;

-- name: InsertAIUsageOperation :one
INSERT INTO ai_usage_operations (user_id, usage_date, operation, idempotency_key)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, usage_date, operation, idempotency_key, created_at;
