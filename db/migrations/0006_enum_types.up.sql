CREATE TYPE student_status AS ENUM ('active', 'inactive', 'suspended');
CREATE TYPE subscription_status AS ENUM ('active', 'ended', 'canceled', 'suspended');
CREATE TYPE payment_status AS ENUM ('confirmed', 'reversed');
CREATE TYPE payment_method AS ENUM ('cash', 'pix', 'card', 'transfer', 'other');
CREATE TYPE payment_kind AS ENUM ('full', 'partial', 'advance', 'credit');
CREATE TYPE billing_period_status AS ENUM ('open', 'paid', 'partial', 'overdue');
CREATE TYPE user_role AS ENUM ('admin', 'operator');

ALTER TABLE students
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE student_status USING status::student_status,
  ALTER COLUMN status SET DEFAULT 'active';

ALTER TABLE subscriptions
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE subscription_status USING status::subscription_status,
  ALTER COLUMN status SET DEFAULT 'active';

ALTER TABLE payments
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN kind DROP DEFAULT,
  ALTER COLUMN method TYPE payment_method USING method::payment_method,
  ALTER COLUMN status TYPE payment_status USING status::payment_status,
  ALTER COLUMN kind TYPE payment_kind USING kind::payment_kind,
  ALTER COLUMN status SET DEFAULT 'confirmed',
  ALTER COLUMN kind SET DEFAULT 'full';

ALTER TABLE billing_periods
  ALTER COLUMN status DROP DEFAULT,
  ALTER COLUMN status TYPE billing_period_status USING status::billing_period_status,
  ALTER COLUMN status SET DEFAULT 'open';

ALTER TABLE users
  ALTER COLUMN role TYPE user_role USING role::user_role;
