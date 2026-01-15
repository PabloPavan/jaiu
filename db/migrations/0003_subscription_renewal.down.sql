DROP INDEX IF EXISTS billing_periods_subscription_period_start_idx;
DROP INDEX IF EXISTS subscriptions_auto_renew_idx;

ALTER TABLE subscriptions DROP COLUMN IF EXISTS auto_renew;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS payment_day;
