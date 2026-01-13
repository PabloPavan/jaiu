-- name: CreatePayment :one
INSERT INTO payments (
  subscription_id,
  paid_at,
  amount_cents,
  method,
  reference,
  notes,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: ListPaymentsBySubscription :many
SELECT * FROM payments WHERE subscription_id = $1 ORDER BY paid_at DESC;

-- name: ListPaymentsByPeriod :many
SELECT *
FROM payments
WHERE paid_at >= $1 AND paid_at < $2
ORDER BY paid_at DESC;
