ALTER TABLE payments
  ADD COLUMN idempotency_key text;

CREATE UNIQUE INDEX payments_idempotency_key_idx ON payments (idempotency_key);
