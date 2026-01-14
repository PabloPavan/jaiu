-- name: CreateBillingPeriod :one
INSERT INTO billing_periods (
  subscription_id,
  period_start,
  period_end,
  amount_due_cents,
  amount_paid_cents,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateBillingPeriod :one
UPDATE billing_periods
SET
  amount_paid_cents = $2,
  status = $3,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListBillingPeriodsBySubscription :many
SELECT * FROM billing_periods WHERE subscription_id = $1 ORDER BY period_start;

-- name: ListOpenBillingPeriodsBySubscription :many
SELECT *
FROM billing_periods
WHERE subscription_id = $1
  AND status IN ('open', 'partial', 'overdue')
ORDER BY period_end;

-- name: MarkBillingPeriodsOverdue :exec
UPDATE billing_periods
SET status = 'overdue',
    updated_at = now()
WHERE subscription_id = $1
  AND status IN ('open', 'partial')
  AND period_end < $2;
