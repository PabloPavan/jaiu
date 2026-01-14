-- name: CreatePaymentAllocation :exec
INSERT INTO payment_allocations (payment_id, billing_period_id, amount_cents)
VALUES ($1, $2, $3);

-- name: ListPaymentAllocationsByPayment :many
SELECT * FROM payment_allocations WHERE payment_id = $1 ORDER BY created_at;

-- name: DeletePaymentAllocationsByPayment :exec
DELETE FROM payment_allocations WHERE payment_id = $1;
