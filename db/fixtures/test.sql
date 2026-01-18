BEGIN;

INSERT INTO users (id, name, email, password_hash, role, active)
VALUES ('99999999-9999-9999-9999-999999999999', 'Admin', 'admin@example.com', 'hash', 'admin', true);

INSERT INTO students (id, full_name, birth_date, gender, phone, email, cpf, address, notes, photo_object_key, status)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'Alice Example', '2000-01-02', 'F', '+5511999999999', 'alice@example.com', '12345678900', 'Rua 1', 'Notes', 'photos/alice', 'active'),
  ('22222222-2222-2222-2222-222222222222', 'Bob Example', '1999-05-10', NULL, NULL, 'bob@example.com', NULL, NULL, NULL, NULL, 'inactive');

INSERT INTO plans (id, name, duration_days, price_cents, active, description)
VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Basic', 30, 1000, true, 'Basic plan'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Legacy', 30, 2000, false, 'Legacy plan');

INSERT INTO subscriptions (id, student_id, plan_id, start_date, end_date, status, price_cents, payment_day, auto_renew)
VALUES
  ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '2024-01-01', '2024-02-01', 'active', 1000, 1, true),
  ('44444444-4444-4444-4444-444444444444', '22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '2024-01-05', '2024-02-05', 'ended', 2000, 5, false);

INSERT INTO subscription_balances (subscription_id, credit_cents)
VALUES ('33333333-3333-3333-3333-333333333333', 0);

INSERT INTO payments (id, subscription_id, paid_at, amount_cents, method, reference, notes, status, kind, credit_cents, idempotency_key)
VALUES
  ('55555555-5555-5555-5555-555555555555', '33333333-3333-3333-3333-333333333333', '2024-01-02T10:00:00Z', 1000, 'cash', 'ref-1', 'note-1', 'confirmed', 'full', 0, 'idem-1');

INSERT INTO billing_periods (id, subscription_id, period_start, period_end, amount_due_cents, amount_paid_cents, status)
VALUES
  ('66666666-6666-6666-6666-666666666666', '33333333-3333-3333-3333-333333333333', '2024-01-01', '2024-01-31', 1000, 1000, 'paid'),
  ('88888888-8888-8888-8888-888888888888', '33333333-3333-3333-3333-333333333333', '2024-02-01', '2024-02-29', 1000, 0, 'open');

INSERT INTO payment_allocations (payment_id, billing_period_id, amount_cents)
VALUES ('55555555-5555-5555-5555-555555555555', '66666666-6666-6666-6666-666666666666', 1000);

COMMIT;
