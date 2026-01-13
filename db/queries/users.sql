-- name: CreateUser :one
INSERT INTO users (
  name,
  email,
  password_hash,
  role,
  active
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;
