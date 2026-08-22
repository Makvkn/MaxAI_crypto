-- name: CreateSubscription :one
INSERT INTO subscriptions (user_id, plan, status)
VALUES ($1, $2, $3)
RETURNING id, user_id, plan, status, current_period_end, created_at, updated_at;

-- name: GetSubscriptionByUser :one
SELECT id, user_id, plan, status, current_period_end, created_at, updated_at
FROM subscriptions
WHERE user_id = $1;
