ALTER TABLE users
  ALTER COLUMN role TYPE text USING role::text;

ALTER TABLE billing_periods
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE text USING status::text,
  ALTER COLUMN status SET DEFAULT 'open';

ALTER TABLE payments
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN kind DROP DEFAULT,
  ALTER COLUMN method TYPE text USING method::text,
  ALTER COLUMN status TYPE text USING status::text,
  ALTER COLUMN kind TYPE text USING kind::text,
  ALTER COLUMN status SET DEFAULT 'confirmed',
  ALTER COLUMN kind SET DEFAULT 'full';

ALTER TABLE subscriptions
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE text USING status::text,
  ALTER COLUMN status SET DEFAULT 'active';

ALTER TABLE students
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE text USING status::text,
  ALTER COLUMN status SET DEFAULT 'active';

DROP TYPE user_role;
DROP TYPE billing_period_status;
DROP TYPE payment_kind;
DROP TYPE payment_method;
DROP TYPE payment_status;
DROP TYPE subscription_status;
DROP TYPE student_status;
