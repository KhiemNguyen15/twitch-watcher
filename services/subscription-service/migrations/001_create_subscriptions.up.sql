CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    discord_webhook TEXT NOT NULL,
    watch_type      VARCHAR(10) NOT NULL CHECK (watch_type IN ('game', 'streamer')),
    watch_target    TEXT NOT NULL,
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_active
    ON subscriptions (active) WHERE active = TRUE;

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_unique
    ON subscriptions (discord_webhook, watch_type, watch_target);
