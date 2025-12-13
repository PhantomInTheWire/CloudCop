-- name: CreateUser :one
INSERT INTO users (email, full_name)
VALUES ($1, $2)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserExternalID :one
SELECT external_id FROM users WHERE email = $1;

-- name: CreateAccountConnection :one
INSERT INTO account_connections (user_id, external_id, account_id, account_arn, verified)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAccountConnections :many
SELECT * FROM account_connections WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetAccountConnectionByID :one
SELECT * FROM account_connections WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: VerifyAccountConnection :one
UPDATE account_connections
SET verified = true, last_verified_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND account_id = $2
RETURNING *;

-- name: DeleteAccountConnection :exec
DELETE FROM account_connections WHERE user_id = $1 AND id = $2;

-- name: GetAccountConnectionByAccountID :one
SELECT * FROM account_connections
WHERE user_id = $1 AND account_id = $2
LIMIT 1;
