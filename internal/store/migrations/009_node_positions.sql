CREATE TABLE IF NOT EXISTS node_positions (
    node_id TEXT PRIMARY KEY,
    x REAL NOT NULL,
    y REAL NOT NULL,
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
