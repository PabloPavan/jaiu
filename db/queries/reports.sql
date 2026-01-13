-- name: RevenueByPeriod :one
SELECT
  $1::timestamptz AS start,
  $2::timestamptz AS end,
  COALESCE(SUM(amount_cents), 0)::bigint AS total_cents
FROM payments
WHERE paid_at >= $1
  AND paid_at < $2
  AND status = 'confirmed';

-- name: StudentsByStatus :many
SELECT status, COUNT(*)::bigint AS total
FROM students
GROUP BY status
ORDER BY status;

-- name: DelinquentSubscriptions :many
SELECT
  s.id AS subscription_id,
  s.student_id,
  s.plan_id,
  s.end_date,
  GREATEST(0, DATE_PART('day', $1::date - s.end_date))::int AS days_overdue
FROM subscriptions s
WHERE s.status = 'active'
  AND s.end_date < $1::date
ORDER BY s.end_date;

-- name: UpcomingDue :many
SELECT
  s.id AS subscription_id,
  s.student_id,
  s.plan_id,
  s.end_date
FROM subscriptions s
WHERE s.status = 'active'
  AND s.end_date BETWEEN $1::date AND $2::date
ORDER BY s.end_date;
