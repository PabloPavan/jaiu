ALTER TABLE payments
  ADD COLUMN kind text NOT NULL DEFAULT 'full',
  ADD COLUMN credit_cents bigint NOT NULL DEFAULT 0;

CREATE TABLE subscription_balances (
  subscription_id uuid PRIMARY KEY REFERENCES subscriptions(id),
  credit_cents bigint NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE billing_periods (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  subscription_id uuid NOT NULL REFERENCES subscriptions(id),
  period_start date NOT NULL,
  period_end date NOT NULL,
  amount_due_cents bigint NOT NULL,
  amount_paid_cents bigint NOT NULL DEFAULT 0,
  status text NOT NULL DEFAULT 'open',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE payment_allocations (
  payment_id uuid NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
  billing_period_id uuid NOT NULL REFERENCES billing_periods(id) ON DELETE CASCADE,
  amount_cents bigint NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (payment_id, billing_period_id)
);

CREATE INDEX billing_periods_subscription_idx ON billing_periods (subscription_id);
CREATE INDEX billing_periods_status_idx ON billing_periods (status);
CREATE INDEX billing_periods_period_end_idx ON billing_periods (period_end);

CREATE INDEX payment_allocations_payment_idx ON payment_allocations (payment_id);
CREATE INDEX payment_allocations_period_idx ON payment_allocations (billing_period_id);

CREATE TRIGGER billing_periods_updated_at
  BEFORE UPDATE ON billing_periods
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER subscription_balances_updated_at
  BEFORE UPDATE ON subscription_balances
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();
