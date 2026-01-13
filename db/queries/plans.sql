-- name: CreatePlan :one
INSERT INTO plans (
  name,
  duration_days,
  price_cents,
  active,
  description
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdatePlan :one
UPDATE plans
SET
  name = $2,
  duration_days = $3,
  price_cents = $4,
  active = $5,
  description = $6,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetPlan :one
SELECT * FROM plans WHERE id = $1 LIMIT 1;

-- name: ListActivePlans :many
SELECT * FROM plans WHERE active = true ORDER BY name;
