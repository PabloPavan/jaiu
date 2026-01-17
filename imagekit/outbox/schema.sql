-- Postgres schema for the imagekit outbox table.
CREATE TABLE IF NOT EXISTS imagekit_outbox (
    id BIGSERIAL PRIMARY KEY,
    payload JSONB NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    locked_at TIMESTAMPTZ,
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS imagekit_outbox_available_at_idx
    ON imagekit_outbox (available_at)
    WHERE attempts < 10;
