CREATE TABLE IF NOT EXISTS playbooks (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    hash TEXT NOT NULL,
    imported_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(name, version)
);

CREATE TABLE IF NOT EXISTS pb_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    playbook_id TEXT NOT NULL REFERENCES playbooks(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    UNIQUE(playbook_id, name)
);

CREATE TABLE IF NOT EXISTS pb_fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type_id INTEGER NOT NULL REFERENCES pb_types(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    notation TEXT NOT NULL,
    UNIQUE(type_id, name)
);

CREATE TABLE IF NOT EXISTS pb_sections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type_id INTEGER NOT NULL REFERENCES pb_types(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    required BOOLEAN NOT NULL DEFAULT 0,
    "order" INTEGER NOT NULL DEFAULT 0,
    UNIQUE(type_id, name)
);

CREATE TABLE IF NOT EXISTS pb_edges (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    playbook_id TEXT NOT NULL REFERENCES playbooks(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    notation TEXT NOT NULL,
    UNIQUE(playbook_id, name)
);

CREATE TABLE IF NOT EXISTS pb_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    playbook_id TEXT NOT NULL REFERENCES playbooks(id) ON DELETE CASCADE,
    scope_type TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    severity TEXT NOT NULL DEFAULT 'error',
    expr TEXT NOT NULL,
    UNIQUE(playbook_id, scope_type, name)
);

CREATE TABLE IF NOT EXISTS pb_examples (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id INTEGER NOT NULL REFERENCES pb_rules(id) ON DELETE CASCADE,
    pass BOOLEAN NOT NULL,
    data_json TEXT NOT NULL,
    "order" INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_pb_types_playbook ON pb_types(playbook_id);
CREATE INDEX IF NOT EXISTS idx_pb_fields_type ON pb_fields(type_id);
CREATE INDEX IF NOT EXISTS idx_pb_sections_type ON pb_sections(type_id);
CREATE INDEX IF NOT EXISTS idx_pb_edges_playbook ON pb_edges(playbook_id);
CREATE INDEX IF NOT EXISTS idx_pb_rules_playbook ON pb_rules(playbook_id);
CREATE INDEX IF NOT EXISTS idx_pb_examples_rule ON pb_examples(rule_id);
