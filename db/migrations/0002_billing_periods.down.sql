DROP TRIGGER IF EXISTS subscription_balances_updated_at ON subscription_balances;
DROP TRIGGER IF EXISTS billing_periods_updated_at ON billing_periods;

DROP TABLE IF EXISTS payment_allocations;
DROP TABLE IF EXISTS billing_periods;
DROP TABLE IF EXISTS subscription_balances;

ALTER TABLE payments
  DROP COLUMN IF EXISTS credit_cents,
  DROP COLUMN IF EXISTS kind;
