-- name: CreateSubscription :one
INSERT INTO subscriptions (
  student_id,
  plan_id,
  start_date,
  end_date,
  status,
  price_cents,
  payment_day,
  auto_renew
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateSubscription :one
UPDATE subscriptions
SET
  start_date = $2,
  end_date = $3,
  status = $4,
  price_cents = $5,
  payment_day = $6,
  auto_renew = $7,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetSubscription :one
SELECT * FROM subscriptions WHERE id = $1 LIMIT 1;

-- name: ListSubscriptionsByStudent :many
SELECT * FROM subscriptions WHERE student_id = $1 ORDER BY start_date DESC;

-- name: ListSubscriptionsDueBetween :many
SELECT *
FROM subscriptions
WHERE status = 'active'
  AND end_date BETWEEN $1 AND $2
ORDER BY end_date;

-- name: ListAutoRenewSubscriptions :many
SELECT *
FROM subscriptions
WHERE status = 'active'
  AND auto_renew = true
ORDER BY start_date;
