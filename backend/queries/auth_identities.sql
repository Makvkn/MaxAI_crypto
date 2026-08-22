-- name: CreateIdentity :one
INSERT INTO auth_identities (user_id, provider, subject, email, password_hash, email_verified)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, user_id, provider, subject, email, password_hash, email_verified, created_at, updated_at;

-- name: GetIdentityByProviderSubject :one
SELECT id, user_id, provider, subject, email, password_hash, email_verified, created_at, updated_at
FROM auth_identities
WHERE provider = $1
  AND subject = $2;

-- name: ListAuthProvidersByUser :many
SELECT provider
FROM auth_identities
WHERE user_id = $1
ORDER BY provider;
