-- name: CreateUser :one
INSERT INTO users (email, password, username)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1
LIMIT 1;


-- name: GetAllUsers :many
SELECT id, email FROM users;

-- name: UpdateUserPassword :exec
UPDATE users
SET password = $1, updated_at = NOW()
WHERE email = $2;

-- name: SavePasswordReset :exec
INSERT INTO password_resets (email, otp, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (email)
DO UPDATE SET
    otp = EXCLUDED.otp,
    expires_at = EXCLUDED.expires_at,
    created_at = NOW();

-- name: GetPasswordReset :one
SELECT otp, expires_at FROM password_resets
WHERE email = $1;

-- name: DeletePasswordReset :exec
DELETE FROM password_resets
WHERE email = $1;