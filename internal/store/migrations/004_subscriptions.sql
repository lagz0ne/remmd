CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY,
    subscriber_ref TEXT NOT NULL,
    tag_expression TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_by TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_subscriptions_tag ON subscriptions(tag_expression);

CREATE TABLE IF NOT EXISTS subscription_fires (
    id TEXT PRIMARY KEY,
    subscription_id TEXT NOT NULL REFERENCES subscriptions(id),
    triggering_doc_id TEXT NOT NULL,
    triggering_ref TEXT NOT NULL DEFAULT '',
    matched_tag TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    fired_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_subscription_fires_sub ON subscription_fires(subscription_id);
CREATE INDEX IF NOT EXISTS idx_subscription_fires_status ON subscription_fires(status);
