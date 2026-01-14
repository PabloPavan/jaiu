-- name: GetSubscriptionBalance :one
SELECT * FROM subscription_balances WHERE subscription_id = $1 LIMIT 1;

-- name: UpsertSubscriptionBalance :one
INSERT INTO subscription_balances (subscription_id, credit_cents)
VALUES ($1, $2)
ON CONFLICT (subscription_id)
DO UPDATE SET credit_cents = EXCLUDED.credit_cents, updated_at = now()
RETURNING *;

-- name: AddSubscriptionBalance :one
INSERT INTO subscription_balances (subscription_id, credit_cents)
VALUES ($1, GREATEST(sqlc.arg(credit_cents), 0))
ON CONFLICT (subscription_id)
DO UPDATE SET credit_cents = GREATEST(subscription_balances.credit_cents + sqlc.arg(credit_cents), 0),
              updated_at = now()
RETURNING *;
