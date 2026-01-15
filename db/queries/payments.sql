-- name: CreatePayment :one
INSERT INTO payments (
  subscription_id,
  paid_at,
  amount_cents,
  method,
  reference,
  notes,
  status,
  kind,
  credit_cents,
  idempotency_key
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdatePayment :one
UPDATE payments
SET
  subscription_id = $2,
  paid_at = $3,
  amount_cents = $4,
  method = $5,
  reference = $6,
  notes = $7,
  status = $8,
  kind = $9,
  credit_cents = $10
WHERE id = $1
RETURNING *;

-- name: GetPayment :one
SELECT * FROM payments WHERE id = $1 LIMIT 1;

-- name: GetPaymentByIdempotencyKey :one
SELECT * FROM payments WHERE idempotency_key = $1 LIMIT 1;

-- name: ListPaymentsBySubscription :many
SELECT * FROM payments WHERE subscription_id = $1 ORDER BY paid_at DESC;

-- name: ListPaymentsByPeriod :many
SELECT *
FROM payments
WHERE paid_at >= $1 AND paid_at < $2
ORDER BY paid_at DESC;
