-- name: CreateUser :one
INSERT INTO users (id, email, name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    name = EXCLUDED.name
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: CreateAccount :one
INSERT INTO aws_accounts (team_id, account_id, external_id, role_arn, verified, last_verified_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAccountByExternalID :one
SELECT * FROM aws_accounts
WHERE external_id = $1 LIMIT 1;

-- name: GetAccountsByTeamID :many
SELECT * FROM aws_accounts
WHERE team_id = $1;

-- name: DeleteAccount :exec
DELETE FROM aws_accounts
WHERE account_id = $1 AND team_id = $2;

-- name: CreateTeam :one
INSERT INTO teams (name, slug, owner_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTeamByOwnerID :one
SELECT * FROM teams
WHERE owner_id = $1 LIMIT 1;

-- name: AddTeamMember :one
INSERT INTO team_members (team_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;
