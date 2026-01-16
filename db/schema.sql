CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE students (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  full_name text NOT NULL,
  birth_date date,
  gender text,
  phone text,
  email text,
  cpf text,
  address text,
  notes text,
  photo_url text,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE plans (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  duration_days integer NOT NULL,
  price_cents bigint NOT NULL,
  active boolean NOT NULL DEFAULT true,
  description text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE subscriptions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id uuid NOT NULL REFERENCES students(id),
  plan_id uuid NOT NULL REFERENCES plans(id),
  start_date date NOT NULL,
  end_date date NOT NULL,
  status text NOT NULL DEFAULT 'active',
  price_cents bigint NOT NULL,
  payment_day integer NOT NULL,
  auto_renew boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE payments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  subscription_id uuid NOT NULL REFERENCES subscriptions(id),
  paid_at timestamptz NOT NULL,
  amount_cents bigint NOT NULL,
  method text NOT NULL,
  reference text,
  notes text,
  status text NOT NULL DEFAULT 'confirmed',
  kind text NOT NULL DEFAULT 'full',
  credit_cents bigint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);

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

CREATE TABLE audit_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id uuid,
  actor_role text,
  action text NOT NULL,
  entity_type text NOT NULL,
  entity_id uuid,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  ip text,
  user_agent text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  email text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  role text NOT NULL,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX students_full_name_idx ON students (full_name);
CREATE INDEX students_phone_idx ON students (phone);
CREATE INDEX students_cpf_idx ON students (cpf);
CREATE INDEX students_status_idx ON students (status);

CREATE INDEX plans_active_idx ON plans (active);

CREATE INDEX subscriptions_student_idx ON subscriptions (student_id);
CREATE INDEX subscriptions_status_idx ON subscriptions (status);
CREATE INDEX subscriptions_end_date_idx ON subscriptions (end_date);
CREATE INDEX subscriptions_auto_renew_idx ON subscriptions (status, auto_renew);

CREATE INDEX payments_subscription_idx ON payments (subscription_id);
CREATE INDEX payments_paid_at_idx ON payments (paid_at);

CREATE INDEX billing_periods_subscription_idx ON billing_periods (subscription_id);
CREATE INDEX billing_periods_status_idx ON billing_periods (status);
CREATE INDEX billing_periods_period_end_idx ON billing_periods (period_end);
CREATE UNIQUE INDEX billing_periods_subscription_period_start_idx ON billing_periods (subscription_id, period_start);

CREATE INDEX payment_allocations_payment_idx ON payment_allocations (payment_id);
CREATE INDEX payment_allocations_period_idx ON payment_allocations (billing_period_id);

CREATE INDEX audit_events_actor_idx ON audit_events (actor_id, created_at);
CREATE INDEX audit_events_entity_idx ON audit_events (entity_type, entity_id, created_at);

CREATE INDEX users_active_idx ON users (active);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER students_updated_at
  BEFORE UPDATE ON students
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER plans_updated_at
  BEFORE UPDATE ON plans
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER subscriptions_updated_at
  BEFORE UPDATE ON subscriptions
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER billing_periods_updated_at
  BEFORE UPDATE ON billing_periods
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER subscription_balances_updated_at
  BEFORE UPDATE ON subscription_balances
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER users_updated_at
  BEFORE UPDATE ON users
  FOR EACH ROW
  EXECUTE FUNCTION set_updated_at();
