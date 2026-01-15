ALTER TABLE subscriptions ADD COLUMN payment_day integer;
ALTER TABLE subscriptions ADD COLUMN auto_renew boolean NOT NULL DEFAULT false;

UPDATE subscriptions
SET payment_day = EXTRACT(DAY FROM end_date)::int
WHERE payment_day IS NULL;

ALTER TABLE subscriptions ALTER COLUMN payment_day SET NOT NULL;

CREATE INDEX subscriptions_auto_renew_idx ON subscriptions (status, auto_renew);
CREATE UNIQUE INDEX billing_periods_subscription_period_start_idx ON billing_periods (subscription_id, period_start);
